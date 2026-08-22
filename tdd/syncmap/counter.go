package syncmap

import (
	"sync"
	"sync/atomic"
)

// AtomicCounter 是基于 atomic.Int64 的并发计数器，零值可用。
// 用类型化原子值（Go 1.19+）而不是 atomic.AddInt64(&n) 自由函数——
// 类型把"这个字段只能原子访问"变成编译器强制（uber·使用 atomic）
type AtomicCounter struct {
	n atomic.Int64
}

// Inc 原子加一：一条 CPU 原子指令（x86 上 LOCK XADD），无锁无阻塞
func (c *AtomicCounter) Inc() {
	c.n.Add(1)
}

// Load 原子读当前值
func (c *AtomicCounter) Load() int64 {
	return c.n.Load()
}

// MutexCounter 是 Mutex 保护的并发计数器，AtomicCounter 的对照组。
// 零值可用——mu 是值字段（uber·零值 Mutex），不是 *sync.Mutex
// （指针零值是 nil，Lock 直接 panic）。
// 只以指针形式传递和使用（值拷贝会复制锁状态）
type MutexCounter struct {
	mu sync.Mutex
	n  int64
}

// Inc 加锁加一：至少两次原子操作（加锁+解锁），竞争时还有挂起/唤醒——
// 对比 AtomicCounter.Inc 的一条指令，临界区越短差距越夸张
func (c *MutexCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

// Load 加锁读当前值
func (c *MutexCounter) Load() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
