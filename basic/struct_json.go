package main

import (
	"encoding/json"
	"fmt"
)

type PerJson struct {
	ID   int `json:"id"` //tag实现指定json序列化该字段的key
	Name string
	Age  int
	sex  string //私有属性（全小写）不能被json包访问，公有（首字母大写）才可以
}

//嵌套结构体的序列化/反序列化

type Student struct {
	ID     int
	Name   string
	Gender string
}
type Class struct {
	Title   string
	Student []Student
}

func main() {
	var s1 = PerJson{
		ID:   1,
		Name: "Jason",
		Age:  20,
		sex:  "male",
	}
	fmt.Printf("%#v\n", s1)

	//结构体转Json
	jsonBytes, _ := json.Marshal(s1)

	jsonStr := string(jsonBytes)
	fmt.Printf("%v\n", jsonStr)
	fmt.Println("---------------------------------------------")
	var s2 PerJson
	err := json.Unmarshal([]byte(jsonStr), &s2)
	if err == nil {
		fmt.Printf("%#v\n", s2)
	}

	fmt.Println("---------------------------------------------")
	c := Class{
		Title:   "2班",
		Student: make([]Student, 0),
	}
	for i := 1; i <= 10; i++ {
		s := Student{
			ID:     i,
			Name:   fmt.Sprintf("Student %d", i),
			Gender: "male",
		}
		c.Student = append(c.Student, s)
	}
	fmt.Printf("%#v\n", c)

	strBytes, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("序列化成功", string(strBytes))
	}

	//反序列化
	fmt.Println("---------------------------------------------")
	strFancy := string(strBytes)
	var cFancy = &Class{}
	err = json.Unmarshal([]byte(strFancy), cFancy)
	if err == nil {
		fmt.Printf("反序列化成功：%#v\n", cFancy)
	}
}
