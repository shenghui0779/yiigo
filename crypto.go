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

// AESPadding aes padding
type AESPadding string

const (
	// PKCS5_PADDING PKCS#5 padding
	PKCS5_PADDING AESPadding = "PKCS#5"
	// PKCS7_PADDING PKCS#7 padding
	PKCS7_PADDING AESPadding = "PKCS#7"
)

// AESCrypto aes crypto
type AESCrypto struct {
	block cipher.Block
	key   []byte
	iv    []byte
}

// NewAESCrypto returns new aes crypto
func NewAESCrypto(key []byte, iv ...byte) (*AESCrypto, error) {
	b, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	r := &AESCrypto{
		block: b,
		key:   key,
		iv:    iv,
	}

	if len(iv) == 0 {
		r.iv = key[:b.BlockSize()]
	}

	return r, nil
}

// CBCEncrypt aes-cbc encryption
func (a *AESCrypto) CBCEncrypt(plainText []byte, padding AESPadding) []byte {
	switch padding {
	case PKCS5_PADDING:
		plainText = a.padding(plainText, a.block.BlockSize())
	case PKCS7_PADDING:
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
	case PKCS5_PADDING:
		plainText = a.unPadding(plainText, a.block.BlockSize())
	case PKCS7_PADDING:
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

// RSACrypto rsa crypto
type RSACrypto struct {
	publicKey  []byte
	privateKey []byte
}

// NewRSACrypto returns new rsa crypto
func NewRSACrypto(privateKey, publicKey []byte) *RSACrypto {
	return &RSACrypto{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// SignWithSha256 returns rsa sha256-signature with private key
func (r *RSACrypto) SignWithSha256(data []byte) ([]byte, error) {
	block, _ := pem.Decode(r.privateKey)

	if block == nil {
		return nil, errors.New("yiigo: invalid rsa private key")
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

// VerifyWithSha256 verifies rsa sha256-signature with public key
func (r *RSACrypto) VerifyWithSha256(data, signature []byte) error {
	block, _ := pem.Decode(r.publicKey)

	if block == nil {
		return errors.New("yiigo: invalid rsa public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return err
	}

	key, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		return errors.New("yiigo: invalid rsa public key")
	}

	hashed := sha256.Sum256(data)

	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature)
}

// Encrypt rsa encryption with public key
func (r *RSACrypto) Encrypt(data []byte) ([]byte, error) {
	block, _ := pem.Decode(r.publicKey)

	if block == nil {
		return nil, errors.New("yiigo: invalid rsa public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	key, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		return nil, errors.New("yiigo: invalid rsa public key")
	}

	return rsa.EncryptPKCS1v15(rand.Reader, key, data)
}

// Decrypt rsa decryption with private key
func (r *RSACrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, _ := pem.Decode(r.privateKey)

	if block == nil {
		return nil, errors.New("yiigo: invalid rsa private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, key, cipherText)
}
