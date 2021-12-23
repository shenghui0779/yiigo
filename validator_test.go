package yiigo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type ParamsValidate struct {
	ID   int64          `valid:"required"`
	Name string         `valid:"required"`
	Desc sql.NullString `valid:"required"`
}

func TestValidator(t *testing.T) {
	params := new(ParamsValidate)

	err := validate.ValidateStruct(params)

	assert.NotNil(t, err)

	logger.Info("err validate params", zap.Error(err))
}
