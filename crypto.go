package yiigo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AESCBCEncrypt AES CBC encrypt
func AESCBCEncrypt(plainText, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	plainText = PKCS7Padding(plainText, blockSize)

	cipherText := make([]byte, len(plainText))

	if len(iv) == 0 {
		iv = key[:blockSize]
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// AESCBCDecrypt AES CBC decrypt
func AESCBCDecrypt(cipherText, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	plainText := make([]byte, len(cipherText))

	if len(iv) == 0 {
		iv = key[:blockSize]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(plainText, cipherText)

	return PKCS7UnPadding(plainText), nil
}

// PKCS7Padding PKCS7 padding
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(cipherText, padText...)
}

// PKCS7UnPadding PKCS7 unpadding
func PKCS7UnPadding(decryptedData []byte) []byte {
	l := len(decryptedData)
	unpadding := int(decryptedData[l-1])

	return decryptedData[:(l - unpadding)]
}
