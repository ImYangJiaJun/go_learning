package syncmap

import "sync"

// Dict 是 sync.Map 的薄封装：键、值都是 any。
// 零值即可用——var d Dict 之后直接 Store，没有也不需要构造函数，
// 因为内嵌的 sync.Map 本身就零值可用。
// Dict 含锁，第一次使用后禁止拷贝（全部方法用指针接收者，
// go vet 的 copylocks 检查会抓值拷贝）。
// （真实项目里这一层会把 any 收窄成具体类型；本练习保留 any，
//
//	是为了让每个方法的签名与 sync.Map 一一对应。）
type Dict struct {
	m sync.Map
}

// Store 存入键值对；键已存在则覆盖旧值。
// 键必须是可比较类型（内部就是 map 的键），否则 panic
func (d *Dict) Store(key, value any) {
	d.m.Store(key, value)
}

// Load 按键取值：存在 → (值, true)；不存在 → (nil, false)。
// 与普通 map 的 v, ok := m[k] 是同一套 ok-idiom
func (d *Dict) Load(key any) (value any, ok bool) {
	return d.m.Load(key)
}

// LoadOrStore 键已存在 → 返回旧值，loaded=true，本次不写入（传入的 value 被丢弃）；
// 键不存在 → 存入 value，返回 (value, false)。
// "取不到就存"是一步原子操作，不是 Load + Store 两步
func (d *Dict) LoadOrStore(key, value any) (actual any, loaded bool) {
	return d.m.LoadOrStore(key, value)
}

// Delete 删除键；键不存在时什么也不发生（不 panic）
func (d *Dict) Delete(key any) {
	d.m.Delete(key)
}

// Range 遍历全部键值对，逐个调用 f；f 返回 false 立即提前终止。
// 顺序不保证，不要依赖顺序写断言
func (d *Dict) Range(f func(key, value any) bool) {
	d.m.Range(func(key, value any) bool {
		return f(key, value)
	})
}

// CompareAndSwap （sync.Map 自 Go 1.20 起支持）：
// 键存在且当前值 == old → 替换为 new，返回 true；
// 否则（含键不存在）原样不动，返回 false。
// old 和当前值都必须可比较，否则 == 在运行时 panic
func (d *Dict) CompareAndSwap(key, old, new any) bool {
	return d.m.CompareAndSwap(key, old, new)
}

// Len 返回键值对总数。sync.Map 没有内置 Len——这是它的知名取舍，
// 只能用 Range 全表数一遍（O(n)）
func (d *Dict) Len() int {
	count := 0
	d.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}
