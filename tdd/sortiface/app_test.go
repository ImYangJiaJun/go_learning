package sortiface

import (
	"cmp"
	"slices"
	"sort"
	"testing"
)

func TestSortByDownloads(t *testing.T) {
	apps := []App{{3, 500}, {1, 1200}, {2, 800}}
	want := []App{{1, 1200}, {2, 800}, {3, 500}}

	sort.Sort(ByDownloads(apps))
	for i := range apps {
		if apps[i] != want[i] {
			t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
		}
	}
}

func TestReverseAndByID(t *testing.T) {
	t.Run("Reverse 得下载量升序", func(t *testing.T) {
		apps := []App{{1, 1200}, {3, 500}, {2, 800}}
		want := []App{{3, 500}, {2, 800}, {1, 1200}}
		sort.Sort(sort.Reverse(ByDownloads(apps)))
		for i := range apps {
			if apps[i] != want[i] {
				t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
			}
		}
	})

	t.Run("ByID 升序", func(t *testing.T) {
		apps := []App{{3, 500}, {1, 1200}, {2, 800}}
		want := []App{{1, 1200}, {2, 800}, {3, 500}}
		sort.Sort(ByID(apps))
		for i := range apps {
			if apps[i] != want[i] {
				t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
			}
		}
	})

	t.Run("Reverse(ByID) 得降序", func(t *testing.T) {
		apps := []App{{3, 500}, {1, 1200}, {2, 800}}
		want := []App{{3, 500}, {2, 800}, {1, 1200}}
		sort.Sort(sort.Reverse(ByID(apps)))
		for i := range apps {
			if apps[i] != want[i] {
				t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
			}
		}
	})
}

func TestStableKeepsOriginalOrder(t *testing.T) {
	apps := []App{{1, 100}, {2, 300}, {3, 100}}
	want := []App{{2, 300}, {1, 100}, {3, 100}}
	sort.Stable(ByDownloads(apps))
	for i := range apps {
		if apps[i] != want[i] {
			t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
		}
	}
}

func TestModernSort(t *testing.T) {
	t.Run("sort.Slice 降序", func(t *testing.T) {
		apps := []App{{3, 500}, {1, 1200}, {2, 800}}
		want := []App{{1, 1200}, {2, 800}, {3, 500}}
		// sort.Slice（Go 1.8）：免定义命名类型，规则写进匿名闭包。
		// 注意闭包参数是下标 i、j 不是元素本身；内部靠反射操作切片，比接口写法慢。
		sort.Slice(apps, func(i, j int) bool {
			return apps[i].Downloads > apps[j].Downloads
		})
		for i := range apps {
			if apps[i] != want[i] {
				t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
			}
		}
	})

	t.Run("已序校验", func(t *testing.T) {
		apps := []App{{1, 1200}, {2, 800}, {3, 500}}
		// SliceIsSorted 只校验不排序：用同一个 less 闭包问"是否已经有序"。
		less := func(i, j int) bool {
			return apps[i].Downloads > apps[j].Downloads
		}
		if !sort.SliceIsSorted(apps, less) {
			t.Errorf("expected apps to be sorted by downloads desc, got %+v", apps)
		}
	})

	t.Run("slices.SortFunc 降序", func(t *testing.T) {
		apps := []App{{3, 500}, {1, 1200}, {2, 800}}
		want := []App{{1, 1200}, {2, 800}, {3, 500}}
		// slices 包（Go 1.21，泛型）：比较函数拿元素本身，返回三态 int——
		// 负数 = a 在前，0 = 相等，正数 = b 在前。写降序只需交换实参。
		// 泛型编译期实例化，零反射开销，是新代码的首选写法。
		slices.SortFunc(apps, func(a, b App) int {
			return cmp.Compare(b.Downloads, a.Downloads)
		})
		for i := range apps {
			if apps[i] != want[i] {
				t.Errorf("apps[%d]: got %+v, want %+v", i, apps[i], want[i])
			}
		}
	})

	t.Run("slices.Sort 升序", func(t *testing.T) {
		nums := []int{3, 1, 2}
		want := []int{1, 2, 3}
		// 元素满足 cmp.Ordered（int、string 等内置有序类型）时的最简写法：连比较函数都不用给。
		slices.Sort(nums)
		for i := range nums {
			if nums[i] != want[i] {
				t.Errorf("nums[%d]: got %d, want %d", i, nums[i], want[i])
			}
		}
	})
}
