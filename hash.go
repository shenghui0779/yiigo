package yiigo

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MD5 calculates the md5 hash of a string.
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// SHA1 calculates the sha1 hash of a string.
func SHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 calculates the sha256 hash of a string.
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// Hash generates a hash value.
func Hash(hash crypto.Hash, s string) (string, error) {
	if !hash.Available() {
		return "", fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	h := hash.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HMacSHA256 generates a keyed sha256 hash value.
func HMacSHA256(key, s string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(s))

	return hex.EncodeToString(mac.Sum(nil))
}

// HMac generates a keyed hash value.
func HMac(hash crypto.Hash, key, s string) (string, error) {
	if !hash.Available() {
		return "", fmt.Errorf("crypto: requested hash function (%s) is unavailable", hash.String())
	}

	mac := hmac.New(hash.New, []byte(key))
	mac.Write([]byte(s))

	return hex.EncodeToString(mac.Sum(nil)), nil
}
