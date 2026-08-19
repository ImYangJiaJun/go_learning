# SyncMap TDD —— sync.Map：机制、并发验证与性能对照

目标：通过 6 个行为切片吃透 `sync.Map`——ok-idiom 读写、`LoadOrStore` 的原子语义、
`Delete`/`Range`、Go 1.20 的 `CompareAndSwap`，再用 `-race` 验证并发安全，最后用
benchmark 对照 `RWMutex+map` 与 `atomic` 计数器，把"什么时候该用 sync.Map"从背诵变成实测。
对应教程 ch18（Sync.Map 全章）和 uber 规范的零值 Mutex、使用 atomic 两条。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上——
> `Dict` 的每个方法就是 `sync.Map` 同名方法的 1:1 委托，注意力全部放在机制上。
> 用法：第一节看需求规格（接口契约固定，照此实现）；第二节是纯任务单——只给行为目标、
> 用例表和验收命令，测试代码全部自己写；第三节是知识点讲解，做之前通读或卡壳时查阅，
> 做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/syncmap` 就是它的运行方式，验收者是测试，不是人。

包名定为 `syncmap` 而不是 `sync`：本包代码要 `import "sync"`（内嵌 `sync.Map`、
并发测试要用 `sync.WaitGroup`），包名若与要导入的标准库同名，文件里的 `sync.Xxx`
会让人和工具都分不清指谁——本仓库约定一律避开（同理 net/http 练习叫 httplab、
context 练习叫 contextlab）。

这个包对外提供三组能力：

1. **`Dict`**：`sync.Map` 的薄封装（键、值都是 `any`），把本练习要用的 6 个操作
   收敛成固定契约，外加一个 `Len`；零值即可用
2. **`RWMutexMap`**：手工版"读写锁 + 普通 map"，benchmark 对照组；
   零值**不可用**，必须构造函数——与 Dict 形成对照
3. **`AtomicCounter` / `MutexCounter`**：两个并发计数器，atomic 与 Mutex 的对照组

### 调用关系（谁在调用谁）

```text
测试代码 ──► Dict.Store / Load / ... ──► 内嵌的 sync.Map          行为 1~5
benchmark ──► RWMutexMap.Get / Set ──► RWMutex + 普通 map          行为 6 对照组
benchmark ──► AtomicCounter.Inc / MutexCounter.Inc ──► atomic.Int64 / Mutex
```

`Dict` 的每个方法都是对 `sync.Map` 同名方法的 1:1 委托——封装层本身没有逻辑，
逻辑全在 `sync.Map` 里。这正是"机制学习"的含义：要验证的是标准库的行为，
不是自己写的算法。`RWMutexMap` 和两个计数器只为行为 6 的 benchmark 对照存在。

### 文件计划（共 5 个文件，分三次建）

最终目录长这样：

```text
tdd/syncmap/
├── dict_test.go    # 行为 1~5 的全部测试（含并发测试）
├── dict.go         # Dict：sync.Map 的薄封装（本练习的主角）
├── bench_test.go   # 行为 6 的 4 个 benchmark
├── rwmutexmap.go   # RWMutexMap：benchmark 对照组（含构造函数）
└── counter.go      # AtomicCounter / MutexCounter：benchmark 对照组
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `dict_test.go` | 行为 1~5 的全部测试 | `TestDictStoreLoad`、行为 2~4 的测试（命名自定）、`TestDictConcurrentStore` | **第 1 个建** |
| 2 | `dict.go` | `Dict` 的 6 个委托方法 + `Len` | `Dict`、`Dict.Store`、`Dict.Load`、`Dict.LoadOrStore`、`Dict.Delete`、`Dict.Range`、`Dict.CompareAndSwap`、`Dict.Len` | 测试编译报错时 |
| 3 | `bench_test.go` | 行为 6 的 4 个 benchmark | `BenchmarkDictReadMostly`、`BenchmarkRWMutexMapReadMostly`、`BenchmarkAtomicCounter`、`BenchmarkMutexCounter` | 行为 6 开始时 |
| 4 | `rwmutexmap.go` | benchmark 对照组：读写锁 + 普通 map | `RWMutexMap`、`NewRWMutexMap`、`RWMutexMap.Get`、`RWMutexMap.Set` | benchmark 编译报错时 |
| 5 | `counter.go` | benchmark 对照组：两个并发计数器 | `AtomicCounter`、`AtomicCounter.Inc`、`AtomicCounter.Load`、`MutexCounter`、`MutexCounter.Inc`、`MutexCounter.Load` | 同上 |

