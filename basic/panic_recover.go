package main

import (
	"errors"
	"fmt"
)

/*
使用panic/recover模式处理异常

panic可以在任何地方触发，recover只有在defer调用的函数中触发
*/

func fn1() {
	fmt.Println("fn1")
}

func fn2() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	panic("fn2 抛出一个异常")
}

// 模拟实际使用
func readFileTest(fileName string) error {
	if fileName == "write file" {
		return nil
	}
	return errors.New("read file fail")
}

func myFn() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("send email to admin")
		}
	}()
	err := readFileTest("wrong file")
	if err != nil {
		panic(err)
	}
}

func main() {
	fn1()
	fn2()

	myFn()
	fmt.Println("main end")
}
