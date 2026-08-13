package main

import "fmt"

func main() {
	//常量-不可改变值
	//定义的时候必须赋值
	const Pi = 3.1415926
	fmt.Println(Pi)

	/*
		多个常量一起声明方法同var

		const (
			n1=100
			n2=1
			n3
		)

		！！！注：同时声明多个常量，如果没有赋值表示和上一行的值相同
	*/

	const (
		n1 = 1
		n2
		n3 = 3
		n4
	)
	fmt.Printf("n1=%d, n2=%d, n3=%d, n4= %d\n", n1, n2, n3, n4)

	/*
		const 结合 iota

		iota - go中的常量计数器，同时定义多个常量的时候，iota出现的时候为当前是本次定义中当前的变量的序号，下一行不赋值就默认是累加，可以使用匿名变量跳过某一个位置的值
	*/

	const (
		i = 1
		i1
		i2 = iota
		_
		i3
		i4 = 1
		i5 = iota
	)
	fmt.Printf("i=%d , i1=%d, i2=%d, i3=%d, i4=%d, i5=%d\n", i, i1, i2, i3, i4, i5)

	//特殊情况-多个iota定义在同一行,每一列单独累加
	const (
		x1, x2 = iota + 1, iota
		x3, x4
		x5, x6
	)
	fmt.Printf("x1=%d , x2=%d \n", x1, x2)
	fmt.Printf("x3=%d , x4=%d \n", x3, x4)
	fmt.Printf("x5=%d , x6=%d \n", x5, x6)
}
