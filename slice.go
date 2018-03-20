package yiigo

import (
	"sort"
)

// Int64Slice attaches the methods of Interface to []int64, sorting a increasing order.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt64s sorts a slice of int64s in increasing order.
func SortInt64s(a []int64) {
	sort.Sort(Int64Slice(a))
}

// SearchInt64s searches for x in a sorted slice of ints and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// InSliceInt checks if x exists in a slice of ints and
// returns TRUE if x is found.
func InSliceInt(x int, a []int) bool {
	l := len(a)

	if l == 0 {
		return false
	}

	sort.Ints(a)

	i := sort.SearchInts(a, x)

	if i < l && a[i] == x {
		return true
	}

	return false
}

// InSliceInt64 checks if x exists in a slice of int64s and
// returns TRUE if x is found.
func InSliceInt64(x int64, a []int64) bool {
	l := len(a)

	if l == 0 {
		return false
	}

	SortInt64s(a)

	i := SearchInt64s(a, x)

	if i < l && a[i] == x {
		return true
	}

	return false
}

// InSliceFloat64 checks if x exists in a slice of float64s and
// returns TRUE if x is found.
func InSliceFloat64(x float64, a []float64) bool {
	l := len(a)

	if l == 0 {
		return false
	}

	sort.Float64s(a)

	i := sort.SearchFloat64s(a, x)

	if i < l && a[i] == x {
		return true
	}

	return false
}

// InSliceString checks if x exists in a slice of strings and
// returns TRUE if x is found.
func InSliceString(x string, a []string) bool {
	l := len(a)

	if l == 0 {
		return false
	}

	sort.Strings(a)

	i := sort.SearchStrings(a, x)

	if i < l && a[i] == x {
		return true
	}

	return false
}

// UniqueInt takes an input slice of ints and
// returns a new sorted slice of ints without duplicate values.
func UniqueInt(a []int) []int {
	r := make([]int, 0, len(a))

	sort.Ints(a)

	r = append(r, a[0])
	i := a[0]

	for _, v := range a {
		if v != i {
			r = append(r, v)
			i = v
		}
	}

	return r
}

// UniqueInt64 takes an input slice of int64s and
// returns a new sorted slice of int64s without duplicate values.
func UniqueInt64(a []int64) []int64 {
	r := make([]int64, 0, len(a))

	SortInt64s(a)

	r = append(r, a[0])
	i := a[0]

	for _, v := range a {
		if v != i {
			r = append(r, v)
			i = v
		}
	}

	return r
}

// UniqueFloat64 takes an input slice of float64s and
// returns a new sorted slice of float64s without duplicate values.
func UniqueFloat64(a []float64) []float64 {
	r := make([]float64, 0, len(a))

	sort.Float64s(a)

	r = append(r, a[0])
	i := a[0]

	for _, v := range a {
		if v != i {
			r = append(r, v)
			i = v
		}
	}

	return r
}

// UniqueString takes an input slice of strings and
// returns a new sorted slice of strings without duplicate values.
func UniqueString(a []string) []string {
	r := make([]string, 0, len(a))

	sort.Strings(a)

	r = append(r, a[0])
	i := a[0]

	for _, v := range a {
		if v != i {
			r = append(r, v)
			i = v
		}
	}

	return r
}
