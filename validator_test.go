package yiigo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type ParamsValidate struct {
	ID   sql.NullInt64  `valid:"nullint_gte=10"`
	Desc sql.NullString `valid:"nullstring_required"`
}

func TestValidator(t *testing.T) {
	validator := NewValidator(
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

	err := validator.ValidateStruct(params1)

	assert.NotNil(t, err)

	logger.Info("err validate params", zap.Error(err))

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

	err = validator.ValidateStruct(params2)

	assert.Nil(t, err)
}
