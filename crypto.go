package yiigo

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AESPaddingMode AES填充模式
type AESPaddingMode int

const (
	AES_ZERO  AESPaddingMode = iota // 0
	AES_PKCS5                       // PKCS#5
	AES_PKCS7                       // PKCS#7
)

// RSAPaddingMode RSA PEM 填充模式
type RSAPaddingMode int

const (
	RSA_PKCS1 RSAPaddingMode = iota // PKCS#1 (格式：`RSA PRIVATE KEY` 和 `RSA PUBLIC KEY`)
	RSA_PKCS8                       // PKCS#8 (格式：`PRIVATE KEY` 和 `PUBLIC KEY`)
)

// AESCrypto AES加解密
type AESCrypto interface {
	// Encrypt 加密
	Encrypt(plainText []byte) ([]byte, error)

	// Decrypt 解密
	Decrypt(cipherText []byte) ([]byte, error)
}

// ------------------------------------ AES-CBC ------------------------------------

type cbccrypto struct {
	key  []byte
	iv   []byte
	mode AESPaddingMode
}

func (c *cbccrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	switch c.mode {
	case AES_ZERO:
		plainText = ZeroPadding(plainText, block.BlockSize())
	case AES_PKCS5:
		plainText = PKCS5Padding(plainText, block.BlockSize())
	case AES_PKCS7:
		plainText = PKCS5Padding(plainText, len(c.key))
	}

	cipherText := make([]byte, len(plainText))

	blockMode := cipher.NewCBCEncrypter(block, c.iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

func (c *cbccrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	blockMode := cipher.NewCBCDecrypter(block, c.iv)
	blockMode.CryptBlocks(plainText, cipherText)

	switch c.mode {
	case AES_ZERO:
		plainText = ZeroUnPadding(plainText)
	case AES_PKCS5:
		plainText = PKCS5Unpadding(plainText, block.BlockSize())
	case AES_PKCS7:
		plainText = PKCS5Unpadding(plainText, len(c.key))
	}

	return plainText, nil
}

// NewCBCCrypto 生成 AES-CBC 加密模式
func NewCBCCrypto(key, iv []byte, mode AESPaddingMode) AESCrypto {
	return &cbccrypto{
		key:  key,
		iv:   iv,
		mode: mode,
	}
}

// ------------------------------------ AES-ECB ------------------------------------

type ecbcrypto struct {
	key  []byte
	mode AESPaddingMode
}

func (c *ecbcrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	switch c.mode {
	case AES_ZERO:
		plainText = ZeroPadding(plainText, block.BlockSize())
	case AES_PKCS5:
		plainText = PKCS5Padding(plainText, block.BlockSize())
	case AES_PKCS7:
		plainText = PKCS5Padding(plainText, len(c.key))
	}

	cipherText := make([]byte, len(plainText))

	blockMode := NewECBEncrypter(block)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

func (c *ecbcrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	plainText := make([]byte, len(cipherText))

	blockMode := NewECBDecrypter(block)
	blockMode.CryptBlocks(plainText, cipherText)

	switch c.mode {
	case AES_ZERO:
		plainText = ZeroUnPadding(plainText)
	case AES_PKCS5:
		plainText = PKCS5Unpadding(plainText, block.BlockSize())
	case AES_PKCS7:
		plainText = PKCS5Unpadding(plainText, len(c.key))
	}

	return plainText, nil
}

// NewECBCrypto 生成 AES-ECB 加密模式
func NewECBCrypto(key []byte, mode AESPaddingMode) AESCrypto {
	return &ecbcrypto{
		key:  key,
		mode: mode,
	}
}

// ------------------------------------ AES-CFB ------------------------------------

type cfbcrypto struct {
	key []byte
	iv  []byte
}

func (c *cfbcrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewCFBEncrypter(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

func (c *cfbcrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewCFBDecrypter(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// NewCFBCrypto returns a new aes-cfb crypto
func NewCFBCrypto(key, iv []byte) AESCrypto {
	return &cfbcrypto{
		key: key,
		iv:  iv,
	}
}

// ------------------------------------ AES-OFB ------------------------------------

type ofbcrypto struct {
	key []byte
	iv  []byte
}

func (c *ofbcrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewOFB(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

func (c *ofbcrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewOFB(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// NewOFBCrypto returns a new aes-ofb crypto
func NewOFBCrypto(key, iv []byte) AESCrypto {
	return &ofbcrypto{
		key: key,
		iv:  iv,
	}
}

// ------------------------------------ AES-CTR ------------------------------------

type ctrcrypto struct {
	key []byte
	iv  []byte
}

func (c *ctrcrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	cipherText := make([]byte, len(plainText))

	stream := cipher.NewCTR(block, c.iv)
	stream.XORKeyStream(cipherText, plainText)

	return cipherText, nil
}

func (c *ctrcrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plainText := make([]byte, len(cipherText))

	stream := cipher.NewCTR(block, c.iv)
	stream.XORKeyStream(plainText, cipherText)

	return plainText, nil
}

// NewCTRCrypto returns a new aes-ctr crypto
func NewCTRCrypto(key, iv []byte) AESCrypto {
	return &ctrcrypto{
		key: key,
		iv:  iv,
	}
}

// ------------------------------------ AES-GCM ------------------------------------

type gcmcrypto struct {
	key   []byte
	nonce []byte
}

func (c *gcmcrypto) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aesgcm.NonceSize() {
		return nil, errors.New("nonce length must equal gcm standard nonce size")
	}

	return aesgcm.Seal(nil, c.nonce, plainText, nil), nil
}

func (c *gcmcrypto) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aesgcm.NonceSize() {
		return nil, errors.New("nonce length must equal gcm standard nonce size")
	}

	return aesgcm.Open(nil, c.nonce, cipherText, nil)
}

// NewGCMCrypto returns a new aes-gcm crypto
func NewGCMCrypto(key, nonce []byte) AESCrypto {
	return &gcmcrypto{
		key:   key,
		nonce: nonce,
	}
}

// ------------------------------------ RSA ------------------------------------

// GenerateRSAKey returns rsa private and public key.
func GenerateRSAKey(bitSize int, mode RSAPaddingMode) (privateKey, publicKey []byte, err error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bitSize)

	if err != nil {
		return
	}

	switch mode {
	case RSA_PKCS1:
		privateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		})

		publicKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&privKey.PublicKey),
		})
	case RSA_PKCS8:
		privBlock := &pem.Block{
			Type: "PRIVATE KEY",
		}
		privBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(privKey)

		if err != nil {
			return
		}

		pubBlock := &pem.Block{
			Type: "PUBLIC KEY",
		}
		pubBlock.Bytes, err = x509.MarshalPKIXPublicKey(&privKey.PublicKey)

		if err != nil {
			return
		}

		privateKey = pem.EncodeToMemory(privBlock)
		publicKey = pem.EncodeToMemory(pubBlock)
	}

	return
}

// PrivateKey RSA private key
type PrivateKey struct {
	key *rsa.PrivateKey
}

// Decrypt rsa decrypt with PKCS #1 v1.5
func (pk *PrivateKey) Decrypt(cipherText []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, pk.key, cipherText)
}

// DecryptOAEP rsa decrypt with PKCS #1 OAEP.
func (pk *PrivateKey) DecryptOAEP(hash crypto.Hash, cipherText []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	return rsa.DecryptOAEP(hash.New(), rand.Reader, pk.key, cipherText, nil)
}

// Sign returns sha-with-rsa signature.
func (pk *PrivateKey) Sign(hash crypto.Hash, data []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	h := hash.New()
	h.Write(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, pk.key, hash, h.Sum(nil))

	if err != nil {
		return nil, err
	}

	return signature, nil
}

// NewPrivateKeyFromPemBlock returns new private key with pem block.
func NewPrivateKeyFromPemBlock(mode RSAPaddingMode, pemBlock []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(pemBlock)

	if block == nil {
		return nil, errors.New("no PEM data is found")
	}

	var (
		pk  any
		err error
	)

	switch mode {
	case RSA_PKCS1:
		pk, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case RSA_PKCS8:
		pk, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}

	if err != nil {
		return nil, err
	}

	return &PrivateKey{key: pk.(*rsa.PrivateKey)}, nil
}

// NewPrivateKeyFromPemFile returns new private key with pem file.
func NewPrivateKeyFromPemFile(mode RSAPaddingMode, pemFile string) (*PrivateKey, error) {
	keyPath, err := filepath.Abs(pemFile)

	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(keyPath)

	if err != nil {
		return nil, err
	}

	return NewPrivateKeyFromPemBlock(mode, b)
}

// NewPrivateKeyFromPfxFile returns private key with pfx(p12) file.
func NewPrivateKeyFromPfxFile(pfxFile, password string) (*PrivateKey, error) {
	cert, err := LoadCertFromPfxFile(pfxFile, password)

	if err != nil {
		return nil, err
	}

	return &PrivateKey{key: cert.PrivateKey.(*rsa.PrivateKey)}, nil
}

// PublicKey RSA public key
type PublicKey struct {
	key *rsa.PublicKey
}

// Encrypt rsa encrypt with PKCS #1 v1.5
func (pk *PublicKey) Encrypt(plainText []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pk.key, plainText)
}

// EncryptOAEP rsa encrypt with PKCS #1 OAEP.
func (pk *PublicKey) EncryptOAEP(hash crypto.Hash, plainText []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	return rsa.EncryptOAEP(hash.New(), rand.Reader, pk.key, plainText, nil)
}

// Verify verifies the sha-with-rsa signature.
func (pk *PublicKey) Verify(hash crypto.Hash, data, signature []byte) error {
	if !hash.Available() {
		return fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	h := hash.New()
	h.Write(data)

	return rsa.VerifyPKCS1v15(pk.key, hash, h.Sum(nil), signature)
}

// NewPublicKeyFromPemBlock returns new public key with pem block.
func NewPublicKeyFromPemBlock(mode RSAPaddingMode, pemBlock []byte) (*PublicKey, error) {
	block, _ := pem.Decode(pemBlock)

	if block == nil {
		return nil, errors.New("no PEM data is found")
	}

	var (
		pk  any
		err error
	)

	switch mode {
	case RSA_PKCS1:
		pk, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case RSA_PKCS8:
		pk, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return nil, err
	}

	return &PublicKey{key: pk.(*rsa.PublicKey)}, nil
}

// NewPublicKeyFromPemFile returns new public key with pem file.
func NewPublicKeyFromPemFile(mode RSAPaddingMode, pemFile string) (*PublicKey, error) {
	keyPath, err := filepath.Abs(pemFile)

	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(keyPath)

	if err != nil {
		return nil, err
	}

	return NewPublicKeyFromPemBlock(mode, b)
}

// NewPublicKeyFromDerBlock returns public key with DER block.
// NOTE: PEM format with -----BEGIN CERTIFICATE----- | -----END CERTIFICATE-----
// CMD: openssl x509 -inform der -in cert.cer -out cert.pem
func NewPublicKeyFromDerBlock(pemBlock []byte) (*PublicKey, error) {
	block, _ := pem.Decode(pemBlock)

	if block == nil {
		return nil, errors.New("no PEM data is found")
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, err
	}

	return &PublicKey{key: cert.PublicKey.(*rsa.PublicKey)}, nil
}

// NewPublicKeyFromDerFile returns public key with DER file.
// NOTE: PEM format with -----BEGIN CERTIFICATE----- | -----END CERTIFICATE-----
// CMD: openssl x509 -inform der -in cert.cer -out cert.pem
func NewPublicKeyFromDerFile(pemFile string) (*PublicKey, error) {
	keyPath, err := filepath.Abs(pemFile)

	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(keyPath)

	if err != nil {
		return nil, err
	}

	return NewPublicKeyFromDerBlock(b)
}

func ZeroPadding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{0}, padding)

	return append(cipherText, padText...)
}

func ZeroUnPadding(plainText []byte) []byte {
	return bytes.TrimRightFunc(plainText, func(r rune) bool {
		return r == rune(0)
	})
}

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize

	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

func PKCS5Unpadding(plainText []byte, blockSize int) []byte {
	length := len(plainText)
	unpadding := int(plainText[length-1])

	if unpadding < 1 || unpadding > blockSize {
		unpadding = 0
	}

	return plainText[:(length - unpadding)]
}

// --------------------------------- AES-256-ECB ---------------------------------

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter returns a BlockMode which encrypts in electronic code book mode, using the given Block.
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}

	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}

	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter returns a BlockMode which decrypts in electronic code book mode, using the given Block.
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}

	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}

	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])

		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}
