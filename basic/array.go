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

	//数组以及其余基本类型是值类型
	arr7 := arr5 //arr7会单独分配内存
	fmt.Println("before:", "arr5:", arr5, "arr7", arr7)
	arr5[0] = 11 //修改arr5不会影响arr7
	fmt.Println("after:", "arr5:", arr5, "arr7", arr7)

	//切片是引用类型
	slice1 := []int{1, 2, 3} //[]内部不定义长度就是切片
	slice2 := slice1         //不会分配新内存，指向同一份内存,修改其中一个会影响所有指向这个内存的变量
	fmt.Println("before:", "slice1:", slice1, "slice2:", slice2)
	slice1[0] = 11
	fmt.Println("after:", "slice1:", slice1, "slice2:", slice2)
	slice2[1] = 22
	fmt.Println("after:", "slice1:", slice1, "slice2:", slice2)

	//多维数组
	var arrs1 = [3][2]string{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	}
	fmt.Println("arrs1:", arrs1, "len:", len(arrs1), "arrs1[0]", arrs1[0], "len_in:", len(arrs1[0]), "arrs[0][0]:", arrs1[0][0])

	println("--------------------------------------")

	//循环遍历
	for k1, v1 := range arrs1 {
		for k2, v2 := range v1 {
			fmt.Printf("arrs1[%d][%d]=%v\n", k1, k2, v2)
		}
	}
	println("--------------------------------------")
	for i := 0; i < len(arrs1); i++ {
		for j := 0; j < len(arrs1[i]); j++ {
			fmt.Printf("arrs1[%d][%d]=%v\n", i, j, arrs1[i][j])
		}
	}

	println("--------------------------------------")
	var arrs2 = [...][2]string{ //[...]只能用在外层
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
		{"g", "h"},
	}
	fmt.Println("arrs2:", arrs2, "len:", len(arrs2))
}
