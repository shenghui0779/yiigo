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

	cbc := NewAESCBCCrypto(key, iv)

	// PKCS5_PADDING
	e5b, err := cbc.Encrypt([]byte(plainText), PKCS5)
	assert.Nil(t, err)

	d5b, err := cbc.Decrypt(e5b, PKCS5)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d5b))

	// PKCS7_PADDING
	e7b, err := cbc.Encrypt([]byte(plainText), PKCS7)
	assert.Nil(t, err)

	d7b, err := cbc.Decrypt(e7b, PKCS7)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(d7b))
}

func TestAESCFBCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	cfb := NewAESCFBCrypto(key, iv)

	eb, err := cfb.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := cfb.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESOFBCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	ofb := NewAESOFBCrypto(key, iv)

	eb, err := ofb.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := ofb.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESCTRCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "Iloveyiigo"

	ctr := NewAESCTRCrypto(key, iv)

	eb, err := ctr.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := ctr.Decrypt(eb)
	assert.Nil(t, err)

	assert.Equal(t, plainText, string(db))
}

func TestAESGCMCrypto(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	nonce := key[:12]
	plainText := "Iloveyiigo"

	gcm := NewAESGCMCrypto(key, nonce)

	eb, err := gcm.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := gcm.Decrypt(eb)
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
