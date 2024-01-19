package util

import (
	"math/rand"
)

// SliceUniq 切片去重
func SliceUniq[T ~int | ~int64 | ~float64 | ~string](a []T) []T {
	ret := make([]T, 0)
	if len(a) == 0 {
		return ret
	}

	m := make(map[T]struct{}, 0)

	for _, v := range a {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}

	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(a)，则返回打乱的切片
func SliceRand[T any](a []T, n int) []T {
	if n == 0 || n < -1 {
		return make([]T, 0)
	}

	count := len(a)
	ret := make([]T, count)

	copy(ret, a)

	rand.Shuffle(count, func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})

	if n == -1 || n >= count {
		return ret
	}

	return ret[:n]
}
