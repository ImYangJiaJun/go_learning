package main

import "fmt"

type Animaler1 interface {
	SetName(string)
}
type Animaler2 interface {
	GetName() string
}

// 接口嵌套，表示要实现其那套接口的所有方法
type Animaler interface {
	Animaler1
	Animaler2
}

type Dog0 struct {
	Name string
}

func (d *Dog0) SetName(name string) {
	d.Name = name
}

func (d Dog0) GetName() string {
	return d.Name
}

func main() {
	d := &Dog0{
		Name: "小黑",
	}
	var a1 Animaler1 = d
	var a2 Animaler2 = d
	fmt.Println(a2.GetName())
	a1.SetName("test")
	fmt.Println(a2.GetName())
	//一个结构体实现多个接口

	//接口嵌套
	var a Animaler = d
	fmt.Println(a.GetName())
	a.SetName("小灰")
	fmt.Println(a.GetName())

}
