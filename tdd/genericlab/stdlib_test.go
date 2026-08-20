package genericlab

import (
	"maps"
	"reflect"
	"slices"
	"testing"
)

func TestSlicesEqual(t *testing.T) {
	if !slices.Equal([]int{1, 4, 9}, []int{1, 4, 9}) {
		t.Error("内容相同的两个切片应判定相等")
	}
}

func TestSlicesEqualNilFriendly(t *testing.T) {
	if !slices.Equal(nil, []int{}) {
		t.Error("slices.Equal 应把 nil 切片与空切片视为相等")
	}
	// 对照：同样的比较用 reflect.DeepEqual 是 false，这正是行为 1 的 nil 坑
	if reflect.DeepEqual([]int(nil), []int{}) {
		t.Error("预期 DeepEqual 对 nil 与空切片判定不等")
	}
}

func TestSlicesSort(t *testing.T) {
	a := []int{3, 1, 2}
	slices.Sort(a) // 原地排序，不返回新切片
	if want := []int{1, 2, 3}; !slices.Equal(a, want) {
		t.Errorf("got %v, want %v", a, want)
	}
}

func TestSlicesConcat(t *testing.T) {
	a, b, c := []int{1, 2}, []int{3}, []int{4, 5}
	got := slices.Concat(a, b, c)
	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	// 三个入参都不被修改
	if !slices.Equal(a, []int{1, 2}) || !slices.Equal(b, []int{3}) || !slices.Equal(c, []int{4, 5}) {
		t.Errorf("入参被修改: a=%v b=%v c=%v", a, b, c)
	}
}

func TestMapsKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	// maps.Keys 返回迭代器（Go 1.23 起），map 遍历顺序随机，收集后必须排序再断言
	got := slices.Sorted(maps.Keys(m))
	if want := []string{"a", "b"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMapsValues(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	got := slices.Sorted(maps.Values(m))
	if want := []int{1, 2}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
