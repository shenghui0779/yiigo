package yiigo

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortUints(t *testing.T) {
	a := []uint{4, 2, 7, 9, 1}
	SortUints(a)

	assert.Equal(t, true, sort.IsSorted(UintSlice(a)))
}

func TestSearchUints(t *testing.T) {
	assert.Equal(t, 2, SearchUints([]uint{4, 1, 7, 2, 9}, 4))
}

func TestSortInt8s(t *testing.T) {
	a := []int8{4, 2, 7, 9, 1}
	SortInt8s(a)

	assert.Equal(t, true, sort.IsSorted(Int8Slice(a)))
}

func TestSearchInt8s(t *testing.T) {
	assert.Equal(t, 2, SearchInt8s([]int8{4, 1, 7, 2, 9}, 4))
}

func TestSortUint8s(t *testing.T) {
	a := []uint8{4, 2, 7, 9, 1}
	SortUint8s(a)

	assert.Equal(t, true, sort.IsSorted(Uint8Slice(a)))
}

func TestSearchUint8s(t *testing.T) {
	assert.Equal(t, 2, SearchUint8s([]uint8{4, 1, 7, 2, 9}, 4))
}

func TestSortInt16s(t *testing.T) {
	a := []int16{4, 2, 7, 9, 1}
	SortInt16s(a)

	assert.Equal(t, true, sort.IsSorted(Int16Slice(a)))
}

func TestSearchInt16s(t *testing.T) {
	assert.Equal(t, 2, SearchInt16s([]int16{4, 1, 7, 2, 9}, 4))
}

func TestSortUint16s(t *testing.T) {
	a := []uint16{4, 2, 7, 9, 1}
	SortUint16s(a)

	assert.Equal(t, true, sort.IsSorted(Uint16Slice(a)))
}

func TestSearchUint16s(t *testing.T) {
	assert.Equal(t, 2, SearchUint16s([]uint16{4, 1, 7, 2, 9}, 4))
}

func TestSortInt32s(t *testing.T) {
	a := []int32{4, 2, 7, 9, 1}
	SortInt32s(a)

	assert.Equal(t, true, sort.IsSorted(Int32Slice(a)))
}

func TestSearchInt32s(t *testing.T) {
	assert.Equal(t, 2, SearchInt32s([]int32{4, 1, 7, 2, 9}, 4))
}

func TestSortUint32s(t *testing.T) {
	a := []uint32{4, 2, 7, 9, 1}
	SortUint32s(a)

	assert.Equal(t, true, sort.IsSorted(Uint32Slice(a)))
}

func TestSearchUint32s(t *testing.T) {
	assert.Equal(t, 2, SearchUint32s([]uint32{4, 1, 7, 2, 9}, 4))
}

func TestSortInt64s(t *testing.T) {
	a := []int64{4, 2, 7, 9, 1}
	SortInt64s(a)

	assert.Equal(t, true, sort.IsSorted(Int64Slice(a)))
}

func TestSearchInt64s(t *testing.T) {
	assert.Equal(t, 2, SearchInt64s([]int64{4, 1, 7, 2, 9}, 4))
}

func TestSortUint64s(t *testing.T) {
	a := []uint64{4, 2, 7, 9, 1}
	SortUint64s(a)

	assert.Equal(t, true, sort.IsSorted(Uint64Slice(a)))
}

func TestSearchUint64s(t *testing.T) {
	assert.Equal(t, 2, SearchUint64s([]uint64{4, 1, 7, 2, 9}, 4))
}

func TestInInts(t *testing.T) {
	assert.Equal(t, true, InInts(4, []int{2, 4, 6, 7, 1, 3}...))
}

func TestInUints(t *testing.T) {
	assert.Equal(t, true, InUints(4, []uint{2, 4, 6, 7, 1, 3}...))
}

