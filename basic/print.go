package main

import "fmt"

func main() {
	fmt.Println("==========================Println=============================")

	fmt.Println("默认换行")
	fmt.Println("输出", "多个", "中间", "有", "空格")

	fmt.Println("==========================Print=============================")

	fmt.Print("默认不换行")
	fmt.Print("输出", "多个", "中间", "没有", "空格")

	fmt.Println("")
	fmt.Println("==========================Printf=============================")

	fmt.Printf("默认不换行")
	var a = "格式化"
	var b = "输出"
	var c = "类似C中的printf"
	fmt.Printf(" %v%v \n ,%v\n", a, b, c)

	/*
		%v 输出值
		%+v 结构体：输出字段名+值；字符串：效果同%v
		%#v 输出字面量表示
		%T 输出变量类型
	*/

	fmt.Println("===字符串")
	var test string = "测试字符串"
	fmt.Printf("%%v -> %v\n", test)
	fmt.Printf("%%+v -> %+v\n", test)
	fmt.Printf("%%#v -> %#v\n", test)
	fmt.Printf("%%T -> %T\n", test)

	fmt.Println("===结构体")
	type Person struct {
		name string
		age  int
	}
	var p1 = Person{
		name: "张三",
		age:  23,
	}
	fmt.Printf("%%v -> %v\n", p1)
	fmt.Printf("%%+v -> %+v\n", p1)
	fmt.Printf("%%#v -> %#v\n", p1)
	fmt.Printf("%%T -> %T\n", p1)

	fmt.Println("===整数")
	var i int = 10
	fmt.Printf("%%v -> %v\n", i)
	fmt.Printf("%%+v -> %+v\n", i)
	fmt.Printf("%%#v -> %#v\n", i)
	fmt.Printf("%%T -> %T\n", i)

	fmt.Println("===浮点数")
	var f float64 = 12.34567
	fmt.Printf("%%f -> %f\n", f)
	fmt.Printf("%%.2f -> %.2f\n", f)
	fmt.Printf("%%.0f -> %.0f\n", f)
	fmt.Printf("%%.10f -> %.10f\n", f)

	fmt.Printf("%%.3e -> %.3e\n", f)
	fmt.Printf("%%.3g -> %.3g\n", f)

}
