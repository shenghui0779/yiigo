package yiigo

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhcn "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"
)

// ValidatorOption 验证器选项
type ValidatorOption func(validate *validator.Validate, trans ut.Translator)

// WithValidateTag 设置Tag名称，默认：valid
func WithValidateTag(tagname string) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		validate.SetTagName(tagname)
	}
}

// WithValuerType 注册自定义验证类型
func WithValuerType(types ...driver.Valuer) ValidatorOption {
	customTypes := make([]any, 0, len(types))

	for _, t := range types {
		customTypes = append(customTypes, t)
	}

	return func(validate *validator.Validate, trans ut.Translator) {
		validate.RegisterCustomTypeFunc(func(field reflect.Value) any {
			if valuer, ok := field.Interface().(driver.Valuer); ok {
				v, _ := valuer.Value()
				return v
			}

			return nil
		}, customTypes...)
	}
}

// WithValidation 注册自定义验证器
func WithValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		if err := validate.RegisterValidation(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Error("err register validation", zap.Error(err))
		}
	}
}

// WithValidationCtx 注册带Context的自定义验证器
func WithValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		if err := validate.RegisterValidationCtx(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Error("err register validation with ctx", zap.Error(err))
		}
	}
}

// WithTranslation 注册自定义错误翻译
// 参数 `text` 示例：{0}为必填字段 或 {0}必须大于{1}
func WithTranslation(tag, text string, override bool) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, text, override)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field(), fe.Param())
			return t
		})
	}
}

// Validator 可被用于Gin框架的验证器
// 具体支持的验证规则，可以参考：https://pkg.go.dev/github.com/go-playground/validator/v10
type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(obj any) error {
	if reflect.Indirect(reflect.ValueOf(obj)).Kind() != reflect.Struct {
		return nil
	}

	if err := v.validator.Struct(obj); err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		errM := e.Translate(v.translator)
		msgs := make([]string, 0, len(errM))

		for _, v := range errM {
			msgs = append(msgs, v)
		}

		return errors.New(strings.Join(msgs, ";"))
	}

	return nil
}

// ValidateStructCtx 验证结构体，带Context
func (v *Validator) ValidateStructCtx(ctx context.Context, obj any) error {
	if reflect.Indirect(reflect.ValueOf(obj)).Kind() != reflect.Struct {
		return nil
	}

	if err := v.validator.StructCtx(ctx, obj); err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		errM := e.Translate(v.translator)
		msgs := make([]string, 0, len(errM))

		for _, v := range errM {
			msgs = append(msgs, v)
		}

		return errors.New(strings.Join(msgs, ";"))
	}

	return nil
}

// Engine 实现Gin验证器接口
func (v *Validator) Engine() any {
	return v.validator
}

// NewValidator 生成一个验证器实例
// 在Gin中使用：binding.Validator = yiigo.NewValidator()
func NewValidator(options ...ValidatorOption) *Validator {
	validate := validator.New()
	validate.SetTagName("valid")

	zhTrans := zh.New()
	trans, _ := ut.New(zhTrans, zhTrans).GetTranslator("zh")

	if err := zhcn.RegisterDefaultTranslations(validate, trans); err != nil {
		logger.Error("err validation translator", zap.Error(err))
	}

	for _, f := range options {
		f(validate, trans)
	}

	return &Validator{
		validator:  validate,
		translator: trans,
	}
}
