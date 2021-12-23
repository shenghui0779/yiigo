package yiigo

import (
	"context"
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
type ValidatorOption func(validate *validator.Validate)

// SetValidateTag allows for changing of the default tag name of 'validate'.
func SetValidateTag(tagname string) ValidatorOption {
	return func(validate *validator.Validate) {
		validate.SetTagName(tagname)
	}
}

// WithStructValidation registers a StructLevelFunc against a number of types.
func WithStructValidation(fn validator.StructLevelFunc, types ...interface{}) ValidatorOption {
	return func(validate *validator.Validate) {
		validate.RegisterStructValidation(fn, types...)
	}
}

// WithStructValidationCtx registers a StructLevelFuncCtx against a number of types and allows passing of contextual validation information via context.Context.
func WithStructValidationCtx(fn validator.StructLevelFuncCtx, types ...interface{}) ValidatorOption {
	return func(validate *validator.Validate) {
		validate.RegisterStructValidationCtx(fn, types...)
	}
}

// WithCustomValidateType registers a CustomTypeFunc against a number of types.
func WithCustomValidateType(fn validator.CustomTypeFunc, types ...interface{}) ValidatorOption {
	return func(validate *validator.Validate) {
		validate.RegisterCustomTypeFunc(fn, types...)
	}
}

// WithCustomValidation adds a validation with the given tag.
func WithCustomValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate) {
		if err := validate.RegisterValidation(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Warn("err register validation", zap.Error(err))
		}
	}
}

// WithCustomValidationCtx does the same as WithCustomValidation on accepts a FuncCtx validation allowing context.Context validation support.
func WithCustomValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) ValidatorOption {
	return func(validate *validator.Validate) {
		if err := validate.RegisterValidationCtx(tag, fn, callValidationEvenIfNull...); err != nil {
			logger.Warn("err register validation with ctx", zap.Error(err))
		}
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

// NewValidator returns a new validator with default tag: valid.
// Used for Gin: binding.Validator = yiigo.NewValidator()
func NewValidator(options ...ValidatorOption) *Validator {
	locale := zh.New()
	uniTrans := ut.New(locale)

	validate := validator.New()
	validate.SetTagName("valid")

	for _, f := range options {
		f(validate)
	}

	translator, _ := uniTrans.GetTranslator("zh")

	if err := zhcn.RegisterDefaultTranslations(validate, translator); err != nil {
		logger.Warn("err validation translator", zap.Error(err))
	}

	return &Validator{
		validator:  validate,
		translator: translator,
	}
}
