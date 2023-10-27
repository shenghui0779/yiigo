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

// AESPadding AES 填充模式
type AESPadding interface {
	// BlockSize 填充字节数
	BlockSize() int

	// Padding 填充
	Padding(data []byte) []byte

	// UnPadding 移除填充
	UnPadding(data []byte) []byte
}

// RSAPadding RSA PEM 填充模式
type RSAPadding int

const (
	RSA_PKCS1 RSAPadding = 1 // PKCS#1 (格式：`RSA PRIVATE KEY` 和 `RSA PUBLIC KEY`)
	RSA_PKCS8 RSAPadding = 8 // PKCS#8 (格式：`PRIVATE KEY` 和 `PUBLIC KEY`)
)

// ------------------------------------ AES ------------------------------------

// AES-CBC 加密模式
type AesCBC struct {
	key  []byte
	iv   []byte
	mode AESPadding
}

// Encrypt AES-CBC 加密
func (c *AesCBC) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plainText = c.mode.Padding(plainText)

	bm := cipher.NewCBCEncrypter(block, c.iv)
	if len(plainText)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	cipherText := make([]byte, len(plainText))
	bm.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// Decrypt AES-CBC 解密
func (c *AesCBC) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(c.iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	bm := cipher.NewCBCDecrypter(block, c.iv)
	if len(cipherText)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	plainText := make([]byte, len(cipherText))
	bm.CryptBlocks(plainText, cipherText)

	return c.mode.UnPadding(plainText), nil
}

// NewAesCBC 生成 AES-CBC 加密模式
func NewAesCBC(key, iv []byte, padding AESPadding) *AesCBC {
	return &AesCBC{
		key:  key,
		iv:   iv,
		mode: padding,
	}
}

// AES-ECB 加密模式
type AesECB struct {
	key  []byte
	mode AESPadding
}

// Encrypt AES-ECB 加密
func (c *AesECB) Encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	plainText = c.mode.Padding(plainText)

	bm := NewECBEncrypter(block)
	if len(plainText)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	cipherText := make([]byte, len(plainText))
	bm.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// Decrypt AES-ECB 解密
func (c *AesECB) Decrypt(cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	bm := NewECBDecrypter(block)
	if len(cipherText)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	plainText := make([]byte, len(cipherText))
	bm.CryptBlocks(plainText, cipherText)

	return c.mode.UnPadding(plainText), nil
}

// NewAesECB 生成 AES-ECB 加密模式
func NewAesECB(key []byte, padding AESPadding) *AesECB {
	return &AesECB{
		key:  key,
		mode: padding,
	}
}

// AES-CFB 加密模式
type AesCFB struct {
	key []byte
	iv  []byte
}

