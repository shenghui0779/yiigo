package yiigo

import (
	"reflect"
	"sort"
)

const (
	numberUniqueThreshold = 1024
	stringUniqueThreshold = 256
)

// UintSlice attaches the methods of Interface to []uint, sorting a increasing order.
type UintSlice []uint

func (p UintSlice) Len() int           { return len(p) }
func (p UintSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p UintSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortUints sorts []uints in increasing order.
func SortUints(a []uint) {
	sort.Sort(UintSlice(a))
}

// SearchUints searches for x in a sorted slice of uints and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUints(a []uint, x uint) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int8Slice attaches the methods of Interface to []int8, sorting a increasing order.
type Int8Slice []int8

func (p Int8Slice) Len() int           { return len(p) }
func (p Int8Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int8Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt8s sorts []int8s in increasing order.
func SortInt8s(a []int8) {
	sort.Sort(Int8Slice(a))
}

// SearchInt8s searches for x in a sorted slice of int8s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt8s(a []int8, x int8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint8Slice attaches the methods of Interface to []uint8, sorting a increasing order.
type Uint8Slice []uint8

func (p Uint8Slice) Len() int           { return len(p) }
func (p Uint8Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint8Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortUint8s sorts []uint8s in increasing order.
func SortUint8s(a []uint8) {
	sort.Sort(Uint8Slice(a))
}

// SearchUint8s searches for x in a sorted slice of uint8s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint8s(a []uint8, x uint8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int16Slice attaches the methods of Interface to []int16, sorting a increasing order.
type Int16Slice []int16

func (p Int16Slice) Len() int           { return len(p) }
func (p Int16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt16s sorts []int16s in increasing order.
func SortInt16s(a []int16) {
	sort.Sort(Int16Slice(a))
}

// SearchInt16s searches for x in a sorted slice of int16s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt16s(a []int16, x int16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint16Slice attaches the methods of Interface to []uint16, sorting a increasing order.
type Uint16Slice []uint16

func (p Uint16Slice) Len() int           { return len(p) }
func (p Uint16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortUint16s sorts []uint16s in increasing order.
func SortUint16s(a []uint16) {
	sort.Sort(Uint16Slice(a))
}

// SearchUints searches for x in a sorted slice of uint16s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint16s(a []uint16, x uint16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int32Slice attaches the methods of Interface to []int32, sorting a increasing order.
type Int32Slice []int32

func (p Int32Slice) Len() int           { return len(p) }
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt32s sorts []int32s in increasing order.
func SortInt32s(a []int32) {
	sort.Sort(Int32Slice(a))
}

// SearchInt32s searches for x in a sorted slice of int32s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt32s(a []int32, x int32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint32Slice attaches the methods of Interface to []uint, sorting a increasing order.
type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortUint32s sorts []uint32s in increasing order.
func SortUint32s(a []uint32) {
	sort.Sort(Uint32Slice(a))
}

// SearchUint32s searches for x in a sorted slice of uint32s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint32s(a []uint32, x uint32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

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

// Uint64Slice attaches the methods of Interface to []uint64, sorting a increasing order.
type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortUint64s sorts []uint64s in increasing order.
func SortUint64s(a []uint64) {
	sort.Sort(Uint64Slice(a))
}

// SearchUint64s searches for x in a sorted slice of uint64s and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint64s(a []uint64, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// InInts checks if x exists in []ints and returns TRUE if x is found.
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

// InUints checks if x exists in []uints and returns TRUE if x is found.
func InUints(x uint, y ...uint) bool {
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

// InInt8s checks if x exists in []int8s and returns TRUE if x is found.
func InInt8s(x int8, y ...int8) bool {
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

// InUint8s checks if x exists in []uint8s and returns TRUE if x is found.
func InUint8s(x uint8, y ...uint8) bool {
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

// InInt16s checks if x exists in []int16s and returns TRUE if x is found.
func InInt16s(x int16, y ...int16) bool {
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

// InUint16s checks if x exists in []uint16s and returns TRUE if x is found.
func InUint16s(x uint16, y ...uint16) bool {
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

// InInt32s checks if x exists in []int32s and returns TRUE if x is found.
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

// InUint32s checks if x exists in []uint32s and returns TRUE if x is found.
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

// InInt64s checks if x exists in []int64s and returns TRUE if x is found.
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

// InUint64s checks if x exists in []uint64s and returns TRUE if x is found.
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

// InFloat64s checks if x exists in []float64s and returns TRUE if x is found.
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

// InStrings checks if x exists in []strings and returns TRUE if x is found.
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

// UintsUnique takes an input slice of uints and
// returns a new slice of uints without duplicate values.
func UintsUnique(a []uint) []uint {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]uint, 0, l)

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
	m := make(map[uint]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Int8sUnique takes an input slice of int8s and
// returns a new slice of int8s without duplicate values.
func Int8sUnique(a []int8) []int8 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]int8, 0, l)

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
	m := make(map[int8]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Uint8sUnique takes an input slice of uint8s and
// returns a new slice of uint8s without duplicate values.
func Uint8sUnique(a []uint8) []uint8 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]uint8, 0, l)

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
	m := make(map[uint8]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Int16sUnique takes an input slice of int16s and
// returns a new slice of int16s without duplicate values.
func Int16sUnique(a []int16) []int16 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]int16, 0, l)

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
	m := make(map[int16]byte, l)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = 0
			r = append(r, v)
		}
	}

	return r
}

// Uint16sUnique takes an input slice of uint16s and
// returns a new slice of uint16s without duplicate values.
func Uint16sUnique(a []uint16) []uint16 {
	l := len(a)

	if l <= 1 {
		return a
	}

	r := make([]uint16, 0, l)

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
	m := make(map[uint16]byte, l)

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
