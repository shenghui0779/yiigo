package yiigo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AESCBCEncrypt AES CBC encrypt
func AESCBCEncrypt(data, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	data = PKCS7Padding(data, blockSize)

	encryptedData := make([]byte, len(data))

	if len(iv) == 0 {
		iv = key[:blockSize]
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(encryptedData, data)

	return encryptedData, nil
}

// AESCBCDecrypt AES CBC decrypt
func AESCBCDecrypt(encryptedData, key []byte, iv ...byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	decryptedData := make([]byte, len(encryptedData))

	if len(iv) == 0 {
		iv = key[:blockSize]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(decryptedData, encryptedData)

	return PKCS7UnPadding(decryptedData), nil
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
