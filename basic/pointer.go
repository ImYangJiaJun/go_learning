package main

import "fmt"

func fnPointer1(x int) {
	x += 10
	fmt.Println("result in fn1", x)
}
func fnPointer2(x *int) {
	*x += 10
	fmt.Println("result in fn2", *x)
}

func main() {
	var a = 0
	fmt.Printf("a 值：%v 类型：%T 地址：%p\n", a, a, &a)

	var p = &a //指针变量
	fmt.Printf("p 值：%v 类型：%T 地址：%p\n", p, p, &p)

	//*p表示取出这个地址的值
	fmt.Println(p)
	fmt.Println(*p)

	*p = 30 //改变这个内存地址的值
	fmt.Printf("a 值：%v 类型：%T 地址：%p\n", a, a, &a)

	fmt.Println("--------------------------------------")
	fnPointer1(a)
	fmt.Printf("After fn1 -> a 值：%v 类型：%T 地址：%p\n", a, a, &a)
	fnPointer2(&a)
	fmt.Printf("After fn2 -> a 值：%v 类型：%T 地址：%p\n", a, a, &a)

	fmt.Println("--------------------------------------")
	//指针变量本身存的是地址，它指向的内存必须先分配空间才可以使用
	var p1 *int = new(int)
	fmt.Printf("p1 值：%v 地址对应值：%v 类型：%T 地址：%p\n", p1, *p1, p1, &p1)

	/*
		new与make
		相同点：都是用于内存分配
		不同点：make只用于slice,map,channel三个引用类型的初始化，返回的是这三个引用类型本身
			  new可用于任意类型的内存分配，返回的是该类型的指针（值为该类型零值）

	*/

}
