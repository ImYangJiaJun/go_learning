package main

import "fmt"

type Person struct {
	name string
	age  int
	sex  string
}

func main() {
	var p1 Person
	p1.name = "Jason"
	p1.age = 20
	p1.sex = "male"
	fmt.Printf("p1  value: %v   value#:%#v   type:%T\n", p1, p1, p1)

	//通过new实例化 指针类型
	//注：Go中支持对结构体指针直接使用 . 访问结构体成员 p2.name底层是  (*p2).name
	var p2 = new(Person)
	p2.name = "Violet"
	p2.age = 20
	(*p2).sex = "female"
	fmt.Printf("p2  value: %v   value#:%#v   type:%T\n", p2, p2, p2)

	//通过&实例化 指针类型
	var p3 = &Person{}
	p3.name = "John"
	p3.age = 21
	p3.sex = "male"
	fmt.Printf("p3  value: %v   value#:%#v   type:%T\n", p3, p3, p3)

	//键值对初始化实例化 不是指针类型		不初始化的字段会是对应类型的默认值
	var p4 = Person{
		name: "Alice",
		sex:  "female",
	}
	fmt.Printf("p4  value: %v   value#:%#v   type:%T\n", p4, p4, p4)

	//通过&+键值对实例化 指针类型	不初始化的字段会是对应类型的默认值
	var p5 = &Person{
		name: "Paradox",
		age:  21,
	}
	fmt.Printf("p5  value: %v   value#:%#v   type:%T\n", p5, p5, p5)

	//可以省略key，顺序要与定义的一致,不能少
	var p6 = Person{
		"Peter",
		21,
		"male",
	}
	fmt.Printf("p6  value: %v   value#:%#v   type:%T\n", p6, p6, p6)

	p7 := p6
	fmt.Printf("before  p6.age:%v   p7.age:%v\n", p6.age, p7.age) //21 21
	p7.age = 33
	fmt.Printf("after   p6.age:%v   p7.age:%v\n", p6.age, p7.age) //21 33
	//结构体是值类型
}
