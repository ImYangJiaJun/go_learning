package main

import "fmt"

func main() {
	//！！！注：
	//1.if中的语句只有一句 {}也不能省略
	//2.if 与函数体的 { 必须在同一行

	//两种写法
	//1
	a := 1
	if a > 0 {
		fmt.Println("a>0")
	}
	//a的作用域还没有结束

	//2
	if b := 4; b < 3 {
		fmt.Println("b<3")
	} else {
		fmt.Println(b)
	}
	//出了if else的作用域后b就被释放了

}
