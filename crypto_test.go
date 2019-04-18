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
				key:  []byte("1234567890abcdef"),
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
