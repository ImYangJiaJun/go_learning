package main

import (
	"fmt"
	"reflect"
)

type myIntR int
type PersonR struct {
	Name string
	Age  int
}

// reflect获取任意变量的类型对象
func reflectType(x interface{}) {
	// TypeOf 获取动态类型,返回 reflect.Type
	t := reflect.TypeOf(x)
	// t:具体类型,如 main.PersonR
	// t.Name():类型名,指针等非命名类型为空
	// t.Kind():底层分类,如 myIntR 的 Kind 是 int
	fmt.Printf("TypeOf:%-14v Name:%-8v Kind:%-8v\n", t, t.Name(), t.Kind())
}

// 反射获取变量的原始值,返回 reflect.Value
func reflectValue(x interface{}) {
	v := reflect.ValueOf(x)

	// 根据底层分类取值,不同类型用不同的方法
	kind := v.Kind()
	switch kind {
	case reflect.Int:
		fmt.Printf("int value:%v\n", v.Int())
	case reflect.String:
		fmt.Printf("string value:%v\n", v.String())
	case reflect.Float32:
		fmt.Printf("Float32 value:%v\n", v)
	case reflect.Float64:
		fmt.Printf("Float64 value:%v\n", v)
	case reflect.Struct:
		fmt.Printf("Struct value:%v\n", v)
	case reflect.Ptr:
		fmt.Printf("Ptr value:%v\n", v)
	case reflect.Slice:
		fmt.Printf("Slice value:%v\n", v)
	case reflect.Array:
		fmt.Printf("Array value:%v\n", v)
	default:
		fmt.Printf("unknown type:%v\n", kind)
	}
}

// 通过反射修改变量的值
func reflectSetValue(x interface{}, v reflect.Value) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	// 必须传指针,Elem() 拿到指针指向的值,传值进来的是副本,修改无效
	a := reflect.ValueOf(x)
	// CanSet 判断值是否可修改
	if !a.Elem().CanSet() {
		return
	}
	// 根据类型选择对应的 Set 方法
	switch a.Elem().Kind() {
	case reflect.Int:
		a.Elem().SetInt(v.Int())
	case reflect.String:
		a.Elem().SetString(v.String())
	case reflect.Bool:
		a.Elem().SetBool(v.Bool())
	case reflect.Float64:
		a.Elem().SetFloat(v.Float())
	default:
		panic("unhandled default case")
	}
}

func main() {
	var a myIntR = 10
	var b = PersonR{
		Name: "Bob",
		Age:  20,
	}
	var c = 10
	var d = [3]int{1, 2, 3}
	var e = []int{1, 2, 3}
	// 反射获取类型
	reflectType("1")
	reflectType(a)
	reflectType(b)
	reflectType(c)
	reflectType(&c)
	reflectType(d)
	reflectType(e)
	fmt.Println("---------------------------------")
	// 反射获取原始值
	reflectValue(a)
	reflectValue(b)
	reflectValue(c)
	reflectValue(d)
	reflectValue(e)
	fmt.Println("---------------------------------")
	// 反射修改值,必须传指针
	var f = 0
	reflectSetValue(&f, reflect.ValueOf(34))
	fmt.Println("f:", f)
	var g = ""
	reflectSetValue(&g, reflect.ValueOf("hello"))
	fmt.Println("g:", g)
}
