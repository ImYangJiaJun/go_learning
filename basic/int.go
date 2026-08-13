package main

import (
	"fmt"
	"unsafe"
)

func main() {
	/*
		int类型

		有符号
		int8	-2^7	~	2^7-1
		int16	-2^15	~	2^15-1
		int32	-2^31	~	2^31-1
		int64	-2^63	~	2^63-1

		无符号
		uint8	0	~	2^8-1
		uint16	0	~	2^16-1
		uint32	0	~	2^32-1
		uint64	0	~	2^64-1

		注：单独使用int/uint时，根据电脑操作系统更改，32位操作系统就是int32/uint32，64位操作系统就是int64/uint64
		int默认值为0
	*/

	//unsafe.Sizeof 查看在内存占用的空间（多少字节,一个字节就是8位）
	var a int32 = 42
	fmt.Printf("a=%d  类型：%T  占用字节：%v\n", a, a, unsafe.Sizeof(a))

	//不同长度的int转换
	var a1 int32 = 42
	var a2 int64 = 31
	println(int64(a1) + a2)
	//!!!注意：高转低如果超限的话不会报错，会给一个错误值，要手动进行校验

	//数字字面量语法

	var c int32 = 15
	fmt.Printf("原样输出 -> %v\n", c)
	fmt.Printf("10进制输出 -> %d\n", c)
	fmt.Printf("2进制输出 -> %b\n", c)
	fmt.Printf("8进制输出 -> %o\n", c)
	fmt.Printf("16进制输出 -> %x\n", c)
}
