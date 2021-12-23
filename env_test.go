package yiigo

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"testing"
)

var (
	builder  SQLBuilder
	validate *Validator

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	builder = NewMySQLBuilder(WithSQLDebug())

	validate = NewValidator(WithCustomValidateType(ValidateValuer, sql.NullString{}))

	privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSAPKCS8)

	m.Run()
}

func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		v, _ := valuer.Value()

		return v
	}

	return nil
}
