package main

import (
	"fmt"
	"sync"
	"time"
)

/*
channel管道是一种特殊的类型，管道像是传送带/队列，遵循先入先出规则，声明channel的时候要指定元素类型
channel是引用数据类型
声明的channel需要使用make初始化后才能使用
*/

var wg1 sync.WaitGroup

// 写数据
func fnWrite(ch chan int) {
	defer wg1.Done()
	for i := 0; i < 10; i++ {
		ch <- i
		fmt.Println("write:", i)
		time.Sleep(time.Millisecond * 50)
	}
	close(ch)
}

func fnRead(ch chan int) {
	defer wg1.Done()
	for v := range ch {
		fmt.Println("read", v)
		time.Sleep(time.Millisecond * 500) //读写速度不一致的时候会自动等待，不会触发报错
	}
}

// 向intChan放入 2 到 n-1 的数
func putNum(intChan chan<- int, n int) {
	defer wg1.Done()
	for i := 2; i < n; i++ {
		intChan <- i
	}
	close(intChan)
}

// 从intChan取出数据，并判断是否为素数，如果是就把得到的素数放在primeChan
func primeNum(intChan <-chan int, primeChan chan<- int, exitChan chan<- bool) { //使用只读/只写channel 限制数据操作方向
	defer wg1.Done()
	for num := range intChan {
		flag := true
		for i := 2; i < num; i++ {
			if num%i == 0 {
				flag = false
				break
			}
		}
		if flag {
			primeChan <- num
		}
	}
	exitChan <- true
}

func exitPrime(exitChan chan bool, primeChan chan int) {
	defer wg1.Done()
	for i := 0; i < 16; i++ {
		<-exitChan
	}
	close(exitChan)
	close(primeChan)
}

func printPrime(primeChan chan int) {
	defer wg1.Done()
	for num := range primeChan {
		fmt.Printf("%d\t\t是素数\n", num)
	}
}

func main() {
	//创建channel
	ch := make(chan int, 3)
	//给channel存入数据
	ch <- 1
	ch <- 2
	ch <- 3
	//获取channel的内容
	a := <-ch
	fmt.Println(a)
	<-ch
	c := <-ch
	fmt.Println(c)

	fmt.Printf("ch 值：%v 类型：%T 容量：%v 长度：%v\n", ch, ch, cap(ch), len(ch)) //是引用数据类型，所以值是地址

	ch1 := make(chan int, 4)
	ch1 <- 0
	ch1 <- 1
	ch1 <- 2
	ch2 := ch1
	ch2 <- 3
	<-ch1
	<-ch1
	<-ch1
	d := <-ch1
	fmt.Println(d) //3  证明channel为引用数据类型

	//管道阻塞	channel为空时取、满时存会阻塞等待；当所有goroutine都阻塞时会报 fatal error: deadlock
	ch3 := make(chan int, 1)
	ch3 <- 0
	//ch3 <- 1

	fmt.Println("---------------------------------------------------")
	//遍历channel
	var ch4 = make(chan int, 10)
	for i := 0; i < 10; i++ {
		ch4 <- i
	}
	//通过for循环遍历不需要关闭channel
	for i := 0; i < 10; i++ {
		fmt.Println(<-ch4)
	}
	for i := 0; i < 10; i++ {
		ch4 <- i * 10
	}
	//使用for range循环遍历， 注意channel没有key
	//遍历前需要关闭channel，不然会报错 fatal error: all goroutines are asleep - deadlock!
	close(ch4)
	for val := range ch4 {
		fmt.Println(val)
	}

	fmt.Println("---------------------------------------------------")
	var ch5 = make(chan int, 1)

	wg1.Add(2)
	go fnWrite(ch5)
	go fnRead(ch5)
	wg1.Wait()

	fmt.Println("---------------------------------------------------")
	intChan := make(chan int, 1000)
	primeChan := make(chan int, 1000)
	exitChan := make(chan bool, 16)

	wg1.Add(1)
	go putNum(intChan, 100)
	for i := 0; i < 16; i++ {
		wg1.Add(1)
		go primeNum(intChan, primeChan, exitChan)
	}
	wg1.Add(2)
	go exitPrime(exitChan, primeChan)
	go printPrime(primeChan)

	wg1.Wait()

	fmt.Println("---------------------------------------------------")

	//默认情况下管道是双向的（可读可写）
	//声明为只写
	ch6 := make(chan<- int, 2)
	ch6 <- 0
	ch6 <- 1
	//声明为只读
	//ch7 := make(<-chan int, 2)
	//用处-限制函数的数据操作方向

	fmt.Println("---------------------------------------------------")
	//select一次读取多个channel的数据
	intChanS := make(chan int, 10)
	for i := 0; i < 10; i++ {
		intChanS <- i
	}
	stringChanS := make(chan string, 10)
	for i := 0; i < 10; i++ {
		stringChanS <- fmt.Sprint(i)
	}

	//使用select获取channel的数据的时候不需要关闭channel
	for {
		select {
		case v := <-intChanS:
			fmt.Printf("intChanS:%v\n", v)
			time.Sleep(time.Millisecond * 50)
		case v := <-stringChanS:
			fmt.Printf("stringChanS:%v\n", v)
			time.Sleep(time.Millisecond * 50)
		default:
			fmt.Printf("数据获取完毕\n")
			return

		}
	}

}
