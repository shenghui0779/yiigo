package utils

import (
	"crypto/md5"
	"fmt"
	"time"
)

// MD5 获取字符串md5值
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Date 时间戳格式化日期，默认：2006-01-02 15:04:05
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}