要实现的方法就是下面契约里的全部：7 + 3 + 4 = 14 个，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**你要写的每一个类型、每一个签名都在下面**，按实现文件分组。
你唯一需要自己实现的是函数体；如果写代码时发现要发明契约之外的类型或函数，说明走偏了。

**写在 `dict.go`：**（需要 `import "sync"`）

```go
package syncmap

// Dict 是 sync.Map 的薄封装：键、值都是 any。
// 零值即可用——var d Dict 之后直接 Store，没有也不需要构造函数，
// 因为内嵌的 sync.Map 本身就零值可用。
// Dict 含锁，第一次使用后禁止拷贝（全部方法用指针接收者，
// go vet 的 copylocks 检查会抓值拷贝）。
// （真实项目里这一层会把 any 收窄成具体类型；本练习保留 any，
//   是为了让每个方法的签名与 sync.Map 一一对应。）
type Dict struct {
	m sync.Map
}

// Store 存入键值对；键已存在则覆盖旧值。
// 键必须是可比较类型（内部就是 map 的键），否则 panic
func (d *Dict) Store(key, value any)

// Load 按键取值：存在 → (值, true)；不存在 → (nil, false)。
// 与普通 map 的 v, ok := m[k] 是同一套 ok-idiom
func (d *Dict) Load(key any) (value any, ok bool)

// LoadOrStore：键已存在 → 返回旧值，loaded=true，本次不写入（传入的 value 被丢弃）；
// 键不存在 → 存入 value，返回 (value, false)。
// "取不到就存"是一步原子操作，不是 Load + Store 两步
func (d *Dict) LoadOrStore(key, value any) (actual any, loaded bool)

// Delete 删除键；键不存在时什么也不发生（不 panic）
func (d *Dict) Delete(key any)

// Range 遍历全部键值对，逐个调用 f；f 返回 false 立即提前终止。
// 顺序不保证，不要依赖顺序写断言
func (d *Dict) Range(f func(key, value any) bool)

// CompareAndSwap（sync.Map 自 Go 1.20 起支持）：
// 键存在且当前值 == old → 替换为 new，返回 true；
// 否则（含键不存在）原样不动，返回 false。
// old 和当前值都必须可比较，否则 == 在运行时 panic
func (d *Dict) CompareAndSwap(key, old, new any) bool

// Len 返回键值对总数。sync.Map 没有内置 Len——这是它的知名取舍，
// 只能用 Range 全表数一遍（O(n)）
func (d *Dict) Len() int
```

**写在 `rwmutexmap.go`：**（需要 `import "sync"`）

```go
// RWMutexMap 是手工版"读写锁 + 普通 map"，sync.Map 的 benchmark 对照组。
// 与 Dict 相反：零值不可用——m 是 nil map，直接写会 panic，必须走构造函数。
// 只以指针形式传递和使用
type RWMutexMap struct {
	mu sync.RWMutex
	m  map[string]int
}

// NewRWMutexMap 返回初始化好的 *RWMutexMap（m 已 make）
func NewRWMutexMap() *RWMutexMap

// Get 读（RLock）：存在返回 (值, true)，否则 (0, false)
func (m *RWMutexMap) Get(key string) (int, bool)

// Set 写（Lock）：键已存在则覆盖
func (m *RWMutexMap) Set(key string, value int)
```

**写在 `counter.go`：**（需要 `import "sync"` 和 `"sync/atomic"`）

