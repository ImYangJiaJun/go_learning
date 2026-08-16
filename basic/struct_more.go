package main

import "fmt"

/*
结构体的匿名字段：
	结构体允许成员字段在生命的时候只有类型，这种没有名字的字段就称为匿名字段
	匿名字段默认类型名为字段名，结构体要求字段名唯一，所以一个结构体中同种类型的匿名字段只能有一个
*/

type per_noname struct {
	string
	int
}

/*
结构体字段类型可以是：基本字段类型、切片、Map、结构体
如果结构体的字段类型是：指针、slice和Map则默认值为nil，如果没有在定义的时候初始化，后续要分配空间（make）才能使用
*/
type per_1 struct {
	name    string
	age     int
	hobbies []string
	map1    map[string]string
}

/*
结构体嵌套
*/
type User struct {
	Username string
	Password string
	Address  //User结构体嵌套Address结构体
	Email
}

type Address struct {
	Name  string
	Phone string
	City  string
}

type Email struct {
	Email string
	Phone string
}

/*
结构体的继承（使用嵌套实现）
*/
type Animal struct {
	Name string
}

func (a Animal) Speak() {
	fmt.Println(a.Name, "is speaking")
}

type Dog struct {
	Age int
	Animal
}

func (d Dog) Wang() {
	fmt.Println(d.Name, "is wang wang")
}

type Cat struct {
	Age int
	*Animal
}

func (c Cat) Miao() {
	fmt.Println(c.Name, "is miao")
}

func main() {
	p := per_noname{
		"Jason",
		1,
	}
	fmt.Println(p)
	fmt.Println("---------------------------------------------")

	var p1 per_1
	p1.name = "Jason"
	p1.age = 1
	p1.hobbies = make([]string, 6, 6) //必须分配空间才能使用
	p1.hobbies[0] = "Reading"
	p1.hobbies[1] = "Writing"
	p1.map1 = make(map[string]string)
	p1.map1["address"] = "Sichuan"

	fmt.Printf("%#v\n", p1)

	fmt.Println("---------------------------------------------")
	var u User
	u.Username = "Jason"
	u.Password = "123456"
	u.Address.Name = "Jason"
	u.Address.Phone = "123456" //两个嵌套的匿名结构体中都有相同的字段的时候不能直接使用内层的字段名
	u.Email.Phone = "123456"
	u.City = "Sichuan" //嵌套的匿名结构体在外层可以直接使用内层的字段名进行访问
	fmt.Printf("%#v\n", u)

	fmt.Println("---------------------------------------------")
	var d = Dog{
		Age: 18,
		Animal: Animal{
			Name: "Niuniu",
		},
	}
	d.Speak()
	d.Wang()

	var c = Cat{
		Age: 18,
		Animal: &Animal{ //嵌套的是结构体指针，所以需要7
			Name: "Caibao",
		},
	}
	c.Speak()
	c.Miao()

}
