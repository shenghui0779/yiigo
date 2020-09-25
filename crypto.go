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
	"encoding/hex"
	"encoding/pem"
	"errors"
)

// AESPadding aes padding
type AESPadding string

const (
	// PKCS5_PADDING PKCS#5 padding
	PKCS5_PADDING AESPadding = "PKCS#5"
	// PKCS7_PADDING PKCS#7 padding
	PKCS7_PADDING AESPadding = "PKCS#7"
)

// AESCBCEncrypt AES CBC encrypt
func AESCBCEncrypt(plainText []byte, padding AESPadding, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	switch padding {
	case PKCS5_PADDING:
		plainText = aesPadding(plainText, block.BlockSize())
	case PKCS7_PADDING:
		plainText = aesPadding(plainText, len(key))
	}

	cipherText := make([]byte, len(plainText))

	if len(iv) == 0 {
		iv = key[:block.BlockSize()]
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// AESCBCDecrypt AES CBC decrypt
func AESCBCDecrypt(cipherText []byte, padding AESPadding, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	plainText := make([]byte, len(cipherText))

	if len(iv) == 0 {
		iv = key[:block.BlockSize()]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(plainText, cipherText)

	switch padding {
	case PKCS5_PADDING:
		plainText = aesUnPadding(plainText, block.BlockSize())
	case PKCS7_PADDING:
		plainText = aesUnPadding(plainText, len(key))
	}

	return plainText, nil
}

func aesPadding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize

	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

func aesUnPadding(plainText []byte, blockSize int) []byte {
	l := len(plainText)
	unpadding := int(plainText[l-1])

	if unpadding < 0 || unpadding > blockSize {
		unpadding = 0
	}

	return plainText[:(l - unpadding)]
}

// AESGCMEncrypt AES GCM encrypt
func AESGCMEncrypt(plainText, key, nonce []byte) ([]byte, error) {
	nonceHex, err := hex.DecodeString(string(nonce))

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonceHex, plainText, nil), nil
}

// AESGCMDecrypt AES GCM decrypt
func AESGCMDecrypt(cipherText, key, nonce []byte) ([]byte, error) {
	nonceHex, err := hex.DecodeString(string(nonce))

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonceHex, cipherText, nil)
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
