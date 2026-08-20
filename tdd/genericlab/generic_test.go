package genericlab

import (
	"slices"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		f    func(int) int
		want []int
	}{
		{"平方映射", []int{1, 2, 3}, func(x int) int { return x * x }, []int{1, 4, 9}},
		{"空切片返回非 nil 空切片", []int{}, func(x int) int { return x * x }, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.in, tt.f)
			if got == nil {
				t.Error("got nil, want 非 nil 切片")
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	// R ≠ T：结果类型与元素类型不同，无法与上表共用，单独成例
	t.Run("int 映射成 string", func(t *testing.T) {
		got := Map([]int{1, 2}, func(x int) string { return strconv.Itoa(x * 10) })
		if want := []string{"10", "20"}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestFilter(t *testing.T) {
	isEven := func(x int) bool { return x%2 == 0 }
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"筛出偶数", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"全部不满足", []int{1, 3, 5}, []int{}},
		{"空切片输入", []int{}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.in, isEven)
			if got == nil {
				t.Error("got nil, want 非 nil 空切片")
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("换类型：按长度筛字符串", func(t *testing.T) {
		got := Filter([]string{"go", "java", "c"}, func(s string) bool { return len(s) > 1 })
		if want := []string{"go", "java"}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		init int
		f    func(int, int) int
		want int
	}{
		{"求和", []int{1, 2, 3, 4}, 0, func(acc, x int) int { return acc + x }, 10},
		{"空切片返回 init", []int{}, 5, func(acc, x int) int { return acc + x }, 5},
		{"求积（init 是单位元 1）", []int{2, 3, 4}, 1, func(acc, x int) int { return acc * x }, 24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reduce(tt.in, tt.init, tt.f); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	// R ≠ T：R 由 init 推断，与上表不同类型，单独成例
	t.Run("R≠T：int 切片拼成 string", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3}, "", func(acc string, x int) string { return acc + strconv.Itoa(x) })
		if want := "123"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
