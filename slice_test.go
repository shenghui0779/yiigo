package yiigo

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortInt64s(t *testing.T) {
	a := []int64{4, 2, 7, 9, 1}
	SortInt64s(a)

	assert.Equal(t, true, sort.IsSorted(Int64Slice(a)))
}

func TestSearchInt64s(t *testing.T) {
	assert.Equal(t, 2, SearchInt64s([]int64{4, 1, 7, 2, 9}, 4))
}

func TestInInts(t *testing.T) {
	assert.Equal(t, true, InInts(4, []int{2, 4, 6, 7, 1, 3}))
}

func TestInInt64s(t *testing.T) {
	assert.Equal(t, true, InInt64s(4, []int64{2, 4, 6, 7, 1, 3}))
}

func TestInFloat64s(t *testing.T) {
	assert.Equal(t, true, InFloat64s(4.4, []float64{2.3, 4.4, 6.7, 7.2, 1.9, 3.5}))
}

func TestInStrings(t *testing.T) {
	assert.Equal(t, true, InStrings("shenghui0779", []string{"hello", "test", "shenghui0779", "yiigo", "world"}))
}

func TestInArray(t *testing.T) {
	assert.Equal(t, true, InArray("shenghui0779", []interface{}{1, "test", "shenghui0779", 2.9, true}))
}

func TestIntsUnique(t *testing.T) {
	assert.Equal(t, []int{2, 4, 6, 7, 1, 3, 9}, IntsUnique([]int{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestInt64sUnique(t *testing.T) {
	assert.Equal(t, []int64{2, 4, 6, 7, 1, 3, 9}, Int64sUnique([]int64{2, 4, 6, 7, 1, 3, 4, 9, 7}))
}

func TestFloat64sUnique(t *testing.T) {
	assert.Equal(t, []float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 9.2}, Float64sUnique([]float64{2.2, 4.2, 6.2, 7.2, 1.2, 3.2, 4.2, 9.2, 7.2}))
}

func TestStringsUnique(t *testing.T) {
	assert.Equal(t, []string{"a", "c", "d", "e", "x", "f"}, StringsUnique([]string{"a", "c", "d", "a", "e", "d", "x", "f", "c"}))
}
