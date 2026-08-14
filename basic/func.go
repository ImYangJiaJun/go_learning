package main

import "fmt"

func sumFn(x int, y int) int {
	return x + y
}

// 参数可简写，(x, y int)表示 x,y都是int类型，只能简写前面
func subFn(x, y int) int {
	return x - y
}

// 函数的可变参数，参数名后加 ... 表示参数数量不固定
func sumFn1(x ...int) int { //x是一个切片
	sum := 0
	for _, v := range x {
		sum += v
	}
	return sum
}

func sumFn2(x int, y ...int) int { //表示第一个参数给x，其余给y	注：可变参数只能写在后面
	sum := x
	for _, v := range y {
		sum += v
	}
	return sum
}

// return 可以一次返回多个值
func calc(x, y int) (int, int) {
	return x + y, x - y
}

// 返回值命名：函数定义的时候给返回值命名，并在函数体中直接使用这些变量，最后通过return返回
func calc1(x, y int) (sum, sub int) {
	sum = x + y
	sub = x - y
	return
}

func main() {
	sum1 := sumFn(1, 2)
	fmt.Println(sum1)

	sub1 := subFn(1, 2)
	fmt.Println(sub1)

	sum2 := sumFn1(1, 2, 3, 4)
	fmt.Println(sum2)

	sum3 := sumFn2(1, 2, 3, 4, 5, 6)
	fmt.Println(sum3)

	sum4, sub2 := calc(100, 150)
	fmt.Println(sum4, sub2)

	a, b := calc1(20, 31)
	fmt.Println(a, b)
}
