package yiigo

import (
	"reflect"
	"sort"
	"testing"
)

func TestSortUints(t *testing.T) {
	type args struct {
		a []uint
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []uint{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUints(tt.args.a)

			if !sort.IsSorted(UintSlice(tt.args.a)) {
				t.Error("SortUints test failed")
			}
		})
	}
}

func TestSearchUints(t *testing.T) {
	type args struct {
		a []uint
		x uint
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []uint{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchUints(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchUints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortInt8s(t *testing.T) {
	type args struct {
		a []int8
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []int8{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortInt8s(tt.args.a)

			if !sort.IsSorted(Int8Slice(tt.args.a)) {
				t.Error("SortInt8s test failed")
			}
		})
	}
}

func TestSearchInt8s(t *testing.T) {
	type args struct {
		a []int8
		x int8
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []int8{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchInt8s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchInt8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortUint8s(t *testing.T) {
	type args struct {
		a []uint8
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []uint8{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUint8s(tt.args.a)

			if !sort.IsSorted(Uint8Slice(tt.args.a)) {
				t.Error("SortUint8s test failed")
			}
		})
	}
}

func TestSearchUint8s(t *testing.T) {
	type args struct {
		a []uint8
		x uint8
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []uint8{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchUint8s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchUint8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortInt16s(t *testing.T) {
	type args struct {
		a []int16
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []int16{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortInt16s(tt.args.a)

			if !sort.IsSorted(Int16Slice(tt.args.a)) {
				t.Error("SortInt16s test failed")
			}
		})
	}
}

func TestSearchInt16s(t *testing.T) {
	type args struct {
		a []int16
		x int16
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []int16{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchInt16s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchInt16s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortUint16s(t *testing.T) {
	type args struct {
		a []uint16
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []uint16{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUint16s(tt.args.a)

			if !sort.IsSorted(Uint16Slice(tt.args.a)) {
				t.Error("SortUint16s test failed")
			}
		})
	}
}

func TestSearchUint16s(t *testing.T) {
	type args struct {
		a []uint16
		x uint16
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []uint16{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchUint16s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchUint16s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortInt32s(t *testing.T) {
	type args struct {
		a []int32
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []int32{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortInt32s(tt.args.a)

			if !sort.IsSorted(Int32Slice(tt.args.a)) {
				t.Error("SortInt32s test failed")
			}
		})
	}
}

func TestSearchInt32s(t *testing.T) {
	type args struct {
		a []int32
		x int32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []int32{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchInt32s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchInt32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortUint32s(t *testing.T) {
	type args struct {
		a []uint32
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []uint32{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUint32s(tt.args.a)

			if !sort.IsSorted(Uint32Slice(tt.args.a)) {
				t.Error("SortUint32s test failed")
			}
		})
	}
}

func TestSearchUint32s(t *testing.T) {
	type args struct {
		a []uint32
		x uint32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []uint32{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchUint32s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchUint32s() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

			if !sort.IsSorted(Int64Slice(tt.args.a)) {
				t.Error("SortInt64s test failed")
			}
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

func TestSortUint64s(t *testing.T) {
	type args struct {
		a []uint64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{a: []uint64{4, 2, 7, 9, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortUint64s(tt.args.a)

			if !sort.IsSorted(Uint64Slice(tt.args.a)) {
				t.Error("SortUint64s test failed")
			}
		})
	}
}

func TestSearchUint64s(t *testing.T) {
	type args struct {
		a []uint64
		x uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "t1",
			args: args{
				a: []uint64{4, 1, 7, 2, 9},
				x: 4,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchUint64s(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("SearchUint64s() = %v, want %v", got, tt.want)
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

func TestInUints(t *testing.T) {
	type args struct {
		x uint
		y []uint
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
				y: []uint{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InUints(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InUints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInInt8s(t *testing.T) {
	type args struct {
		x int8
		y []int8
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
				y: []int8{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InInt8s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InInt8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInUint8s(t *testing.T) {
	type args struct {
		x uint8
		y []uint8
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
				y: []uint8{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InUint8s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InUint8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInInt16s(t *testing.T) {
	type args struct {
		x int16
		y []int16
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
				y: []int16{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InInt16s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InInt16s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInUint16s(t *testing.T) {
	type args struct {
		x uint16
		y []uint16
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
				y: []uint16{2, 4, 6, 7, 1, 3},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InUint16s(tt.args.x, tt.args.y...); got != tt.want {
				t.Errorf("InUint16s() = %v, want %v", got, tt.want)
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

func TestIntsUnique(t *testing.T) {
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
			if got := IntsUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IntsUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUintsUnique(t *testing.T) {
	type args struct {
		a []uint
	}
	tests := []struct {
		name string
		args args
		want []uint
	}{
		{
			name: "t1",
			args: args{a: []uint{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []uint{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UintsUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UintsUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt8sUnique(t *testing.T) {
	type args struct {
		a []int8
	}
	tests := []struct {
		name string
		args args
		want []int8
	}{
		{
			name: "t1",
			args: args{a: []int8{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []int8{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int8sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int8sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint8sUnique(t *testing.T) {
	type args struct {
		a []uint8
	}
	tests := []struct {
		name string
		args args
		want []uint8
	}{
		{
			name: "t1",
			args: args{a: []uint8{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []uint8{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint8sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint8sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt16sUnique(t *testing.T) {
	type args struct {
		a []int16
	}
	tests := []struct {
		name string
		args args
		want []int16
	}{
		{
			name: "t1",
			args: args{a: []int16{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []int16{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int16sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int16sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint16sUnique(t *testing.T) {
	type args struct {
		a []uint16
	}
	tests := []struct {
		name string
		args args
		want []uint16
	}{
		{
			name: "t1",
			args: args{a: []uint16{2, 4, 6, 7, 1, 3, 4, 9, 7}},
			want: []uint16{2, 4, 6, 7, 1, 3, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint16sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint16sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt32sUnique(t *testing.T) {
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
			if got := Int32sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int32sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint32sUnique(t *testing.T) {
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
			if got := Uint32sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint32sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64sUnique(t *testing.T) {
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
			if got := Int64sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int64sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint64sUnique(t *testing.T) {
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
			if got := Uint64sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint64sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64sUnique(t *testing.T) {
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
			if got := Float64sUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Float64sUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringsUnique(t *testing.T) {
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
			if got := StringsUnique(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringsUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}
