package yiigo

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"math/big"
	"net"
	"time"
)

// X is a convenient alias for a map[string]interface{}.
type X map[string]interface{}

// MD5 calculate the md5 hash of a string.
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// Date format a local time/date and
// returns a string formatted according to the given format string using the given timestamp of int64.
// The default format is: 2006-01-02 15:04:05.
func Date(timestamp int64, format ...string) string {
	layout := "2006-01-02 15:04:05"

	if len(format) > 0 {
		layout = format[0]
	}

	date := time.Unix(timestamp, 0).Format(layout)

	return date
}

// IP2Long converts a string containing an (IPv4) Internet Protocol dotted address into a long integer.
func IP2Long(ip string) int64 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	ret := big.NewInt(0)
	ret.SetBytes(ipv4)

	return ret.Int64()
}

// Long2IP converts an long integer address into a string in (IPv4) Internet standard dotted format.
func Long2IP(ip int64) string {
	ret := big.NewInt(ip)
	ipv4 := net.IP(ret.Bytes())

	return ipv4.String()
}

// AddSlashes returns a string with backslashes added before characters that need to be escaped.
func AddSlashes(s string) string {
	var buf bytes.Buffer

	for _, ch := range s {
		if ch == '\'' || ch == '"' || ch == '\\' {
			buf.WriteRune('\\')
		}

		buf.WriteRune(ch)
	}

	return buf.String()
}

// StripSlashes returns a string with backslashes stripped off. (\' becomes ' and so on.) Double backslashes (\\) are made into a single backslash (\).
func StripSlashes(s string) string {
	var buf bytes.Buffer

	l, skip := len(s), false

	for i, ch := range s {
		if skip {
			buf.WriteRune(ch)
			skip = false

			continue
		}

		if ch == '\\' {
			if i+1 < l && s[i+1] == '\\' {
				skip = true
			}

			continue
		}

		buf.WriteRune(ch)
	}

	return buf.String()
}

// QuoteMeta returns a version of str with a backslash character (\) before every character that is among these: . \ + * ? [ ^ ] ( $ )
func QuoteMeta(s string) string {
	var buf bytes.Buffer

	for _, ch := range s {
		switch ch {
		case '.', '+', '\\', '(', '$', ')', '[', '^', ']', '*', '?':
			buf.WriteRune('\\')
		}

		buf.WriteRune(ch)
	}

	return buf.String()
}
