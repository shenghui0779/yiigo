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

// ValidatorOption configures how we set up the validator.
type ValidatorOption func(validate *validator.Validate, trans ut.Translator)

// SetValidateTag allows for changing of the default validate tag name: valid.
func SetValidateTag(tagname string) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		validate.SetTagName(tagname)
	}
}

// WithValuerType registers a number of custom validate types which implement the driver.Valuer interface.
func WithValuerType(types ...driver.Valuer) ValidatorOption {
	customTypes := make([]interface{}, 0, len(types))

	for _, t := range types {
		customTypes = append(customTypes, t)
	}

	return func(validate *validator.Validate, trans ut.Translator) {
		validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if valuer, ok := field.Interface().(driver.Valuer); ok {
				v, _ := valuer.Value()

				return v
			}

			return nil
		}, customTypes...)
	}
}

// WithValidation adds a custom validation with the given tag.
func WithValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		if err := validate.RegisterValidation(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Warn("err register validation", zap.Error(err))
		}
	}
}

// WithValidationCtx does the same as WithValidation on accepts a FuncCtx validation allowing context.Context validation support.
func WithValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate, trans ut.Translator) {
		if err := validate.RegisterValidationCtx(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Warn("err register validation with ctx", zap.Error(err))
		}
	}
}

// WithTranslation registers custom validate translation against the provided tag.
// Param text, eg: {0}为必填字段 或 {0}必须大于{1}
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

// Validator a validator which can be used for Gin.
type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *Validator) ValidateStruct(obj interface{}) error {
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

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type and allows passing of context.Context for contextual validation information.
func (v *Validator) ValidateStructCtx(ctx context.Context, obj interface{}) error {
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

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://pkg.go.dev/github.com/go-playground/validator/v10
func (v *Validator) Engine() interface{} {
	return v.validator
}

// NewValidator returns a new validator with default tag name: valid.
// Used for Gin: binding.Validator = yiigo.NewValidator()
func NewValidator(options ...ValidatorOption) *Validator {
	validate := validator.New()
	validate.SetTagName("valid")

	zhTrans := zh.New()
	trans, _ := ut.New(zhTrans, zhTrans).GetTranslator("zh")

	if err := zhcn.RegisterDefaultTranslations(validate, trans); err != nil {
		logger.Warn("err validation translator", zap.Error(err))
	}

	for _, f := range options {
		f(validate, trans)
	}

	return &Validator{
		validator:  validate,
		translator: trans,
	}
}
