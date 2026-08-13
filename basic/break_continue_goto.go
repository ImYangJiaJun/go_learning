package main

import "fmt"

func main() {

	/*
		break		默认跳出当前循环，多重循环中可以用标号label标出想break的循环
		continue	结束当前循环，开始下一次循环，仅限for中使用
		go			跳转到指定标签
	*/

	//break	默认跳出当前循环
	for i := 0; i < 3; i++ {
		for j := 0; ; j++ {
			if j == 3 {
				fmt.Println("break")
				break
			}
			fmt.Printf("%d:%d\n", i, j)
		}
	}

	switch ext := ".css"; ext {
	case ".css":
		fmt.Println("text/css")
		break
		fmt.Println("text/css") //break提前跳出，不会执行
	case ".js":
		fmt.Println("text/javascript")
		break
	default:
		fmt.Println("没找到该类型：", ext)
	}

	println("------------------------------")

	//多重循环中可以用标号label标出想break的循环
lable11:
	for i := 0; i < 3; i++ {
		for j := 0; ; j++ {
			if j == 3 {
				fmt.Println("break lable11")
				break lable11
			}
			fmt.Printf("%d:%d\n", i, j)
		}
	}

	println("------------------------------")

	//continue
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Println(i)
	}

	//在continue后添加标签的时候，表示开始标签对应的循环
label2:
	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			if j == 3 {
				continue label2
			}
			fmt.Printf("%d:%d\n", i, j)
		}
	}

	println("------------------------------")

	//goto
	var x = 20
	if x < 30 {
		fmt.Println("x<30")
		goto label4
	}
	fmt.Println("a")
	fmt.Println("b")
label4:
	fmt.Println("c")
	fmt.Println("d")
	fmt.Println("e")

}
