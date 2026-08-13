package main

func main() {
	/*
		bool类型注意事项

		1.bool类型默认值为false
		2.不允许将int转换为bool
		例如
		````
			var a int = 1
			if a{
				...
			}
		````
		是错误写法会报错
		3.bool无法参与数值运算，无法与其它类型进行转换
	*/

	var t bool
	println(t)
}
