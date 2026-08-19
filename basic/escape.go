package main

/*
逃逸分析观察实验（教程 ch13）

运行方式（编译期日志输出在 stderr，程序的 fmt 输出在 stdout）：

	go run -gcflags="-m -l" basic/escape.go

	-gcflags="-m"  打印逃逸分析结果
	-gcflags="-l"  禁止函数内联（排除内联对逃逸判断的干扰）

观察任务：对照下面 5 个案例的注释，在编译输出里找到对应的行，
确认哪些变量 "moved to heap"（逃逸到堆），哪些 "does not escape"（留在栈上）。
*/
import "fmt"

// 案例 1：返回局部变量的指针 → x 在函数返回后仍被引用，逃逸到堆
// 预期输出：moved to heap: x
func escapeReturnPtr() *int {
	x := 42
	return &x
}

// 案例 2：局部变量只在函数内使用 → 不逃逸，分配在栈上，函数返回即释放
// 预期输出：does not escape
func noEscape() int {
	x := 42
	return x
}

// 案例 3：interface 装箱 → 实参被包进 any（interface{}）时逃逸
// fmt.Println 的参数是 ...any，n 装箱后逃逸到堆
// 预期输出：n escapes to heap
func escapeInterface() {
	n := 42
	fmt.Println("装箱逃逸:", n)
}

// 案例 4：闭包捕获外部变量 → 被捕获的变量逃逸
// 闭包在 escapeClosure 返回后仍存活，n 必须比函数栈帧活得久
// 预期输出：moved to heap: n
func escapeClosure() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// 案例 5：动态大小的切片 → 编译期无法确定大小，分配在堆上
// 预期输出：make([]int, n) escapes to heap
func escapeDynamicSize(n int) []int {
	return make([]int, n)
}

func main() {
	p := escapeReturnPtr()
	fmt.Println("案例1 返回指针:", *p)

	fmt.Println("案例2 栈上分配:", noEscape())

	escapeInterface()

	next := escapeClosure()
	next()
	fmt.Println("案例4 闭包计数:", next())

	fmt.Println("案例5 动态切片长度:", len(escapeDynamicSize(10)))
}
