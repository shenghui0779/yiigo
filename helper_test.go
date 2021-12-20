package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	assert.Equal(t, "2016-03-19 15:03:19", Date(1458370999))
}

func TestStrToTime(t *testing.T) {
	assert.Equal(t, int64(1562910319), StrToTime("2019-07-12 13:45:19"))
}

func TestWeekAround(t *testing.T) {
	monday, sunday := WeekAround(time.Date(2020, 12, 12, 0, 0, 0, 0, time.Local))

	assert.Equal(t, "20201207", monday)
	assert.Equal(t, "20201213", sunday)
}

func TestIP2Long(t *testing.T) {
	assert.Equal(t, uint32(3221234342), IP2Long("192.0.34.166"))
}

func TestLong2IP(t *testing.T) {
	assert.Equal(t, "192.0.34.166", Long2IP(uint32(3221234342)))
}

func TestMD5(t *testing.T) {
	assert.Equal(t, "483367436bc9a6c5256bfc29a24f955e", MD5("iiinsomnia"))
}

func TestSHA1(t *testing.T) {
	assert.Equal(t, "7a4082bd79f2086af2c2b792c5e0ad06e729b9c4", SHA1("iiinsomnia"))
}

func TestHash(t *testing.T) {
	type args struct {
		algo HashAlgo
		s    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "md5",
			args: args{algo: AlgoMD5, s: "iiinsomnia"},
			want: "483367436bc9a6c5256bfc29a24f955e",
		},
		{
			name: "sha1",
			args: args{algo: AlgoSha1, s: "iiinsomnia"},
			want: "7a4082bd79f2086af2c2b792c5e0ad06e729b9c4",
		},
		{
			name: "sha224",
			args: args{algo: AlgoSha224, s: "iiinsomnia"},
			want: "c29117a2d94338daaab2315a7d896e05c1c04c9bf8525ac82d2c759f",
		},
		{
			name: "sha256",
			args: args{algo: AlgoSha256, s: "iiinsomnia"},
			want: "efed14231acf19fdca03adfac049171c109c922008e64dbaaf51a0c2cf11306b",
		},
		{
			name: "sha384",
			args: args{algo: AlgoSha384, s: "iiinsomnia"},
			want: "a0f3339d799e465d66c48d00dc101d4cfa343bf73eadd3e0713173924a0dea8d94f9b360c73da39612ecf495e6f7fa6d",
		},
		{
			name: "sha512",
			args: args{algo: AlgoSha512, s: "iiinsomnia"},
			want: "06d5c64c737b9b57a38aaa5289721f7954c18a85174c56410beba7331ba161c07e9cdf615c6f78c9b32999fd57745ab030cf83d6afa34bbbc9030f948849c19e",
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Hash(tt.args.algo, tt.args.s))
	}
}

func TestAddSlashes(t *testing.T) {
	assert.Equal(t, `Is your name O\'Reilly?`, AddSlashes("Is your name O'Reilly?"))
}

func TestStripSlashes(t *testing.T) {
	assert.Equal(t, "Is your name O'Reilly?", StripSlashes(`Is your name O\'Reilly?`))
}

func TestQuoteMeta(t *testing.T) {
	assert.Equal(t, `Hello world\. \(can you hear me\?\)`, QuoteMeta("Hello world. (can you hear me?)"))
}

func TestVersionCompare(t *testing.T) {
	ok, err := VersionCompare("1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.4")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">2.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "3.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)
}
