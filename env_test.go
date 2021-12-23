package yiigo

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
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
		WithCustomValidateType(ValidateValuer, sql.NullString{}),
		WithCustomValidation("nullrequired", ValidateNullStringRequired),
	)

	privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS8)

	m.Run()
}

func ValidateNullStringRequired(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) == 0
}

func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		v, _ := valuer.Value()

		return v
	}

	return nil
}
