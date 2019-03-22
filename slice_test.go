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

func TestInInts(t *testing.T) {
	type args struct {
		x int
		y []int
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
				y: []int{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InInts(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InInts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInInt32s(t *testing.T) {
	type args struct {
		x int32
		y []int32
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
				y: []int32{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InInt32s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InInt32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInUint32s(t *testing.T) {
	type args struct {
		x uint32
		y []uint32
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
				y: []uint32{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InUint32s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InUint32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInInt64s(t *testing.T) {
	type args struct {
		x int64
		y []int64
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
				y: []int64{5, 2, 4, 7, 6, 1},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InInt64s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InInt64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInUint64s(t *testing.T) {
	type args struct {
		x uint64
		y []uint64
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
				y: []uint64{5, 2, 4, 7, 6, 1},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InUint64s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InUint64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInFloat64s(t *testing.T) {
	type args struct {
		x float64
		y []float64
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
				y: []float64{2.3, 4.4, 6.7, 7.2, 1.9, 3.5},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InFloat64s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InFloat64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInStrings(t *testing.T) {
	type args struct {
		x string
		y []string
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
				y: []string{"hello", "test", "iiinsomnia", "yiigo", "world"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InStrings(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInArray(t *testing.T) {
	type args struct {
		x interface{}
		y []interface{}
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
				y: []interface{}{1, "test", "iiinsomnia", 2.9, true},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InArray(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueInts(t *testing.T) {
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
			if got := UniqueInts(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueInts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueInt32s(t *testing.T) {
	type args struct {
		a []int32
	}
	tests := []struct {
		name string
		args args
		want []int32
	}{
		{
			name: "t1",
			args: args{a: []int32{3, 4, 5, 7, 5, 4, 3, 2}},
			want: []int32{3, 4, 5, 7, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueInt32s(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueInt32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueUint32s(t *testing.T) {
	type args struct {
		a []uint32
	}
	tests := []struct {
		name string
		args args
		want []uint32
	}{
		{
			name: "t1",
			args: args{a: []uint32{3, 4, 5, 7, 5, 4, 3, 2}},
			want: []uint32{3, 4, 5, 7, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueUint32s(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueUint32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueInt64s(t *testing.T) {
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
			if got := UniqueInt64s(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueInt64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueUint64s(t *testing.T) {
	type args struct {
		a []uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "t1",
			args: args{a: []uint64{3, 4, 5, 7, 5, 4, 3, 2}},
			want: []uint64{3, 4, 5, 7, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueUint64s(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueUint64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueFloat64s(t *testing.T) {
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
			if got := UniqueFloat64s(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueFloat64s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
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
			if got := UniqueStrings(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
