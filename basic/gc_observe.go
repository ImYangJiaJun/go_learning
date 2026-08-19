package main

/*
GC 观察实验（教程 spec·02）

玩法（任选组合）：

	GODEBUG=gctrace=1 go run basic/gc_observe.go          # 每次 GC 打一行日志
	GOGC=50 GODEBUG=gctrace=1 go run basic/gc_observe.go  # 对比：堆增长 50% 就触发 GC，GC 更频繁
	GOGC=off GODEBUG=gctrace=1 go run basic/gc_observe.go # 关闭自动 GC，只剩 runtime.GC() 那一次

gctrace 每行重点字段：
	gc 3@0.012s 6%: ... 1.0->1.5->0.8 MB ...
	  第 3 次 GC @ 程序启动后 12ms，GC 占用 CPU 6%，
	  GC 前堆存活->GC 时堆总量->GC 后堆存活

程序自身的 ReadMemStats 输出重点字段：
	HeapAlloc     当前堆上存活对象占用字节
	HeapObjects   堆上对象个数
	NumGC         已完成的 GC 轮数
注意：Go 1.26 默认启用 Green Tea GC（按内存 span 扫描，官方称开销降约 40%），
可用 GOEXPERIMENT=nogreenteagc 跑一遍对比旧回收器的 gctrace 输出差异。
*/
import (
	"fmt"
	"runtime"
)

func printStats(tag string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s: HeapAlloc=%.1fMB HeapObjects=%d NumGC=%d\n",
		tag, float64(m.HeapAlloc)/1024/1024, m.HeapObjects, m.NumGC)
}

func main() {
	printStats("启动后")

	// 制造垃圾：每轮分配一个 1MB 的切片，下轮即成为不可达垃圾
	for i := 0; i < 20; i++ {
		garbage := make([]byte, 1<<20)
		_ = garbage
	}
	printStats("分配 20MB 垃圾后")

	runtime.GC() // 主动触发一轮 GC（阻塞式），观察 NumGC +1、HeapAlloc 回落
	printStats("runtime.GC() 后")
}
