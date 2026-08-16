package tools

import "fmt"
import T "go_learning/basic/mod/tools_by"

func init() {
	fmt.Println("tools/calc.go init")
}

func Mul(a, b int) int {
	T.Print()
	return a * b
}

func Div(a, b int) int {
	return a / b
}
