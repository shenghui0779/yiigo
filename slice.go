package yiigo

import (
	"sort"
)

// Int64Slice attaches the methods of Interface to []int64, sorting a increasing order.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortInt64s 将 Int64 切变递增排序
func SortInt64s(a []int64) {
	sort.Sort(Int64Slice(a))
}

// SearchInt64s 在递增顺序的切片a中搜索x，返回x的索引。如果查找不到，返回值是x应该插入a的位置「以保证a的递增顺序」，返回值可以是len(a)
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// UniqueInt Int 切片去重「返回一个递增排序的切片」
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

// UniqueInt64 Int64 切片去重「返回一个递增排序的切片」
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

// UniqueFloat64 Float64 切片去重「返回一个递增排序的切片」
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

// UniqueString String 切片去重「返回一个递增排序的切片」
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

// InSliceInt 检查 Int 值是否存在于 Int 切片中
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

// InSliceInt64 检查 Int64 值是否存在于 Int64 切片中
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

// InSliceFloat64 检查 Float64 值是否存在于 Float64 切片中
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

// InSliceString 检查 String 值是否存在于 String 切片中
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
