package main

import "fmt"

func main() {
	//变量命名规则同其它编程语言，由字母、数字、下划线组成，不能以数字开头，不能是关键字
	//同一个作用域不能重复定义
	//=======变量定义方式=======
	//var 变量名 (类型) = 表达式
	var a string = "a"
	fmt.Println(a)

	/*
		一次定义多个变量
		var 变量名称,变量名称 类型 （变量都限定最后的类型）

		var (
			变量名称 类型
			变量名称 类型
		)
	*/

	var c1, c2 string
	c1 = "c1"
	c2 = "c2"
	fmt.Println(c1, c2)

	var (
		d1 string = "d1"
		d2 int    = 12
	)
	fmt.Printf("d1 -> %v  %T\nd2 -> %v  %T", d1, d1, d2, d2)

	//短变量声明-只能在当前作用域生效（局部变量）
	//变量名 := 表达式
	e := "E"
	fmt.Println(e)

	//匿名变量-用于忽略函数返回的多个值中的值
	//写法   _
	//注：匿名变量不占用命名空间和内存，不存在重复声明的问题

	var x, y, _ = getData()
	fmt.Println(x, y)

}
func getData() (string, int, float64) {
	return "张三", 10, 3.14
}
