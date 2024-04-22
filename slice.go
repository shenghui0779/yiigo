package yiigo

import (
	"math/rand"
)

// SliceUniq 切片去重
func SliceUniq[T ~[]E, E comparable](a T) T {
	if len(a) == 0 {
		return a
	}

	m := make(map[E]struct{})
	for _, v := range a {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}

	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(a)，则返回打乱的切片
func SliceRand[T any](a []T, n int) []T {
	if n == 0 || n < -1 {
		return nil
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
