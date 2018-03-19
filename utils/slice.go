package utils

import (
	"sort"
)

// Int64Slice attaches the methods of Interface to []int64, sorting in increasing order.
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

// UniqueInt Int 切片去重
func UniqueInt(in []int) []int {
	out := make([]int, 0, len(in))

	sort.Ints(in)

	out = append(out, in[0])
	i := in[0]

	for _, v := range in {
		if v != i {
			out = append(out, v)
			i = v
		}
	}

	return out
}

// UniqueInt64 Int64 切片去重
func UniqueInt64(in []int64) []int64 {
	out := make([]int64, 0, len(in))

	SortInt64s(in)

	out = append(out, in[0])
	i := in[0]

	for _, v := range in {
		if v != i {
			out = append(out, v)
			i = v
		}
	}

	return out
}

// UniqueFloat64 Float64 切片去重
func UniqueFloat64(in []float64) []float64 {
	out := make([]float64, 0, len(in))

	sort.Float64s(in)

	out = append(out, in[0])
	i := in[0]

	for _, v := range in {
		if v != i {
			out = append(out, v)
			i = v
		}
	}

	return out
}

// UniqueString String 切片去重
func UniqueString(in []string) []string {
	out := make([]string, 0, len(in))

	sort.Strings(in)

	out = append(out, in[0])
	i := in[0]

	for _, v := range in {
		if v != i {
			out = append(out, v)
			i = v
		}
	}

	return out
}

// InSliceInt 检查 Int 值是否存在于 Int 切片中
func InSliceInt(needle int, haystack []int) bool {
	count := len(haystack)

	if count == 0 {
		return false
	}

	sort.Ints(haystack)

	i := sort.SearchInts(haystack, needle)

	if i < count && haystack[i] == needle {
		return true
	}

	return false
}

// InSliceInt64 检查 Int64 值是否存在于 Int64 切片中
func InSliceInt64(needle int64, haystack []int64) bool {
	count := len(haystack)

	if count == 0 {
		return false
	}

	SortInt64s(haystack)

	i := SearchInt64s(haystack, needle)

	if i < count && haystack[i] == needle {
		return true
	}

	return false
}

// InSliceFloat64 检查 Float64 值是否存在于 Float64 切片中
func InSliceFloat64(needle float64, haystack []float64) bool {
	count := len(haystack)

	if count == 0 {
		return false
	}

	sort.Float64s(haystack)

	i := sort.SearchFloat64s(haystack, needle)

	if i < count && haystack[i] == needle {
		return true
	}

	return false
}

// InSliceString 检查 String 值是否存在于 String 切片中
func InSliceString(needle string, haystack []string) bool {
	count := len(haystack)

	if count == 0 {
		return false
	}

	sort.Strings(haystack)

	i := sort.SearchStrings(haystack, needle)

	if i < count && haystack[i] == needle {
		return true
	}

	return false
}
