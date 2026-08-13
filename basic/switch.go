package main

import "fmt"

func main() {
	/*
		switch expression{
			case condition:
			default:
		}

		！！！注：Go中不写break也不会穿透
	*/

	//1
	var ext = ".html"
	switch ext {
	case ".html":
		fmt.Println("text/html")
	case ".css":
		fmt.Println("text/css")
		break
	case ".js", ".jsx":
		fmt.Println("text/javascript")
		break
	default:
		fmt.Println("没找到该类型：", ext)
	}

	//2
	switch ext := ".css"; ext {
	case ".html":
		fmt.Println("text/html")
		break
	case ".css":
		fmt.Println("text/css")
	case ".js":
		fmt.Println("text/javascript")
		break
	default:
		fmt.Println("没找到该类型：", ext)
	}
	//两种写法区别同if，也是作用域的区别

	//case后使用表达式的时候switch后不用写变量
	var age = 30
	switch {
	case age < 24:
		fmt.Println("好好学习")
	case age >= 24 && age < 60:
		fmt.Println("努力工作")
	case age >= 60:
		fmt.Println("注意身体")
	default:
		fmt.Println("输入错误")
	}

	//switch的穿透 fallthrough
	//fallthrough可用执行满足条件的下一个case,一个fallthrough只会穿透一次
	println("------------------------------")
	switch {
	case age < 24:
		fmt.Println("好好学习")
	case age >= 24 && age < 60:
		fmt.Println("努力工作")
		fallthrough
	case age >= 60:
		fmt.Println("注意身体")
	default:
		fmt.Println("输入错误")
	}
}
