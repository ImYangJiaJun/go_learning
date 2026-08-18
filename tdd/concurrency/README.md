# Concurrency TDD —— 并发机制刻画：close、同步、超时与 WaitGroup

目标：并发轨道（阶段 3）第一站。要实现的代码故意极少（只有一个 `FetchWithTimeout`），
**注意力全部放在并发原语的语义本身**：close 三规则、无缓冲 channel 的同步性、
`time.After` 超时模式、WaitGroup 陷阱与 Go 1.25 的 `wg.Go`。全程 `-race` 必须通过。

> 本任务是**机制学习型**练习：接口契约已固定（仅行为 3 一个函数；行为 1/2/4 用测试
> 刻画 Go 语言本身的语义，没有产品代码可设计）。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/concurrency` 就是它的运行方式，验收者是测试，不是人。

这个包做两件事：

- 用测试**刻画** Go 并发原语的既有语义：close 三规则（行为 1）、无缓冲 channel 的
  同步性（行为 2）、WaitGroup 计数规则（行为 4）——这些测试没有对应的产品代码，
  测试本身就是产出物
- 实现一个超时等待工具 `FetchWithTimeout`（行为 3）——本练习唯一的产品函数

包名用 `concurrency`：本练习只导入 `testing`/`time`/`sync`，与标准库无重名，不需要改名。

### 文件计划（共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `concurrency_test.go` | 全部测试（行为 1/2/4 只测 Go 语义，不需要产品代码） | **第 1 个建** |
| `timeout.go` | `FetchWithTimeout`（本练习唯一要实现的函数） | 行为 3 测试编译报错时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package concurrency

import "time"

// FetchWithTimeout 在时限 d 内等待 ch 上的值：
// 收到则返回 (值, true)；超时返回 ("", false)，d 从调用时刻起算。
// ch 是只读定向 channel——接收方无权发送，更无权关闭它（关闭权属于发送方）。
func FetchWithTimeout(ch <-chan string, d time.Duration) (string, bool)
```

### 第一步：手把手起步（行为 1：close 三规则）

1. 在 `tdd/concurrency/` 下新建 `concurrency_test.go`，写入（本练习唯一给全代码的行为）：

```go
package concurrency

import "testing"

// 规则 1：从已关闭的 channel 接收，立即得到零值，ok == false（ok-idiom）
func TestReceiveFromClosedChannel(t *testing.T) {
	ch := make(chan string)
	close(ch)

	got, ok := <-ch
	if got != "" {
		t.Errorf("已关闭 channel 应返回零值 \"\"，得到 %q", got)
	}
	if ok {
		t.Error("已关闭 channel 的 ok 应为 false")
	}
}

// 规则 2：重复 close 会 panic（close of closed channel）
func TestDoubleClosePanics(t *testing.T) {
	ch := make(chan int)
	close(ch)

	defer func() {
		if recover() == nil {
			t.Error("重复 close 没有 panic")
		}
	}()
	close(ch) // 期望这一行 panic
}

// 规则 3：向已关闭的 channel 发送会 panic（send on closed channel）
func TestSendOnClosedChannelPanics(t *testing.T) {
	ch := make(chan int, 1) // 带缓冲：排除"无人接收"的干扰，panic 只能来自"已关闭"
	close(ch)

	defer func() {
		if recover() == nil {
			t.Error("向已关闭 channel 发送没有 panic")
		}
	}()
	ch <- 1 // 期望这一行 panic
}
```

2. 运行 `go test ./tdd/concurrency` → **直接全绿**。这不是漏了步骤：行为 1/2/4 刻画的
   是 Go 语言本身的语义，不依赖任何待实现函数，这类"刻画测试"没有 RED 阶段。
3. 本练习的 RED 在行为 3：第一次运行它的测试会**编译失败** `undefined: FetchWithTimeout`
   ——编译失败就是 RED；然后写最少实现到变绿（详见行为 3）。

---

## 二、任务单（边做边学）

### 行为 1：close 三规则（测试代码已在"第一步"完整给出）