```go
// AtomicCounter 是基于 atomic.Int64 的并发计数器，零值可用
type AtomicCounter struct {
	n atomic.Int64
}

// Inc 原子加一
func (c *AtomicCounter) Inc()

// Load 原子读当前值
func (c *AtomicCounter) Load() int64

// MutexCounter 是 Mutex 保护的并发计数器，AtomicCounter 的对照组。
// 零值可用——mu 是值字段（uber·零值 Mutex），不是 *sync.Mutex。
// 只以指针形式传递和使用
type MutexCounter struct {
	mu sync.Mutex
	n  int64
}

// Inc 加锁加一
func (c *MutexCounter) Inc()

// Load 加锁读当前值
func (c *MutexCounter) Load() int64
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 4 个类型：`Dict`、`RWMutexMap`、`AtomicCounter`、`MutexCounter`
- 14 个函数/方法：`Dict` 的 7 个（`Store`、`Load`、`LoadOrStore`、`Delete`、`Range`、`CompareAndSwap`、`Len`）、
  `NewRWMutexMap` 加 `RWMutexMap` 的 2 个（`Get`、`Set`）、两个计数器各 2 个（`Inc`、`Load`）
- 0 个哨兵错误/常量：本练习不涉及——并发不安全的暴露方式是 panic 或 DATA RACE，不是错误返回值

---

## 二、任务单

每个行为 = 一轮完整的 RED → GREEN，**先把测试写出来再实现**：先写测试，编译失败即
RED，再写最少实现变绿。行为 1 从 `TestDictStoreLoad` 开始——零值 `var d Dict`，
`Store("lang", "go")` 后分别断言 `Load` 命中（返回 `(值, true)`）与未命中
（返回 `(nil, false)`）；跑 `go test ./tdd/syncmap`，先看到 `undefined: Dict`
的编译失败（RED），再照契约在 `dict.go` 里写出让它通过的最少代码（`Dict` 内嵌
`sync.Map`，`Store`/`Load` 各一行委托，`Len` 等其余方法等对应行为的测试编译报错时
再补）。之后每个行为如法炮制：测试先行，实现跟上。

### 行为 2：LoadOrStore——"取不到就存"是一步原子操作

| 用例 | 前置状态 | 调用 | 期望（逐条断言） |
|---|---|---|---|
| 键不存在则存入 | 空 `Dict` | `actual, loaded := d.LoadOrStore("a", 1)` | ① `actual == 1` ② `loaded == false` ③ 之后 `Load("a")` 得 `(1, true)` |
| 键已存在返回旧值 | 已 `Store("a", 1)` | `d.LoadOrStore("a", 999)` | ① `actual == 1`（旧值）② `loaded == true` ③ 再 `Load("a")` 仍是 `1`——999 被丢弃 |

提示：并发原语的测试不只验返回值，还要验**最终状态**——每个用例第 ③ 条都回头
`Load` 一次，确认 map 里的实际内容。

### 行为 3：Delete 与 Range

| 用例 | 操作序列 | 期望 |
|---|---|---|
| Delete 后 Load 失败 | `Store("a",1)` → `Delete("a")` → `Load("a")` | 返回 `(nil, false)` |
| Delete 不存在的键 | 空 `Dict` 上 `Delete("ghost")` | 不 panic；`Len()` 仍为 0 |
| Range 收集全部 | 存入 `a/1 b/2 c/3`，Range 回调里断言类型后收集进 `map[string]int` | 与 `map[string]int{"a":1,"b":2,"c":3}` 完全一致（`reflect.DeepEqual`） |
| Range 提前终止 | 同上三对，回调里计数并 `return false` | 回调只被调用 **1** 次；收集到的条目少于 3（具体是哪条不保证） |

提示：`Len()` 在这一步的测试里第一次被调用——它也要等你写出来（用 Range 数一遍）。
Range 遍历顺序不保证：回调里把键值收集进普通 map、最后整体比较，不要按顺序断言。

### 行为 4：CompareAndSwap（Go 1.20+）——旧值匹配才替换

| 用例 | 前置状态 | 调用 | 期望 |
|---|---|---|---|
| 旧值匹配则替换 | 已 `Store("n", 1)` | `d.CompareAndSwap("n", 1, 2)` | 返回 `true`；`Load("n")` 得 `(2, true)` |
| 旧值不匹配不动 | 已 `Store("n", 1)` | `d.CompareAndSwap("n", 999, 2)` | 返回 `false`；`Load("n")` 仍是 `1` |
| 键不存在 | 空 `Dict` | `d.CompareAndSwap("n", 1, 2)` | 返回 `false`；`Load("n")` 得 `(nil, false)` |

### 行为 5：并发行为测试——10 个 goroutine 写不相交的 key

目标：起 10 个 goroutine，各自往 `Dict` 写 100 个互不相交的 key（goroutine g 只写
`g<g>-k*` 区间，读写集互不重叠），`wg.Wait()` 之后断言最终键总数 == 1000
（用 `Len()`）。并发测试的骨架写法（WaitGroup 三步、闭包传参）见第三节
「行为 5」知识点，照着模式自己把 `TestDictConcurrentStore` 写出来。

需要的 import：`fmt`、`sync`、`testing`。
跑 `go test ./tdd/syncmap -race -run TestDictConcurrentStore -v`：全绿且没有 `WARNING: DATA RACE` 才算过。

### 机制实验（做完行为 5 后必做）

新建一个临时测试文件，把同样的并发写场景改成裸 `map[string]int`（不加任何锁），跑 `-race`：
你会看到 `WARNING: DATA RACE` 报告，或直接 `fatal error: concurrent map writes`——
两种结局都证明普通 map 并发写不安全。看完把临时文件删掉，体会 `-race` 把
"偶发玄学 bug"变成"每次必现的报告"的价值。

### 行为 6：Benchmark 对照——用数据回答"什么时候该用 sync.Map"

| Benchmark | 被测对象 | 场景 |
|---|---|---|
| `BenchmarkDictReadMostly` | `Dict` | 预填 100 个键；`b.RunParallel` 里每 20 次操作 19 次读 1 次写 |
| `BenchmarkRWMutexMapReadMostly` | `RWMutexMap` | 同上一模一样的场景 |
| `BenchmarkAtomicCounter` | `AtomicCounter.Inc` | `b.RunParallel` 纯递增 |
| `BenchmarkMutexCounter` | `MutexCounter.Inc` | 同上 |

四个 benchmark 的代码都自己写：先预填数据（计数器组不用），再用 `b.RunParallel`
并行施压，按上表场景控制读写比例——并行 benchmark 的写法示例见第三节
「行为 6」知识点，其余三组照换被测对象。

需要的 import：`strconv`、`testing`。
跑 `go test ./tdd/syncmap -bench=. -benchmem -run '^$'`，把两组的 ns/op 抄下来对比。

---

## 三、知识点总结

各行为用到的知识点按行为归置在前面，随后是机制图解、速查表与书目对应。

### 行为 1：Store 与 Load——ok-idiom 读写

1. **零值可用的原理**：`sync.Map` 的内部字段全是指针、map 引用和锁（read 是原子持有的指针、dirty 是 map 引用、mu 零值即未锁状态），零值时它们为 nil/空，首次写入时再惰性分配——所以不需要任何构造函数。对比普通 map：零值是 nil map，**读**返回零值但**写**直接 panic（`assignment to entry in nil map`），必须 `make`。行为 6 的 `RWMutexMap` 内含普通 map，所以它必须有构造函数——同一个练习里两种类型对照着记。
2. **ok-idiom 统一**：`Load` 返回 `(value, ok)`，与普通 map 的 `v, ok := m[k]` 是同一套肌肉记忆——不像某些语言的并发容器用 nil/异常表示"不存在"。区别只在值是 `any`：ok=false 时拿到的是 `nil`（any 的零值），不是某个具体类型的零值。
3. **any 换来的代价**：键和值都是 `any`，编译器完全不检查类型，两条铁律：①**键必须可比较**——内部就是 map 的键，`Store` 一个切片键直接 panic（`hash of unhashable type`）；②读出来要**类型断言**才能当具体类型用。这是"通用容器"的账单，行为 3、4 还会再付两期。
4. **不能拷贝**：`Dict` 内含 `sync.Map`（进而内含 `Mutex`），第一次使用后拷贝它等于复制锁的当前状态——所以全部方法用指针接收者；真不小心值拷贝了，`go vet`（copylocks）会报错。
5. **RED 复习**：`undefined: Dict` 是编译失败型 RED（testbasic 的第一课）；"最少实现"就是每个方法一行委托——这层封装没有任何逻辑可写错，逻辑都在 `sync.Map` 里，这正是本练习"机制学习"的含义。

### 行为 2：LoadOrStore——"取不到就存"是一步原子操作

1. **两个返回值的名字即语义**：官方签名是 `(actual, loaded)`。`loaded=true` 表示"这次实际是 Load"——值是取出来的旧值，你传的 value 没写进去；`loaded=false` 表示"这次实际是 Store"——值是你刚放的。一个方法两种角色，`loaded` 告诉你这次演了哪出。
2. **它解决的竞态叫 check-then-act**：并发下"先 `Load` 判断没有、再 `Store` 写入"是两步，中间别的 goroutine 可能已经写了——结果要么覆盖别人的值，要么重复初始化。`LoadOrStore` 把检查+写入合成**一步原子操作**。典型用途：并发惰性初始化（第一个到的 goroutine 建对象，后面的人全用现成的）、并发去重集合。
3. **易错点**：`loaded=true` 时你构造的 value 被**丢弃**。value 构造昂贵时（大对象、占资源）要意识到这份浪费。本练习值是 int，无感。
4. **断言要点**：并发原语的测试不只验返回值，还要验**最终状态**——所以每个用例第 ③ 条都回头 `Load` 一次，确认 map 里的实际内容。

### 行为 3：Delete 与 Range

1. **Range 回调的 bool 是"继续吗"**：返回 `false` 立即停止遍历（效果类似 break），不是错误信号。"提前终止"用例断言的是**回调被调用的次数**，不是内容。
2. **顺序不保证**：Range 的遍历顺序未定义。正确姿势：回调里把键值收集进普通 map，最后整体比较；错误姿势：假设"第一个遍历到的是 a"——按顺序断言的测试会随机挂。
3. **非一致快照**：Range 期间别的 goroutine 增删，结果可能看到也可能看不到，但同一个键不会被访问两次。并发代码里不要依赖 Range 的"实时性"。（实现上 Range 开始前可能先做一次 dirty 提升，见下文「机制图解」。）
4. **Delete 是标记删除**：Delete 后 `Load` 立刻失败——逻辑上已删；但内部条目可能还在内存里躺着，等批量清理，所以 Delete 很便宜。机制见下文「机制图解」的延迟删除一段。
5. **any 的账单第二期**：回调拿到的是 `(any, any)`，要 `k.(string)`、`v.(int)` 断言回具体类型；测试里用逗号-ok 断言（`s, ok := k.(string)`）更稳，断言失败用 `t.Fatal` 而不是让它 panic。

### 行为 4：CompareAndSwap（Go 1.20+）——旧值匹配才替换

1. **CAS = 乐观锁一句话**："我以为当前值是 old；真是，就换成 new；不是，就告诉我失败，我什么都不动。"比较和交换是**一步原子操作**，对 read 里的键甚至无锁。对比 Mutex 的悲观思路：先独占，再操作。
2. **典型用法是重试循环**：`for { old, _ := d.Load(k); if d.CompareAndSwap(k, old, old.(int)+1) { break } }`——失败说明有人抢先改了，读到新值再试。无锁计数器、并发状态机都这么写。行为 6 会用 `atomic.Int64` 实现同样的事，到时候对比代码量。
3. **易错点：值必须可比较**：判断"当前值是否等于 old"用的是 `==`，`any` 里装了切片/map/函数这类不可比较类型时，`==` 在**运行时 panic**，编译器不报错——any 的账单第三期。
4. **版本背景**：`CompareAndSwap`、`CompareAndDelete`、`Swap` 是 Go 1.20 才加进 `sync.Map` 的，更早只有 Load/Store/LoadOrStore/Delete/Range 五个方法。本仓库 Go 1.26，放心用。

### 行为 5：并发行为测试——10 个 goroutine 写不相交的 key

1. **`-race` 的原理一句话**：编译时给每次内存访问插入记录，运行时维护访问之间的 happens-before 关系；发现两个 goroutine **无同步地**访问同一地址且至少一个是写，就打印 `WARNING: DATA RACE` 并列出两个访问点的栈。代价是数倍的变慢和内存放大（官方典型值：内存 5~10 倍）——所以只用于测试/预发，生产不开。
2. **WaitGroup 三步铁律**：`Add` 必须在 goroutine 启动**之前**（循环内先 `Add(1)` 再 `go`，或循环外一次 `Add(10)`）；goroutine 里 `defer wg.Done()`；主 goroutine `wg.Wait()`。把 `Add` 放进 goroutine 体内是经典 bug——`Wait` 可能先于 `Add` 执行直接返回，测试失去等待。并发测试骨架示例：

```go
func TestDictConcurrentStore(t *testing.T) {
	var d Dict
	var wg sync.WaitGroup

for g := 0; g < 10; g++ {
		wg.Add(1) // Add 在 go 之前
		go func(g int) {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				key := fmt.Sprintf("g%d-k%d", g, k) // goroutine 各自写自己的 key 区间
				d.Store(key, k)
			}
		}(g)
	}
	wg.Wait()

