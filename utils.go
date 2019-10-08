package yiigo

import (
	"encoding/xml"
	"math"
	"net"
	"time"
)

// AsDefault alias for "default"
const AsDefault = "default"

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

	if len(layout) > 0 {
		l = layout[0]
	}

	date := time.Unix(timestamp, 0).Format(l)

	return date
}

// IP2Long converts a string containing an (IPv4) Internet Protocol dotted address into a long integer.
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP converts an long integer address into a string in (IPv4) Internet standard dotted format.
func Long2IP(ip uint32) string {
	if ip > math.MaxUint32 {
		return ""
	}

	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
}
