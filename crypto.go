package yiigo

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type AESPadding string

const (
	PKCS5 AESPadding = "PKCS#5"
	PKCS7 AESPadding = "PKCS#7"
)

type AESCrypto struct {
	block cipher.Block
	key   []byte
	iv    []byte
}

// NewAESCrypto returns new aes crypto
func NewAESCrypto(key []byte, iv ...byte) (*AESCrypto, error) {
	cb, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	r := &AESCrypto{
		block: cb,
		key:   key,
		iv:    iv,
	}

	if len(iv) == 0 {
		r.iv = key[:cb.BlockSize()]
	}

	return r, nil
}

// CBCEncrypt aes-cbc encryption
func (a *AESCrypto) CBCEncrypt(plainText []byte, padding AESPadding) []byte {
	switch padding {
	case PKCS5:
		plainText = a.padding(plainText, a.block.BlockSize())
	case PKCS7:
		plainText = a.padding(plainText, len(a.key))
	}

	cipherText := make([]byte, len(plainText))

	blockMode := cipher.NewCBCEncrypter(a.block, a.iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText
}

// CBCDecrypt aes-cbc decryption
func (a *AESCrypto) CBCDecrypt(cipherText []byte, padding AESPadding) []byte {
	plainText := make([]byte, len(cipherText))

	blockMode := cipher.NewCBCDecrypter(a.block, a.iv)
	blockMode.CryptBlocks(plainText, cipherText)

	switch padding {
	case PKCS5:
		plainText = a.unPadding(plainText, a.block.BlockSize())
	case PKCS7:
		plainText = a.unPadding(plainText, len(a.key))
	}

	return plainText
}

func (a *AESCrypto) padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize

	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

func (a *AESCrypto) unPadding(plainText []byte, blockSize int) []byte {
	l := len(plainText)
	unpadding := int(plainText[l-1])

	if unpadding < 1 || unpadding > blockSize {
		unpadding = 0
	}

	return plainText[:(l - unpadding)]
}

// GenerateRSAKey returns rsa private and public key
func GenerateRSAKey(bits int) (privateKey, publicKey []byte, err error) {
	prvKey, err := rsa.GenerateKey(rand.Reader, bits)

	if err != nil {
		return
	}

	pkixb, err := x509.MarshalPKIXPublicKey(&prvKey.PublicKey)

	if err != nil {
		return
	}

	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(prvKey),
	})

	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixb,
	})

	return
}

// RSASignWithSha256 returns rsa signature with sha256
func RSASignWithSha256(data, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)

	if block == nil {
		return nil, errors.New("invalid rsa private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h.Sum(nil))

	if err != nil {
		return nil, err
	}

	return signature, nil
}

// RSAVerifyWithSha256 verifies rsa signature with sha256
func RSAVerifyWithSha256(data, signature, publicKey []byte) error {
	block, _ := pem.Decode(publicKey)

	if block == nil {
		return errors.New("invalid rsa public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return err
	}

	key, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		return errors.New("invalid rsa public key")
	}

	hashed := sha256.Sum256(data)

	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature)
}

// RSAEncrypt rsa encrypt with public key
func RSAEncrypt(data, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)

	if block == nil {
		return nil, errors.New("invalid rsa public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	key, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		return nil, errors.New("invalid rsa public key")
	}

	return rsa.EncryptPKCS1v15(rand.Reader, key, data)
}

// RSADecrypt rsa decrypt with private key
func RSADecrypt(cipherText, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)

	if block == nil {
		return nil, errors.New("invalid rsa private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, key, cipherText)
}
