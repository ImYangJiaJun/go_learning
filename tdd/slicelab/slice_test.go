package slicelab

import (
	"slices"
	"testing"
)

func TestClassicSharedArray(t *testing.T) {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	s1 := slice[2:5]
	s2 := s1[2:6:7]
	if !slices.Equal(s1, []int{2, 3, 4}) {
		t.Errorf("s1 result error: %v", s1)
	}
	if !slices.Equal(s2, []int{4, 5, 6, 7}) {
		t.Errorf("s2 result error: %v", s2)
	}

	s2 = append(s2, 100)
	if !slices.Equal(s2, []int{4, 5, 6, 7, 100}) {
		t.Errorf("s2 result error: %v", s2)
	}
	if !slices.Equal(slice, []int{0, 1, 2, 3, 4, 5, 6, 7, 100, 9}) {
		t.Errorf("slice result error: %v", slice)
	}

	s2 = append(s2, 200)
	if !slices.Equal(s2, []int{4, 5, 6, 7, 100, 200}) {
		t.Errorf("s2 result error: %v", s2)
	}
	if slices.Contains(slice, 200) {
		t.Errorf("slice result error: %v", slice)
	}

	s1[2] = 20
	if !slices.Equal(s1, []int{2, 3, 20}) {
		t.Errorf("s1 result error: %v", s1)
	}
	if !(s2[0] == 4) {
		t.Errorf("s2 result error: %v", s2)
	}
	if !slices.Equal(slice, []int{0, 1, 2, 3, 20, 5, 6, 7, 100, 9}) {
		t.Errorf("slice result error: %v", slice)
	}
}

// 行为 2：三段切片 s[low:high:max]——len 和 cap 各是多少
func TestSliceExpressions(t *testing.T) {
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	cases := []struct {
		name    string
		got     []int
		want    []int
		wantLen int
		wantCap int
	}{
		{"两段切片 s[2:5]", s[2:5], []int{2, 3, 4}, 3, 8},
		{"三段切片 s[2:5:7]", s[2:5:7], []int{2, 3, 4}, 3, 5},
		{"省略 low s[:3]", s[:3], []int{0, 1, 2}, 3, 10},
		{"省略 high s[5:]", s[5:], []int{5, 6, 7, 8, 9}, 5, 5},
		{"空切片且锁死容量 s[0:0:0]", s[0:0:0], []int{}, 0, 0},
		{"空切片但留容量 s[2:2:5]", s[2:2:5], []int{}, 0, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if len(c.got) != c.wantLen {
				t.Errorf("len = %d, want %d", len(c.got), c.wantLen)
			}
			if cap(c.got) != c.wantCap {
				t.Errorf("cap = %d, want %d", cap(c.got), c.wantCap)
			}
			if !slices.Equal(c.got, c.want) {
				t.Errorf("内容 = %v, want %v", c.got, c.want)
			}
		})
	}

	t.Run("非法表达式 s[2:7:5] 运行期 panic", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("期望 panic，但没有")
			}
		}()
		// 下标用变量：常量下标的非法三段式是编译期错误，轮不到运行期 panic；
		// 值从 len(s) 推算，避免编译器和 IDE 常量折叠后提前拦截。即 s[2:7:5]，high > maxIdx
		low, high, maxIdx := 2, len(s)-3, len(s)-5
		_ = s[low:high:maxIdx]
	})
}

// 行为 3：共享底层数组互污 vs append 扩容脱钩（对照实验）
// 每个用例都重建 base，否则前一个用例的写入会残留——测试本身就互污了。
func TestSharedArrayVsDetach(t *testing.T) {
	t.Run("A 改元素互污", func(t *testing.T) {
		base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		s1 := base[0:3]
		s1[0] = 99
		if base[0] != 99 {
			t.Errorf("base[0] = %d, want 99", base[0])
		}
	})

	t.Run("B cap 未满 append 互污", func(t *testing.T) {
		base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		s2 := base[0:3]
		s2 = append(s2, 88)
		if base[3] != 88 {
			t.Errorf("base[3] = %d, want 88", base[3])
		}
		if len(base) != 10 {
			t.Errorf("len(base) = %d, want 10", len(base))
		}
	})

	t.Run("C cap 已满 append 脱钩", func(t *testing.T) {
		base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		s3 := base[:]
		s3 = append(s3, 77)
		if slices.Contains(base, 77) {
			t.Errorf("base 不应包含 77: %v", base)
		}
	})
}

// 行为 4：函数内 append 为什么"失效"——slice 头按值传递
func TestGrowNoReturn(t *testing.T) {
	t.Run("调用方 len 不变", func(t *testing.T) {
		s := []int{1, 2, 3}
		growNoReturn(s, 99)
		if len(s) != 3 {
			t.Errorf("len(s) = %d, want 3", len(s))
		}
	})

	t.Run("长度不可见、元素可见", func(t *testing.T) {
		s := make([]int, 2, 4)
		s[0], s[1] = 1, 2
		growNoReturn(s, 99)
		if len(s) != 2 {
			t.Errorf("len(s) = %d, want 2", len(s))
		}
		if s[:cap(s)][2] != 99 {
			t.Errorf("s[:cap(s)][2] = %d, want 99", s[:cap(s)][2])
		}
	})
}

func TestGrowReturned(t *testing.T) {
	s := make([]int, 2, 4)
	s[0], s[1] = 1, 2
	s2 := growReturned(s, 99)
	if len(s2) != 3 {
		t.Errorf("len(s2) = %d, want 3", len(s2))
	}
	if s2[2] != 99 {
		t.Errorf("s2[2] = %d, want 99", s2[2])
	}
	if len(s) != 2 {
		t.Errorf("len(s) = %d, want 2", len(s))
	}
}

// 行为 5：nil 是有效的 slice
func TestNilSlice(t *testing.T) {
	var s []int
	if s != nil {
		t.Error("var s []int 应为 nil")
	}
	if len(s) != 0 {
		t.Errorf("len(s) = %d, want 0", len(s))
	}
	if cap(s) != 0 {
		t.Errorf("cap(s) = %d, want 0", cap(s))
	}

	s = append(s, 1)
	if !slices.Equal(s, []int{1}) {
		t.Errorf("nil append 后 s = %v, want [1]", s)
	}

	var r []int
	count := 0
	for range r {
		count++
	}
	if count != 0 {
		t.Errorf("for range nil 执行了 %d 次, want 0", count)
	}

	f := func() []int { return nil }
	if len(f()) != 0 {
		t.Errorf("len(f()) = %d, want 0", len(f()))
	}
}
