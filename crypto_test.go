package yiigo

import (
	"reflect"
	"testing"
)

func TestAESCBCCrypt(t *testing.T) {
	type args struct {
		data []byte
		key  []byte
		iv   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				data: []byte("IIInsomnia"),
				key:  []byte("c510be34b0466938eace8edee61255c0"),
			},
			want:    "IIInsomnia",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesData, err := AESCBCEncrypt(tt.args.data, tt.args.key, tt.args.iv...)

			if (err != nil) != tt.wantErr {
				t.Errorf("AESCBCEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := AESCBCDecrypt(aesData, tt.args.key, tt.args.iv...)

			if (err != nil) != tt.wantErr {
				t.Errorf("AESCBCDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("AESCBCDecrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRSASign(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				data: []byte("IIInsomnia"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := RSASignWithSha256(tt.args.data, privateKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("RSASignWithSha256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = RSAVerifyWithSha256(tt.args.data, signature, publicKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("RSAVerifyWithSha256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRSACrypt(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				data: []byte("IIInsomnia"),
			},
			want:    "IIInsomnia",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsaData, err := RSAEncrypt(tt.args.data, publicKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("RSAEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := RSADecrypt(rsaData, privateKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("RSADecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
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
