package yiigo

import "testing"

func TestMD5(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{s: "iiinsomnia"},
			want: "483367436bc9a6c5256bfc29a24f955e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.args.s); got != tt.want {
				t.Errorf("MD5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA1(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{s: "iiinsomnia"},
			want: "7a4082bd79f2086af2c2b792c5e0ad06e729b9c4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA1(tt.args.s); got != tt.want {
				t.Errorf("SHA1() = %v, want %v", got, tt.want)
			}
		})
	}
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
		t.Run(tt.name, func(t *testing.T) {
			if got := Hash(tt.args.algo, tt.args.s); got != tt.want {
				t.Errorf("Hash(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAddSlashes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{s: "Is your name O'Reilly?"},
			want: `Is your name O\'Reilly?`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddSlashes(tt.args.s); got != tt.want {
				t.Errorf("AddSlashes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripSlashes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{s: `Is your name O\'reilly?`},
			want: "Is your name O'reilly?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripSlashes(tt.args.s); got != tt.want {
				t.Errorf("StripSlashes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuoteMeta(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{s: "Hello world. (can you hear me?)"},
			want: `Hello world\. \(can you hear me\?\)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuoteMeta(tt.args.s); got != tt.want {
				t.Errorf("QuoteMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
