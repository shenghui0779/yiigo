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

// Date 时间戳格式化日期，默认：2006-01-02 15:04:05
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}

// UniqueInt int切片去重
func UniqueInt(in []int) []int {
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

// UniqueInt64 int64切片去重
func UniqueInt64(in []int64) []int64 {
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

// UniqueString string切片去重
func UniqueString(in []string) []string {
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

// InArrayInt Checks if a int value exists in an int slice
func InArrayInt(needle int, haystack []int) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InArrayInt64 Checks if a int64 value exists in an int64 slice
func InArrayInt64(needle int64, haystack []int64) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InArrayFloat64 Checks if a float64 value exists in a float64 slice
func InArrayFloat64(needle float64, haystack []float64) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InArrayString Checks if a string value exists in a string slice
func InArrayString(needle string, haystack []string) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
