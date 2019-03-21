package yiigo

import "sort"

const (
	uniqueNumberThreshold = 1024
	uniqueStringThreshold = 256
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

// SearchInt64s searches for x in a sorted slice of int64s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// InInts checks if x exists in a slice of ints and returns TRUE if x is found.
func InInts(x int, a ...int) bool {
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

// InInt32s checks if x exists in a slice of int32s and returns TRUE if x is found.
func InInt32s(x int32, a ...int32) bool {
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

// InUint32s checks if x exists in a slice of uint32s and returns TRUE if x is found.
func InUint32s(x uint32, a ...uint32) bool {
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

// InInt64s checks if x exists in a slice of int64s and returns TRUE if x is found.
func InInt64s(x int64, a ...int64) bool {
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

// InUint64s checks if x exists in a slice of uint64s and returns TRUE if x is found.
func InUint64s(x uint64, a ...uint64) bool {
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

// InFloat64s checks if x exists in a slice of float64s and returns TRUE if x is found.
func InFloat64s(x float64, a ...float64) bool {
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

// InStrings checks if x exists in a slice of strings and returns TRUE if x is found.
func InStrings(x string, a ...string) bool {
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

// UniqueInts takes an input slice of ints and
// returns a new slice of ints without duplicate values.
func UniqueInts(a []int) []int {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueIntByLoop(a, l)
	}

	return uniqueIntByMap(a, l)
}

// UniqueInt32s takes an input slice of int32s and
// returns a new slice of int32s without duplicate values.
func UniqueInt32s(a []int32) []int32 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueInt32ByLoop(a, l)
	}

	return uniqueInt32ByMap(a, l)
}

// UniqueUint32s takes an input slice of uint32s and
// returns a new slice of uint32s without duplicate values.
func UniqueUint32s(a []uint32) []uint32 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueUint32ByLoop(a, l)
	}

	return uniqueUint32ByMap(a, l)
}

// UniqueInt64s takes an input slice of int64s and
// returns a new slice of int64s without duplicate values.
func UniqueInt64s(a []int64) []int64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueInt64ByLoop(a, l)
	}

	return uniqueInt64ByMap(a, l)
}

// UniqueUint64s takes an input slice of uint64s and
// returns a new slice of uint64s without duplicate values.
func UniqueUint64s(a []uint64) []uint64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueUint64ByLoop(a, l)
	}

	return uniqueUint64ByMap(a, l)
}

// UniqueFloat64s takes an input slice of float64s and
// returns a new slice of float64s without duplicate values.
func UniqueFloat64s(a []float64) []float64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueNumberThreshold {
		return uniqueFloat64ByLoop(a, l)
	}

	return uniqueFloat64ByMap(a, l)
}

// UniqueStrings takes an input slice of strings and
// returns a new slice of strings without duplicate values.
func UniqueStrings(a []string) []string {
	l := len(a)

	if l <= 1 {
		return a
	}

	if l < uniqueStringThreshold {
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

func uniqueInt32ByLoop(a []int32, l int) []int32 {
	r := make([]int32, 0, l)

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

func uniqueInt32ByMap(a []int32, l int) []int32 {
	r := make([]int32, 0, l)
	m := make(map[int32]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

func uniqueUint32ByLoop(a []uint32, l int) []uint32 {
	r := make([]uint32, 0, l)

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

func uniqueUint32ByMap(a []uint32, l int) []uint32 {
	r := make([]uint32, 0, l)
	m := make(map[uint32]byte, l)

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

func uniqueUint64ByLoop(a []uint64, l int) []uint64 {
	r := make([]uint64, 0, l)

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

func uniqueUint64ByMap(a []uint64, l int) []uint64 {
	r := make([]uint64, 0, l)
	m := make(map[uint64]byte, l)

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
