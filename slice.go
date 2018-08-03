package yiigo

import "sort"

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

// InSliceInt checks if x exists in a slice of ints and returns TRUE if x is found.
func InSliceInt(x int, a []int) bool {
	if len(a) == 0 {
		return false
	}

	for _, v := range a {
		if x == v {
			return true
		}
	}

	return false
}

// InSliceInt64 checks if x exists in a slice of int64s and returns TRUE if x is found.
func InSliceInt64(x int64, a []int64) bool {
	if len(a) == 0 {
		return false
	}

	for _, v := range a {
		if x == v {
			return true
		}
	}

	return false
}

// InSliceFloat64 checks if x exists in a slice of float64s and returns TRUE if x is found.
func InSliceFloat64(x float64, a []float64) bool {
	if len(a) == 0 {
		return false
	}

	for _, v := range a {
		if x == v {
			return true
		}
	}

	return false
}

// InSliceString checks if x exists in a slice of strings and returns TRUE if x is found.
func InSliceString(x string, a []string) bool {
	if len(a) == 0 {
		return false
	}

	for _, v := range a {
		if x == v {
			return true
		}
	}

	return false
}

// UniqueInt takes an input slice of ints and
// returns a new slice of ints without duplicate values.
func UniqueInt(a []int) []int {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < 1024 {
		return uniqueIntByLoop(a, l)
	}

	return uniqueIntByMap(a, l)
}

// UniqueInt64 takes an input slice of int64s and
// returns a new slice of int64s without duplicate values.
func UniqueInt64(a []int64) []int64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < 1024 {
		return uniqueInt64ByLoop(a, l)
	}

	return uniqueInt64ByMap(a, l)
}

// UniqueFloat64 takes an input slice of float64s and
// returns a new slice of float64s without duplicate values.
func UniqueFloat64(a []float64) []float64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < 1024 {
		return uniqueFloat64ByLoop(a, l)
	}

	return uniqueFloat64ByMap(a, l)
}

// UniqueString takes an input slice of strings and
// returns a new slice of strings without duplicate values.
func UniqueString(a []string) []string {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < 256 {
		return uniqueStringByLoop(a, l)
	}

	return uniqueStringByMap(a, l)
}

func uniqueIntByLoop(a []int, l int) []int {
	r := make([]int, 0, l)

	for _, v := range a {
		exist := false

		for _, u := range r {
			if v == u {
				exist = true
				break
			}
		}

		if !exist {
			r = append(r, v)
		}
	}

	return r
}

func uniqueIntByMap(a []int, l int) []int {
	r := make([]int, 0, l)
	m := make(map[int]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

func uniqueInt64ByLoop(a []int64, l int) []int64 {
	r := make([]int64, 0, l)

	for _, v := range a {
		exist := false

		for _, u := range r {
			if v == u {
				exist = true
				break
			}
		}

		if !exist {
			r = append(r, v)
		}
	}

	return r
}

func uniqueInt64ByMap(a []int64, l int) []int64 {
	r := make([]int64, 0, l)
	m := make(map[int64]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

func uniqueFloat64ByLoop(a []float64, l int) []float64 {
	r := make([]float64, 0, l)

	for _, v := range a {
		exist := false

		for _, u := range r {
			if v == u {
				exist = true
				break
			}
		}

		if !exist {
			r = append(r, v)
		}
	}

	return r
}

func uniqueFloat64ByMap(a []float64, l int) []float64 {
	r := make([]float64, 0, l)
	m := make(map[float64]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

func uniqueStringByLoop(a []string, l int) []string {
	r := make([]string, 0, l)

	for _, v := range a {
		exist := false

		for _, u := range r {
			if v == u {
				exist = true
				break
			}
		}

		if !exist {
			r = append(r, v)
		}
	}

	return r
}

func uniqueStringByMap(a []string, l int) []string {
	r := make([]string, 0, l)
	m := make(map[string]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}
