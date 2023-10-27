package yiigo

import (
	"crypto"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesCBC(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "IloveYiigo"

	// ZERO_PADDING
	zero := NewAesCBC(key, iv, AES_ZERO())

	e0b, err := zero.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d0b, err := zero.Decrypt(e0b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d0b))

	// PKCS5_PADDING
	pkcs5 := NewAesCBC(key, iv, AES_PKCS5())

	e5b, err := pkcs5.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d5b, err := pkcs5.Decrypt(e5b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d5b))

	// PKCS7_PADDING
	pkcs7 := NewAesCBC(key, iv, AES_PKCS7(32))

	e7b, err := pkcs7.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d7b, err := pkcs7.Decrypt(e7b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d7b))
}

func TestAesECB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	plainText := "IloveYiigo"

	// ZERO_PADDING
	zero := NewAesECB(key, AES_ZERO())

	e0b, err := zero.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d0b, err := zero.Decrypt(e0b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d0b))

	// PKCS5_PADDING
	pkcs5 := NewAesECB(key, AES_PKCS5())

	e5b, err := pkcs5.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d5b, err := pkcs5.Decrypt(e5b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d5b))

	// PKCS7_PADDING
	pkcs7 := NewAesECB(key, AES_PKCS7(32))

	e7b, err := pkcs7.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	d7b, err := pkcs7.Decrypt(e7b)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(d7b))
}

func TestAesCFB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "IloveYiigo"

	cfb := NewAesCFB(key, iv)

	eb, err := cfb.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := cfb.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))
}

func TestAesOFB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "IloveYiigo"

	ofb := NewAesOFB(key, iv)

	eb, err := ofb.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := ofb.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))
}

func TestAesCTR(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	plainText := "IloveYiigo"

	ctr := NewAesCTR(key, iv)

	eb, err := ctr.Encrypt([]byte(plainText))
	assert.Nil(t, err)

	db, err := ctr.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))
}

func TestAesGCM(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	nonce := key[:12]
	plainText := "IloveYiigo"
	additionalData := "IIInsomnia"

	gcm := NewAesGCM(key, nonce)

	eb, err := gcm.Encrypt([]byte(plainText), []byte(additionalData))
	assert.Nil(t, err)

	db, err := gcm.Decrypt(eb, []byte(additionalData))
	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))
}

func TestRSACrypto(t *testing.T) {
	plainText := "IloveYiigo"

	pvtKey, err := NewPrivateKeyFromPemBlock(RSA_PKCS1, privateKey)

	assert.Nil(t, err)

	pubKey, err := NewPublicKeyFromPemBlock(RSA_PKCS1, publicKey)

	assert.Nil(t, err)

	eb, err := pubKey.Encrypt([]byte(plainText))

	assert.Nil(t, err)

	db, err := pvtKey.Decrypt(eb)

	assert.Nil(t, err)
	assert.Equal(t, plainText, string(db))

	eboeap, err := pubKey.EncryptOAEP(crypto.SHA256, []byte(plainText))

	assert.Nil(t, err)

	dboeap, err := pvtKey.DecryptOAEP(crypto.SHA256, eboeap)

	assert.Nil(t, err)
	assert.Equal(t, plainText, string(dboeap))

	signSHA256, err := pvtKey.Sign(crypto.SHA256, []byte(plainText))

	assert.Nil(t, err)
	assert.Nil(t, pubKey.Verify(crypto.SHA256, []byte(plainText), signSHA256))

	signSHA1, err := pvtKey.Sign(crypto.SHA1, []byte(plainText))

	assert.Nil(t, err)
	assert.Nil(t, pubKey.Verify(crypto.SHA1, []byte(plainText), signSHA1))
}
