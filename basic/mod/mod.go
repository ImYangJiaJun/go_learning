package main

import (
	"fmt"
	"go_learning/basic/mod/calc"
	T "go_learning/basic/mod/tools" //定义别名T，后续使用T调用这个包里的方法
)

/*
一个文件夹下包含的文件只能归属于一个package，同一个package的文件不能在多个文件夹下
包名可以不和文件夹的名字一样，包名不能含有 -
包名为main的包为应用程序的入口包，这种包编译后会得到一个可执行文件，而编译不含main包的源代码不会得到可执行文件
*/

/*
导入第三方包的步骤
一：
1.go mod init 项目名称	初始化
2.配置第三方包（直接在代码中使用）
3.go mod tidy 下载当前项目缺少的依赖
4.可以运行
二：
1.初始化同一.1
2.go get 包名		下载包
3.代码中使用
*/

/*
按照导入顺序初始化 calc->tools->main
tools中导入了tools_by会先初始化tools_by
包中有多个文件里面都有init的话会依次执行
*/

// main 包中init()先于main()
func init() {
	fmt.Println("mod.go init")
}

func main() {
	fmt.Println(calc.Add(1, 2))
	fmt.Println(calc.Pub)

	fmt.Println(T.Mul(2, 20)) //使用别名调用方法
	T.PrintInfo()
}
