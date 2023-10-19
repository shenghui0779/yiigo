package yiigo

import (
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
)

var (
	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	privateKey, publicKey, _ = GenerateRSAKey(2048, RSA_PKCS1)
	// privateKey, publicKey, _ = GenerateRSAKey(2048, RSA_PKCS8)

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
