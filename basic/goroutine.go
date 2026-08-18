package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

/*
goroutine(协程)
多协程的开销比多线程少
*/

var wg sync.WaitGroup

func test() {
	for i := 0; i < 10; i++ {
		fmt.Println("test hello world", i)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Done() //协程计数器+1
}
func test1() {
	for i := 0; i < 10; i++ {
		fmt.Println("test1 hello world", i)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Done() //协程计数器+1
}

func testShow(i int) {
	defer wg.Done()
	for j := 0; j < 10; j++ {
		fmt.Printf("协程-%v  testShow %v \n", i, j)
		time.Sleep(time.Millisecond * 100)
	}
}

func panicTest() {
	//使用defer+recover处理协程中的panic
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panicTest()发生错误：", err)
		}
		wg.Done()
	}()

	var myMap map[int]string
	myMap[0] = "hello" //没有给myMap初始化内存导致panic报错
}

func main() {
	wg.Add(2) //协程计数器+2
	go test() //开启一个协程
	go test1()
	for i := 0; i < 10; i++ {
		fmt.Println("main hello world", i)
		time.Sleep(time.Millisecond * 50) //主进程执行完毕，无论协程是否执行完毕都会被中止
	}
	//time.Sleep(time.Second)//解决方案1-主线程增加等待时间
	wg.Wait() //解决方案2-使用sync包，wg.Wait表示等待协程执行完毕

	fmt.Println("--------------------")
	//获取当前计算机上Cpu个数
	cpuNum := runtime.NumCPU()
	fmt.Println("cpuNum:", cpuNum)
	//可以自己设置使用多个cpu
	runtime.GOMAXPROCS(cpuNum - 1)

	fmt.Println("--------------------")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go testShow(i)
	}
	wg.Wait()

	fmt.Println("--------------------")
	wg.Add(2)
	go testShow(5)
	go panicTest()
	wg.Wait()

}
