package xvalidator

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhcn "github.com/go-playground/validator/v10/translations/zh"
)

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

// New 生成一个验证器实例
// 在Gin中使用：binding.Validator = validator.New()
func New(opts ...Option) *Validator {
	validate := validator.New()
	validate.SetTagName("valid")

	zhTrans := zh.New()
	trans, _ := ut.New(zhTrans, zhTrans).GetTranslator("zh")
	zhcn.RegisterDefaultTranslations(validate, trans)

	for _, fn := range opts {
		fn(validate, trans)
	}

	return &Validator{
		validator:  validate,
		translator: trans,
	}
}

// v 默认验证器
var v = New()

// ValidateStruct 验证结构体
func ValidateStruct(obj any) error {
	return v.ValidateStruct(obj)
}

// ValidateStructCtx 验证结构体，带Context
func ValidateStructCtx(ctx context.Context, obj any) error {
	return v.ValidateStructCtx(ctx, obj)
}
