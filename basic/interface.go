package main

import "fmt"

// Usber 接口是一个规范
type Usber interface {
	start()
	stop()
}

// 如果接口中有方法，必须通过结构体/自定义类型实现这个接口
type Phone struct {
	Name string
}

// Phone 要实现Usber接口必须实现Usber中所有方法
func (p Phone) start() { //值接收者，实例化后的结构体值类型和结构体指针类型都可以实现接口
	fmt.Println(p.Name, "start")
}
func (p Phone) stop() {
	fmt.Println(p.Name, "stop")
}

type Camera struct{}

func (c *Camera) start() { //指针接收者，只有结构体指针类型可以，值类型不行
	fmt.Println("camera start")
}
func (c *Camera) stop() { //指针接收者
	fmt.Println("camera stop")
}
func (c *Camera) run() { //接口规定的必须实现之外，还可以实现其余方法
	fmt.Println("camera run")
}

type Computer struct{}

// 使用示例
func (c Computer) work(usb Usber) { //使用接口限定传入的结构体必须满足接口规定的方法
	//使用类型断言判断接口传入类型
	if _, ok := usb.(Phone); ok {
		fmt.Println("type : Phone")
	} else if _, ok := usb.(*Camera); ok {
		fmt.Println("type : Camera")
	}
	usb.start()
	usb.stop()
}

func main() {
	p := Phone{
		Name: "VIVO",
	}
	p.start()

	var u Usber //Go中接口就是一个数据类型
	u = p       //表示Phone实现Usber接口
	u.start()
	u.stop()

	c := Camera{}
	var u1 Usber
	u1 = &c //指针接收者实现接口要使用&
	u1.start()
	u1.stop()

	c.run() //接口规定的方法之外的方法不能使用接口变量调用，只能通过结构体变量本事调用

	fmt.Println("-------------------------------------------")
	var computer = Computer{}
	computer.work(&c) //限制传入的结构体必须实现Usber接口的方法
	computer.work(p)

	fmt.Println("-------------------------------------------")

}
