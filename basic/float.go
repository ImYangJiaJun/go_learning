package main

import (
	"fmt"
	"unsafe"

	"github.com/shopspring/decimal"
)

func main() {
	var x float64 = 3.141
	var y float32 = 3.14123123123
	fmt.Printf("x\t类型：%T\t值：%v\t|%f\t占用字节：%v\n", x, x, x, unsafe.Sizeof(x))
	fmt.Printf("y\t类型：%T\t值：%v\t|%f\t占用字节：%v\n", y, y, y, unsafe.Sizeof(y))
	//输出时，使用%v会原样输出，使用%f默认6位小数，少补0，多截断（四舍五入）

	var f float64 = 12.34567
	fmt.Printf("%%f -> %f\n", f)
	fmt.Printf("%%.2f -> %.2f\n", f)
	fmt.Printf("%%.0f -> %.0f\n", f)
	fmt.Printf("%%.10f -> %.10f\n", f)

	fmt.Printf("%%.3e -> %.3e\n", f)
	fmt.Printf("%%.3g -> %.3g\n", f)

	//科学计数法
	var f1 float32 = 3.14e2 //表示3.14*10的2次方
	fmt.Printf("y\t类型：%T\t值：%v\t|%f\t占用字节：%v\n", f1, f1, f1, unsafe.Sizeof(f1))
	var f2 float32 = 3.14e-2 //表示3.14*10的-2次方
	fmt.Printf("y\t类型：%T\t值：%v\t|%f\t占用字节：%v\n", f2, f2, f2, unsafe.Sizeof(f2))

	//float精度丢失，出现原因-在二进制中一些数为类似十进制中1/3的循环小数
	m1 := 8.2
	m2 := 3.8
	fmt.Println(m1 - m2) // 期望4.4，实际结果为4.3999999999999995

	var m3 float64 = 1129.6
	fmt.Println(m3 * 100) //期望112960，实际112959.99999999999

	//使用第三方包解决
	fmt.Println(decimal.NewFromFloat(m1).Sub(decimal.NewFromFloat(m2)))
	fmt.Println(decimal.NewFromFloat(m3).Mul(decimal.NewFromInt(100)))
}
