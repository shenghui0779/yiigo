package yiigo

import (
	"reflect"
	"sort"
)

const (
	numberUniqueThreshold = 1024
	stringUniqueThreshold = 256
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
func InInts(x int, y ...int) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InInt32s checks if x exists in a slice of int32s and returns TRUE if x is found.
func InInt32s(x int32, y ...int32) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InUint32s checks if x exists in a slice of uint32s and returns TRUE if x is found.
func InUint32s(x uint32, y ...uint32) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InInt64s checks if x exists in a slice of int64s and returns TRUE if x is found.
func InInt64s(x int64, y ...int64) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InUint64s checks if x exists in a slice of uint64s and returns TRUE if x is found.
func InUint64s(x uint64, y ...uint64) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InFloat64s checks if x exists in a slice of float64s and returns TRUE if x is found.
func InFloat64s(x float64, y ...float64) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InStrings checks if x exists in a slice of strings and returns TRUE if x is found.
func InStrings(x string, y ...string) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if x == v {
			return true
		}
	}

	return false
}

// InArray checks if x exists in a slice and returns TRUE if x is found.
func InArray(x interface{}, y ...interface{}) bool {
	if len(y) == 0 {
		return false
	}

	for _, v := range y {
		if reflect.DeepEqual(x, v) {
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

	r := make([]int, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[int]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Int32sUnique takes an input slice of int32s and
// returns a new slice of int32s without duplicate values.
func Int32sUnique(a []int32) []int32 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]int32, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[int32]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Uint32sUnique takes an input slice of uint32s and
// returns a new slice of uint32s without duplicate values.
func Uint32sUnique(a []uint32) []uint32 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]uint32, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[uint32]byte, l)

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

	r := make([]int64, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[int64]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Uint64sUnique takes an input slice of uint64s and
// returns a new slice of uint64s without duplicate values.
func Uint64sUnique(a []uint64) []uint64 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]uint64, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[uint64]byte, l)

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

	r := make([]float64, 0, l)

	// remove duplicates with loop
	if l < numberUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[float64]byte, l)

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

	r := make([]string, 0, l)

	// remove duplicates with loop
	if l < stringUniqueThreshold {
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

	// remove duplicates with map
	m := make(map[string]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}
