package yiigo

import (
	"crypto/aes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesCBC(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	cipher, err := AESEncryptCBC([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "inYubOX1oU15tRN8itajQw==", cipher.String())

	plain, err := AESDecryptCBC([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesECB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	data := "IloveYiigo"

	cipher, err := AESEncryptECB([]byte(key), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "3/UUhzaz+sjn3UW64/reaw==", cipher.String())

	plain, err := AESDecryptECB([]byte(key), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))

	cipher2, err := AESEncryptECBWithPaddingSize([]byte(key), []byte(data), 32)
	assert.Nil(t, err)
	assert.Equal(t, "aXvKx3jVNrUWzzHBI+az+rpl6eN3wP/L8phJTP4aWFE=", cipher2.String())

	plain2, err := AESDecryptECB([]byte(key), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain2))
}

func TestAesCFB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	cipher, err := AESEncryptCFB([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", cipher.String())

	plain, err := AESDecryptCFB([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesOFB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	cipher, err := AESEncryptOFB([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", cipher.String())

	plain, err := AESDecryptOFB([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesCTR(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	cipher, err := AESEncryptCTR([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", cipher.String())

	plain, err := AESDecryptCTR([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesGCM(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	nonce := key[:12]
	data := "IloveYiigo"
	aad := "IIInsomnia"

	cipher, err := AESEncryptGCM([]byte(key), []byte(nonce), []byte(data), []byte(aad))
	assert.Nil(t, err)
	assert.Equal(t, "qeiumnRZKY42HZjRPuwOm7wTdj/FKddd5uI=", cipher.String())
	assert.Equal(t, "qeiumnRZKY42HQ==", base64.StdEncoding.EncodeToString(cipher.Data()))
	assert.Equal(t, "mNE+7A6bvBN2P8Up113m4g==", base64.StdEncoding.EncodeToString(cipher.Tag()))

	plain, err := AESDecryptGCM([]byte(key), []byte(nonce), cipher.Bytes(), []byte(aad))
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}
