package xvalue

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
func (v V) Encode(sym, sep string, opts ...Option) string {
	if len(v) == 0 {
		return ""
	}

	o := &options{
		ignoreKeys: make(map[string]struct{}),
	}
	for _, fn := range opts {
		fn(o)
	}

	keys := make([]string, 0, len(v))
	for k := range v {
		if _, ok := o.ignoreKeys[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf strings.Builder

	for _, k := range keys {
		val := v[k]
		if len(val) == 0 && o.emptyMode == EmptyIgnore {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString(sep)
		}
		if o.escape {
			buf.WriteString(url.QueryEscape(k))
		} else {
			buf.WriteString(k)
		}
		if len(val) != 0 {
			buf.WriteString(sym)
			if o.escape {
				buf.WriteString(url.QueryEscape(val))
			} else {
				buf.WriteString(val)
			}
			continue
		}
		// 保留符号
		if o.emptyMode != EmptyOnlyKey {
			buf.WriteString(sym)
		}
	}

	return buf.String()
}
