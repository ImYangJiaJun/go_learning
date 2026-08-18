package main

import "fmt"

func f1() {
	fmt.Println("f1 begin")

	defer func() {
		fmt.Println("f1 defer func begin")
		fmt.Println("f1 defer func end")
	}()

	fmt.Println("f1 end")
}

// 匿名返回值会返回defer修改前的值
func f2() int {
	var a int
	defer func() {
		a++
	}()
	return a
}

// 命名返回值返回命名的变量时会返回defer语句修改后的值
func f3() (a int) {
	defer func() {
		a++
	}()
	return
}

// 命名返回值返回没有命名的变量时会返回defer语句修改前的值
func f4() (b int) {
	a := 0
	defer func() {
		a++
	}()
	return
}
func f5() (x int) {
	defer func(y int) {
		y++ //修改不到f5作用域的x
	}(x)
	return
}

// 注：defer注册要执行的函数的时候，函数所有的值都要确定
func f6(index string, a, b int) int {
	ret := a + b
	fmt.Println(index, a, b, ret)
	return ret
}

func main() {
	//defer后面的语句会延迟处理，在defer归属的函数即将返回的时候将所有带defer关键字的语句按照定义顺序从后往前执行
	//执行顺序-相当于把defer的语句放进栈，先进后出
	fmt.Println("-------------main begin-----------------")

	fmt.Println("开始")
	defer fmt.Println(1)
	fmt.Println(2)
	defer fmt.Println(3)
	fmt.Println("结束")

	fmt.Println("----------------------------------------")

	f1()

	fmt.Println("----------------------------------------")

	fmt.Println(f2())

	fmt.Println("----------------------------------------")

	fmt.Println(f3())

	fmt.Println("----------------------------------------")

	fmt.Println(f4())

	fmt.Println("----------------------------------------")

	fmt.Println(f5())

	fmt.Println("----------------------------------------")

	x := 1
	y := 2
	//注：defer注册要执行的函数的时候，函数所有的值都要确定
	defer f6("AA", x, f6("A", x, y)) //为了确定函数参数的值，会执行f6("A", x, y)
	x = 10
	defer f6("BB", x, f6("B", x, y)) //注册的时候就确定了值，所以下面的y=20不会影响
	y = 20

	fmt.Println("-----------main end---------------------")
}
