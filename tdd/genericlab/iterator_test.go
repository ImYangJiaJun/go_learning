package genericlab

import (
	"slices"
	"testing"
)

func TestAll(t *testing.T) {
	t.Run("完整遍历", func(t *testing.T) {
		var got []int
		for v := range All([]int{1, 2, 3}) {
			got = append(got, v)
		}
		if want := []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("空切片", func(t *testing.T) {
		yieldCount := 0
		for range All([]int{}) {
			yieldCount++
		}
		if yieldCount != 0 {
			t.Errorf("yield 被调用 %d 次, want 0", yieldCount)
		}
	})

	t.Run("换类型：string", func(t *testing.T) {
		var got []string
		for v := range All([]string{"a", "b"}) {
			got = append(got, v)
		}
		if want := []string{"a", "b"}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	/*
		验证协议的关键用例：break 翻译成 yield 返回 false，
		迭代器必须立刻停止，不得多产出
	*/
	t.Run("提前 break", func(t *testing.T) {
		count := 0
		for range All([]int{1, 2, 3, 4, 5}) {
			count++
			if count == 2 {
				break
			}
		}
		if count != 2 {
			t.Errorf("循环体执行 %d 次, want 恰好 2 次", count)
		}
	})

	// All 的返回类型就是 iter.Seq[T]，与标准库协议互通
	t.Run("对接标准库", func(t *testing.T) {
		got := slices.Collect(All([]int{1, 2, 3}))
		if want := []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
