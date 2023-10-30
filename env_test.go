package yiigo

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

func NullStringRequired(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) != 0
}

func NullIntGTE(fl validator.FieldLevel) bool {
	i, err := strconv.ParseInt(fl.Param(), 0, 64)
	if err != nil {
		return false
	}

	return fl.Field().Int() >= i
}
