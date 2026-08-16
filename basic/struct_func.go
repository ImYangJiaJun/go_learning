package main

import "fmt"

type person struct {
	name string
	age  int
	sex  string
}

func (p person) PrintInfo() {
	fmt.Printf("name:%v, age:%v\n", p.name, p.age)
}

// SetInfo 要修改结构体内部的值必须使用指针
func (p *person) SetInfo(name string, age int) {
	p.name = name
	p.age = age
}

func main() {
	//结构体实例相互独立不会相互影响
	var p1 = person{
		name: "Jason",
		age:  20,
		sex:  "male",
	}
	p1.PrintInfo()

	var p2 = person{
		name: "Violet",
		age:  20,
		sex:  "female",
	}
	p2.PrintInfo()
	p2.SetInfo("Jason", 33) //调用指针类型的接收者不需要使用 (&p2)
	p2.PrintInfo()
}
