package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCBCCrypt(t *testing.T) {
	aesCrypto, err := NewAESCrypto([]byte("c510be34b0466938eace8edee61255c0"))

	assert.Nil(t, err)

	// PKCS5_PADDING
	e5b := aesCrypto.CBCEncrypt([]byte("shenghui0779"), PKCS5_PADDING)
	d5b := aesCrypto.CBCDecrypt(e5b, PKCS5_PADDING)

	assert.Equal(t, "shenghui0779", string(d5b))

	// PKCS7_PADDING
	e7b := aesCrypto.CBCEncrypt([]byte("Iloveyiigo"), PKCS7_PADDING)
	d7b := aesCrypto.CBCDecrypt(e7b, PKCS7_PADDING)

	assert.Equal(t, "Iloveyiigo", string(d7b))
}

func TestRSASign(t *testing.T) {
	rsaCrypto := NewRSACrypto(privateKey, publicKey)

	signature, err := rsaCrypto.SignWithSha256([]byte("Iloveyiigo"))

	assert.Nil(t, err)
	assert.Nil(t, rsaCrypto.VerifyWithSha256([]byte("Iloveyiigo"), signature))
}

func TestRSACrypt(t *testing.T) {
	rsaCrypto := NewRSACrypto(privateKey, publicKey)

	eb, err := rsaCrypto.Encrypt([]byte("Iloveyiigo"))

	assert.Nil(t, err)

	db, err := rsaCrypto.Decrypt(eb)

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
