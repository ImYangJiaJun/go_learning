package main

import (
	"fmt"
	"strconv"
)

func main() {

	//使用fmt.Sprintf() 拼接字符串/将其它类型转换成string
	var a int = 24
	var b bool = true
	var c byte = 'a'
	var d float32 = 3.14159265359

	strT := fmt.Sprintf("%d %t %c %f", a, b, c, d)
	fmt.Printf("\n结果:%v\t类型:%T\n", strT, strT)

	//使用strconv包转换类型
	//int->string
	aStr := strconv.FormatInt(int64(a), 10) //(int64类型的要转换的数值，表示进制的int)
	fmt.Printf("\n结果:%v\t类型:%T\n", aStr, aStr)

	//float->string
	/*
		参数：f 要转换的浮点数、fmt 格式、prec 精度、bitSize 位数（32/64）。
		fmt 常用格式：
		  'b'：二进制指数形式
		  'e'：科学计数法，如 1.23e+03
		  'E'：科学计数法，如 1.23E+03
		  'f'：普通小数形式，如 123.45
		  'g'：根据数值自动选择 'e' 或 'f'
		  'G'：根据数值自动选择 'E' 或 'f'
		  'x'：十六进制浮点形式
		  'X'：大写十六进制浮点形式
		prec 表示精度：'f' 为小数位数，'e'/'g' 等为有效数字位数；-1 表示尽可能使用最少数字且保证精确还原。
		例：FormatFloat(3.14159, 'f', 2, 64) → "3.14"
	*/
	dStr := strconv.FormatFloat(float64(d), 'f', 3, 64)
	fmt.Printf("结果:%v\t类型:%T\n", dStr, dStr)

	//string转换成数值

	//string->int
	aStrInt, _ := strconv.ParseInt(aStr, 10, 64) //(要转换的字符串，进制，bitSize取值 0/8/16/32/64)
	fmt.Printf("\n结果：%v\t结果类型：%T\n", aStrInt, aStrInt)

	//string->float
	dStrFloat, _ := strconv.ParseFloat(dStr, 64) //(要转换的字符串，位数)
	fmt.Printf("\n结果：%v\t结果类型：%T\n", dStrFloat, dStrFloat)

}
