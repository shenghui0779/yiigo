package xhash

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	assert.Equal(t, "483367436bc9a6c5256bfc29a24f955e", MD5("iiinsomnia"))
}

func TestSHA1(t *testing.T) {
	assert.Equal(t, "7a4082bd79f2086af2c2b792c5e0ad06e729b9c4", SHA1("iiinsomnia"))
}

func TestSHA256(t *testing.T) {
	assert.Equal(t, "efed14231acf19fdca03adfac049171c109c922008e64dbaaf51a0c2cf11306b", SHA256("iiinsomnia"))
}

func TestHash(t *testing.T) {
	type args struct {
		hash crypto.Hash
		s    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "md5",
			args: args{hash: crypto.MD5, s: "iiinsomnia"},
			want: "483367436bc9a6c5256bfc29a24f955e",
		},
		{
			name: "sha1",
			args: args{hash: crypto.SHA1, s: "iiinsomnia"},
			want: "7a4082bd79f2086af2c2b792c5e0ad06e729b9c4",
		},
		{
			name: "sha224",
			args: args{hash: crypto.SHA224, s: "iiinsomnia"},
			want: "c29117a2d94338daaab2315a7d896e05c1c04c9bf8525ac82d2c759f",
		},
		{
			name: "sha256",
			args: args{hash: crypto.SHA256, s: "iiinsomnia"},
			want: "efed14231acf19fdca03adfac049171c109c922008e64dbaaf51a0c2cf11306b",
		},
		{
			name: "sha384",
			args: args{hash: crypto.SHA384, s: "iiinsomnia"},
			want: "a0f3339d799e465d66c48d00dc101d4cfa343bf73eadd3e0713173924a0dea8d94f9b360c73da39612ecf495e6f7fa6d",
		},
		{
			name: "sha512",
			args: args{hash: crypto.SHA512, s: "iiinsomnia"},
			want: "06d5c64c737b9b57a38aaa5289721f7954c18a85174c56410beba7331ba161c07e9cdf615c6f78c9b32999fd57745ab030cf83d6afa34bbbc9030f948849c19e",
		},
	}
	for _, tt := range tests {
		v, err := Hash(tt.args.hash, tt.args.s)

		assert.Nil(t, err)
		assert.Equal(t, tt.want, v)
	}
}
