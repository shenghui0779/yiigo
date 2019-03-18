package yiigo

import (
	"math/big"
	"net"
	"time"
)

// AsDefault alias for "default"
const AsDefault = "default"

// X is a convenient alias for a map[string]interface{}.
type X map[string]interface{}

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