// 断言：最终键总数 == 1000（用 Len()）
}
```

3. **循环变量与闭包**：Go 1.22 起 for 循环变量每次迭代都是新变量，闭包直接捕获 `g` 也不会串号；上面示例里仍用 `go func(g int)` 显式传参——意图更清晰，也是旧代码里的常见写法，两种都要认得。
4. **不相交 key 集的用意**：这是教程 ch18 指出的 `sync.Map` 优势场景之二——每个 goroutine 只写自己的 key 区间（`g0-k*` 到 `g9-k*`），读写集互不重叠时内部提升开销最小。最终 `Len()==1000` 既验证正确性（一个键都没丢），也为行为 6 的场景设计埋下伏笔。
5. **`Len()` 在这里第一次派大用场**：它是 Range 数出来的 O(n)。1000 个键数一次很快，但要知道这不是 `len(m)` 那种 O(1)——sync.Map 为快读付出的代价之一就是放弃了廉价计数。

### 行为 6：Benchmark 对照——用数据回答"什么时候该用 sync.Map"

1. **`b.RunParallel`**：默认起 GOMAXPROCS 个 goroutine 并行跑被测体，每个 goroutine 各跑自己的 `for pb.Next()` 循环——专门测并发吞吐和锁竞争。串行的 `b.N` 循环测不出竞争：锁没人抢，永远显得很快。读多写少场景的 benchmark 写法示例（其余三组照换被测对象）：

```go
func BenchmarkDictReadMostly(b *testing.B) {
	var d Dict
	for i := 0; i < 100; i++ {
		d.Store(strconv.Itoa(i), i)
	}
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i % 100)
			if i%20 == 0 {
				d.Store(key, i)
			} else {
				d.Load(key)
			}
			i++
		}
	})
}
```

2. **场景设计即结论**：本对照的关键参数是"100 个固定键、95% 读"——键集稳定且读多写少，`sync.Map` 的读全部落在无锁的 read 上。把写比例调到 50% 再跑一次，观察差距缩小甚至反转——benchmark 的参数就是"适用场景"的量化，别只跑一次就下结论。
3. **`-run '^$'` 小技巧**：只跑 benchmark 不跑普通测试——正则 `^$` 匹配不到任何 `TestXxx`。串行 benchmark 的现代写法是 `b.Loop()`（见 testbasic 行为 3），并行场景仍用 `pb.Next()`。
4. **atomic 组为什么总是赢**：`Inc` 编译成一条 CPU 原子指令（x86 上是 LOCK XADD），无锁、无 goroutine 阻塞；Mutex 版至少两次原子操作（加锁+解锁）外加竞争时的挂起/唤醒。临界区越短差距越夸张——这就是 uber"使用 atomic"条目的实证：单个数值能用 atomic 表达，就别用 Mutex。
5. **atomic 的边界**：`atomic.Int64` 只能保护**这一个** int64；不变量一旦跨多个字段（"a 和 b 必须一起改"），原子类型救不了你，必须回到 Mutex。本组两个计数器能力完全相同才可比——别把结论误推广成"到处用 atomic"。
6. **uber·零值 Mutex 落地**：`MutexCounter` 的 `mu` 是**值字段**不是指针（`*sync.Mutex` 零值是 nil，`Lock` 直接 panic）；`RWMutexMap`、`MutexCounter` 都只以指针形式传递（构造函数返回 `*RWMutexMap`），杜绝值拷贝复制锁状态——`go vet` 的 copylocks 会替你检查。
7. **诚实看数据**：不同机器结果会有差异。典型结果是 Dict 组 ns/op 低于 RWMutexMap 组、atomic 组显著低于 Mutex 组；若你机器上反转，先检查写比例和键数——这正是"适用场景判断"要亲手量而不是背的原因。（`Dict` 的方法委托会被编译器内联，测到的就是 `sync.Map` 本身；好奇可跑 `go build -gcflags='-m' ./tdd/syncmap` 看内联报告。）

### 机制图解：read / dirty 双结构（教程 ch18 核心）

`sync.Map` 的内部不是一个加了锁的 map，而是**两个 map 加一套状态机**：

```text
sync.Map
├── read   原子持有的只读快照（readOnly{m, amended}）   ← 读优先走这里，无锁
├── dirty  可写的 map[any]*entry                        ← 新键写这里，需 mu 保护
├── misses int                                          ← read 未命中的累计次数
└── mu     sync.Mutex                                   ← 只保护 dirty 和提升/重建
```

read 和 dirty 的 value 都是 `*entry`，指向**同一份条目对象**——所以"改已存在的键的值"
只改 entry（CAS），两个 map 同时可见，连锁都不用加。

**读路径**：先原子加载 read 查键 → 命中，完事（全程无锁）。未命中且 `amended=true`
（说明 dirty 里有 read 没有的键）→ 加锁查 dirty，`misses+1`。

**misses 提升**：当 `misses` 累计到 ≥ `len(dirty)`，说明"新键老是绕过 read 被找"，
就把 dirty 整体提升为新 read——一次原子指针替换，O(1)；dirty 置 nil、misses 清零。
经常访问的键就这样"沉淀"进 read，读路径重新全无锁。

**写路径**：键已在 read → CAS 改 entry，无锁。键不在 → 加锁写 dirty；若 dirty 为 nil，
还要先从 read 把存活条目**复制**一份建 dirty（O(n)）——注意：提升是 O(1)，重建 dirty
才是复制发生的地方。

**延迟删除（本练习最重要的思想）**：`Delete` 一个 read 里的键时，**不做** map 条目移除，
只把 entry 的值 CAS 标记为 nil——逻辑上已删（`Load` 立刻失败），物理上还留在内存里。
真正的清除推迟到下次提升/重建 dirty 时**批量**完成。

**这就是"空间换时间"**：宁可让已删除的条目在内存里多躺一会儿（付出空间），也不为每次
删除支付加锁+复制 map 的 O(n) 成本（省下时间）；清理攒起来批处理，把成本摊销到少数
几次提升里。同一个思想在别处反复出现：GC 的标记-清除、LSM 树的墓碑（tombstone）
删除、切片的容量预分配——"标记 + 惰性批量处理"是系统设计的通用手法。

**代价清单**（选型时必须知道）：没有 `Len`、没有"清空"、Range 开始前可能先触发一次
提升、写多/键集剧变时 dirty 反复重建复制，可能比 RWMutex+map 还慢。

### sync.Map 方法速查

| 方法 | 语义 | 注意 |
|---|---|---|
| `Store(k, v)` | 存/覆盖 | 键必须可比较，否则 panic |
| `Load(k)` | `(v, true)` / `(nil, false)` | ok-idiom，与普通 map 统一 |
| `LoadOrStore(k, v)` | 存在→`(旧, true)`；否则存→`(v, false)` | 一步原子；loaded=true 时传入的 v 被丢弃 |
| `Delete(k)` | 删除；键不在也无事 | 内部是标记删除，物理清除推迟 |
| `Range(f)` | 遍历；`f` 返 false 提前终止 | 顺序不保证；非一致快照；可能先触发提升 |
| `CompareAndSwap(k, old, new)` | 当前值 == old 才替换 | Go 1.20+；值必须可比较，否则 panic |
| （没有）`Len` | —— | 要计数只能 Range 全表，O(n) |

### 选型速查：sync.Map vs RWMutex+map vs atomic

| 场景 | 选谁 | 理由 |
|---|---|---|
| 读多写少、键集稳定（缓存） | `sync.Map` | 读走 read，全程无锁 |
| 多 goroutine 各写各的 key 区间 | `sync.Map` | 不相交 key 集，提升频率低（行为 5 的场景） |
| 写多 / 需要 Len / 需要一致快照 | `RWMutex+map` | dirty 重建要复制；`len(m)` 是 O(1) |
| 单个计数器 / 标志位 | `atomic.Int64` 等 | 原子读改写，避免 Mutex；具体指令和成本随架构与竞争情况变化 |
| 跨多个字段的不变量 | `Mutex` | atomic 只能保护一个值 |

官方文档只推荐前两种场景用 `sync.Map`——行为 6 的数据就是这两条结论的实证。

### 零值可用与 uber 两条规则

- **零值可用**是标准库的刻意设计：本练习里 `sync.Map`、`Mutex`、`RWMutex`、
  `WaitGroup`、`atomic.Int64` 全都零值可用；普通 map 不是（nil map 写 panic）。
  判断一个自定义类型零值可不可用，看它内含的字段——含普通 map，就得提供构造函数。
- **uber·零值 Mutex**：Mutex 内嵌为**值字段**（`mu sync.Mutex`），不要 `*sync.Mutex`
  （零值是 nil，`Lock` 即 panic，违背零值可用）；含锁结构体只传指针，值拷贝会复制锁
  状态——`go vet` 的 copylocks 会抓。
- **uber·使用 atomic**：单数值并发操作优先 `atomic.Int64` 这类**类型化**原子值
  （Go 1.19+），而不是 `atomic.AddInt64(&n)` 自由函数——类型把"这个字段只能原子
  访问"变成编译器强制；裸 `&int64` 谁都能普通读写，等于没保护。

### 与书目的对应

- 教程 ch18 Sync.Map 全章（read/dirty、misses 提升、延迟删除、适用场景）——
  本练习行为 1~6 与上文「机制图解」全部落地
- uber 规范·零值 Mutex——行为 6 的 `RWMutexMap`/`MutexCounter` 写法
- uber 规范·使用 atomic——行为 6 的 `AtomicCounter` 与对照结论
- 前置回顾：[basic/goroutine_lock.go](../../basic/goroutine_lock.go)（Mutex/RWMutex 基础）、
  [tdd/testbasic](../testbasic/README.md)（benchmark 工具）

---

## 四、验收标准

```bash
go test ./tdd/syncmap -v                             # 全绿
go test ./tdd/syncmap -race                          # 并发练习必跑：无任何 DATA RACE
go vet ./tdd/syncmap                                 # 无警告（copylocks 也在其中）
go test ./tdd/syncmap -cover                         # Dict 的方法应全覆盖
go test ./tdd/syncmap -bench=. -benchmem -run '^$'   # 两组对照数据都跑得出
```

`-race` 需要 cgo 可用（Windows 上装有 gcc/MinGW 即可）；若报
"race detector not supported"，先查 `go env CGO_ENABLED`。

## 五、完成后自查（能口头回答才算过）

1. `LoadOrStore` 的 `loaded` 两个取值各代表这次实际做了什么？它解决的
   check-then-act 竞态用一句话说清。
2. read 和 dirty 的分工是什么？misses 累计到阈值时发生什么，为什么之后读重新变快？
3. 延迟删除：`Delete` 当下做了什么、把什么推迟了？"空间换时间"换掉的和付出的
   各是什么？
4. 官方推荐 `sync.Map` 的两种场景是什么？写多场景为什么可能输给 RWMutex+map？
   （用你行为 6 跑出的数据回答）
5. 为什么 `Dict`、`MutexCounter` 零值可用，而 `RWMutexMap` 必须有构造函数？
   nil map 写入的下场是什么？
6. uber 为什么要求 Mutex 用值字段内嵌？`atomic.Int64` 比 `atomic.AddInt64`
   自由函数"类型安全"在哪？
7. `-race` 的检测原理一句话？为什么只用于测试环境？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把 ch18"Sync.Map"从 ❌ 划掉（改为 ✅）。