| 用例 | 操作 | 具体断言 |
|---|---|---|
| 已关闭后接收 | `close(make(chan string))` 后 `got, ok := <-ch` | `got == ""`（零值）且 `ok == false` |
| 重复 close | 对已关闭的 channel 再次 `close(ch)` | recover 到 panic（`close of closed channel`） |
| 向已关闭发送 | `close(ch)` 后 `ch <- 1` | recover 到 panic（`send on closed channel`） |

**这一步用到的知识点：**

1. **close 语义速查全表**（本练习的根基，背下来）：

| 操作 | 对象状态 | 结果 |
|---|---|---|
| 接收 `<-ch` | 已关闭、缓冲已读空 | 立即返回零值，`ok == false` |
| 接收 `<-ch` | 已关闭、缓冲还有数据 | 照常读出数据，`ok == true`；读空后才出现上一行 |
| 发送 `ch <- v` | 已关闭（无论缓冲满不满） | panic: `send on closed channel` |
| `close(ch)` | 已关闭 | panic: `close of closed channel` |
| `close(ch)` | nil channel | panic: `close of nil channel` |
| 发送 / 接收 | nil channel | 永远阻塞（**不是 panic**；select 里可用 nil 分支禁用某个 case） |

原则一句话：**关闭是发送方的职责**，接收方靠 ok-idiom 或 `for range` 感知关闭，
绝不反过来（接收方一 close，发送方下次发送就 panic）。

