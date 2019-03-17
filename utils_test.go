package yiigo

import "testing"

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

func TestIP2Long(t *testing.T) {
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
			if got := IP2Long(tt.args.ip); got != tt.want {
				t.Errorf("IP2Long() = %v, want %v", got, tt.want)
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
