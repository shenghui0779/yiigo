package yiigo

import (
	"encoding/xml"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
)

var timezone = time.FixedZone("CST", 8*3600)

const (
	layoutdate = "2006-01-02"
	layouttime = "2006-01-02 15:04:05"
)

const (
	// Default defines for `default` name
	Default = "default"

	// OK
	OK = "OK"
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

// SetTimezone sets timezone for time display.
// The default timezone is GMT+8.
func SetTimezone(loc *time.Location) {
	timezone = loc
}

// Date format a local time/date and
// returns a string formatted according to the given layout using the given timestamp of int64.
// If timestamp < 0, use `time.Now()` to format.
// The default layout is: 2006-01-02 15:04:05.
func Date(timestamp int64, layout ...string) string {
	l := layouttime

	if len(layout) != 0 {
		l = layout[0]
	}

	if timestamp < 0 {
		return time.Now().In(timezone).Format(l)
	}

	return time.Unix(timestamp, 0).In(timezone).Format(l)
}

// StrToTime Parse English textual datetime description into a Unix timestamp.
// The default layout is: 2006-01-02 15:04:05
func StrToTime(datetime string, layout ...string) int64 {
	l := layouttime

	if len(layout) != 0 {
		l = layout[0]
	}

	t, err := time.ParseInLocation(l, datetime, timezone)

	if err != nil {
		logger.Error("[yiigo] err parse time", zap.Error(err), zap.String("datetime", datetime), zap.String("layout", l))

		return 0
	}

	return t.Unix()
}

// WeekAround returns the monday and sunday of the week for the given time.
// The default layout is: 2006-01-02
func WeekAround(timestamp int64, layout ...string) (monday, sunday string) {
	t := time.Unix(timestamp, 0).In(timezone)

	weekday := t.Weekday()

	// monday
	offset := int(time.Monday - weekday)

	if offset > 0 {
		offset = -6
	}

	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timezone)

	l := layoutdate

	if len(layout) != 0 {
		l = layout[0]
	}

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

// CreateFile creates or truncates the named file.
// If the file already exists, it is truncated.
// If the directory or file does not exist, it is created with mode 0775
func CreateFile(filename string) (*os.File, error) {
	abspath, err := filepath.Abs(filename)

	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(path.Dir(abspath), 0775); err != nil {
		return nil, err
	}

	return os.OpenFile(abspath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0775)
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
