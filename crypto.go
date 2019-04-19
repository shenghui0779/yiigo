package yiigo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AESCBCEncrypt AES CBC encrypt with PKCS#7 padding
func AESCBCEncrypt(plainText, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	plainText = PKCS7Padding(plainText, len(key))

	cipherText := make([]byte, len(plainText))

	if len(iv) == 0 {
		iv = key[:block.BlockSize()]
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// AESCBCDecrypt AES CBC decrypt with PKCS#7 unpadding
func AESCBCDecrypt(cipherText, key []byte, iv ...byte) ([]byte, error) {
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

	return PKCS7UnPadding(plainText, len(key)), nil
}

// PKCS7Padding PKCS7 padding
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize

	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

// PKCS7UnPadding PKCS7 unpadding
func PKCS7UnPadding(plainText []byte, blockSize int) []byte {
	l := len(plainText)
	unpadding := int(plainText[l-1])

	if unpadding < 1 || unpadding > blockSize {
		unpadding = 0
	}

	return plainText[:(l - unpadding)]
}
