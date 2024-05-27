package xhash

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMacSHA1(t *testing.T) {
	assert.Equal(t, "30c0c496355c2bb9308c63159cc4b726f1205dfc", HMacSHA1("iiinsomnia", "ILoveYiigo"))
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
