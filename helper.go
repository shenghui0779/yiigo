package yiigo

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/xml"
	"hash"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
)

// Default defines for `default` name
const Default = "default"

// HashAlgo hash algorithm
type HashAlgo string

const (
	AlgoMD5    HashAlgo = "md5"
	AlgoSha1   HashAlgo = "sha1"
	AlgoSha224 HashAlgo = "sha224"
	AlgoSha256 HashAlgo = "sha256"
	AlgoSha384 HashAlgo = "sha384"
	AlgoSha512 HashAlgo = "sha512"
)

// X is a convenient alias for a map[string]interface{}.
type X map[string]interface{}

// CDATA XML CDATA section which is defined as blocks of text that are not parsed by the parser, but are otherwise recognized as markup.
type CDATA string

// MarshalXML encodes the receiver as zero or more XML elements.
func (c CDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		string `xml:",cdata"`
	}{string(c)}, start)
}

// Date format a local time/date and
// returns a string formatted according to the given format string using the given timestamp of int64.
// The default layout is: 2006-01-02 15:04:05.
func Date(timestamp int64, layout ...string) string {
	l := "2006-01-02 15:04:05"

	if len(layout) != 0 {
		l = layout[0]
	}

	date := time.Unix(timestamp, 0).Local().Format(l)

	return date
}

// StrToTime Parse English textual datetime description into a Unix timestamp.
// The default layout is: 2006-01-02 15:04:05
func StrToTime(datetime string, layout ...string) int64 {
	l := "2006-01-02 15:04:05"

	if len(layout) != 0 {
		l = layout[0]
	}

	t, err := time.ParseInLocation(l, datetime, time.Local)

	if err != nil {
		logger.Error("[yiigo] err parse time", zap.Error(err), zap.String("datetime", datetime), zap.String("layout", l))

		return 0
	}

	return t.Unix()
}

// WeekAround returns the date of monday and sunday for current week.
// The default layout is: 2006-01-02
func WeekAround(t time.Time, layout ...string) (monday, sunday string) {
	l := "2006-01-02"

	if len(layout) != 0 {
		l = layout[0]
	}

	weekday := t.Local().Weekday()

	// monday
	offset := int(time.Monday - weekday)

	if offset > 0 {
		offset = -6
	}

	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	monday = today.AddDate(0, 0, offset).Format(l)

	// sunday
	offset = int(time.Sunday - weekday)

	if offset < 0 {
		offset += 7
	}

	sunday = today.AddDate(0, 0, offset).Format(l)

	return
}

// IP2Long converts a string containing an (IPv4) Internet Protocol dotted address into an uint32 integer.
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP converts an uint32 integer address into a string in (IPv4) Internet standard dotted format.
func Long2IP(ip uint32) string {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
}

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

// Hash generates a hash value, expects: MD5, SHA1, SHA224, SHA256, SHA384, SHA512.
func Hash(algo HashAlgo, s string) string {
	var h hash.Hash

	switch algo {
	case AlgoMD5:
		h = md5.New()
	case AlgoSha1:
		h = sha1.New()
	case AlgoSha224:
		h = sha256.New224()
	case AlgoSha256:
		h = sha256.New()
	case AlgoSha384:
		h = sha512.New384()
	case AlgoSha512:
		h = sha512.New()
	default:
		return s
	}

	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// HMAC generates a keyed hash value, expects: MD5, SHA1, SHA224, SHA256, SHA384, SHA512.
func HMAC(algo HashAlgo, s, key string) string {
	var mac hash.Hash

	switch algo {
	case AlgoMD5:
		mac = hmac.New(md5.New, []byte(key))
	case AlgoSha1:
		mac = hmac.New(sha1.New, []byte(key))
	case AlgoSha224:
		mac = hmac.New(sha256.New224, []byte(key))
	case AlgoSha256:
		mac = hmac.New(sha256.New, []byte(key))
	case AlgoSha384:
		mac = hmac.New(sha512.New384, []byte(key))
	case AlgoSha512:
		mac = hmac.New(sha512.New, []byte(key))
	default:
		return s
	}

	mac.Write([]byte(s))

	return hex.EncodeToString(mac.Sum(nil))
}

// AddSlashes returns a string with backslashes added before characters that need to be escaped.
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

// StripSlashes returns a string with backslashes stripped off. (\' becomes ' and so on.) Double backslashes (\\) are made into a single backslash (\).
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

// QuoteMeta returns a version of str with a backslash character (\) before every character that is among these: . \ + * ? [ ^ ] ( $ )
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

// VersionCompare compares semantic versions range, support: >, >=, =, !=, <, <=, | (or), & (and).
// Param `rangeVer` eg: 1.0.0, =1.0.0, >2.0.0, >=1.0.0&<2.0.0, <2.0.0|>3.0.0, !=4.0.4
func VersionCompare(rangeVer, curVer string) (bool, error) {
	semVer, err := version.NewVersion(curVer)

	// invalid semantic version
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
