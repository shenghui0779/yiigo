package yiigo

import (
	"bytes"
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

// AesCBCEncrypt AES-CBC 加密
func AesCBCEncrypt(block cipher.Block, iv, data []byte) (*CipherText, error) {
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

// AesCBCDecrypt AES-CBC 解密
func AesCBCDecrypt(block cipher.Block, iv, data []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	bm := cipher.NewCBCDecrypter(block, iv)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7unpadding(out, block.BlockSize()), nil
}

// ------------------------------------ AES-ECB ------------------------------------

// AesECBEncrypt AES-ECB 加密
func AesECBEncrypt(block cipher.Block, data []byte) (*CipherText, error) {
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

// AesECBDecrypt AES-ECB 解密
func AesECBDecrypt(block cipher.Block, data []byte) ([]byte, error) {
	bm := NewECBDecrypter(block)
	if len(data)%bm.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	out := make([]byte, len(data))
	bm.CryptBlocks(out, data)

	return pkcs7unpadding(out, block.BlockSize()), nil
}

// ------------------------------------ AES-CFB ------------------------------------

// AesCFBEncrypt AES-CFB 加密
func AesCFBEncrypt(block cipher.Block, iv, data []byte) (*CipherText, error) {
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

// AesCFBDecrypt AES-CFB 解密
func AesCFBDecrypt(block cipher.Block, iv, data []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-OFB ------------------------------------

// AesOFBEncrypt AES-OFB 加密
func AesOFBEncrypt(block cipher.Block, iv, data []byte) (*CipherText, error) {
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

// AesOFBDecrypt AES-OFB 解密
func AesOFBDecrypt(block cipher.Block, iv, data []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-CTR ------------------------------------

// Encrypt AES-CTR 加密
func AesCTREncrypt(block cipher.Block, iv, data []byte) (*CipherText, error) {
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

// AesCTRDecrypt AES-CTR 解密
func AesCTRDecrypt(block cipher.Block, iv, data []byte) ([]byte, error) {
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	out := make([]byte, len(data))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(out, data)

	return out, nil
}

// ------------------------------------ AES-GCM ------------------------------------

// AesGCMEncrypt AES-GCM 加密 (NonceSize = 12 & TagSize = 16)
func AesGCMEncrypt(block cipher.Block, nonce, data, aad []byte) (*CipherText, error) {
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

// AesGCMDecrypt AES-GCM 解密 (NonceSize = 12 & TagSize = 16)
func AesGCMDecrypt(block cipher.Block, nonce []byte, data, aad []byte) ([]byte, error) {
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aead.Open(nil, nonce, data, aad)
}

// AesGCMEncryptWithTagSize AES-GCM 指定TagSize加密 (12 <= TagSize <= 16)
func AesGCMEncryptWithTagSize(block cipher.Block, nonce, data, aad []byte, tagSize int) (*CipherText, error) {
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

// AesGCMDecryptWithTagSize AES-GCM 指定TagSize解密 (12 <= TagSize <= 16)
func AesGCMDecryptWithTagSize(block cipher.Block, nonce []byte, data, aad []byte, tagSize int) ([]byte, error) {
	aead, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("incorrect nonce length given to GCM")
	}

	return aead.Open(nil, nonce, data, aad)
}

// AesGCMEncryptWithNonceSize AES-GCM 指定NonceSize加密 (NonceSize > 0)
func AesGCMEncryptWithNonceSize(block cipher.Block, nonce, data, aad []byte, nonceSize int) (*CipherText, error) {
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

// AesGCMDecryptWithNonceSize AES-GCM 指定NonceSize解密 (NonceSize > 0)
func AesGCMDecryptWithNonceSize(block cipher.Block, nonce []byte, data, aad []byte, nonceSize int) ([]byte, error) {
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

func pkcs7unpadding(data []byte, blockSize int) []byte {
	length := len(data)
	padding := int(data[length-1])

	if padding < 1 || padding > blockSize {
		padding = 0
	}

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
