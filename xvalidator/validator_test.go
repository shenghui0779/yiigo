package xvalidator

import (
	"database/sql"
	"log"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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

type ParamsValidate struct {
	ID   sql.NullInt64  `valid:"nullint_gte=10"`
	Desc sql.NullString `valid:"nullstring_required"`
}

func TestValidator(t *testing.T) {
	v := New(
		WithValuerType(sql.NullString{}, sql.NullInt64{}),
		WithValidation("nullint_gte", NullIntGTE),
		WithTranslation("nullint_gte", "{0}必须大于或等于{1}", true),
		WithValidation("nullstring_required", NullStringRequired),
		WithTranslation("nullstring_required", "{0}为必填字段", true),
	)

	params1 := new(ParamsValidate)
	params1.ID = sql.NullInt64{
		Int64: 9,
		Valid: true,
	}
	err := v.ValidateStruct(params1)
	assert.NotNil(t, err)
	log.Println("err validate params:", err.Error())

	params2 := &ParamsValidate{
		ID: sql.NullInt64{
			Int64: 13,
			Valid: true,
		},
		Desc: sql.NullString{
			String: "yiigo",
			Valid:  true,
		},
	}
	err = v.ValidateStruct(params2)
	assert.Nil(t, err)
}