// Encrypt AES-CFB 加密
func (c *AesCFB) Encrypt(plainText []byte) ([]byte, error) {
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

// Decrypt AES-CFB 解密
func (c *AesCFB) Decrypt(cipherText []byte) ([]byte, error) {
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

// NewAesCFB 生成 AES-CFB 加密模式
func NewAesCFB(key, iv []byte) *AesCFB {
	return &AesCFB{
		key: key,
		iv:  iv,
	}
}

// AES-OFB 加密模式
type AesOFB struct {
	key []byte
	iv  []byte
}

// Encrypt AES-OFB 加密
func (c *AesOFB) Encrypt(plainText []byte) ([]byte, error) {
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

// Decrypt AES-OFB 解密
func (c *AesOFB) Decrypt(cipherText []byte) ([]byte, error) {
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

// NewAesOFB 生成 AES-OFB 加密模式
func NewAesOFB(key, iv []byte) *AesOFB {
	return &AesOFB{
		key: key,
		iv:  iv,
	}
}

// AES-CTR 加密模式
type AesCTR struct {
	key []byte
	iv  []byte
}

// Encrypt AES-CTR 加密
func (c *AesCTR) Encrypt(plainText []byte) ([]byte, error) {
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

// Decrypt AES-CTR 解密
func (c *AesCTR) Decrypt(cipherText []byte) ([]byte, error) {
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

// NewAesCTR 生成 AES-CTR 加密模式
func NewAesCTR(key, iv []byte) *AesCTR {
	return &AesCTR{
		key: key,
		iv:  iv,
	}
}

// AES-GCM 加密模式
type AesGCM struct {
	key   []byte
	nonce []byte
}

// Encrypt AES-GCM 加密
func (c *AesGCM) Encrypt(plainText, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	if uint64(len(plainText)) > ((1<<32)-2)*uint64(block.BlockSize()) {
		return nil, errors.New("message too large for GCM")
	}

	return aead.Seal(nil, c.nonce, plainText, additionalData), nil
}

// Decrypt AES-GCM 解密
func (c *AesGCM) Decrypt(cipherText, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(c.nonce) != aesgcm.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aesgcm.Open(nil, c.nonce, cipherText, additionalData)
}

// NewAesGCM 生成 AES-GCM 加密模式
func NewAesGCM(key, nonce []byte) *AesGCM {
	return &AesGCM{
		key:   key,
		nonce: nonce,
	}
}

// ------------------------------------ RSA ------------------------------------

// GenerateRSAKey 生成RSA私钥和公钥
func GenerateRSAKey(bitSize int, padding RSAPadding) (privateKey, publicKey []byte, err error) {
	prvKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return
	}

	switch padding {
	case RSA_PKCS1:
		privateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(prvKey),
		})

		publicKey = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&prvKey.PublicKey),
		})
	case RSA_PKCS8:
		prvBlock := &pem.Block{
			Type: "PRIVATE KEY",
		}

		prvBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(prvKey)
		if err != nil {
			return
		}

		pubBlock := &pem.Block{
			Type: "PUBLIC KEY",
		}

		pubBlock.Bytes, err = x509.MarshalPKIXPublicKey(&prvKey.PublicKey)
		if err != nil {
			return
		}

		privateKey = pem.EncodeToMemory(prvBlock)
		publicKey = pem.EncodeToMemory(pubBlock)
	}

	return
}

// PrivateKey RSA私钥
type PrivateKey struct {
	key *rsa.PrivateKey
}

// Decrypt RSA私钥 PKCS#1 v1.5 解密
func (pk *PrivateKey) Decrypt(cipherText []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, pk.key, cipherText)
}

// DecryptOAEP RSA私钥 PKCS#1 OAEP 解密
func (pk *PrivateKey) DecryptOAEP(hash crypto.Hash, cipherText []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	return rsa.DecryptOAEP(hash.New(), rand.Reader, pk.key, cipherText, nil)
}

// Sign RSA私钥签名
func (pk *PrivateKey) Sign(hash crypto.Hash, data []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	h := hash.New()
	h.Write(data)

	return rsa.SignPKCS1v15(rand.Reader, pk.key, hash, h.Sum(nil))
}

// NewPrivateKeyFromPemBlock 通过PEM字节生成RSA私钥
func NewPrivateKeyFromPemBlock(padding RSAPadding, pemBlock []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, errors.New("no PEM data is found")
	}

	var (
		pk  any
		err error
	)

	switch padding {
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

// NewPrivateKeyFromPemFile  通过PEM文件生成RSA私钥
func NewPrivateKeyFromPemFile(padding RSAPadding, pemFile string) (*PrivateKey, error) {
	keyPath, err := filepath.Abs(pemFile)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return NewPrivateKeyFromPemBlock(padding, b)
}

// NewPrivateKeyFromPfxFile 通过pfx(p12)证书生成RSA私钥
// 注意：证书需采用「TripleDES-SHA1」加密方式
func NewPrivateKeyFromPfxFile(pfxFile, password string) (*PrivateKey, error) {
	cert, err := LoadCertFromPfxFile(pfxFile, password)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key: cert.PrivateKey.(*rsa.PrivateKey)}, nil
}

