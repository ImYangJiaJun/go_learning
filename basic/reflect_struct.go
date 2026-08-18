package main

import (
	"fmt"
	"reflect"
)

// StudentRe 演示用的结构体，字段上带有 json / form 标签
type StudentRe struct {
	Name  string `json:"name" form:"username"`
	Age   int    `json:"age"`
	Score int    `json:"score"`
}

// GetInfo 值接收者方法：返回格式化的学生信息
func (s StudentRe) GetInfo() string {
	var str = fmt.Sprintf("Student Name:%v Age:%v Score:%v", s.Name, s.Age, s.Score)
	return str
}

// SetInfo 指针接收者方法：修改结构体字段
func (s *StudentRe) SetInfo(name string, age, score int) {
	s.Name = name
	s.Age = age
	s.Score = score
}

// Print 普通打印方法
func (s StudentRe) Print() {
	fmt.Println("这是一个打印方法...")
}

// PrintStructField 通过反射打印结构体（或结构体指针）的字段信息
func PrintStructField(s interface{}) {
	t := reflect.TypeOf(s)
	// 传入的是指针时，通过 Elem() 解引用拿到指向的元素类型
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// 解引用后必须是结构体，否则直接返回（同时避免了 t.Elem() 的 panic）
	if t.Kind() != reflect.Struct {
		fmt.Println("传入的参数不是结构体")
		return
	}

	// 通过类型变量的 Field 获取指定下标的字段信息
	field0 := t.Field(0)
	fmt.Println("字段名称：", field0.Name)
	fmt.Println("字段类型：", field0.Type)
	fmt.Println("字段Tag：", field0.Tag.Get("json"))

	// 通过 FieldByName 按名称获取字段，第二个返回值表示字段是否存在
	fmt.Println("-----------------------------------")
	field1, ok := t.FieldByName("Age")
	if ok {
		fmt.Println("字段名称：", field1.Name)
		fmt.Println("字段类型：", field1.Type)
		fmt.Println("字段Tag：", field1.Tag.Get("json"))
		fmt.Println("字段Tag：", field1.Tag.Get("form"))
	}

	// 通过 NumField 获取字段个数，并遍历打印所有字段
	fmt.Println("-----------------------------------")
	fieldCount := t.NumField()
	fmt.Println("结构体有", fieldCount, "个属性")
	for i := 0; i < fieldCount; i++ {
		f := t.Field(i)
		fmt.Printf("字段%d 名称:%v 类型:%v jsonTag:%v\n", i, f.Name, f.Type, f.Tag.Get("json"))
	}

	// 通过值变量获取结构体字段的具体值
	fmt.Println("-----------------------------------")
	v := reflect.ValueOf(s)
	// 值也是指针时同样先解引用
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	fmt.Println("Name 字段的值：", v.FieldByName("Name"))
	fmt.Println("Age 字段的值：", v.FieldByName("Age"))

	// 通过反射获取并执行结构体的方法
	// 注意：传入指针时才能看到指针接收者方法（如 SetInfo）
	fmt.Println("-----------------------------------")
	mv := reflect.ValueOf(s)
	fmt.Println("结构体有", mv.NumMethod(), "个方法")
	for i := 0; i < mv.NumMethod(); i++ {
		m := mv.Type().Method(i)
		// Method.Type 的 NumIn 包含接收者本身，等于 1 即无参方法
		fmt.Println("方法名称：", m.Name, "参数个数：", m.Type.NumIn()-1)
		if m.Type.NumIn() == 1 {
			// Call 执行方法，入参和返回值都是 []reflect.Value
			results := mv.Method(i).Call(nil)
			for _, r := range results {
				fmt.Println("  返回值：", r)
			}
		}
	}

	// 通过 MethodByName 按名称获取方法并传参调用
	fmt.Println("-----------------------------------")
	setInfo := mv.MethodByName("SetInfo")
	if setInfo.IsValid() {
		setInfo.Call([]reflect.Value{reflect.ValueOf("Tom"), reflect.ValueOf(20), reflect.ValueOf(99)})
		fmt.Println("调用 SetInfo 后的值：", s)
	}
}

func main() {
	stu1 := StudentRe{
		Name:  "Jason",
		Age:   18,
		Score: 20,
	}
	// 传入结构体指针，反射内部会自动解引用
	PrintStructField(&stu1)
}
