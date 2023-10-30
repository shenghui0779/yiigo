package yiigo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// CipherText 加密文本
type CipherText struct {
	bytes   []byte
	tagsize int
}

// Bytes 返回加密数据base64字符串
func (ct *CipherText) String() string {
	return base64.StdEncoding.EncodeToString(ct.bytes)
}

// Bytes 获取加密数据的bytes
func (ct *CipherText) Bytes() []byte {
	return ct.bytes
}

// Data 获取GCM加密数据的真实数据
func (ct *CipherText) Data() []byte {
	return ct.bytes[:len(ct.bytes)-ct.tagsize]
}

// Tag 获取GCM加密数据的tag
func (ct *CipherText) Tag() []byte {
	return ct.bytes[len(ct.bytes)-ct.tagsize:]
}

// ------------------------------------ AES-CBC ------------------------------------

// AESEncryptCBC AES-CBC 加密
func AESEncryptCBC(key, iv, data []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	data = pkcs7padding(data, block.BlockSize())

	bm := cipher.NewCBCEncrypter(block, iv)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESEncryptCBCWithPaddingSize AES-CBC 加密(指定填充字节大小)
func AESEncryptCBCWithPaddingSize(key, iv, data []byte, paddingSize uint8) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	data = pkcs7padding(data, int(paddingSize))

	bm := cipher.NewCBCEncrypter(block, iv)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESDecryptCBC AES-CBC 解密
func AESDecryptCBC(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	bm := cipher.NewCBCDecrypter(block, iv)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7unpadding(out), nil
}

// ------------------------------------ AES-ECB ------------------------------------

// AESEncryptECB AES-ECB 加密
func AESEncryptECB(key, data []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data = pkcs7padding(data, block.BlockSize())

	bm := NewECBEncrypter(block)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESEncryptECBWithPaddingSize AES-ECB 加密(指定填充字节大小)
func AESEncryptECBWithPaddingSize(key, data []byte, paddingSize uint8) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data = pkcs7padding(data, int(paddingSize))

	bm := NewECBEncrypter(block)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESDecryptECB AES-ECB 解密
func AESDecryptECB(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bm := NewECBDecrypter(block)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7unpadding(out), nil
}

// ------------------------------------ AES-CFB ------------------------------------

// AESEncryptCFB AES-CFB 加密
func AESEncryptCFB(key, iv, data []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESDecryptCFB AES-CFB 解密
func AESDecryptCFB(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-OFB ------------------------------------

// AESEncryptOFB AES-OFB 加密
func AESEncryptOFB(key, iv, data []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESDecryptOFB AES-OFB 解密
func AESDecryptOFB(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-CTR ------------------------------------

// AESEncryptCTR AES-CTR 加密
func AESEncryptCTR(key, iv, data []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(out, data)

	return &CipherText{
		bytes: out,
	}, nil
}

// AESDecryptCTR AES-CTR 解密
func AESDecryptCTR(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-GCM ------------------------------------

// AESEncryptGCM AES-GCM 加密 (NonceSize = 12 & TagSize = 16)
func AESEncryptGCM(key, nonce, data, aad []byte) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	if uint64(len(data)) > ((1<<32)-2)*uint64(block.BlockSize()) {
		return nil, errors.New("message too large for GCM")
	}

	return &CipherText{
		bytes:   aead.Seal(nil, nonce, data, aad),
		tagsize: aead.Overhead(),
	}, nil
}

// AESDecryptGCM AES-GCM 解密 (NonceSize = 12 & TagSize = 16)
func AESDecryptGCM(key, nonce []byte, data, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aead.Open(nil, nonce, data, aad)
}

// AESEncryptGCMWithTagSize AES-GCM 指定TagSize加密 (12 <= TagSize <= 16)
func AESEncryptGCMWithTagSize(key, nonce, data, aad []byte, tagSize int) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	if uint64(len(data)) > ((1<<32)-2)*uint64(block.BlockSize()) {
		return nil, errors.New("message too large for GCM")
	}

	return &CipherText{
		bytes:   aead.Seal(nil, nonce, data, aad),
		tagsize: aead.Overhead(),
	}, nil
}

// AESDecryptGCMWithTagSize AES-GCM 指定TagSize解密 (12 <= TagSize <= 16)
func AESDecryptGCMWithTagSize(key, nonce []byte, data, aad []byte, tagSize int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aead.Open(nil, nonce, data, aad)
}

// AESEncryptGCMWithNonceSize AES-GCM 指定NonceSize加密 (NonceSize > 0)
func AESEncryptGCMWithNonceSize(key, nonce, data, aad []byte, nonceSize int) (*CipherText, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	if uint64(len(data)) > ((1<<32)-2)*uint64(block.BlockSize()) {
		return nil, errors.New("message too large for GCM")
	}

	return &CipherText{
		bytes:   aead.Seal(nil, nonce, data, aad),
		tagsize: aead.Overhead(),
	}, nil
}

// AESDecryptWithNonceSize AES-GCM 指定NonceSize解密 (NonceSize > 0)
func AESDecryptGCMWithNonceSize(key, nonce []byte, data, aad []byte, nonceSize int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aead.Open(nil, nonce, data, aad)
}

// --------------------------------- AES Padding ---------------------------------

// func ZeroPadding(data []byte, blockSize int) []byte {
// 	padding := blockSize - len(data)%blockSize
// 	b := bytes.Repeat([]byte{0}, padding)

// 	return append(data, b...)
// }

// func ZeroUnPadding(data []byte) []byte {
// 	return bytes.TrimRightFunc(data, func(r rune) bool {
// 		return r == rune(0)
// 	})
// }

func pkcs7padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	if padding == 0 {
		padding = blockSize
	}

	b := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(data, b...)
}

func pkcs7unpadding(data []byte) []byte {
	length := len(data)
	padding := int(data[length-1])

	return data[:(length - padding)]
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
