package main

import (
	"fmt"
	"sync"
	"time"
)

var count int

// 等待组,用于等待所有 goroutine 执行完毕
var wgLock sync.WaitGroup

// 互斥锁
var mutex sync.Mutex

// 读写锁
var rwMutex sync.RWMutex

// lockTest 演示互斥锁:同一时间只允许一个 goroutine 访问临界区
func lockTest(n int) {
	mutex.Lock()
	count++
	fmt.Println("goroutine-", n, "count-", count)
	time.Sleep(time.Millisecond)
	mutex.Unlock()
	wgLock.Done()
}

var m = make(map[int]int)

// lockTest1 演示互斥锁保护 map 的并发写入:计算阶乘并写入 map
func lockTest1(num int) {
	mutex.Lock()
	sum := 1
	for i := 1; i <= num; i++ {
		sum *= i
	}
	m[num] = sum
	fmt.Println("key-", num, "value-", sum)
	time.Sleep(time.Millisecond)
	mutex.Unlock()
	wgLock.Done()
}

// rwLockWrite 演示写锁:写锁是独占的,加锁期间其他读写操作都会被阻塞
func rwLockWrite(n int) {
	rwMutex.Lock()
	count += 10
	fmt.Println("write goroutine-", n, "count-", count)
	time.Sleep(time.Millisecond)
	rwMutex.Unlock()
	wgLock.Done()
}

// rwLockRead 演示读锁:多个 goroutine 可以同时持有读锁,互不阻塞
func rwLockRead(n int) {
	rwMutex.RLock()
	fmt.Println("read goroutine-", n, "count-", count)
	rwMutex.RUnlock()
	wgLock.Done()
}

func main() {
	// 互斥锁示例:20 个 goroutine 并发对 count 自增
	for i := 0; i < 20; i++ {
		wgLock.Add(1)
		go lockTest(i)
	}
	wgLock.Wait()

	fmt.Println("--------------------------------------------------")
	// 互斥锁示例:20 个 goroutine 并发计算阶乘并写入 map
	for i := 0; i < 20; i++ {
		wgLock.Add(1)
		go lockTest1(i)
	}
	wgLock.Wait()

	fmt.Println("--------------------------------------------------")
	// 读写锁示例:写操作独占,读操作并发,读多写少时性能优于互斥锁
	for i := 0; i < 5; i++ {
		wgLock.Add(1)
		go rwLockWrite(i)
	}
	for i := 0; i < 20; i++ {
		wgLock.Add(1)
		go rwLockRead(i)
	}
	wgLock.Wait()
}
