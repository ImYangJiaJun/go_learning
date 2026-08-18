package main

import (
	"fmt"
	"unsafe"
)

func main() {
	/*
		Go中字符有两种：
		1.uint8，又叫byte型，代表ASCII码的一个字符
		2.rune，代表一个UTF-8字符
		处理汉字、日文等复合字符（例如一个汉字占用三字节也就是3byte）时需要用rune
	*/
	// 定义字符用 '' 字符属于int类型 默认输出ASCII码
	var a = 'a'
	fmt.Printf("%%v输出ASCII码：%v\t原样输出用%%c：%c\t\t类型：%T\n", a, a, a)

	//输出一个字符串里面的字符
	var str = "abc一二三"
	fmt.Printf("\n%%v输出ASCII码：%v\t原样输出用%%c：%c\t\t类型：%T\n", str[1], str[1], str[1])

	//注：一个汉字占用3字节，一个字母占用1字节
	//unsafe.Sizeof() 返回的是string头部（指针+长度）的固定大小（64位系统16字节），与内容长度无关；字符串内容的字节数用len()查看
	fmt.Printf("\nerror\tunsafe.Sizeof()\t->\t%v\nright\tlen()\t\t->\t%v\n", unsafe.Sizeof(str), len(str))

	//汉字字符，使用的是utf8编码，编码后的值就是int类型，直接输出的是Unicode编码的10进制
	var c = '杨'
	fmt.Printf("\n%%v输出Unicode码：%v\t原样输出用%%c：%c\t\t类型：%T\n", c, c, c)

	//通过循环输出字符串里面的字符
	println("------for(byte)循环")
	for i := 0; i < len(str); i++ { //byte类型循环,由于一个汉字占用3个字节导致输出结果中的汉字不正确
		fmt.Printf("%%v输出编码：%v\t原样输出用%%c：%c\t\t类型：%T\n", str[i], str[i], str[i])
	}

	println("------range(rune)循环")
	for _, r := range str {
		fmt.Printf("%%v输出编码：%v\t原样输出用%%c：%c\t\t类型：%T\n", r, r, r)
	}

	//修改字符串
	s1 := "big"                  //字符串不能直接修改
	byteStr := []byte(s1)        //没有汉字就先转换成byte数组
	byteStr[0] = 'p'             //再修改数组
	fmt.Println(string(byteStr)) //最后转回字符串

	s2 := "你好 Go"
	runeStr := []rune(s2)        //有汉字要转换成rune数组
	runeStr[0] = '我'             //再修改数组
	fmt.Println(string(runeStr)) //然后转回字符串

}
