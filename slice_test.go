package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceIn(t *testing.T) {
	assert.True(t, SliceIn([]int{1, 2, 3, 4, 5}, 2))
	assert.True(t, SliceIn([]int64{1, 2, 3, 4, 5}, 2))
	assert.True(t, SliceIn([]float64{1.01, 2.02, 3.03, 4.04, 5.05}, 2.02))
	assert.True(t, SliceIn([]string{"h", "e", "l", "l", "o"}, "e"))
}

func TestSliceUniq(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3, 4}, SliceUniq([]int{1, 2, 1, 3, 4, 3}))
	assert.Equal(t, []int64{1, 2, 3, 4}, SliceUniq([]int64{1, 2, 1, 3, 4, 3}))
	assert.Equal(t, []float64{1.01, 2.02, 3.03, 4.04}, SliceUniq([]float64{1.01, 2.02, 1.01, 3.03, 4.04, 3.03}))
	assert.Equal(t, []string{"h", "e", "l", "o"}, SliceUniq([]string{"h", "e", "l", "l", "o"}))
}

func TestSliceDiff(t *testing.T) {
	left1, right1 := SliceDiff([]int{0, 1, 2, 3, 4, 5}, []int{0, 2, 6})
	assert.Equal(t, []int{1, 3, 4, 5}, left1)
	assert.Equal(t, []int{6}, right1)

	left2, right2 := SliceDiff([]int{1, 2, 3, 4, 5}, []int{0, 6})
	assert.Equal(t, []int{1, 2, 3, 4, 5}, left2)
	assert.Equal(t, []int{0, 6}, right2)

	left3, right3 := SliceDiff([]int{0, 1, 2, 3, 4, 5}, []int{0, 1, 2, 3, 4, 5})
	assert.Equal(t, []int{}, left3)
	assert.Equal(t, []int{}, right3)
}

func TestSliceWithout(t *testing.T) {
	result1 := SliceWithout([]int{0, 2, 10}, 0, 1, 2, 3, 4, 5)
	result2 := SliceWithout([]int{0, 7}, 0, 1, 2, 3, 4, 5)
	result3 := SliceWithout([]int{}, 0, 1, 2, 3, 4, 5)
	result4 := SliceWithout([]int{0, 1, 2}, 0, 1, 2)
	result5 := SliceWithout([]int{})
	assert.Equal(t, []int{10}, result1)
	assert.Equal(t, []int{7}, result2)
	assert.Equal(t, []int{}, result3)
	assert.Equal(t, []int{}, result4)
	assert.Equal(t, []int{}, result5)
}

func TestSliceIntersect(t *testing.T) {
	result1 := SliceIntersect([]int{0, 1, 2, 3, 4, 5}, []int{0, 2})
	result2 := SliceIntersect([]int{0, 1, 2, 3, 4, 5}, []int{0, 6})
	result3 := SliceIntersect([]int{0, 1, 2, 3, 4, 5}, []int{-1, 6})
	result4 := SliceIntersect([]int{0, 6}, []int{0, 1, 2, 3, 4, 5})
	result5 := SliceIntersect([]int{0, 6, 0}, []int{0, 1, 2, 3, 4, 5})

	assert.Equal(t, []int{0, 2}, result1)
	assert.Equal(t, []int{0}, result2)
	assert.Equal(t, []int{}, result3)
	assert.Equal(t, []int{0}, result4)
	assert.Equal(t, []int{0}, result5)
}

func TestSliceUnion(t *testing.T) {
	result1 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{0, 2, 10})
	result2 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{6, 7})
	result3 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{})
	result4 := SliceUnion([]int{0, 1, 2}, []int{0, 1, 2})
	result5 := SliceUnion([]int{}, []int{})
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 10}, result1)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, result2)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, result3)
	assert.Equal(t, []int{0, 1, 2}, result4)
	assert.Equal(t, []int{}, result5)

	result11 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{0, 2, 10}, []int{0, 1, 11})
	result12 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{6, 7}, []int{8, 9})
	result13 := SliceUnion([]int{0, 1, 2, 3, 4, 5}, []int{}, []int{})
	result14 := SliceUnion([]int{0, 1, 2}, []int{0, 1, 2}, []int{0, 1, 2})
	result15 := SliceUnion([]int{}, []int{}, []int{})
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 10, 11}, result11)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, result12)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, result13)
	assert.Equal(t, []int{0, 1, 2}, result14)
	assert.Equal(t, []int{}, result15)
}

func TestSliceRand(t *testing.T) {
	a1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ret1 := SliceRand(a1, 2)
	assert.Equal(t, 2, len(ret1))
	assert.NotEqual(t, a1[:2], ret1)

	a2 := []float64{1.01, 2.02, 3.03, 4.04, 5.05, 6.06, 7.07, 8.08, 9.09, 10.10}
	ret2 := SliceRand(a2, 2)
	assert.Equal(t, 2, len(ret2))
	assert.NotEqual(t, a2[:2], ret2)

	a3 := []string{"h", "e", "l", "l", "o", "w", "o", "r", "l", "d"}
	ret3 := SliceRand(a3, 2)
	assert.Equal(t, 2, len(ret3))
	assert.NotEqual(t, a3[:2], ret3)

	type User struct {
		ID   int64
		Name string
	}

	a4 := []User{
		{
			ID:   1,
			Name: "h",
		},
		{
			ID:   2,
			Name: "e",
		},
		{
			ID:   3,
			Name: "l",
		},
		{
			ID:   4,
			Name: "l",
		},
		{
			ID:   5,
			Name: "o",
		},
		{
			ID:   6,
			Name: "w",
		},
		{
			ID:   7,
			Name: "o",
		},
		{
			ID:   8,
			Name: "r",
		},
		{
			ID:   9,
			Name: "l",
		},
		{
			ID:   10,
			Name: "d",
		},
	}

	ret4 := SliceRand(a4, 2)
	assert.Equal(t, 2, len(ret4))
	assert.NotEqual(t, a4[:2], ret4)

	ret5 := SliceRand(a4, -1)
	assert.Equal(t, len(a4), len(ret5))
	assert.NotEqual(t, a4, ret5)
}