// PublicKey RSA公钥
type PublicKey struct {
	key *rsa.PublicKey
}

// Encrypt RSA公钥 PKCS#1 v1.5 加密
func (pk *PublicKey) Encrypt(plainText []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pk.key, plainText)
}

// EncryptOAEP RSA公钥 PKCS#1 OAEP 加密
func (pk *PublicKey) EncryptOAEP(hash crypto.Hash, plainText []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	return rsa.EncryptOAEP(hash.New(), rand.Reader, pk.key, plainText, nil)
}

// Verify RSA公钥验签
func (pk *PublicKey) Verify(hash crypto.Hash, data, signature []byte) error {
	if !hash.Available() {
		return fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	h := hash.New()
	h.Write(data)

	return rsa.VerifyPKCS1v15(pk.key, hash, h.Sum(nil), signature)
}

// NewPublicKeyFromPemBlock 通过PEM字节生成RSA公钥
func NewPublicKeyFromPemBlock(padding RSAPadding, pemBlock []byte) (*PublicKey, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, errors.New("no PEM data is found")
	}

	var (
		pk  any
		err error
	)

	switch padding {
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

// NewPublicKeyFromPemFile 通过PEM文件生成RSA公钥
func NewPublicKeyFromPemFile(padding RSAPadding, pemFile string) (*PublicKey, error) {
	keyPath, err := filepath.Abs(pemFile)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return NewPublicKeyFromPemBlock(padding, b)
}

// NewPublicKeyFromDerBlock 通过DER字节生成RSA公钥
// 注意PEM格式: -----BEGIN CERTIFICATE----- | -----END CERTIFICATE-----
// DER转换命令: openssl x509 -inform der -in cert.cer -out cert.pem
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

// NewPublicKeyFromDerFile 通过DER证书生成RSA公钥
// 注意PEM格式: -----BEGIN CERTIFICATE----- | -----END CERTIFICATE-----
// DER转换命令: openssl x509 -inform der -in cert.cer -out cert.pem
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

// --------------------------------- AES Padding ---------------------------------

type pkcsPadding struct {
	blockSize int
}

func (p *pkcsPadding) BlockSize() int {
	return p.blockSize
}

func (p *pkcsPadding) Padding(data []byte) []byte {
	padding := p.blockSize - len(data)%p.blockSize
	if padding == 0 {
		padding = p.blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(data, padText...)
}

func (p *pkcsPadding) UnPadding(data []byte) []byte {
	length := len(data)

	unpadding := int(data[length-1])
	if unpadding < 1 || unpadding > p.blockSize {
		unpadding = 0
	}

	return data[:(length - unpadding)]
}

// AES_PKCS5 pcks#5填充模式(16个字节)
func AES_PKCS5() AESPadding {
	return &pkcsPadding{
		blockSize: aes.BlockSize,
	}
}

// AES_PKCS7 pcks#7填充模式(自定义字节数)
func AES_PKCS7(blockSize int) AESPadding {
	return &pkcsPadding{
		blockSize: blockSize,
	}
}

type zeroPadding struct {
	blockSize int
}

func (p *zeroPadding) BlockSize() int {
	return p.blockSize
}

func (p *zeroPadding) Padding(data []byte) []byte {
	padding := p.blockSize - len(data)%p.blockSize
	padText := bytes.Repeat([]byte{0}, padding)

	return append(data, padText...)
}

func (p *zeroPadding) UnPadding(plainText []byte) []byte {
	return bytes.TrimRightFunc(plainText, func(r rune) bool {
		return r == rune(0)
	})
}

// AES_ZERO 0填充(16个字节)
func AES_ZERO() AESPadding {
	return &zeroPadding{
		blockSize: aes.BlockSize,
	}
}

// --------------------------------- ECB BlockMode ---------------------------------

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

// NewECBEncrypter 生成ECB加密器
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

// NewECBDecrypter 生成ECB解密器
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
