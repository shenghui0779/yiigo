package yiigo

import (
	"reflect"
	"testing"
)

func TestCryptoAES(t *testing.T) {
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
				key:  []byte("0123456789abcdef"),
			},
			want:    "IIInsomnia",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesData, err := AESEncrypt(tt.args.data, tt.args.key, tt.args.iv...)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := AESDecrypt(aesData, tt.args.key, tt.args.iv...)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("AESDecrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
