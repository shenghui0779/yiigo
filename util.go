package yiigo

import (
	"crypto/md5"
	"fmt"
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

	return fmt.Sprintf("%x", h.Sum(nil))
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

// IP2long converts a string containing an (IPv4) Internet Protocol dotted address into a long integer
func IP2long(ip string) int64 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	ret := big.NewInt(0)
	ret.SetBytes(ipv4)

	return ret.Int64()
}

// Long2IP converts an long integer address into a string in (IPv4) Internet standard dotted format
func Long2IP(ip int64) string {
	ipv4 := fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))

	return ipv4
}
