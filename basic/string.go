package main

import (
	"fmt"
	"strings"
)

func main() {
	/*
		转义符
		\r	回车符
		\n	换行符
		\t	制表符
		\'	单引号
		\"	双引号
		\\	反斜杠
	*/
	str := "t\rhis \nis\ta string \" \\ \" "
	fmt.Println(str)

	//多行字符串
	strm := `第一行
第二行
第三行
第四行`
	fmt.Println(strm)

	str1 := "test"
	str2 := "string"
	//string常用操作
	//len(str)求长度
	fmt.Printf("len(str) -> 字符串长度 %d\n", len(str1))

	//拼接字符串  + 或 fmt.Sprintf
	fmt.Printf("+ 拼接 %s\n", str1+str2)

	str3 := fmt.Sprintf("fmt.Sprintf 拼接 %v %v", str1, str2)
	fmt.Println(str3)

	//strings.Split 分割字符串   需要引入strings包
	var str4 string = "123-456-789"
	arr := strings.Split(str4, "-")
	fmt.Println(arr)

	//strings.Join(a[]string,sep string) join操作，切片连接成字符串
	str5 := strings.Join(arr, "_")
	fmt.Println(str5)

	//strings.Contains(str5, "a") 判断是否包含
	flag := strings.Contains(str5, "a")
	fmt.Printf("\nstr5是否包含a %v\nstrings.Contains 函数返回类型%T\n", flag, flag)

	//strings.HasPrefix(str5, "a")/strings.HasSuffix(str5, "a")  前缀/后缀判断
	flagPre := strings.HasPrefix(str5, "a")
	flagSuf := strings.HasSuffix(str5, "789")
	fmt.Printf("\nstr5前缀是否是a\t\t%v\t\tstrings.HasPrefix函数返回类型 %T\n", flagPre, flagPre)
	fmt.Printf("str5后缀是否是789\t%v\t\tstrings.HasSuffix函数返回类型 %T\n", flagSuf, flagSuf)

	//strings.Index(),strings.LastIndex() 字串出现的位置(结果是目标字符的下标，子串是子串开头的字符的下标)，找不到返回-1
	index := strings.Index(str5, "23")
	lastIndex := strings.LastIndex(str5, "_")
	fmt.Printf("\nstr5中23出现的位置 \t%v\tstrings.Index函数返回类型 %T", index, index)
	fmt.Printf("\nstr5中-最后出现的位置 \t%v\tstrings.LastIndex函数返回类型 %T", lastIndex, lastIndex)

}
