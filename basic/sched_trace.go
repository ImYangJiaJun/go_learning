package main

/*
调度与抢占观察实验（教程 ch10 / ch20）

两种玩法：

1. 观察 GPM 调度状态（ch10）：

	GODEBUG=schedtrace=1000 go run basic/sched_trace.go

	每秒打印一行调度器状态，重点字段：
	  gomaxprocs    P 的数量（本程序固定为 1）
	  idleprocs     空闲 P 的数量
	  threads       OS 线程（M）总数
	  runqueue      全局可运行队列里的 G 数量
	  [0]           每个 P 的本地队列里的 G 数量

2. 观察异步抢占（ch20，Go 1.14+ 特性）：

	go run basic/sched_trace.go        # 会生成 trace.output
	go tool trace trace.output         # 浏览器打开，看 View trace

	GOMAXPROCS(1) 意味着同一时刻只有一个 P 在跑 Go 代码。
	10 个纯计算 goroutine 没有任何 IO/函数调用让出点，Go 1.13 及更早会
	一个跑完才轮到下一个（协作式抢占管不了"流氓协程"）；
	Go 1.14+ 通过异步信号强制抢占，trace 里能看到每个 goroutine
	被切成许多 ~20ms 的小段轮流执行。
*/
import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
)

// 纯计算函数：循环内没有 IO、没有 channel 操作，是"不主动让出"的典型
func calculateSum(w *sync.WaitGroup, p int) {
	defer w.Done()
	var sum, n int64
	for ; n < 100000000; n++ {
		sum += n
	}
	fmt.Println(p, sum)
}

func main() {
	runtime.GOMAXPROCS(1) // 单处理器场景：任何时刻只有一个 P 执行 Go 代码

	f, err := os.Create("trace.output")
	if err != nil {
		fmt.Println("创建 trace.output 失败:", err)
		return
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		fmt.Println("启动 trace 失败:", err)
		return
	}
	defer trace.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go calculateSum(&wg, i)
	}
	wg.Wait()
}
