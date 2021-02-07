package yiigo

import (
	"reflect"
	"sort"
)

// Int64Slice attaches the methods of Interface to []int64, sorting a increasing order.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt64s sorts []int64s in increasing order.
func SortInt64s(a []int64) {
	sort.Sort(Int64Slice(a))
}

// SearchInt64s searches for x in a sorted slice of int64s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// InInts checks if a value of int exists in a slice of []int.
func InInts(needle int, haystack []int) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InInt64s checks if a value of int64 exists in a slice of []int64.
func InInt64s(needle int64, haystack []int64) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InFloat64s checks if a value of float64 exists in a slice of []float64.
func InFloat64s(needle float64, haystack []float64) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InStrings checks if a value of string exists in a slice of []string.
func InStrings(needle string, haystack []string) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// InArray checks if a value of interface{} exists in a slice of []interface{}.
func InArray(needle interface{}, haystack []interface{}) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, v := range haystack {
		if reflect.DeepEqual(needle, v) {
			return true
		}
	}

	return false
}

// IntsUnique takes an input slice of ints and
// returns a new slice of ints without duplicate values.
func IntsUnique(a []int) []int {
	l := len(a)

	if l <= 1 {
		return a
	}

	m := make(map[int]byte, l)
	r := make([]int, 0, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Int64sUnique takes an input slice of int64s and
// returns a new slice of int64s without duplicate values.
func Int64sUnique(a []int64) []int64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	m := make(map[int64]byte, l)
	r := make([]int64, 0, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Float64sUnique takes an input slice of float64s and
// returns a new slice of float64s without duplicate values.
func Float64sUnique(a []float64) []float64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	m := make(map[float64]byte, l)
	r := make([]float64, 0, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// StringsUnique takes an input slice of strings and
// returns a new slice of strings without duplicate values.
func StringsUnique(a []string) []string {
	l := len(a)

	if l <= 1 {
		return a
	}

	m := make(map[string]byte, l)
	r := make([]string, 0, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}
