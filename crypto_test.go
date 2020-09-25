package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCBCCrypt(t *testing.T) {
	aesKey := []byte("c510be34b0466938eace8edee61255c0")

	// PKCS5_PADDING
	e5b, err := AESCBCEncrypt([]byte("shenghui0779"), PKCS5_PADDING, aesKey)
	assert.Nil(t, err)

	d5b, err := AESCBCDecrypt(e5b, PKCS5_PADDING, aesKey)
	assert.Nil(t, err)

	assert.Equal(t, "shenghui0779", string(d5b))

	// PKCS7_PADDING
	e7b, err := AESCBCEncrypt([]byte("Iloveyiigo"), PKCS7_PADDING, aesKey)
	assert.Nil(t, err)

	d7b, err := AESCBCDecrypt(e7b, PKCS7_PADDING, aesKey)
	assert.Nil(t, err)

	assert.Equal(t, "Iloveyiigo", string(d7b))
}

func TestRSASign(t *testing.T) {
	signature, err := RSASignWithSha256([]byte("Iloveyiigo"), privateKey)

	assert.Nil(t, err)
	assert.Nil(t, RSAVerifyWithSha256([]byte("Iloveyiigo"), signature, publicKey))
}

func TestRSACrypt(t *testing.T) {
	eb, err := RSAEncrypt([]byte("Iloveyiigo"), publicKey)

	assert.Nil(t, err)

	db, err := RSADecrypt(eb, privateKey)

	assert.Nil(t, err)
	assert.Equal(t, "Iloveyiigo", string(db))
}

var (
	builder *SQLBuilder

	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	builder = NewSQLBuilder(MySQL)

	privateKey, publicKey, _ = GenerateRSAKey(2048)

	m.Run()
}
