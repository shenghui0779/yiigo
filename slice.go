package yiigo

import "math/rand"

// SliceIn 返回指定元素是否在集合中
func SliceIn[T ~[]E, E comparable](list T, elem E) bool {
	if len(list) == 0 {
		return false
	}

	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

// SliceUniq 集合去重
func SliceUniq[T ~[]E, E comparable](list T) T {
	if len(list) == 0 {
		return list
	}

	ret := make(T, 0, len(list))
	m := make(map[E]struct{}, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}
	return ret
}

// SliceDiff 返回两个集合之间的差异
func SliceDiff[T ~[]E, E comparable](list1 T, list2 T) (ret1 T, ret2 T) {
	m1 := map[E]struct{}{}
	m2 := map[E]struct{}{}
	for _, v := range list1 {
		m1[v] = struct{}{}
	}
	for _, v := range list2 {
		m2[v] = struct{}{}
	}

	ret1 = make(T, 0)
	ret2 = make(T, 0)
	for _, v := range list1 {
		if _, ok := m2[v]; !ok {
			ret1 = append(ret1, v)
		}
	}
	for _, v := range list2 {
		if _, ok := m1[v]; !ok {
			ret2 = append(ret2, v)
		}
	}
	return ret1, ret2
}

// SliceWithout 返回不包括所有给定值的切片
func SliceWithout[T ~[]E, E comparable](list T, exclude ...E) T {
	if len(list) == 0 {
		return list
	}

	m := make(map[E]struct{}, len(exclude))
	for _, v := range exclude {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceIntersect 返回两个集合的交集
func SliceIntersect[T ~[]E, E comparable](list1 T, list2 T) T {
	m := make(map[E]struct{})
	for _, v := range list1 {
		m[v] = struct{}{}
	}

	ret := make(T, 0)
	for _, v := range list2 {
		if _, ok := m[v]; ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceUnion 返回两个集合的并集
func SliceUnion[T ~[]E, E comparable](lists ...T) T {
	ret := make(T, 0)
	m := make(map[E]struct{})
	for _, list := range lists {
		for _, v := range list {
			if _, ok := m[v]; !ok {
				ret = append(ret, v)
				m[v] = struct{}{}
			}
		}
	}
	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(list)，则返回打乱的切片
func SliceRand[T ~[]E, E any](list T, n int) T {
	if n == 0 || n < -1 {
		return nil
	}

	l := len(list)
	ret := make(T, l)
	copy(ret, list)
	rand.Shuffle(l, func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})
	if n == -1 || n >= l {
		return ret
	}
	return ret[:n]
}
