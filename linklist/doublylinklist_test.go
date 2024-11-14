package linklist

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	list1 := New[int]()

	if actualValue := list1.Empty(); actualValue != true {
		t.Errorf("Got %v expected %v", actualValue, true)
	}

	list2 := New[int](1, 2)

	if actualValue := list2.Size(); actualValue != 2 {
		t.Errorf("Got %v expected %v", actualValue, 2)
	}

	if actualValue, ok := list2.Get(0); actualValue != 1 || !ok {
		t.Errorf("Got %v expected %v", actualValue, 1)
	}

	if actualValue, ok := list2.Get(1); actualValue != 2 || !ok {
		t.Errorf("Got %v expected %v", actualValue, 2)
	}

	if actualValue, ok := list2.Get(2); actualValue != 0 || ok {
		t.Errorf("Got %v expected %v", actualValue, 0)
	}
}

func TestAdd(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	if actualValue := list.Empty(); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
	if actualValue := list.Size(); actualValue != 3 {
		t.Errorf("Got %v expected %v", actualValue, 3)
	}
	if actualValue, ok := list.Get(2); actualValue != "c" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "c")
	}
}

func TestAppendAndPrepend(t *testing.T) {
	list := New[string]()
	list.Append("b")
	list.Prepend("a")
	list.Append("c")
	if actualValue := list.Empty(); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
	if actualValue := list.Size(); actualValue != 3 {
		t.Errorf("Got %v expected %v", actualValue, 3)
	}
	if actualValue, ok := list.Get(0); actualValue != "a" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "c")
	}
	if actualValue, ok := list.Get(1); actualValue != "b" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "c")
	}
	if actualValue, ok := list.Get(2); actualValue != "c" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "c")
	}
}

func TestGet(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	if actualValue, ok := list.Get(0); actualValue != "a" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "a")
	}
	if actualValue, ok := list.Get(1); actualValue != "b" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "b")
	}
	if actualValue, ok := list.Get(2); actualValue != "c" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "c")
	}
	if actualValue, ok := list.Get(3); actualValue != "" || ok {
		t.Errorf("Got %v expected %v", actualValue, "")
	}
	list.Remove(0)
	if actualValue, ok := list.Get(0); actualValue != "b" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "b")
	}
}

func TestRemove(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	list.Remove(2)
	if actualValue, ok := list.Get(2); actualValue != "" || ok {
		t.Errorf("Got %v expected %v", actualValue, "")
	}
	list.Remove(1)
	list.Remove(0)
	list.Remove(0) // no effect
	if actualValue := list.Empty(); actualValue != true {
		t.Errorf("Got %v expected %v", actualValue, true)
	}
	if actualValue := list.Size(); actualValue != 0 {
		t.Errorf("Got %v expected %v", actualValue, 0)
	}
}

func TestPop(t *testing.T) {
	list := New[string]()
	list.Append("a", "b", "c")
	v, ok := list.Remove(0)
	assert.True(t, ok)
	assert.Equal(t, "a", v)
	assert.Equal(t, 2, list.Size())
	v, ok = list.Remove(0)
	assert.True(t, ok)
	assert.Equal(t, "b", v)
	assert.Equal(t, 1, list.Size())
	v, ok = list.Remove(0)
	assert.True(t, ok)
	assert.Equal(t, "c", v)
	assert.Equal(t, 0, list.Size())
}

func TestEach(t *testing.T) {
	list := New[string]()
	list.Append("a", "b", "c")
	list.Each(func(index int, value string) {
		switch index {
		case 0:
			if actualValue, expectedValue := value, "a"; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		case 1:
			if actualValue, expectedValue := value, "b"; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		case 2:
			if actualValue, expectedValue := value, "c"; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			t.Errorf("Too many")
		}
	})
}

func TestMap(t *testing.T) {
	list := New[string]()
	list.Append("a", "b", "c")
	mappedList := list.Map(func(index int, value string) string {
		return "mapped: " + value
	})
	if actualValue, _ := mappedList.Get(0); actualValue != "mapped: a" {
		t.Errorf("Got %v expected %v", actualValue, "mapped: a")
	}
	if actualValue, _ := mappedList.Get(1); actualValue != "mapped: b" {
		t.Errorf("Got %v expected %v", actualValue, "mapped: b")
	}
	if actualValue, _ := mappedList.Get(2); actualValue != "mapped: c" {
		t.Errorf("Got %v expected %v", actualValue, "mapped: c")
	}
	if mappedList.Size() != 3 {
		t.Errorf("Got %v expected %v", mappedList.Size(), 3)
	}
}

func TestFilter(t *testing.T) {
	list := New[string]()
	list.Append("a", "b", "c", "d", "e")
	values := list.Filter(func(index int, value string) bool {
		if value == "b" || value == "d" {
			return true
		}
		return false
	})
	assert.Equal(t, []string{"b", "d"}, values)
	assert.Equal(t, 3, list.Size())
	assert.Equal(t, []string{"a", "c", "e"}, list.Values())
}

func TestSwap(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	list.Swap(0, 1)
	if actualValue, ok := list.Get(0); actualValue != "b" || !ok {
		t.Errorf("Got %v expected %v", actualValue, "b")
	}
}

func TestSort(t *testing.T) {
	list := New[string]()
	list.Sort(cmp.Compare[string])
	list.Append("e", "f", "g", "a", "b", "c", "d")
	list.Sort(cmp.Compare[string])
	for i := 1; i < list.Size(); i++ {
		a, _ := list.Get(i - 1)
		b, _ := list.Get(i)
		if a > b {
			t.Errorf("Not sorted! %s > %s", a, b)
		}
	}
}

