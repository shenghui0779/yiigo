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

// AESPaddingMode aes padding mode
type AESPaddingMode string

const (
	// PKCS5 PKCS#5 padding mode
	PKCS5 AESPaddingMode = "PKCS#5"
	// PKCS7 PKCS#7 padding mode
	PKCS7 AESPaddingMode = "PKCS#7"
)

// AESCBCCrypto aes-cbc crypto
type AESCBCCrypto struct {
	key []byte
	iv  []byte
}

// NewAESCBCCrypto returns new aes-cbc crypto
func NewAESCBCCrypto(key, iv []byte) *AESCBCCrypto {
	return &AESCBCCrypto{
		key: key,
		iv:  iv,
	}
}

// Encrypt aes-cbc encrypt
func (c *AESCBCCrypto) Encrypt(plainText []byte, mode AESPaddingMode) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	switch mode {
	case PKCS5:
		plainText = c.padding(plainText, block.BlockSize())
	case PKCS7:
		plainText = c.padding(plainText, len(c.key))
	}

	cipherText := make([]byte, len(plainText))

	blockMode := cipher.NewCBCEncrypter(block, c.iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// Decrypt aes-cbc decrypt
func (c *AESCBCCrypto) Decrypt(cipherText []byte, mode AESPaddingMode) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	blockMode := cipher.NewCBCDecrypter(block, c.iv)
	blockMode.CryptBlocks(plainText, cipherText)

	switch mode {
	case PKCS5:
		plainText = c.unpadding(plainText, block.BlockSize())
	case PKCS7:
		plainText = c.unpadding(plainText, len(c.key))
	}

	return plainText, nil
}

func (c *AESCBCCrypto) padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize

	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

func (c *AESCBCCrypto) unpadding(plainText []byte, blockSize int) []byte {
	l := len(plainText)
	unpadding := int(plainText[l-1])

	if unpadding < 1 || unpadding > blockSize {
		unpadding = 0
	}

	return plainText[:(l - unpadding)]
}

// AESCFBCrypto aes-cfb crypto
type AESCFBCrypto struct {
	key []byte
	iv  []byte
}

// NewAESCFBCrypto returns new aes-cfb crypto
func NewAESCFBCrypto(key, iv []byte) *AESCFBCrypto {
	return &AESCFBCrypto{
		key: key,
		iv:  iv,
	}
}

// Encrypt aes-cfb encrypt
func (c *AESCFBCrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewCFBEncrypter(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

// Decrypt aes-cfb decrypt
func (c *AESCFBCrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewCFBDecrypter(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// AESOFBCrypto aes-ofb crypto
type AESOFBCrypto struct {
	key []byte
	iv  []byte
}

// NewAESOFBCrypto returns new aes-ofb crypto
func NewAESOFBCrypto(key, iv []byte) *AESOFBCrypto {
	return &AESOFBCrypto{
		key: key,
		iv:  iv,
	}
}

// Encrypt aes-ofb encrypt
func (c *AESOFBCrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewOFB(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

// Decrypt aes-ofb decrypt
func (c *AESOFBCrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewOFB(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// AESCTRCrypto aes-ctr crypto
type AESCTRCrypto struct {
	key []byte
	iv  []byte
}

// NewAESCTRCrypto returns new aes-ctr crypto
func NewAESCTRCrypto(key, iv []byte) *AESCTRCrypto {
	return &AESCTRCrypto{
		key: key,
		iv:  iv,
	}
}

// Encrypt aes-ctr encrypt
func (c *AESCTRCrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewCTR(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

// Decrypt aes-ctr decrypt
func (c *AESCTRCrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("yiigo: IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewCTR(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// AESGCMCrypto aes-gcm crypto
type AESGCMCrypto struct {
	key   []byte
	nonce []byte
}

// NewAESGCMCrypto returns new aes-gcm crypto
func NewAESGCMCrypto(key, nonce []byte) *AESGCMCrypto {
	return &AESGCMCrypto{
		key:   key,
		nonce: nonce,
	}
}

// Encrypt aes-gcm encrypt
func (c *AESGCMCrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aesgcm.NonceSize() {
		return nil, errors.New("yiigo: Nonce length must equal gcm standard nonce size")
	}

	return aesgcm.Seal(nil, c.nonce, plainText, nil), nil
}

// Decrypt aes-gcm decrypt
func (c *AESGCMCrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aesgcm.NonceSize() {
		return nil, errors.New("yiigo: Nonce length must equal gcm standard nonce size")
	}

	return aesgcm.Open(nil, c.nonce, cipherText, nil)
}

// GenerateRSAKey returns rsa private and public key
func GenerateRSAKey(bitSize int) (privateKey, publicKey []byte, err error) {
	prvKey, err := rsa.GenerateKey(rand.Reader, bitSize)

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
func RSAEncrypt(plainText, publicKey []byte) ([]byte, error) {
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

	return rsa.EncryptPKCS1v15(rand.Reader, key, plainText)
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
