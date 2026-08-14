package main

import "fmt"

func main() {
	//数组长度是类型的一部分,长度在定义后就不能改变,数组内部只能保存定义类型的变量
	var arr1 [3]int
	var arr2 [5]string
	fmt.Printf("type of arr1:%T   type of arr2:%T\n", arr1, arr2)

	//数组的初始化
	var arr3 [3]int
	arr3[0] = 1
	arr3[1] = 2
	arr3[2] = 3
	fmt.Println("arr3:", arr3)

	//初始化方法二
	var arr4 = [3]int{1, 2, 3}
	fmt.Println("arr4:", arr4)

	//初始化方法三,根据初始值个数自动推断数组长度
	var arr5 = [...]int{1, 2, 3, 4, 5}
	fmt.Println("arr5:", arr5, "len:", len(arr5))

	//初始化方法四，指定下标
	var arr6 = [...]string{1: "一", 3: "三", 5: "五"}
	fmt.Println("arr6:", arr6, "len:", len(arr6))

	//数组循环遍历使用 for/for range
	for i := 0; i < len(arr6); i++ {
		fmt.Printf("arr6[%d]=%v\n", i, arr6[i])
	}

	for k, v := range arr6 {
		fmt.Printf("arr6[%d]=%v\n", k, v)
	}

}
