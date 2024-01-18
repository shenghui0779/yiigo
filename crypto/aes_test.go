package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesCBC(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:16]
	data := "ILoveYiigo"

	cipher, err := EncryptCBC([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "kyJ6t0cpUYpoWaewhTwDwQ==", cipher.String())

	plain, err := DecryptCBC([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))

	cipher2, err := EncryptCBC([]byte(key), []byte(iv), []byte(data), 32)
	assert.Nil(t, err)
	assert.Equal(t, "hSXsKUV2fbG8F2JlVcnra876xvKxyXwoJvaebTtWGzQ=", cipher2.String())

	plain2, err := DecryptCBC([]byte(key), []byte(iv), cipher2.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain2))
}

func TestAesECB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	data := "ILoveYiigo"

	cipher, err := EncryptECB([]byte(key), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "8+evCMirn78a5l2mCCdJug==", cipher.String())

	plain, err := DecryptECB([]byte(key), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))

	cipher2, err := EncryptECB([]byte(key), []byte(data), 32)
	assert.Nil(t, err)
	assert.Equal(t, "FqrgSRCY4zBRYBOg4Pe3Vbpl6eN3wP/L8phJTP4aWFE=", cipher2.String())

	plain2, err := DecryptECB([]byte(key), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain2))
}

func TestAesCFB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:16]
	data := "ILoveYiigo"

	cipher, err := EncryptCFB([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KP7OnZjqIdiGlA==", cipher.String())

	plain, err := DecryptCFB([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesOFB(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:16]
	data := "ILoveYiigo"

	cipher, err := EncryptOFB([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KP7OnZjqIdiGlA==", cipher.String())

	plain, err := DecryptOFB([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesCTR(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	iv := key[:16]
	data := "ILoveYiigo"

	cipher, err := EncryptCTR([]byte(key), []byte(iv), []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KP7OnZjqIdiGlA==", cipher.String())

	plain, err := DecryptCTR([]byte(key), []byte(iv), cipher.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}

func TestAesGCM(t *testing.T) {
	key := "AES256Key-32Characters1234567890"
	nonce := key[:12]
	data := "ILoveYiigo"
	aad := "IIInsomnia"

	cipher, err := EncryptGCM([]byte(key), []byte(nonce), []byte(data), []byte(aad), &GCMOption{})
	assert.Nil(t, err)
	assert.Equal(t, "qciumnRZKY42HVjng/cUjd0V+OJZB6ZwRF8=", cipher.String())
	assert.Equal(t, "qciumnRZKY42HQ==", base64.StdEncoding.EncodeToString(cipher.Data()))
	assert.Equal(t, "WOeD9xSN3RX44lkHpnBEXw==", base64.StdEncoding.EncodeToString(cipher.Tag()))

	plain, err := DecryptGCM([]byte(key), []byte(nonce), cipher.Bytes(), []byte(aad), nil)
	assert.Nil(t, err)
	assert.Equal(t, data, string(plain))
}
