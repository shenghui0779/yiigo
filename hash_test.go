package yiigo

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

func TestHMacSHA256(t *testing.T) {
	assert.Equal(t, "a458409cd884140c1ca36ef3013a5c7289c3e057049e3563401094d3f929b93b", HMacSHA256("iiinsomnia", "ILoveYiigo"))
}

func TestHMac(t *testing.T) {
	type args struct {
		hash crypto.Hash
		key  string
		s    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "md5",
			args: args{hash: crypto.MD5, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "319a0496da92188c893a338947173057",
		},
		{
			name: "sha1",
			args: args{hash: crypto.SHA1, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "30c0c496355c2bb9308c63159cc4b726f1205dfc",
		},
		{
			name: "sha224",
			args: args{hash: crypto.SHA224, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "a7ca1021ba1525be6e48d5dfd12543beeaac32a92c8b7536369f5369",
		},
		{
			name: "sha256",
			args: args{hash: crypto.SHA256, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "a458409cd884140c1ca36ef3013a5c7289c3e057049e3563401094d3f929b93b",
		},
		{
			name: "sha384",
			args: args{hash: crypto.SHA384, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "13b5a15d7f2af4fade3fe49dfda2642c72e4cd33918285f3aeb1550dfcf764c1d5969e1506fa177f0b23922855e2cd84",
		},
		{
			name: "sha512",
			args: args{hash: crypto.SHA512, key: "iiinsomnia", s: "ILoveYiigo"},
			want: "4412772b1d9278e04edcbcab20b900a41cf28a1dcf0f2bc7391f354940b9bcad4c6f716e9c6197118c769d2498eb819bc234cae76218aed64cb4e1468b082e1c",
		},
	}
	for _, tt := range tests {
		v, err := HMac(tt.args.hash, tt.args.key, tt.args.s)

		assert.Nil(t, err)
		assert.Equal(t, tt.want, v)
	}
}
