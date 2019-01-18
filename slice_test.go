package yiigo

import (
	"reflect"
	"testing"
)

func TestSortInt64s(t *testing.T) {
	type args struct {
		a []int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []int64{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortInt64s(tt.args.a)
		})
	}
}

func TestSearchInt64s(t *testing.T) {
	type args struct {
		a []int64
		x int64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []int64{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchInt64s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchInt64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInSliceInt(t *testing.T) {
	type args struct {
		x int
		a []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				x: 4,
				a: []int{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InSliceInt(tt.args.x, tt.args.a); got != tt.want {
				t.Errorf("InSliceInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInSliceInt64(t *testing.T) {
	type args struct {
		x int64
		a []int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				x: 4,
				a: []int64{5, 2, 4, 7, 6, 1},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InSliceInt64(tt.args.x, tt.args.a); got != tt.want {
				t.Errorf("InSliceInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInSliceFloat64(t *testing.T) {
	type args struct {
		x float64
		a []float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				x: 4.4,
				a: []float64{2.3, 4.4, 6.7, 7.2, 1.9, 3.5},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InSliceFloat64(tt.args.x, tt.args.a); got != tt.want {
				t.Errorf("InSliceFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInSliceString(t *testing.T) {
	type args struct {
		x string
		a []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				x: "iiinsomnia",
				a: []string{"hello", "test", "iiinsomnia", "yiigo", "world"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InSliceString(tt.args.x, tt.args.a); got != tt.want {
				t.Errorf("InSliceString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueInt(t *testing.T) {
	type args struct {
		a []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "t1",
			args: args{a: []int{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []int{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueInt(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueInt64(t *testing.T) {
	type args struct {
		a []int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "t1",
			args: args{a: []int64{3, 4, 5, 7, 5, 4, 3, 2}},
			want: []int64{3, 4, 5, 7, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueInt64(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueFloat64(t *testing.T) {
	type args struct {
		a []float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "t1",
			args: args{a: []float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 4.2, 9.2, 7.2}},
			want: []float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 9.2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueFloat64(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueString(t *testing.T) {
	type args struct {
		a []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "t1",
			args: args{a: []string{"a", "c", "d", "a", "e", "d", "x", "f", "c"}},
			want: []string{"a", "c", "d", "e", "x", "f"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueString(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueString() = %v, want %v", got, tt.want)
			}
		})
	}
}
