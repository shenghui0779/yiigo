package yiigo

import (
	"net/url"
	"sort"
	"strings"
)

// V 用于处理 k-v 需要格式化的场景，如：签名
type V map[string]string

// Set 设置 k-v
func (v V) Set(key, value string) {
	v[key] = value
}

// Get 获取值
func (v V) Get(key string) string {
	return v[key]
}

// Del 删除Key
func (v V) Del(key string) {
	delete(v, key)
}

// Has 判断Key是否存在
func (v V) Has(key string) bool {
	_, ok := v[key]

	return ok
}

// Encode 通过自定义的符号和分隔符按照key的ASCII码升序格式化为字符串。
// 例如：("=", "&") ---> bar=baz&foo=quux；
// 例如：(":", "#") ---> bar:baz#foo:quux；
func (v V) Encode(sym, sep string, options ...VEncodeOption) string {
	if len(v) == 0 {
		return ""
	}

	setting := new(vencodeSetting)

	for _, f := range options {
		f(setting)
	}

	keys := make([]string, 0, len(v))

	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var buf strings.Builder

	for _, k := range keys {
		val := v[k]

		if len(val) == 0 && setting.emptyMode == EmptyEncodeIgnore {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		if setting.escape {
			buf.WriteString(url.QueryEscape(k))
		} else {
			buf.WriteString(k)
		}

		if len(val) != 0 {
			buf.WriteString(sym)

			if setting.escape {
				buf.WriteString(url.QueryEscape(val))
			} else {
				buf.WriteString(val)
			}

			continue
		}

		// 保留符号
		if setting.emptyMode != EmptyEncodeOnlyKey {
			buf.WriteString(sym)
		}
	}

	return buf.String()
}

// VEmptyEncodeMode 值为空时的Encode模式
type VEmptyEncodeMode int

const (
	EmptyEncodeDefault VEmptyEncodeMode = iota // 默认：bar=baz&foo=
	EmptyEncodeIgnore                          // 忽略：bar=baz
	EmptyEncodeOnlyKey                         // 仅保留Key：bar=baz&foo
)

type vencodeSetting struct {
	escape    bool
	emptyMode VEmptyEncodeMode
}

// VEncodeOption V Encode 选项
type VEncodeOption func(s *vencodeSetting)

// WithEmptyEncodeMode 设置值为空时的Encode模式
func WithEmptyEncodeMode(mode VEmptyEncodeMode) VEncodeOption {
	return func(s *vencodeSetting) {
		s.emptyMode = mode
	}
}

// WithKVEscape 设置K-V是否需要QueryEscape
func WithKVEscape() VEncodeOption {
	return func(s *vencodeSetting) {
		s.escape = true
	}
}
