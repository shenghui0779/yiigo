package yiigo

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"

	"github.com/hashicorp/go-version"
)

// Nonce 生成随机串(size应为偶数)
func Nonce(size uint8) string {
	nonce := make([]byte, size/2)
	io.ReadFull(rand.Reader, nonce)

	return hex.EncodeToString(nonce)
}

// MarshalNoEscapeHTML 不带HTML转义的JSON序列化
func MarshalNoEscapeHTML(v any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	b := buf.Bytes()

	// 去掉 go std 给末尾加的 '\n'
	// @see https://github.com/golang/go/issues/7767
	if l := len(b); l != 0 && b[l-1] == '\n' {
		b = b[:l-1]
	}

	return b, nil
}

// VersionCompare 语义化的版本比较，支持：>, >=, =, !=, <, <=, | (or), & (and).
// 参数 `rangeVer` 示例：1.0.0, =1.0.0, >2.0.0, >=1.0.0&<2.0.0, <2.0.0|>3.0.0, !=4.0.4
func VersionCompare(rangeVer, curVer string) (bool, error) {
	semVer, err := version.NewVersion(curVer)
	if err != nil {
		return false, err
	}

	orVers := strings.Split(rangeVer, "|")

	for _, ver := range orVers {
		andVers := strings.Split(ver, "&")

		constraints, err := version.NewConstraint(strings.Join(andVers, ","))
		if err != nil {
			return false, err
		}

		if constraints.Check(semVer) {
			return true, nil
		}
	}

	return false, nil
}
