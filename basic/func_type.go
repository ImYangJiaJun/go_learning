package main

import "fmt"

type calcType func(int, int) int //表示定义一个calc的函数类型

type myInt int //自定义类型

func add(x, y int) int {
	return x + y
}

func sub(x, y int) int {
	return x - y
}

// 使用函数类型作为函数参数
func calcT(x, y int, cb calcType) int {
	return cb(x, y)
}

// 使用函数作为返回值
func do(o string) calcType {
	switch o {
	case "+":
		return add
	case "-":
		return sub
	case "*":
		return func(c1, c2 int) int {
			return c1 * c2
		}
	case "/":
		return func(c1, c2 int) int {
			return c1 / c2
		}
	default:
		return nil

	}
}

// 递归函数
func sumAll(n int) int {
	if n > 1 {
		return n + sumAll(n-1)
	}
	return n
}
func showAll(n int) {
	if n > 0 {
		fmt.Printf("%d ", n)
		showAll(n - 1)
	}
}

/*
闭包
1.可以让一个变量常驻内存
2.可以让一个变量不污染全局

闭包是指有权访问另一个函数作用域中的变量的函数
创建闭包的常见方式是在一个函数内部创建另一个函数，通过另一个函数访问这个函数这个函数的局部变量

注：由于闭包里作用域返回的局部变量资源不会被立即销毁回收，所以可能占用更多的内存，过度使用闭包会导致性能下降，建议在非常有必要的时候才使用
*/

// 闭包写法：函数里面嵌套一个函数，最后返回里面的函数
func adder1() func() int {
	var i = 10
	return func() int {
		i += 1
		return i
	}
}
func adder2() func(y int) int {
	var i = 10
	return func(y int) int {
		i += y
		return i
	}
}

func main() {
	var c calcType
	c = add                          //赋值的函数必须满足基本结构一致
	fmt.Printf("Type of c: %T\n", c) //main.calcType
	fmt.Println(c(1, 2))

	f := sub
	fmt.Printf("Type of f: %T\n", f) //func(int, int) int
	fmt.Println(f(1, 2))

	var a myInt = 1
	var b int = 2
	fmt.Printf("Type of a: %T\tType of b: %T\n", a, b) //a与b不是同一类型，不能直接相加
	fmt.Println(int(a) + b)                            //转换类型才能计算

	sum := calcT(5, 2, add)
	fmt.Println(sum)

	j := calcT(5, 2, func(x, y int) int {
		return x * y
	})
	fmt.Println(j)

	var m = do("+") //结果是把m赋值成为一个方法
	fmt.Println(m(2, 6))

	//匿名函数 匿名自执行函数
	func() {
		fmt.Println("匿名自执行函数")
	}() //把前面整体当作一个方法进行调用

	fun := func() {
		fmt.Println("匿名函数赋值给变量进行调用")
	}
	fun()

	//匿名函数可以正常接受和返回
	fun1 := func(x, y int) int {
		return x + y
	}
	fmt.Println(fun1(3, 2))

	//n以内正整数的和
	fmt.Println(sumAll(100))
	//打印小于n的所有正整数
	showAll(10)

	//闭包
	fmt.Printf("\n------------------------\n")
	var bb1 = adder1() //注意，有()是执行方法，没有()是把方法赋值给变量
	fmt.Println(bb1())
	fmt.Println(bb1())
	fmt.Println(bb1())
	fmt.Println("---")
	var bb2 = adder2()   //创建变量i常驻内存，但是不会污染全局
	fmt.Println(bb2(10)) //20
	fmt.Println(bb2(10)) //30
	fmt.Println(bb2(10)) //40
	fmt.Println(bb2(10)) //50
	//输出的是同一个变量，是在 adder2() 中创建的变量

}