func TestClear(t *testing.T) {
	list := New[string]()
	list.Append("e", "f", "g", "a", "b", "c", "d")
	list.Clear()
	if actualValue := list.Empty(); actualValue != true {
		t.Errorf("Got %v expected %v", actualValue, true)
	}
	if actualValue := list.Size(); actualValue != 0 {
		t.Errorf("Got %v expected %v", actualValue, 0)
	}
}

func TestContains(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	if actualValue := list.Contains("a"); actualValue != true {
		t.Errorf("Got %v expected %v", actualValue, true)
	}
	if actualValue := list.Contains(""); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
	if actualValue := list.Contains("a", "b", "c"); actualValue != true {
		t.Errorf("Got %v expected %v", actualValue, true)
	}
	if actualValue := list.Contains("a", "b", "c", "d"); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
	list.Clear()
	if actualValue := list.Contains("a"); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
	if actualValue := list.Contains("a", "b", "c"); actualValue != false {
		t.Errorf("Got %v expected %v", actualValue, false)
	}
}

func TestValues(t *testing.T) {
	list := New[string]()
	list.Append("a")
	list.Append("b", "c")
	if actualValue, expectedValue := list.Values(), []string{"a", "b", "c"}; !slices.Equal(actualValue, expectedValue) {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestInsert(t *testing.T) {
	list := New[string]()
	list.Insert(0, "b", "c", "d")
	list.Insert(0, "a")
	list.Insert(10, "x") // ignore
	if actualValue := list.Size(); actualValue != 4 {
		t.Errorf("Got %v expected %v", actualValue, 4)
	}
	list.Insert(4, "g") // append
	if actualValue := list.Size(); actualValue != 5 {
		t.Errorf("Got %v expected %v", actualValue, 5)
	}
	if actualValue, expectedValue := strings.Join(list.Values(), ""), "abcdg"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
	list.Insert(4, "e", "f") // last to first traversal
	if actualValue := list.Size(); actualValue != 7 {
		t.Errorf("Got %v expected %v", actualValue, 7)
	}
	if actualValue, expectedValue := strings.Join(list.Values(), ""), "abcdefg"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestSet(t *testing.T) {
	list := New[string]()
	list.Set(0, "a")
	list.Set(1, "b")
	if actualValue := list.Size(); actualValue != 2 {
		t.Errorf("Got %v expected %v", actualValue, 2)
	}
	list.Set(2, "c") // append
	if actualValue := list.Size(); actualValue != 3 {
		t.Errorf("Got %v expected %v", actualValue, 3)
	}
	list.Set(4, "d")  // ignore
	list.Set(1, "bb") // update
	if actualValue := list.Size(); actualValue != 3 {
		t.Errorf("Got %v expected %v", actualValue, 3)
	}
	if actualValue, expectedValue := list.Values(), []string{"a", "bb", "c"}; !slices.Equal(actualValue, expectedValue) {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestString(t *testing.T) {
	c := New[int]()
	c.Append(1)
	if !strings.HasPrefix(c.String(), "DoublyLinkedList") {
		t.Errorf("String should start with container name")
	}
}

func benchmarkGet(b *testing.B, list *DoublyLinkList[int], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			list.Get(n)
		}
	}
}

func benchmarkAdd(b *testing.B, list *DoublyLinkList[int], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			list.Append(n)
		}
	}
}

func benchmarkRemove(b *testing.B, list *DoublyLinkList[int], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			list.Remove(n)
		}
	}
}

func BenchmarkDoublyLinkedListGet100(b *testing.B) {
	b.StopTimer()
	size := 100
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkGet(b, list, size)
}

func BenchmarkDoublyLinkedListGet1000(b *testing.B) {
	b.StopTimer()
	size := 1000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkGet(b, list, size)
}

func BenchmarkDoublyLinkedListGet10000(b *testing.B) {
	b.StopTimer()
	size := 10000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkGet(b, list, size)
}

func BenchmarkDoublyLinkedListGet100000(b *testing.B) {
	b.StopTimer()
	size := 100000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkGet(b, list, size)
}

func BenchmarkDoublyLinkedListAdd100(b *testing.B) {
	b.StopTimer()
	size := 100
	list := New[int]()
	b.StartTimer()
	benchmarkAdd(b, list, size)
}

func BenchmarkDoublyLinkedListAdd1000(b *testing.B) {
	b.StopTimer()
	size := 1000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkAdd(b, list, size)
}

func BenchmarkDoublyLinkedListAdd10000(b *testing.B) {
	b.StopTimer()
	size := 10000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkAdd(b, list, size)
}

func BenchmarkDoublyLinkedListAdd100000(b *testing.B) {
	b.StopTimer()
	size := 100000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkAdd(b, list, size)
}

func BenchmarkDoublyLinkedListRemove100(b *testing.B) {
	b.StopTimer()
	size := 100
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkRemove(b, list, size)
}

func BenchmarkDoublyLinkedListRemove1000(b *testing.B) {
	b.StopTimer()
	size := 1000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkRemove(b, list, size)
}

func BenchmarkDoublyLinkedListRemove10000(b *testing.B) {
	b.StopTimer()
	size := 10000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkRemove(b, list, size)
}

func BenchmarkDoublyLinkedListRemove100000(b *testing.B) {
	b.StopTimer()
	size := 100000
	list := New[int]()
	for n := 0; n < size; n++ {
		list.Append(n)
	}
	b.StartTimer()
	benchmarkRemove(b, list, size)
}
