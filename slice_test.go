package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceUniq(t *testing.T) {
	assert.Equal(t, 4, len(SliceUniq([]int{1, 2, 1, 3, 4, 3})))
	assert.Equal(t, 4, len(SliceUniq([]int64{1, 2, 1, 3, 4, 3})))
	assert.Equal(t, 4, len(SliceUniq([]float64{1.01, 2.02, 1.01, 3.03, 4.04, 3.03})))
	assert.Equal(t, 4, len(SliceUniq([]string{"h", "e", "l", "l", "o"})))
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
