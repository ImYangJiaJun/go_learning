package main

import "fmt"

func main() {
	//切片初始化方法1
	var slice1 []int
	fmt.Println(slice1 == nil)
	fmt.Printf("slice1 - 值：%v\t\t|长度：%v\t|类型：%T\n", slice1, len(slice1), slice1)
	//切片初始化方法2
	var slice2 = []int{1, 2, 3}
	fmt.Printf("slice2 - 值：%v\t|长度：%v\t|类型：%T\n", slice2, len(slice2), slice2)
	//切片初始化方法3
	var slice3 = []int{1: 1, 3: 3}
	fmt.Printf("slice3 - 值：%v\t|长度：%v\t|类型：%T\n", slice3, len(slice3), slice3)

	fmt.Println("--------------------------------------")
	//切片循环遍历,方法同for
	var strSlice = []string{"a", "b", "c"}
	for i := 0; i < len(strSlice); i++ {
		fmt.Printf("strSlice[%d]=%v\n", i, strSlice[i])
	}
	fmt.Println("--------------------------------------")
	for k, v := range strSlice {
		fmt.Printf("strSlice[%d]=%v\n", k, v)
	}

	//切片的长度和容量，用len()求长度-包含的元素个数，cap()求容量-从第一个元素开始到底层数组元素末尾的个数
	//基于数组定义切片
	fmt.Println("--------------------------------------")
	arr1 := [...]int{0, 1, 2, 3, 4, 5}
	slice4 := arr1[:]   //获取数组里面所有值创建切片
	slice5 := arr1[1:4] //左闭右开
	slice6 := arr1[2:]
	slice7 := arr1[:3]
	fmt.Printf("arr\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", arr1, len(arr1), cap(arr1), arr1)
	fmt.Printf("[:]\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", slice4, len(slice4), cap(slice4), slice4)
	fmt.Printf("[1:4]\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", slice5, len(slice5), cap(slice5), slice5)
	fmt.Printf("[2:]\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", slice6, len(slice6), cap(slice6), slice6)
	fmt.Printf("[:3]\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", slice7, len(slice7), cap(slice7), slice7)

	//基于切片定义切片
	fmt.Println("--------------------------------------")
	slice8 := slice4[:3]
	fmt.Printf("base\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", slice4, len(slice4), cap(slice4), slice4)
	fmt.Printf("[:3]\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", slice8, len(slice8), cap(slice8), slice8)

	//make()函数创建切片		make([]T,size,cap)
	fmt.Println("--------------------------------------")
	var sliceA = make([]int, 4, 8)
	fmt.Printf("sliceA\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", sliceA, len(sliceA), cap(sliceA), sliceA)

	//切片值的修改(同数组)
	fmt.Println("--------------------------------------")
	sliceA[1] = 1
	sliceA[2] = 2
	sliceA[3] = 3
	fmt.Printf("A_m\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", sliceA, len(sliceA), cap(sliceA), sliceA)

	//切片扩容使用append()为切片动态添加元素
	fmt.Println("--------------------------------------")
	var sliceB []int
	fmt.Printf("before\t值：%v\t\t\t长度：%v\t容量：%v\t类型：%T\n", sliceB, len(sliceB), cap(sliceB), sliceB)
	sliceB = append(sliceB, 12)
	fmt.Printf("after1\t值：%v\t\t长度：%v\t容量：%v\t类型：%T\n", sliceB, len(sliceB), cap(sliceB), sliceB)
	sliceB = append(sliceB, 34, 56, 78, 90)
	fmt.Printf("after2\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", sliceB, len(sliceB), cap(sliceB), sliceB)
	//append 扩容时底层数组按内存分配器的 size class 对齐：5 个 int 需要 40 字节，向上取整到 48 字节，而 48 字节能容纳 6 个 int，所以 cap 从 5 变成了 6
	sliceB = append(sliceB, 10)
	fmt.Printf("after3\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", sliceB, len(sliceB), cap(sliceB), sliceB)
	//刚好可以容纳，cap没有增加

	//append方法合并切片
	fmt.Println("--------------------------------------")
	sliceC := []string{"A", "B", "C"}
	sliceD := []string{"D", "E", "F"}
	sliceC = append(sliceC, sliceD...) // sliceD... 是切片的展开写法，等价于将切片的元素一个一个填入，只能用于切片不能用于数组
	fmt.Printf("sliceC\t值：%v\t长度：%v\t容量：%v\t类型：%T\n", sliceC, len(sliceC), cap(sliceC), sliceC)

	/*
		切片的扩容策略（Go 1.18及以后，1024阈值是1.18之前的旧规则）
		IF 新申请的容量大于2倍旧容量 最终容量基准就是新申请的容量
		ELSE IF 旧切片长度小于256 最终容量基准就是旧容量的2倍
		ELSE 旧切片长度大于等于256 最终容量基准从旧容量按约1.25倍循环增长直到大于等于新申请的容量
		注意：上面算出的只是基准值，实际容量还会按内存分配器的size class向上对齐（见上面append的cap从5变6的例子）
	*/

	fmt.Println("--------------------------------------")
	var sliceE []int
	for i := 0; i < 10; i++ {
		sliceE = append(sliceE, i)
		fmt.Printf("append%d\t长度：%v 容量：%v 类型：%T\t值：%v\n", i, len(sliceE), cap(sliceE), sliceE, sliceE)
	}

	//使用copy复制切片，切片默认是引用类型，使用copy深拷贝，修改被复制的变量，复制出来的变量不会被影响
	fmt.Println("--------------------------------------")
	var sliceF = []int{0, 1, 2, 3, 4}
	sliceG := make([]int, len(sliceF), cap(sliceF))
	copy(sliceG, sliceF)
	sliceF[0] = 100
	fmt.Printf("sliceF:%v\t\tsliceG:%v\n", sliceF, sliceG)

	//从切片删除元素
	fmt.Println("--------------------------------------")
	sliceH := []int{0, 1, 2, 3, 4, 5, 6}
	//删除索引为2的元素
	sliceH = append(sliceH[:2], sliceH[3:]...)
	fmt.Printf("sliceH\t长度：%v 容量：%v 类型：%T\t值：%v\n", len(sliceH), cap(sliceH), sliceH, sliceH)

	//修改字符串,本质是修改切片
	fmt.Println("--------------------------------------")
	str := "你好Golang"
	runeStr := []rune(str)
	fmt.Println("Unicode:", runeStr)
	runeStr[0] = '您'
	fmt.Println(string(runeStr))

}
