package main

import "fmt"

type A interface{} //空接口，表示没有任何约束，任意类型都可以实现
//Go中空接口可以作为类型来使用，可以表示任意类型

// 空接口作为函数参数，表示任意类型
func show(a interface{}) {
	fmt.Printf("value:%v type: %T\n", a, a)
}

// 使用类型断言，根据传入变量不同类型实现不同方法
func MyPrint(a interface{}) {
	//!!!注意：a.(type)只能用于switch语句
	switch a.(type) {
	case int:
		fmt.Println("int")
	case string:
		fmt.Println("string")
	case bool:
		fmt.Println("bool")
	case float64:
		fmt.Println("float64")
	default:
		fmt.Println("识别失败")
	}
}
func main() {
	var a A
	var str = "你好"
	a = str //让string类型实现A接口
	fmt.Printf("value:%v type: %T\n", a, a)

	var num = 20
	a = num //让int类型实现空接口
	fmt.Printf("value:%v type: %T\n", a, a)

	var t = []string{"1", "2", "3"}
	show(t)

	fmt.Println("------------------------------------------------")

	//空接口作为map值类型，实现可以存多种类型的map
	var m1 = make(map[string]interface{})
	m1["name"] = "Jason"
	m1["age"] = 23
	m1["alive"] = true
	fmt.Printf("value:%v\tvalue#:%#v\ttype:%T\n", m1, m1, m1)

	//空接口作为slice值类型，实现存多种类型的slice
	var s1 = []interface{}{1, "Jason", true}
	fmt.Printf("value:%v\tvalue#:%#v\ttype:%T\n", s1, s1, s1)

	//注意：slice、map、结构体传入空接口不能直接读取数据，需要使用类型断言赋值
	m1["slice"] = []int{1, 2, 4}
	s2, _ := m1["slice"].([]int)
	fmt.Println(s2)

	fmt.Println("------------------------------------------------")
	//类型断言
	var in interface{}
	in = "hello Go"
	v, ok := in.(string)
	if ok {
		fmt.Printf("value:%v\tvalue#:%#v\ttype:%T\n", v, v, v)
	} else {
		fmt.Println("断言失败")
	}

	MyPrint(true)
	MyPrint(1.1)
}
