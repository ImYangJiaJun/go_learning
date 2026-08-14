package main

import (
	"fmt"
	"sort"
)

func main() {
	/*
		map的定义
		1.无序的基于 key-value 的数据结构
		2.引用类型
		3.必须初始化才能使用
		定义语法: map[KeyType]ValueType
	*/

	//make创建
	var userinfo1 = make(map[string]string)
	userinfo1["name"] = "Jason"
	userinfo1["age"] = "18"
	userinfo1["sex"] = "male"

	fmt.Println("userinfo:", userinfo1)
	fmt.Println("userinfo_name:", userinfo1["name"])

	//创建的时候填充元素
	var userinfo2 = map[string]string{
		"name": "Jason",
		"age":  "18",
		"sex":  "male",
	}
	fmt.Println("userinfo2:", userinfo2)

	//类型推导创建
	userinfo3 := map[string]string{
		"name": "Jason",
		"age":  "18",
		"sex":  "male",
	}
	fmt.Println("userinfo3:", userinfo3)

	//使用 for range循环遍历
	for k, v := range userinfo3 {
		fmt.Println("k:", k, "v:", v)
	}

	//修改map
	userinfo3["age"] = "23"
	fmt.Println("after_edit:", userinfo3)

	//获取/查找 map类型的数据
	value1, exist1 := userinfo3["age"]
	fmt.Println("age", "是否存在：", exist1, "值:", value1)
	value2, exist2 := userinfo3["xxxx"]
	fmt.Println("xxxx", "是否存在：", exist2, "值:", value2)

	//删除map中的键值对
	fmt.Println("before", userinfo3)
	delete(userinfo3, "age")
	fmt.Println("after", userinfo3)

	//值为map类型的slice
	fmt.Println("--------------------------------------")
	var userinfos = make([]map[string]string, 3, 3)
	if userinfos[0] == nil { //map定义不初始化，默认值为nil
		userinfos[0] = make(map[string]string)
		userinfos[0]["name"] = "Jason"
		userinfos[0]["age"] = "18"
	}
	if userinfos[1] == nil {
		userinfos[1] = make(map[string]string)
		userinfos[1]["name"] = "Violet"
		userinfos[1]["age"] = "18"
	}
	fmt.Println("userinfos:", userinfos)
	//同样使用for range进行遍历
	for k, v := range userinfos {
		for k2, v2 := range v {
			fmt.Println("k:", k, "k2:", k2, "v:", v2)
		}
	}

	//值为slice的map
	fmt.Println("--------------------------------------")
	var userinfos2 = make(map[string][]string)
	userinfos2["hobby"] = []string{"sing", "dance", "code"}
	userinfos2["work"] = []string{"C", "C++", "Golang"}
	fmt.Println("userinfos2:", userinfos2)
	//同样使用for range进行遍历
	for k, v := range userinfos2 {
		for k2, v2 := range v {
			fmt.Println("k:", k, "k2:", k2, "v:", v2)
		}
	}

	//map是引用类型
	fmt.Println("--------------------------------------")
	var userinfo4 = make(map[string]string)
	userinfo4["name"] = "Jason"
	userinfo4["age"] = "18"
	userinfo5 := userinfo4
	fmt.Println("before", "userinfo4:", userinfo4, "userinfo5:", userinfo5)
	userinfo5["name"] = "Violet"
	userinfo4["age"] = "21"
	fmt.Println("after", "userinfo4:", userinfo4, "userinfo5:", userinfo5)

	//map的排序
	//按照key升序输出
	fmt.Println("--------------------------------------")
	map1 := map[int]int{
		5:  2,
		1:  2,
		89: 36,
		7:  17,
	}
	var keySlice []int
	for k := range map1 {
		keySlice = append(keySlice, k)
	}
	sort.Ints(keySlice)
	fmt.Println("keySlice:", keySlice)
	for _, k := range keySlice {
		fmt.Println("k:", k, "v:", map1[k])
	}

}
