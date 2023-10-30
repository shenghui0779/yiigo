package yiigo

import (
	"crypto"
	"crypto/aes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesCBC(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")

	block, err := aes.NewCipher(key)
	assert.Nil(t, err)

	out, err := AesCBCEncrypt(block, key[:aes.BlockSize], []byte("IloveYiigo"))
	assert.Nil(t, err)
	assert.Equal(t, "inYubOX1oU15tRN8itajQw==", out.String())
}

func TestAesECB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")

	block, err := aes.NewCipher(key)
	assert.Nil(t, err)

	e5b, err := pkcs5.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "3/UUhzaz+sjn3UW64/reaw==", base64.StdEncoding.EncodeToString(e5b))

	d5b, err := pkcs5.Decrypt(e5b)
	assert.Nil(t, err)
	assert.Equal(t, data, string(d5b))

	// pkcs#7
	pkcs7 := NewAesECB(key, AES_PKCS7(32))

	e7b, err := pkcs7.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "aXvKx3jVNrUWzzHBI+az+rpl6eN3wP/L8phJTP4aWFE=", base64.StdEncoding.EncodeToString(e7b))

	d7b, err := pkcs7.Decrypt(e7b)
	assert.Nil(t, err)
	assert.Equal(t, data, string(d7b))
}

func TestAesCFB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	cfb := NewAesCFB(key, iv)

	eb, err := cfb.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", base64.StdEncoding.EncodeToString(eb))

	db, err := cfb.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, data, string(db))
}

func TestAesOFB(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	ofb := NewAesOFB(key, iv)

	eb, err := ofb.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", base64.StdEncoding.EncodeToString(eb))

	db, err := ofb.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, data, string(db))
}

func TestAesCTR(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	iv := key[:aes.BlockSize]
	data := "IloveYiigo"

	ctr := NewAesCTR(key, iv)

	eb, err := ctr.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "KN7OnZjqIdiGlA==", base64.StdEncoding.EncodeToString(eb))

	db, err := ctr.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, data, string(db))
}

func TestAesGCM(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	nonce := key[:12]
	data := "IloveYiigo"
	aad := "IIInsomnia"

	gcm := NewAesGCM(key, nonce)

	eb, err := gcm.Encrypt([]byte(data), []byte(aad))
	assert.Nil(t, err)
	assert.Equal(t, "qeiumnRZKY42HZjRPuwOm7wTdj/FKddd5uI=", base64.StdEncoding.EncodeToString(eb))
	assert.Equal(t, "qeiumnRZKY42HQ==", base64.StdEncoding.EncodeToString(gcm.RealData(eb)))
	assert.Equal(t, "mNE+7A6bvBN2P8Up113m4g==", base64.StdEncoding.EncodeToString(gcm.Tag(eb)))

	db, err := gcm.Decrypt(eb, []byte(aad))
	assert.Nil(t, err)
	assert.Equal(t, data, string(db))
}