2. **ok-idiom 与 map 取值的统一感**：`v, ok := <-ch` 和 `v, ok := m[k]`、
   `v, ok := x.(T)`（类型断言）完全同构——comma-ok 是 Go 通用的"拿到了吗"协议。
   差异只在语义：map 的 `false` 表示 key 不存在；channel 的 `false` 表示"已关闭
   **且缓冲读空**"。`for v := range ch` 是它的语法糖：一直取到关闭后自动退出循环——
   所以 for range 之前必须有人负责 close（[basic 易错点 12](../../basic/goroutine_channel.go#L113)）。
3. **断言 panic 的写法原理**：panic 会让当前 goroutine 的调用栈逐层展开、沿途执行
   defer；只有**在 defer 函数里直接调用** `recover` 才能截停它
   （[basic/panic_recover.go](../../basic/panic_recover.go)）。`recover` 的返回值是
   panic 的值；若 goroutine 根本没 panic，`recover` 返回 `nil`——所以骨架固定是：

```go
defer func() {
	if recover() == nil {
		t.Error("期望 panic，但没有发生")
	}
}()
// 期望 panic 的那一行写在这里
```

   易错点：`recover` 只能捕获**同一个 goroutine** 的 panic；测试里 `go` 出去的
   goroutine 若 panic，会直接崩掉整个测试进程（对照 [basic/goroutine.go](../../basic/goroutine.go)
   的"协程内 panic 须自己 recover"）。所以 panic 断言必须写在同一个 goroutine 里。
4. **为什么"向已关闭发送"的测试要用缓冲 channel**：`make(chan int, 1)` 缓冲有空位，
   若没有 panic 发送会成功返回——所以 panic 一旦发生必然来自"已关闭"，排除了
   "无人接收导致阻塞"这条干扰路径。设计测试时要想清楚：除了被测机制，还有没有别的
   路径能产生同样的现象。

### 行为 2：无缓冲 channel 的同步语义——证明发送方会一直阻塞

| 用例（场景） | 操作序列 | 具体断言 |
|---|---|---|
| 接收方就绪前，发送一直阻塞 | goroutine 执行 `ch <- "x"` 后 `close(sent)`；主测试**故意不接收** | select `sent` vs `time.After(50ms)`：必须走到 time.After 分支（50ms 内 sent 未被 close） |
| 接收之后，发送立刻完成 | 主测试 `<-ch` 收到 `"x"`，然后再次 select | 必须走到 `sent` 分支（接收后发送方立即解除阻塞；超时保险放宽到 1s） |

骨架如下，**两个 select 按用例表自己写**：

```go
func TestUnbufferedChannelBlocksUntilReceived(t *testing.T) {
	ch := make(chan string)     // 无缓冲
	sent := make(chan struct{}) // 观察"发送是否已完成"的事件 channel

	go func() {
		ch <- "x"
		close(sent) // 只有值被取走后才执行得到
	}()

	// 第一幕：还没人接收——select sent vs time.After(50ms)，必须超时

	// 第二幕：<-ch 接收并断言值为 "x"——再 select，sent 必须就绪
}
```

**这一步用到的知识点：**

1. **无缓冲 channel 是会合点（rendezvous）**：发送方和接收方必须同时到场才成交；
   值被取走之前，发送方一直阻塞。本测试正是利用这一点——`close(sent)` 的执行时机
   就是"阻塞解除"的可观察证据。对照缓冲 channel：只要缓冲有空位，发送立即返回，
   发送方**不知道**值何时被取走。
2. **并发测试的核心手法：把内部状态变成 channel 事件**。测试无法直接断言"某
   goroutine 正在阻塞"，只能观察副作用。这里用第二个 channel `sent` 把"发送完成"
   变成可被 select 捕获的事件——以后所有并发测试都是这个思路。
3. **正/负断言的超时取舍**：断言"不该发生"（第一幕）用短超时 50ms——等再久也不该
   发生，50ms 已远超调度延迟；断言"必须发生"（第二幕）用宽松超时（1s）——正常路径
   立即通过，长超时只是 CI 抖动保险，不拖慢测试。方向用反了，测试要么慢要么抖。
4. **内存模型依据（happens-before）**：channel 上的发送 happens-before 对应的接收
   完成；对无缓冲 channel，接收完成又 happens-before 发送方解除阻塞。正是这两条保证
   接收方能读到发送方写入的数据、发送方能确定值已被取走——它也是 `-race` 判定数据
   竞争用的那把尺子。

### 行为 3：超时模式 `FetchWithTimeout`（本练习的 RED 在这里）

| 用例名 | ch 的行为 | d | 期望返回 |
|---|---|---|---|
| 超时返回零值和 false | goroutine 50ms 后才发送 `"late"` | `10 * time.Millisecond` | `("", false)` |
| 时限内取到值 | goroutine 50ms 后才发送 `"late"` | `1 * time.Second` | `("late", true)` |

骨架如下，**表驱动两行用例按上表自己补**：

```go
// 辅助：返回一个 delay 后才产出 v 的 channel
func slowCh(delay time.Duration, v string) chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(delay)
		ch <- v
	}()
	return ch
}

func TestFetchWithTimeout(t *testing.T) {
	// 表驱动：建议字段 delay / d / want / wantOK，按用例表写两行
	// 循环里调用 FetchWithTimeout(slowCh(c.delay, "late"), c.d) 并断言两个返回值
}
```

加菜（可选）：超时用例里记录调用前后时刻，断言耗时明显小于 50ms——证明它确实
在 10ms 超时返回，而不是傻等值到达。

步骤：先写测试 → `go test ./tdd/concurrency -run TestFetchWithTimeout` → **编译失败
`undefined: FetchWithTimeout`，这就是 RED** → 新建 `timeout.go`，按契约写最少实现
（一个 `select`、两个 `case`）→ 变绿。

**这一步用到的知识点：**

1. **`select` 的语义**：所有 case 一起等，谁先就绪执行谁；多个同时就绪**随机**选一个；
   都不就绪则阻塞（有 `default` 则立即走 default）。`time.After(d)` 返回一个
   `<-chan Time`，d 时刻收到一个值——把它放进 select 的一个 case，就是标准的
   "超时分支"写法。
2. **超时 ≠ 取消**（为 `tdd/context` 练习铺垫）：超时是**被动等闹钟**——d 到期前你
   无法提前叫停这次等待，而且 timer 在触发前不会被 GC（d 很大而函数很快返回时，timer
   白活到触发）；取消是**主动发信号**——调用方随时可以说"我不等了"。本练习只用
   `time.After` 做超时；取消信号沿调用树传播，留给 `tdd/context`。
3. **契约里的 `<-chan string` 是只读定向 channel**：函数体内对它发送或 close 都是
   **编译错误**。把方向写进签名，是在向调用方声明"我只消费，不生产、不关闭"——
   关闭权永远属于发送方（[basic/goroutine_channel.go 定向管道](../../basic/goroutine_channel.go#L37)）。
4. **超时返回的代价：泄漏的 goroutine**。超时用例里 `FetchWithTimeout` 在 10ms 就
   返回了，但 `slowCh` 的 goroutine 在 50ms 醒来后会**永远阻塞**在 `ch <- v`（无人
   接收）——测试进程退出时无害，长命服务里这就是 goroutine 泄漏。解法（缓冲 channel
   或 done 信号）在 `tdd/pipeline` 练。先建立直觉：**每次超时都可能留下一个没救的
   goroutine**。

### 行为 4：WaitGroup 陷阱与 Go 1.25 的 `wg.Go`

| 用例名 | 操作 | 具体断言 |
|---|---|---|
| 计数为负 panic | `wg.Add(1)` → `wg.Done()` → 再多调一次 `wg.Done()` | recover 断言第二次多出的 `Done()` panic（`sync: negative WaitGroup counter`） |
| `wg.Go` 等价手写 Add/Done | 对 3 个任务分别 `wg.Go(func(){...})`，主测试 `wg.Wait()` | Wait 返回后 3 个任务的结果全部可见（缓冲 channel 里 `len == 3`） |

骨架如下，**断言按用例表自己补**：

```go
func TestWaitGroupNegativeCounterPanics(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()

	defer func() {
		if recover() == nil {
			t.Error("计数为负没有 panic")
		}
	}()
	wg.Done() // 期望 panic：sync: negative WaitGroup counter
}

func TestWaitGroupGoEquivalentToAddDone(t *testing.T) {
	var wg sync.WaitGroup
	done := make(chan struct{}, 3) // 缓冲 = 任务数，goroutine 不会卡在发送上

	for i := 0; i < 3; i++ {
		wg.Go(func() { // Go 1.25 新 API；可另写一版 Add(1)+go+defer Done() 对照
			done <- struct{}{}
		})
	}
	wg.Wait()

	// 断言 len(done) == 3
}
```

**这一步用到的知识点：**

1. **WaitGroup 是计数器**：`Add(n)` 加 n、`Done()` 减 1、`Wait()` 阻塞到计数归零。
   `Done()` 多于 `Add` → 计数为负 → panic（`sync: negative WaitGroup counter`）。
   归零后 `Wait` 立即放行，WaitGroup 可复用（等下一轮 `Add`）。
2. **Add 的位置陷阱**：`Add` 必须在**启动 goroutine 之前**调用。先 `go` 后 `Add`，
   `Wait` 可能在 `Add` 之前就把计数看成 0、提前返回——这是**逻辑竞态，不是数据
   竞争，`-race` 抓不到**。所以惯用法是在循环里 `wg.Add(1)` 紧挨 `go` 之前。
3. **Go 1.25 的 `wg.Go(f)`**：语义等价于 `wg.Add(1); go func(){ defer wg.Done(); f() }()`。
   它把 Add/Done 的配对责任收进一个调用：不可能忘记 Add（启动即计数）、不可能忘记
   Done（defer 内置）、任务内 panic 也不会漏掉 Done——两类最常见的不匹配错误从机制
   上消失。对照 [basic/goroutine.go 的手写推荐写法](../../basic/goroutine.go#L32)。
4. **Wait 建立 happens-before**：各 goroutine 在 `Done` 之前的所有写，对 `Wait`
   返回之后的读可见——所以 Wait 之后直接读结果（比如 `len(done)`）不需要再加锁。
   这正是本测试能在 `-race` 下通过的原理。

---

## 三、知识点总结

### close 语义速查全表

| 操作 | 对象状态 | 结果 |
|---|---|---|
| 接收 `<-ch` | 已关闭、缓冲已读空 | 立即返回零值，`ok == false` |
| 接收 `<-ch` | 已关闭、缓冲还有数据 | 照常读出数据，`ok == true` |
| 发送 `ch <- v` | 已关闭 | panic: `send on closed channel` |
| `close(ch)` | 已关闭 | panic: `close of closed channel` |
| `close(ch)` | nil channel | panic: `close of nil channel` |
| 发送 / 接收 | nil channel | 永远阻塞（select 里可用来禁用分支） |

一句话：**发送方负责 close，接收方用 ok 感知**。

### ok-idiom 速查

- `v, ok := <-ch`：`ok == false` ⇔ 已关闭且缓冲读空；与 `v, ok := m[k]`、
  `v, ok := x.(T)` 同构——comma-ok 是 Go 通用协议
- `for v := range ch`：取到关闭后自动退出循环；前提是有人负责 close

### 超时 vs 取消（一句话对比）

- 超时（`time.After`）：**被动闹钟**——到期前不可叫停，timer 触发前不被 GC
- 取消（`context`）：**主动信号**——随时可停，沿调用树传播 → 留给 `tdd/context`

### WaitGroup 速查

| 操作 | 语义 | 陷阱 |
|---|---|---|
| `Add(n)` | 计数 +n | 必须在 `go` 之前调用（否则逻辑竞态，`-race` 抓不到） |
| `Done()` | 计数 -1 | Done 多于 Add → 计数为负 panic；手写时用 defer 防遗漏 |
| `Wait()` | 阻塞到计数归零 | Wait 建立 happens-before，之后读结果无需加锁 |
| `Go(f)`（Go 1.25） | Add(1) + go + defer Done 一体 | 从机制上消除 Add/Done 不匹配 |

### `-race` 工作原理（一句话）

编译期给每次内存访问插入记录，运行时发现"两个 goroutine 访问同一地址、至少一个是
写、且两次访问之间没有 happens-before 关系"就报告——所以它是**运行时**检测器：
测试没走到的路径，竞争它发现不了。

### 与书目的对应

- 教程 ch09 前半：channel 的创建 / 收发 / close——行为 1、2 全部落地；
  ch19 深化：sync.WaitGroup——行为 4 落地并深化（计数为负 panic、Go 1.25 `wg.Go`）
- Effective Go·14 并发："Share by communicating"——本练习的 channel 用法就是这个
  思想的最小实例
- 巩固 [basic/goroutine.go](../../basic/goroutine.go)、
  [basic/goroutine_channel.go](../../basic/goroutine_channel.go)、
  [basic/goroutine_lock.go](../../basic/goroutine_lock.go)

---

## 四、验收标准

```bash
go test ./tdd/concurrency -v        # 全绿
go vet ./tdd/concurrency            # 无警告
go test ./tdd/concurrency -race     # 并发练习硬门槛：无数据竞争
go test ./tdd/concurrency -cover    # FetchWithTimeout 的超时/成功两个分支都要走到
```

## 五、完成后自查（能口头回答才算过）

1. close 一个已关闭的 channel、向已关闭 channel 发送、从已关闭 channel 接收，分别
   是什么结果？close 一个 nil channel 呢？向 nil channel 发送呢？
2. `v, ok := <-ch` 的 `ok == false` 确切含义是什么（缓冲里还有数据时会怎样）？
   和 map 取值的 ok 有什么异同？
3. 为什么断言 panic 必须把 `recover` 放在 defer 里？它能捕获别的 goroutine 的
   panic 吗？
4. 无缓冲 channel 上，发送方什么时候解除阻塞？行为 2 的测试是怎么"证明阻塞存在"的？
5. 超时（`time.After`）和取消（`context`）的本质区别是什么？`time.After` 的 timer
   有什么代价？
6. WaitGroup 计数为负会怎样？`Add` 为什么必须放在启动 goroutine 之前？`wg.Go`
   从机制上消除了哪两类错误？
7. `-race` 的检测原理一句话是什么？为什么"-race 通过"不能证明"没有数据竞争"？

全部答清后，回到 [根 README 的 TDD 总目录](../../README.md#四目录划分评估与-tdd-驱动学习计划)，
把 tdd/concurrency（阶段 3·练习 10）的状态从「待建」划掉（改为「✅ 已建」）——它同时为
`tdd/pipeline` 和 `tdd/context` 两个后续练习打好了地基。
