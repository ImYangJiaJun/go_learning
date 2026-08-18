package main

import (
	"cmp"
	"fmt"
)

func main() {
	// 泛型函数：用类型参数处理多种类型
	genericFunc()

	// 自定义约束：用接口限制可用的类型
	genericConstraint()

	// 泛型类型：定义可以装任意类型的容器
	genericType()
}

// 泛型函数：T 是类型参数，调用时由实参自动推断类型
// comparable 是内置约束，表示支持 == 和 != 比较的类型
func indexOf[T comparable](slice []T, target T) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

func genericFunc() {
	// 找出切片中第一次出现目标值的下标，找不到返回 -1
	// 类型自动推断，也可以显式写成 indexOf[int](...)
	fmt.Println(indexOf([]int{1, 2, 3}, 2))            // 输出 1
	fmt.Println(indexOf([]string{"a", "b", "c"}, "c")) // 输出 2
}

// 自定义约束：接口中用 | 组合多个类型，~ 表示底层类型匹配
type Number interface {
	~int | ~int64 | ~float64
}

// 对比：不写 ~ 的约束，只匹配 int/int64/float64 这三个精确类型
type StrictNumber interface {
	int | int64 | float64
}

// MyintGen 底层类型是 int 的自定义类型
type MyintGen int

// sum 对任意 Number 类型的切片求和
func sum[T Number](slice []T) T {
	var total T // T 的零值，数值类型为 0
	for _, v := range slice {
		total += v
	}
	return total
}

// strictSum 与 sum 相同，但使用不带 ~ 的约束
func strictSum[T StrictNumber](slice []T) T {
	var total T
	for _, v := range slice {
		total += v
	}
	return total
}

func genericConstraint() {
	fmt.Println(sum([]int{1, 2, 3}))      // 输出 6
	fmt.Println(sum([]float64{1.5, 2.5})) // 输出 4

	// ~ 的区别：Number 带 ~，自定义底层类型 MyintGen 也能通过约束
	fmt.Println(sum([]MyintGen{1, 2, 3})) // 输出 6

	// StrictNumber 不带 ~，只接受精确类型 int
	fmt.Println(strictSum([]int{1, 2, 3})) // 输出 6
	// 下面这行会编译报错：MyintGen does not satisfy StrictNumber
	// fmt.Println(strictSum([]MyintGen{1, 2, 3}))

	// 也可以直接使用标准库的约束，如 cmp.Ordered（内部同样用 ~ 定义）
	fmt.Println(maxValue([]int{3, 9, 5}))          // 输出 9
	fmt.Println(maxValue([]string{"a", "c", "b"})) // 输出 c
}

// maxValue 使用 cmp.Ordered 约束，支持所有可比较大小的类型
func maxValue[T cmp.Ordered](slice []T) T {
	max0 := slice[0]
	for _, v := range slice[1:] {
		if v > max0 {
			max0 = v
		}
	}
	return max0
}

// 泛型类型：栈，可以存放任意类型的元素
type Stack[T any] struct {
	items []T
}

// 泛型类型的方法也要带上类型参数 T
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // 空栈时返回 T 的零值
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func genericType() {
	// int 类型的栈
	intStack := Stack[int]{}
	intStack.Push(1)
	intStack.Push(2)
	if v, ok := intStack.Pop(); ok {
		fmt.Println("弹出：", v) // 输出 2
	}

	// string 类型的栈
	strStack := Stack[string]{}
	strStack.Push("hello")
	if v, ok := strStack.Pop(); ok {
		fmt.Println("弹出：", v) // 输出 hello
	}
}
