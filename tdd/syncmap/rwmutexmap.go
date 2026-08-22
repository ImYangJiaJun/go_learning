package syncmap

import "sync"

// RWMutexMap 是手工版"读写锁 + 普通 map"，sync.Map 的 benchmark 对照组。
// 与 Dict 相反：零值不可用——m 是 nil map，直接写会 panic
// （assignment to entry in nil map），必须走构造函数。
// 只以指针形式传递和使用（含锁，值拷贝会复制锁状态，copylocks 会抓）
type RWMutexMap struct {
	mu sync.RWMutex
	m  map[string]int
}

// NewRWMutexMap 返回初始化好的 *RWMutexMap（m 已 make）
func NewRWMutexMap() *RWMutexMap {
	return &RWMutexMap{m: make(map[string]int)}
}

// Get 读（RLock）：存在返回 (值, true)，否则 (0, false)。
// RLock 允许多个读者并发——这正是读多写少场景下它还能勉强跟上 sync.Map 的原因
func (m *RWMutexMap) Get(key string) (int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

// Set 写（Lock）：键已存在则覆盖。
// 写锁是独占的——一个写者会挡住所有读者和其他写者
func (m *RWMutexMap) Set(key string, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}
