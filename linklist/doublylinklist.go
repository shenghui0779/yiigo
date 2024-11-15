package linklist

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// DoublyLinkList 双向链表(并发安全)
type DoublyLinkList[T comparable] struct {
	first *element[T]
	last  *element[T]
	size  int
	mutex sync.RWMutex
}

type element[T comparable] struct {
	value T
	prev  *element[T]
	next  *element[T]
}

// New 初始化一个链表
func New[T comparable](values ...T) *DoublyLinkList[T] {
	list := &DoublyLinkList[T]{}
	if len(values) > 0 {
		list.appendWithoutLock(values...)
	}
	return list
}

// Append 向链表尾部追加值
func (list *DoublyLinkList[T]) Append(values ...T) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	list.appendWithoutLock(values...)
}

// appendWithoutLock 向链表尾部追加值(不持有锁)
func (list *DoublyLinkList[T]) appendWithoutLock(values ...T) {
	for _, value := range values {
		newElement := &element[T]{value: value, prev: list.last}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.last.next = newElement
			list.last = newElement
		}
		list.size++
	}
}

// Prepend 向链表头部追加值
func (list *DoublyLinkList[T]) Prepend(values ...T) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	// in reverse to keep passed order i.e. ["c","d"] -> Prepend(["a","b"]) -> ["a","b","c",d"]
	for v := len(values) - 1; v >= 0; v-- {
		newElement := &element[T]{value: values[v], next: list.first}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.first.prev = newElement
			list.first = newElement
		}
		list.size++
	}
}

// Get 获取指定索引位置元素的值
func (list *DoublyLinkList[T]) Get(index int) (T, bool) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if !list.withinRange(index) {
		var t T
		return t, false
	}

	e := list.getElement(index)
	return e.value, true
}

// Remove 移除指定索引位置的元素并返回该元素的值
func (list *DoublyLinkList[T]) Remove(index int) (T, bool) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	var value T
	if !list.withinRange(index) {
		return value, false
	}

	e := list.getElement(index)
	value = e.value
	list.deleteElement(e)

	return value, true
}

// Each 遍历操作
func (list *DoublyLinkList[T]) Each(fn func(index int, value T)) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	for i, e := 0, list.first; e != nil; i, e = i+1, e.next {
		fn(i, e.value)
	}
}

// Map 根据指定条件，返回一个新链表
func (list *DoublyLinkList[T]) Map(fn func(index int, value T) T) *DoublyLinkList[T] {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	newList := &DoublyLinkList[T]{}
	for i, e := 0, list.first; e != nil; i, e = i+1, e.next {
		newList.Append(fn(i, e.value))
	}
	return newList
}

// Filter 过滤出指定条件的元素，移除并返回包含的值
func (list *DoublyLinkList[T]) Filter(fn func(index int, value T) bool) []T {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	var elements []*element[T]
	for i, e := 0, list.first; e != nil; i, e = i+1, e.next {
		if fn(i, e.value) {
			elements = append(elements, e)
		}
	}

	values := make([]T, 0, len(elements))
	for _, e := range elements {
		values = append(values, e.value)
		list.deleteElement(e)
	}

	return values
}

