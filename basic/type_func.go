package main

import "fmt"

// 非本地类型不能定义方法--不能给其它包的类型定义方法
type MyInt int

func (m MyInt) printInfo() {
	fmt.Println("自定义类型的自定义方法")
}

func main() {
	var m1 MyInt
	m1.printInfo()
}
