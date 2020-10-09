package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCBCCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	plainText := "Iloveyiigo"

	// PKCS5_PADDING
	e5b, err := AESCBCEncrypt([]byte(plainText), key, PKCS5_PADDING)
	assert.Nil(t, err)

	d5b, err := AESCBCDecrypt(e5b, key, PKCS5_PADDING)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d5b))

	// PKCS7_PADDING
	e7b, err := AESCBCEncrypt([]byte(plainText), key, PKCS7_PADDING)
	assert.Nil(t, err)

	d7b, err := AESCBCDecrypt(e7b, key, PKCS7_PADDING)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d7b))
}

func TestAESCFBCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	plainText := "Iloveyiigo"

	eb, err := AESCFBEncrypt([]byte(plainText), key)
	assert.Nil(t, err)

	db, err := AESCFBDecrypt(eb, key)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESGCMCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	plainText := "Iloveyiigo"

	eb, err := AESGCMEncrypt([]byte(plainText), key)
	assert.Nil(t, err)

	db, err := AESGCMDecrypt(eb, key)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestRSASign(t *testing.T) {
	plainText := "Iloveyiigo"

	signature, err := RSASignWithSha256([]byte(plainText), privateKey)

	assert.Nil(t, err)
	assert.Nil(t, RSAVerifyWithSha256([]byte(plainText), signature, publicKey))
}

func TestRSACrypto(t *testing.T) {
	plainText := "Iloveyiigo"

	eb, err := RSAEncrypt([]byte(plainText), publicKey)

	assert.Nil(t, err)

	db, err := RSADecrypt(eb, privateKey)

	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))
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