// Contains 返回链表中是否包含指定值中的一个
func (list *DoublyLinkList[T]) Contains(values ...T) bool {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if len(values) == 0 {
		return true
	}
	if list.size == 0 {
		return false
	}
	for _, value := range values {
		found := false
		for e := list.first; e != nil; e = e.next {
			if e.value == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Values 返回链表所有元素的值
func (list *DoublyLinkList[T]) Values() []T {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	return list.valuesWithoutLock()
}

// valuesWithoutLock 返回链表所有元素的值(不持有锁)
func (list *DoublyLinkList[T]) valuesWithoutLock() []T {
	values := make([]T, 0, list.size)
	for e := list.first; e != nil; e = e.next {
		values = append(values, e.value)
	}
	return values
}

// IndexOf 返回指定值在链表中的索引，-1表示不存在
func (list *DoublyLinkList[T]) IndexOf(value T) int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	if list.size == 0 {
		return -1
	}
	for index, e := range list.Values() {
		if e == value {
			return index
		}
	}
	return -1
}

// Empty 返回链表是否为空
func (list *DoublyLinkList[T]) Empty() bool {
	return list.size == 0
}

// Size 返回链表的大小
func (list *DoublyLinkList[T]) Size() int {
	return list.size
}

// Clear 清空链表
func (list *DoublyLinkList[T]) Clear() {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	list.clearWithoutLock()
}

// clearWithoutLock 清空链表(不持有锁)
func (list *DoublyLinkList[T]) clearWithoutLock() {
	list.size = 0
	list.first = nil
	list.last = nil
}

// Sort 将链表中元素值排序
func (list *DoublyLinkList[T]) Sort(comparator func(x, y T) int) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if list.size < 2 {
		return
	}

	values := list.valuesWithoutLock()
	slices.SortFunc(values, comparator)

	list.clearWithoutLock()
	list.appendWithoutLock(values...)
}

// Swap 交换链表中两个指定索引的元素
func (list *DoublyLinkList[T]) Swap(i, j int) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if list.withinRange(i) && list.withinRange(j) && i != j {
		var element1, element2 *element[T]
		for e, currentElement := 0, list.first; element1 == nil || element2 == nil; e, currentElement = e+1, currentElement.next {
			switch e {
			case i:
				element1 = currentElement
			case j:
				element2 = currentElement
			}
		}
		element1.value, element2.value = element2.value, element1.value
	}
}

// Insert 在链表的指定索引位置插入值
func (list *DoublyLinkList[T]) Insert(index int, values ...T) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.appendWithoutLock(values...)
		}
		return
	}

	var beforeElement *element[T]
	var foundElement *element[T]
	// determine traversal direction, last to first or first to last
	if list.size-index < index {
		foundElement = list.last
		beforeElement = list.last.prev
		for e := list.size - 1; e != index; e, foundElement = e-1, foundElement.prev {
			beforeElement = beforeElement.prev
		}
	} else {
		foundElement = list.first
		for e := 0; e != index; e, foundElement = e+1, foundElement.next {
			beforeElement = foundElement
		}
	}

	if foundElement == list.first {
		oldNextElement := list.first
		for i, value := range values {
			newElement := &element[T]{value: value}
			if i == 0 {
				list.first = newElement
			} else {
				newElement.prev = beforeElement
				beforeElement.next = newElement
			}
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	} else {
		oldNextElement := beforeElement.next
		for _, value := range values {
			newElement := &element[T]{value: value}
			newElement.prev = beforeElement
			beforeElement.next = newElement
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	}

	list.size += len(values)
}

// Set 设置链表指定索引位置索引的值
func (list *DoublyLinkList[T]) Set(index int, value T) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.appendWithoutLock(value)
		}
		return
	}

	var foundElement *element[T]
	// determine traversal direction, last to first or first to last
	if list.size-index < index {
		foundElement = list.last
		for e := list.size - 1; e != index; {
			fmt.Println("Set last", index, value, foundElement, foundElement.prev)
			e, foundElement = e-1, foundElement.prev
		}
	} else {
		foundElement = list.first
		for e := 0; e != index; {
			e, foundElement = e+1, foundElement.next
		}
	}

	foundElement.value = value
}

// String 实现 Stringer Interface
func (list *DoublyLinkList[T]) String() string {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	str := "DoublyLinkList\n"
	var values []string
	for e := list.first; e != nil; e = e.next {
		values = append(values, fmt.Sprintf("%v", e.value))
	}
	str += strings.Join(values, ", ")
	return str
}

func (list *DoublyLinkList[T]) withinRange(index int) bool {
	return index >= 0 && index < list.size
}

func (list *DoublyLinkList[T]) getElement(index int) *element[T] {
	// determine traveral direction, last to first or first to last
	if list.size-index < index {
		e := list.last
		for i := list.size - 1; i != index; i, e = i-1, e.prev {
		}
		return e
	}

	e := list.first
	for i := 0; i != index; i, e = i+1, e.next {
	}
	return e
}

func (list *DoublyLinkList[T]) deleteElement(e *element[T]) {
	if e == list.first {
		list.first = e.next
	}
	if e == list.last {
		list.last = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	}
	e = nil

	list.size--
}
