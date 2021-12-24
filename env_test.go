package yiigo

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
)

var (
	builder  SQLBuilder
	validate *Validator

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	builder = NewMySQLBuilder(WithSQLDebug())

	validate = NewValidator(
		WithValuerType(sql.NullString{}, sql.NullInt64{}),
		WithValidation("nullint_gte", NullIntGTE),
		WithTranslation("nullint_gte", "{0}必须大于或等于{1}", true),
		WithValidation("nullstring_required", NullStringRequired),
		WithTranslation("nullstring_required", "{0}为必填字段", true),
	)

	privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS8)

	m.Run()
}

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
