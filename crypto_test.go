package yiigo

import (
	"reflect"
	"testing"
)

func TestAESCBCCrypt(t *testing.T) {
	type args struct {
		data    []byte
		padding AESPadding
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				data:    []byte("shenghui0779"),
				padding: PKCS5,
			},
			want: "shenghui0779",
		},
		{
			name: "t2",
			args: args{
				data:    []byte("Iloveyiigo"),
				padding: PKCS7,
			},
			want: "Iloveyiigo",
		},
	}

	aesCrypto, err := NewAESCrypto([]byte("c510be34b0466938eace8edee61255c0"))

	if err != nil {
		t.Errorf("RSAVerifyWithSha256() error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := aesCrypto.CBCEncrypt(tt.args.data, tt.args.padding)

			got := aesCrypto.CBCDecrypt(b, tt.args.padding)

			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("AESCBCCrypt() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestRSASign(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{
				data: []byte("IIInsomnia"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := RSASignWithSha256(tt.args.data, privateKey)

			if err != nil {
				t.Errorf("RSASignWithSha256() error = %v", err)
			}

			if err = RSAVerifyWithSha256(tt.args.data, signature, publicKey); err != nil {
				t.Errorf("RSAVerifyWithSha256() error = %v", err)
			}
		})
	}
}

func TestRSACrypt(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				data: []byte("IIInsomnia"),
			},
			want: "IIInsomnia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsaData, err := RSAEncrypt(tt.args.data, publicKey)

			if err != nil {
				t.Errorf("RSAEncrypt() error = %v", err)
			}

			got, err := RSADecrypt(rsaData, privateKey)

			if err != nil {
				t.Errorf("RSADecrypt() error = %v", err)
			}

			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("RSADecrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	privateKey []byte
	publicKey  []byte
)

func TestMain(m *testing.M) {
	privateKey, publicKey, _ = GenerateRSAKey(2048)

	m.Run()
}
