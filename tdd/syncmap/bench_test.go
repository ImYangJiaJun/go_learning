package syncmap

import (
	"strconv"
	"testing"
)

// BenchmarkDictReadMostly 读多写少（95% 读）+ 键集稳定——sync.Map 的主场：
// 读全部落在无锁的 read 上。对照组是场景一模一样的 BenchmarkRWMutexMapReadMostly
func BenchmarkDictReadMostly(b *testing.B) {
	var d Dict // 零值可用，直接预填
	for i := range 100 {
		d.Store(strconv.Itoa(i), i)
	}
	// RunParallel：默认起 GOMAXPROCS 个 goroutine，等齐后同时开跑，
	// b.N 被切分给各 goroutine，pb.Next() 控制各自份额。
	// 串行的 b.N 循环测不出锁竞争——锁没人抢，永远显得很快
	b.RunParallel(func(pb *testing.PB) {
		i := 0 // 每个 goroutine 私有的计数
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			if i%20 == 0 { // 每 20 次操作 1 写 19 读
				d.Store(key, i)
			} else {
				d.Load(key)
			}
			i++
		}
	})
}

// BenchmarkRWMutexMapReadMostly 与 Dict 组一模一样的场景——benchmark 的
// 场景参数必须相同才可比，差一个变量结论就不成立
func BenchmarkRWMutexMapReadMostly(b *testing.B) {
	m := NewRWMutexMap() // 零值不可用（nil map 写 panic），必须构造函数——与 Dict 对照
	for i := range 100 {
		m.Set(strconv.Itoa(i), i)
	}
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			if i%20 == 0 {
				m.Set(key, i)
			} else {
				m.Get(key)
			}
			i++
		}
	})
}

// BenchmarkAtomicCounter 纯递增：Inc 编译成一条 CPU 原子指令
// （x86 上是 LOCK XADD），无锁、无 goroutine 阻塞
func BenchmarkAtomicCounter(b *testing.B) {
	c := &AtomicCounter{} // 零值可用
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

// BenchmarkMutexCounter 与 AtomicCounter 能力完全相同才可比：
// Mutex 版至少两次原子操作（加锁+解锁），竞争时还有挂起/唤醒——
// 临界区越短差距越夸张，这就是 uber"使用 atomic"条目的实证
func BenchmarkMutexCounter(b *testing.B) {
	c := &MutexCounter{} // 零值可用——mu 是值字段（uber·零值 Mutex）
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}
