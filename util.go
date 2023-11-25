package yiigo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"golang.org/x/crypto/pkcs12"
)

var timezone = time.FixedZone("CST", 8*3600)

const (
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

const (
	ContentJSON = "application/json;charset=utf-8"
	ContentForm = "application/x-www-form-urlencoded"
)

const (
	OK      = "OK"
	Default = "default"
)

// X 类型别名
type X map[string]any

// CDATA XML `CDATA` 标记
type CDATA string

// MarshalXML XML 带 `CDATA` 标记序列化
func (c CDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		string `xml:",cdata"`
	}{string(c)}, start)
}

// SetTimeZone 设置时区；默认：GMT+8
func SetTimeZone(loc *time.Location) {
	timezone = loc
}

// TimeToStr 时间戳格式化为时间字符串
// 若 timestamp < 0，则使用 `time.Now()`
func TimeToStr(timestamp int64, layout string) string {
	if timestamp < 0 {
		return time.Now().In(timezone).Format(layout)
	}

	return time.Unix(timestamp, 0).In(timezone).Format(layout)
}

// StrToTime 时间字符串解析为时间戳
func StrToTime(datetime, layout string) time.Time {
	t, _ := time.ParseInLocation(layout, datetime, timezone)

	return t
}

// WeekAround 返回给定时间戳所在周的「周一」和「周日」时间字符串
func WeekAround(timestamp int64, layout string) (monday, sunday string) {
	t := time.Unix(timestamp, 0).In(timezone)

	weekday := t.Weekday()

	// monday
	offset := int(time.Monday - weekday)
	if offset > 0 {
		offset = -6
	}

	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timezone)

	monday = today.AddDate(0, 0, offset).Format(layout)

	// sunday
	offset = int(time.Sunday - weekday)
	if offset < 0 {
		offset += 7
	}

	sunday = today.AddDate(0, 0, offset).Format(layout)

	return
}

// IP2Long IP地址转整数
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP 整数转IP地址
func Long2IP(ip uint32) string {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
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

// AddSlashes 在字符串的每个引号前添加反斜杠
func AddSlashes(s string) string {
	var builder strings.Builder

	for _, ch := range s {
		if ch == '\'' || ch == '"' || ch == '\\' {
			builder.WriteRune('\\')
		}

		builder.WriteRune(ch)
	}

	return builder.String()
}

// StripSlashes 删除字符串中的反斜杠
func StripSlashes(s string) string {
	var builder strings.Builder

	l, skip := len(s), false

	for i, ch := range s {
		if skip {
			builder.WriteRune(ch)
			skip = false

			continue
		}

		if ch == '\\' {
			if i+1 < l && s[i+1] == '\\' {
				skip = true
			}

			continue
		}

		builder.WriteRune(ch)
	}

	return builder.String()
}

// QuoteMeta 在字符串预定义的字符前添加反斜杠
func QuoteMeta(s string) string {
	var builder strings.Builder

	for _, ch := range s {
		switch ch {
		case '.', '+', '\\', '(', '$', ')', '[', '^', ']', '*', '?':
			builder.WriteRune('\\')
		}

		builder.WriteRune(ch)
	}

	return builder.String()
}

// SliceUniq 切片去重
func SliceUniq[T ~int | ~int64 | ~float64 | ~string](a []T) []T {
	ret := make([]T, 0)
	if len(a) == 0 {
		return ret
	}

	m := make(map[T]struct{}, 0)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}

	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(a)，则返回打乱的切片
func SliceRand[T any](a []T, n int) []T {
	if n == 0 || n < -1 {
		return make([]T, 0)
	}

	count := len(a)
	ret := make([]T, count)

	copy(ret, a)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(count, func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})

	if n == -1 || n >= count {
		return ret
	}

	return ret[:n]
}

// CreateFile 创建或清空指定的文件
// 文件已存在，则清空；文件或目录不存在，则以0775权限创建
func CreateFile(filename string) (*os.File, error) {
	abspath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(path.Dir(abspath), 0775); err != nil {
		return nil, err
	}

	return os.OpenFile(abspath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
}

// OpenFile 打开指定的文件
// 文件已存在，则追加内容；文件或目录不存在，则以0775权限创建
func OpenFile(filename string) (*os.File, error) {
	abspath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(path.Dir(abspath), 0775); err != nil {
		return nil, err
	}

	return os.OpenFile(abspath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0775)
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

// LoadCertFromPfxFile 通过pfx(p12)文件生成TLS证书
// 注意：证书需采用「TripleDES-SHA1」加密方式
func LoadCertFromPfxFile(pfxFile, password string) (tls.Certificate, error) {
	fail := func(err error) (tls.Certificate, error) { return tls.Certificate{}, err }

	certPath, err := filepath.Abs(filepath.Clean(pfxFile))
	if err != nil {
		return fail(err)
	}

	b, err := os.ReadFile(certPath)
	if err != nil {
		return fail(err)
	}

	blocks, err := pkcs12.ToPEM(b, password)
	if err != nil {
		return fail(err)
	}

	pemData := make([]byte, 0)
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	return tls.X509KeyPair(pemData, pemData)
}
