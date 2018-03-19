package utils

import (
	"crypto/md5"
	"fmt"
)

// MD5 获取字符串md5值
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
}
