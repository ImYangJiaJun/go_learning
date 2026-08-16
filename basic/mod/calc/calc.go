package calc

import "fmt"

var Pub = "公有变量" //首字母大写，表示为共有变量，可以在其余包中访问
var pri = "私有变量" //首字母小写，表示是私有变量，其余包中无法访问

func init() {
	fmt.Println("calc/calc.go init")
}

func Add(a, b int) int { //首字母大写表示是公有方法，其它包可以使用
	return a + b
}

func Sub(a, b int) int {
	return a - b
}

func pFn() { // 首字母小写表示是私有方法，只能在当前包里使用
	fmt.Println("私有方法")
}
