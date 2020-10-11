package yiigo

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCBCCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	crypto := NewAESCBCCrypto(key, iv)

	// PKCS5_PADDING
	e5b, err := crypto.Encrypt([]byte(plainText), PKCS5)
	assert.Nil(t, err)

	d5b, err := crypto.Decrypt(e5b, PKCS5)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d5b))

	// PKCS7_PADDING
	e7b, err := crypto.Encrypt([]byte(plainText), PKCS7)
	assert.Nil(t, err)

	d7b, err := crypto.Decrypt(e7b, PKCS7)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d7b))
}

func TestAESCFBCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	crypto := NewAESCFBCrypto(key, iv)

	eb, err := crypto.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := crypto.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESOFBCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	crypto := NewAESOFBCrypto(key, iv)

	eb, err := crypto.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := crypto.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESCTRCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	crypto := NewAESCTRCrypto(key, iv)

	eb, err := crypto.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := crypto.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESGCMCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	nonce := key[:12]
	plainText := "Iloveyiigo"

	crypto := NewAESGCMCrypto(key, nonce)

	eb, err := crypto.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := crypto.Decrypt(eb)
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
