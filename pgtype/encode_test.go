package pgtype

import (
	"bytes"
	"testing"
	"time"
)

var timeTests = []struct {
	str     string
	timeval time.Time
}{
	{"22001-02-03", time.Date(22001, time.February, 3, 0, 0, 0, 0, time.FixedZone("", 0))},
	{"2001-02-03", time.Date(2001, time.February, 3, 0, 0, 0, 0, time.FixedZone("", 0))},
	{"0001-12-31 BC", time.Date(0, time.December, 31, 0, 0, 0, 0, time.FixedZone("", 0))},
	{"2001-02-03 BC", time.Date(-2000, time.February, 3, 0, 0, 0, 0, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06", time.Date(2001, time.February, 3, 4, 5, 6, 0, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.000001", time.Date(2001, time.February, 3, 4, 5, 6, 1000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.00001", time.Date(2001, time.February, 3, 4, 5, 6, 10000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.0001", time.Date(2001, time.February, 3, 4, 5, 6, 100000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.001", time.Date(2001, time.February, 3, 4, 5, 6, 1000000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.01", time.Date(2001, time.February, 3, 4, 5, 6, 10000000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.1", time.Date(2001, time.February, 3, 4, 5, 6, 100000000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.12", time.Date(2001, time.February, 3, 4, 5, 6, 120000000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.123", time.Date(2001, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.1234", time.Date(2001, time.February, 3, 4, 5, 6, 123400000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.12345", time.Date(2001, time.February, 3, 4, 5, 6, 123450000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.123456", time.Date(2001, time.February, 3, 4, 5, 6, 123456000, time.FixedZone("", 0))},
	{"2001-02-03 04:05:06.123-07", time.Date(2001, time.February, 3, 4, 5, 6, 123000000,
		time.FixedZone("", -7*60*60))},
	{"2001-02-03 04:05:06-07", time.Date(2001, time.February, 3, 4, 5, 6, 0,
		time.FixedZone("", -7*60*60))},
	{"2001-02-03 04:05:06-07:42", time.Date(2001, time.February, 3, 4, 5, 6, 0,
		time.FixedZone("", -(7*60*60+42*60)))},
	{"2001-02-03 04:05:06-07:30:09", time.Date(2001, time.February, 3, 4, 5, 6, 0,
		time.FixedZone("", -(7*60*60+30*60+9)))},
	{"2001-02-03 04:05:06+07:30:09", time.Date(2001, time.February, 3, 4, 5, 6, 0,
		time.FixedZone("", +(7*60*60+30*60+9)))},
	{"2001-02-03 04:05:06+07", time.Date(2001, time.February, 3, 4, 5, 6, 0,
		time.FixedZone("", 7*60*60))},
	{"0011-02-03 04:05:06 BC", time.Date(-10, time.February, 3, 4, 5, 6, 0, time.FixedZone("", 0))},
	{"0011-02-03 04:05:06.123 BC", time.Date(-10, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0))},
	{"0011-02-03 04:05:06.123-07 BC", time.Date(-10, time.February, 3, 4, 5, 6, 123000000,
		time.FixedZone("", -7*60*60))},
	{"0001-02-03 04:05:06.123", time.Date(1, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0))},
	{"0001-02-03 04:05:06.123 BC", time.Date(1, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0)).AddDate(-1, 0, 0)},
	{"0001-02-03 04:05:06.123 BC", time.Date(0, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0))},
	{"0002-02-03 04:05:06.123 BC", time.Date(0, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0)).AddDate(-1, 0, 0)},
	{"0002-02-03 04:05:06.123 BC", time.Date(-1, time.February, 3, 4, 5, 6, 123000000, time.FixedZone("", 0))},
	{"12345-02-03 04:05:06.1", time.Date(12345, time.February, 3, 4, 5, 6, 100000000, time.FixedZone("", 0))},
	{"123456-02-03 04:05:06.1", time.Date(123456, time.February, 3, 4, 5, 6, 100000000, time.FixedZone("", 0))},
}

// Test that parsing the string results in the expected value.
func TestParseTs(t *testing.T) {
	for i, tt := range timeTests {
		val, err := ParseTimestamp(nil, tt.str)
		if err != nil {
			t.Errorf("%d: got error: %v", i, err)
		} else if val.String() != tt.timeval.String() {
			t.Errorf("%d: expected to parse %q into %q; got %q",
				i, tt.str, tt.timeval, val)
		}
	}
}

var timeErrorTests = []string{
	"BC",
	" BC",
	"2001",
	"2001-2-03",
	"2001-02-3",
	"2001-02-03 ",
	"2001-02-03 B",
	"2001-02-03 04",
	"2001-02-03 04:",
	"2001-02-03 04:05",
	"2001-02-03 04:05 B",
	"2001-02-03 04:05 BC",
	"2001-02-03 04:05:",
	"2001-02-03 04:05:6",
	"2001-02-03 04:05:06 B",
	"2001-02-03 04:05:06BC",
	"2001-02-03 04:05:06.123 B",
}

// Test that parsing the string results in an error.
func TestParseTsErrors(t *testing.T) {
	for i, tt := range timeErrorTests {
		_, err := ParseTimestamp(nil, tt)
		if err == nil {
			t.Errorf("%d: expected an error from parsing: %v", i, tt)
		}
	}
}

var formatTimeTests = []struct {
	time     time.Time
	expected string
}{
	{time.Time{}, "0001-01-01 00:00:00Z"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "2001-02-03 04:05:06.123456789Z"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "2001-02-03 04:05:06.123456789+02:00"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "2001-02-03 04:05:06.123456789-06:00"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "2001-02-03 04:05:06-07:30:09"},

	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "0001-02-03 04:05:06.123456789Z"},
	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "0001-02-03 04:05:06.123456789+02:00"},
	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "0001-02-03 04:05:06.123456789-06:00"},

	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "0001-02-03 04:05:06.123456789Z BC"},
	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "0001-02-03 04:05:06.123456789+02:00 BC"},
	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "0001-02-03 04:05:06.123456789-06:00 BC"},

	{time.Date(1, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "0001-02-03 04:05:06-07:30:09"},
	{time.Date(0, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "0001-02-03 04:05:06-07:30:09 BC"},
}

func TestFormatTs(t *testing.T) {
	for i, tt := range formatTimeTests {
		val := string(formatTs(tt.time))
		if val != tt.expected {
			t.Errorf("%d: incorrect time format %q, want %q", i, val, tt.expected)
		}
	}
}

func TestTextDecodeIntoString(t *testing.T) {
	input := []byte("hello world")
	want := string(input)
	for _, typ := range []Oid{T_char, T_varchar, T_text} {
		got := decode(&parameterStatus{}, input, typ, formatText)
		if got != want {
			t.Errorf("invalid string decoding output for %T(%+v), got %v but expected %v", typ, typ, got, want)
		}
	}
}

func TestByteaOutputFormatEncoding(t *testing.T) {
	input := []byte("\\x\x00\x01\x02\xFF\xFEabcdefg0123")
	want := []byte("\\x5c78000102fffe6162636465666730313233")
	got := encode(&parameterStatus{serverVersion: 90000}, input, T_bytea)
	if !bytes.Equal(want, got) {
		t.Errorf("invalid hex bytea output, got %v but expected %v", got, want)
	}

	want = []byte("\\\\x\\000\\001\\002\\377\\376abcdefg0123")
	got = encode(&parameterStatus{serverVersion: 84000}, input, T_bytea)
	if !bytes.Equal(want, got) {
		t.Errorf("invalid escape bytea output, got %v but expected %v", got, want)
	}
}

func TestAppendEncodedText(t *testing.T) {
	var buf []byte

	buf = appendEncodedText(&parameterStatus{serverVersion: 90000}, buf, int64(10))
	buf = append(buf, '\t')
	buf = appendEncodedText(&parameterStatus{serverVersion: 90000}, buf, 42.0000000001)
	buf = append(buf, '\t')
	buf = appendEncodedText(&parameterStatus{serverVersion: 90000}, buf, "hello\tworld")
	buf = append(buf, '\t')
	buf = appendEncodedText(&parameterStatus{serverVersion: 90000}, buf, []byte{0, 128, 255})

	if string(buf) != "10\t42.0000000001\thello\\tworld\t\\\\x0080ff" {
		t.Fatal(string(buf))
	}
}

func TestAppendEscapedText(t *testing.T) {
	if esc := appendEscapedText(nil, "hallo\tescape"); string(esc) != "hallo\\tescape" {
		t.Fatal(string(esc))
	}
	if esc := appendEscapedText(nil, "hallo\\tescape\n"); string(esc) != "hallo\\\\tescape\\n" {
		t.Fatal(string(esc))
	}
	if esc := appendEscapedText(nil, "\n\r\t\f"); string(esc) != "\\n\\r\\t\f" {
		t.Fatal(string(esc))
	}
}

func TestAppendEscapedTextExistingBuffer(t *testing.T) {
	buf := []byte("123\t")
	if esc := appendEscapedText(buf, "hallo\tescape"); string(esc) != "123\thallo\\tescape" {
		t.Fatal(string(esc))
	}
	buf = []byte("123\t")
	if esc := appendEscapedText(buf, "hallo\\tescape\n"); string(esc) != "123\thallo\\\\tescape\\n" {
		t.Fatal(string(esc))
	}
	buf = []byte("123\t")
	if esc := appendEscapedText(buf, "\n\r\t\f"); string(esc) != "123\t\\n\\r\\t\f" {
		t.Fatal(string(esc))
	}
}

var formatAndParseTimestamp = []struct {
	time     time.Time
	expected string
}{
	{time.Time{}, "0001-01-01 00:00:00Z"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "2001-02-03 04:05:06.123456789Z"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "2001-02-03 04:05:06.123456789+02:00"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "2001-02-03 04:05:06.123456789-06:00"},
	{time.Date(2001, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "2001-02-03 04:05:06-07:30:09"},

	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "0001-02-03 04:05:06.123456789Z"},
	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "0001-02-03 04:05:06.123456789+02:00"},
	{time.Date(1, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "0001-02-03 04:05:06.123456789-06:00"},

	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 0)), "0001-02-03 04:05:06.123456789Z BC"},
	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", 2*60*60)), "0001-02-03 04:05:06.123456789+02:00 BC"},
	{time.Date(0, time.February, 3, 4, 5, 6, 123456789, time.FixedZone("", -6*60*60)), "0001-02-03 04:05:06.123456789-06:00 BC"},

	{time.Date(1, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "0001-02-03 04:05:06-07:30:09"},
	{time.Date(0, time.February, 3, 4, 5, 6, 0, time.FixedZone("", -(7*60*60+30*60+9))), "0001-02-03 04:05:06-07:30:09 BC"},
}

func TestFormatAndParseTimestamp(t *testing.T) {
	for _, val := range formatAndParseTimestamp {
		formattedTime := FormatTimestamp(val.time)
		parsedTime, err := ParseTimestamp(nil, string(formattedTime))
		if err != nil {
			t.Errorf("invalid parsing, err: %v", err.Error())
		}

		if val.time.UTC() != parsedTime.UTC() {
			t.Errorf("invalid parsing from formatted timestamp, got %v; expected %v", parsedTime.String(), val.time.String())
		}
	}
}

func BenchmarkAppendEscapedText(b *testing.B) {
	longString := ""
	for i := 0; i < 100; i++ {
		longString += "123456789\n"
	}
	for i := 0; i < b.N; i++ {
		appendEscapedText(nil, longString)
	}
}

func BenchmarkAppendEscapedTextNoEscape(b *testing.B) {
	longString := ""
	for i := 0; i < 100; i++ {
		longString += "1234567890"
	}
	for i := 0; i < b.N; i++ {
		appendEscapedText(nil, longString)
	}
}
