package yiigo

import (
	"net/url"
	"sort"
	"strings"
)

// Values 用于处理 k-v 需要格式化的场景，如：签名
type Values map[string]string

// Set 设置 k-v
func (v Values) Set(key, value string) {
	v[key] = value
}

// Get 获取值
func (v Values) Get(key string) string {
	return v[key]
}

// Del 删除Key
func (v Values) Del(key string) {
	delete(v, key)
}

// Has 判断Key是否存在
func (v Values) Has(key string) bool {
	_, ok := v[key]

	return ok
}

// Encode 通过自定义的符号和分隔符按照key的ASCII码升序格式化为字符串。
// 例如：("=", "&") ---> bar=baz&foo=quux；
// 例如：(":", "#") ---> bar:baz#foo:quux；
func (v Values) Encode(sym, sep string) string {
	if len(v) == 0 {
		return ""
	}

	keys := make([]string, 0, len(v))

	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var buf strings.Builder

	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		buf.WriteString(k)
		buf.WriteString(sym)
		buf.WriteString(v[k])
	}

	return buf.String()
}

// EncodeEscape 通过自定义的符号和分隔符按照key的ASCII码升序格式化为字符串。
// 例如：("=", "&") ---> bar=baz&foo=quux；
// 例如：(":", "#") ---> bar:baz#foo:quux；
// 注意：这里 key 和 value 会被 QueryEscape；
func (v Values) EncodeEscape(sym, sep string) string {
	if v == nil {
		return ""
	}

	keys := make([]string, 0, len(v))

	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var buf strings.Builder

	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		buf.WriteString(url.QueryEscape(k))
		buf.WriteString(sym)
		buf.WriteString(url.QueryEscape(v[k]))
	}

	return buf.String()
}
