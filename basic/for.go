package main

import "fmt"

func main() {
	/*
		for 初始语句;条件表达式;结束语句{
			循环体语句
		}

		执行顺序：
		初始->条件->循环体->结束->条件->循环体->结束->条件->循环体->…………

		注：Go中没有while，可用for {} 替代
	*/

	//打印1-10
	println("------------------------------")
	for i := 1; i <= 10; i++ {
		fmt.Printf("%d\n", i)
	}

	//省略初始化
	println("------------------------------")
	i := 1
	for ; i <= 10; i++ {
		fmt.Printf("%d\n", i)
	}

	//省略初始化和结束
	println("------------------------------")
	i = 1
	for i <= 10 {
		fmt.Printf("%d\n", i)
		i++
	}

	//全部省略
	println("------------------------------")
	i = 1
	for {
		if i > 10 {
			break
		} else {
			fmt.Printf("%d\n", i)
		}
		i++
	}

	//range
	println("------------------------------")
	str := "你好 Go"
	for k, v := range str {
		fmt.Printf("%d:%c\n", k, v)
	}
}
