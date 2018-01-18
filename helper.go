package yiigo

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

// Date 时间戳格式化日期，format: 2006-01-02 15:04:05
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}

// IntUnique int切片去重
func IntUnique(in []int) []int {
	out := make([]int, 0, len(in))

	for _, i := range in {
		exist := false

		for _, o := range out {
			if i == o {
				exist = true
				break
			}
		}

		if !exist {
			out = append(out, i)
		}
	}

	return out
}

// Int64Unique int64切片去重
func Int64Unique(in []int64) []int64 {
	out := make([]int64, 0, len(in))

	for _, i := range in {
		exist := false

		for _, o := range out {
			if i == o {
				exist = true
				break
			}
		}

		if !exist {
			out = append(out, i)
		}
	}

	return out
}

// StringUnique string切片去重
func StringUnique(in []string) []string {
	out := make([]string, 0, len(in))

	for _, i := range in {
		exist := false

		for _, o := range out {
			if i == o {
				exist = true
				break
			}
		}

		if !exist {
			out = append(out, i)
		}
	}

	return out
}
