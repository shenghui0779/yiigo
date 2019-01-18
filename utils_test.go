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

func TestDate(t *testing.T) {
	type args struct {
		timestamp int64
		format    []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				timestamp: 1458370999,
				format:    []string{"2006-01-02 15:04:05"},
			},
			want: "2016-03-19 15:03:19",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Date(tt.args.timestamp, tt.args.format...); got != tt.want {
				t.Errorf("Date() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIP2long(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "t1",
			args: args{ip: "192.0.34.166"},
			want: 3221234342,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IP2long(tt.args.ip); got != tt.want {
				t.Errorf("IP2long() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLong2IP(t *testing.T) {
	type args struct {
		ip int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{ip: 3221234342},
			want: "192.0.34.166",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Long2IP(tt.args.ip); got != tt.want {
				t.Errorf("Long2IP() = %v, want %v", got, tt.want)
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