func TestInInt8s(t *testing.T) {
	assert.Equal(t, true, InInt8s(4, []int8{2, 4, 6, 7, 1, 3}...))
}

func TestInUint8s(t *testing.T) {
	assert.Equal(t, true, InUint8s(4, []uint8{2, 4, 6, 7, 1, 3}...))
}

func TestInInt16s(t *testing.T) {
	assert.Equal(t, true, InInt16s(4, []int16{2, 4, 6, 7, 1, 3}...))
}

func TestInUint16s(t *testing.T) {
	assert.Equal(t, true, InUint16s(4, []uint16{2, 4, 6, 7, 1, 3}...))
}

func TestInInt32s(t *testing.T) {
	assert.Equal(t, true, InInt32s(4, []int32{2, 4, 6, 7, 1, 3}...))
}

func TestInUint32s(t *testing.T) {
	assert.Equal(t, true, InUint32s(4, []uint32{2, 4, 6, 7, 1, 3}...))
}

func TestInInt64s(t *testing.T) {
	assert.Equal(t, true, InInt64s(4, []int64{2, 4, 6, 7, 1, 3}...))
}

func TestInUint64s(t *testing.T) {
	assert.Equal(t, true, InUint64s(4, []uint64{2, 4, 6, 7, 1, 3}...))
}

func TestInFloat64s(t *testing.T) {
	assert.Equal(t, true, InFloat64s(4.4, []float64{2.3, 4.4, 6.7, 7.2, 1.9, 3.5}...))
}

func TestInStrings(t *testing.T) {
	assert.Equal(t, true, InStrings("shenghui0779", []string{"hello", "test", "shenghui0779", "yiigo", "world"}...))
}

func TestInArray(t *testing.T) {
	assert.Equal(t, true, InArray("shenghui0779", []interface{}{1, "test", "shenghui0779", 2.9, true}...))
}

func TestIntsUnique(t *testing.T) {
	assert.Equal(t, []int{2, 4, 6, 7, 1, 3, 9}, IntsUnique([]int{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestUintsUnique(t *testing.T) {
	assert.Equal(t, []uint{2, 4, 6, 7, 1, 3, 9}, UintsUnique([]uint{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestInt8sUnique(t *testing.T) {
	assert.Equal(t, []int8{2, 4, 6, 7, 1, 3, 9}, Int8sUnique([]int8{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestUint8sUnique(t *testing.T) {
	assert.Equal(t, []uint8{2, 4, 6, 7, 1, 3, 9}, Uint8sUnique([]uint8{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestInt16sUnique(t *testing.T) {
	assert.Equal(t, []int16{2, 4, 6, 7, 1, 3, 9}, Int16sUnique([]int16{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestUint16sUnique(t *testing.T) {
	assert.Equal(t, []uint16{2, 4, 6, 7, 1, 3, 9}, Uint16sUnique([]uint16{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestInt32sUnique(t *testing.T) {
	assert.Equal(t, []int32{2, 4, 6, 7, 1, 3, 9}, Int32sUnique([]int32{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestUint32sUnique(t *testing.T) {
	assert.Equal(t, []uint32{2, 4, 6, 7, 1, 3, 9}, Uint32sUnique([]uint32{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestInt64sUnique(t *testing.T) {
	assert.Equal(t, []int64{2, 4, 6, 7, 1, 3, 9}, Int64sUnique([]int64{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestUint64sUnique(t *testing.T) {
	assert.Equal(t, []uint64{2, 4, 6, 7, 1, 3, 9}, Uint64sUnique([]uint64{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestFloat64sUnique(t *testing.T) {
	assert.Equal(t, []float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 9.2}, Float64sUnique([]float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 4.2, 9.2, 7.2}))
}

func TestStringsUnique(t *testing.T) {
	assert.Equal(t, []string{"a", "c", "d", "e", "x", "f"}, StringsUnique([]string{"a", "c", "d", "a", "e", "d", "x", "f", "c"}))
}