func TestRSACrypto(t *testing.T) {
	publicKey := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAwWVvD3G+O9N1NuBBz44OLb6aq85w8ahoTRepzydJ2qBcaDh+Zj6M
cybRSGHIGBIG0vyzYiPQhLK+s2kzKJ9rUHkQqRc7zDdVfclJhul1n1oBReyue1q9
AyZXhWssZodeQPG5SnlwziCuVhP6WCLF0M1bkvJr0+VOAfSHeTeYx/S/nH8JErmY
1HQTpkPs/fyabzCKoStWg6D62840HA2gn6Xq1MuPFki+BR8xcaM3Tqp2yN2kkIgO
RcGpTUOMk1L8xXRjTbYT48wyXmeMnR1TtmFE2Xc3sMC8y/mn8V7D4r2alfDHDX4d
13hBzo0oap7tugnr9yA2lak4Nvah03ZprwIDAQAB
-----END RSA PUBLIC KEY-----`)

	privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwWVvD3G+O9N1NuBBz44OLb6aq85w8ahoTRepzydJ2qBcaDh+
Zj6McybRSGHIGBIG0vyzYiPQhLK+s2kzKJ9rUHkQqRc7zDdVfclJhul1n1oBReyu
e1q9AyZXhWssZodeQPG5SnlwziCuVhP6WCLF0M1bkvJr0+VOAfSHeTeYx/S/nH8J
ErmY1HQTpkPs/fyabzCKoStWg6D62840HA2gn6Xq1MuPFki+BR8xcaM3Tqp2yN2k
kIgORcGpTUOMk1L8xXRjTbYT48wyXmeMnR1TtmFE2Xc3sMC8y/mn8V7D4r2alfDH
DX4d13hBzo0oap7tugnr9yA2lak4Nvah03ZprwIDAQABAoIBAB80zeHxGaAvs9dC
AnyKUJFjEzQr4J+t6/6cleL+VPV5MNAEZaj76M/f8J88X/w6VG2RJyTr4Ia5DPqI
PCAO8VMP5fdS72w5dYsRgtLJMxieflwZH+J5tsweULsPmx+EMlpKZvq0c9ZfAaKU
IK4+FitmJ6OjiHCtrJO2MHIH3ZhOBxn032BfdyVqhNN+oyn0zSjXvpHg9t/UEsXp
ZA7rHYn7m0RTwynFSaouAhmmZAp2GTYhe0NFu8rCG5afhtw9H2XiIiOhmLcURG+P
oW8v3I/Vt0OoLcqilbjPJs6nd43CAVyGastcBXhDFJJ4mFw5itMV9c+XNsEXPDcD
2g2voqECgYEA38UTnGv1eciGNcYMWUDJIB1c/205GoSpQ2kHXkNbFdN7u9lGlopq
3NwUPpHgbuWR5VxPmZCy1hCpFVXyeF9Ea3mFahiyiFECj4MeYq7i8Yd+UIfDNQ99
4C8TJP2mI4a8DaH7qG1KHfpkgaLsYuIhCmm+aNXsqcSNqRjYJtAE+lECgYEA3UBp
F6asT+ztQXF0QC7JOdaJgW6W4RNaIcU5rdK2vkkfhqQzR/XEFmHqVW7qUnLGm4mW
dTS6QBAoLwyd87KXvTW4y5rW2Un+l0Pc59Kl35BdlwMpXCffeqhamS4B7F4AdVZY
JaCYTCkTuwAx2r5nyOlkTcMIEGeDL676dRHII/8CgYEA3gZq+O9dd2JxV/WT1xMi
/ExmM8IpwJgUYiBaATuPqs5VnQNuuHvKoC11oMeZCi+aXRsEl/gsmZ2aRuMqXCka
eBDxQV4T9pF6mu6cPYoM/11TBZBPLdybJs9OjYtnRySuflBUpL8bpTcGdmIzbcG0
yuI03Uw1MBUoAbn27jvEVKECgYBiWxXc671CMqMuKo9xUNsnmRW7sjvkhsPUq2Z+
vWN7p+oZ4rjhToIDKTgRDqOgT2G3Fy0JoY0CmawjbkpxYX1PIaiq6oSER/6jpAl6
DQysG/NfBIrIavlP/7N20RsNxqQRhXbeE0xg3wnkYavIAEkG6aorX34gPMP22KSC
kosUZQKBgDKPXK4tnOC4HzYFlkiRxBuCMxU8bTG1+qKFvp+O4BbniDcUkZGJP/Gp
t6RsET7ZhCU8m8/6gIS5lZRoJt1aoqL3UyfFdWVA8pZwihDnEHvp1+0yl2BBaAN1
Vv8zI7kt+uZxD5mBGglKs2wzaHqADBXa5kSznIvkcZSg07UQQYU6
-----END RSA PRIVATE KEY-----`)

	data := "IloveYiigo"

	pvtKey, err := NewPrivateKeyFromPemBlock(RSA_PKCS1, privateKey)
	assert.Nil(t, err)

	pubKey, err := NewPublicKeyFromPemBlock(RSA_PKCS1, publicKey)
	assert.Nil(t, err)

	eb, err := pubKey.Encrypt([]byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "", base64.StdEncoding.EncodeToString(eb))

	db, err := pvtKey.Decrypt(eb)
	assert.Nil(t, err)
	assert.Equal(t, data, string(db))

	eboeap, err := pubKey.EncryptOAEP(crypto.SHA256, []byte(data))
	assert.Nil(t, err)
	assert.Equal(t, "", base64.StdEncoding.EncodeToString(eb))

	dboeap, err := pvtKey.DecryptOAEP(crypto.SHA256, eboeap)
	assert.Nil(t, err)
	assert.Equal(t, data, string(dboeap))

	signSHA256, err := pvtKey.Sign(crypto.SHA256, []byte(data))
	assert.Nil(t, err)
	assert.Nil(t, pubKey.Verify(crypto.SHA256, []byte(data), signSHA256))

	signSHA1, err := pvtKey.Sign(crypto.SHA1, []byte(data))
	assert.Nil(t, err)
	assert.Nil(t, pubKey.Verify(crypto.SHA1, []byte(data), signSHA1))
}
