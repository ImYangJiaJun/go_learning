# Pipeline TDD —— 并发管线：扇出/扇入与 goroutine 生命周期

目标：这是 TDD 轨道阶段 3（并发）的第二站。用教程 ch09 的经典管线三件套
（gen → sq → merge）做骨架，把**并发管线的四个核心机制**逐个练透：
扇出/扇入、`close(done)` 广播取消、worker pool、Ticker 限流。
被测函数签名全部固定，注意力全部放在 **channel 关闭语义与 goroutine 退出路径**上。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/pipeline` 就是它的运行方式，验收者是测试，不是人。
这是并发练习，**全部测试必须加 `-race` 通过**。

这个包对外提供四个能力（函数全部小写未导出，测试与被测代码同包，可直接调用）：

1. **gen**：把一批整数变成 channel 数据流（管线的生产入口）
2. **sq**：从 channel 读数、平方后写入新 channel（管线的加工阶段）
3. **merge**：把任意多个 channel 合并成一个（扇入）
4. **workerPool**：n 个 worker 共同消费一个任务 channel，输出每个任务的平方

四个函数共守一条铁律：**收到 `done` 关闭信号后，能立刻停手并关闭自己的输出 channel**。

### 文件计划（共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `pipeline_test.go` | 全部 4 个行为的测试（行为 4 的限流代码直接写在测试里，是模式演示，不进生产代码） | **第 1 个建** |
| `pipeline.go` | `gen` / `sq` / `merge` / `workerPool` 四个函数 | 测试编译报错时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package pipeline

// gen 把 nums 逐个发送到新建的 channel 上并返回它。
// 发送在内部 goroutine 中进行：发完所有数后关闭 out；
// 中途 done 关闭则立即停止发送并关闭 out（谁生产谁关闭）。
func gen(done <-chan struct{}, nums ...int) <-chan int

// sq 从 in 读整数，把平方值发送到新建的 channel 返回。
// in 关闭（自然结束）或 done 关闭（被取消）时，退出并关闭 out。
func sq(done <-chan struct{}, in <-chan int) <-chan int

// merge 把 cs 里任意多个 channel 的数据搬到同一个 channel 返回（扇入）。
// 每个输入 channel 配一个搬运 goroutine；全部搬完（或 done 关闭）后关闭 out。
func merge(done <-chan struct{}, cs ...<-chan int) <-chan int

// workerPool 启动 n 个 worker goroutine 共同从 jobs 读任务，
// 把每个任务的平方值发送到新建的 channel 返回。
// jobs 关闭且全部结果发完（或 done 关闭）后关闭 out。
// 注意：out 只能由 workerPool 在"所有 worker 都退出后"统一关闭，
// 任何单个 worker 都不得 close(out)——其他 worker 还在往上面发。
func workerPool(done <-chan struct{}, jobs <-chan int, n int) <-chan int
```

### 第一步：手把手起步（行为 1 的测试直接给你当模板）

1. 在 `tdd/pipeline/` 下新建 `pipeline_test.go`，写入：

```go
package pipeline

import (
	"slices"
	"testing"
)

// 整条管线：gen 生产 2、3 → 两路 sq 从同一个 channel 扇出 → merge 扇入回一路。
// 扇出后哪个数走哪条支路由调度器决定，输出顺序不确定——收集后排序再断言。
func TestPipeline(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	in := gen(done, 2, 3)
	out1 := sq(done, in)
	out2 := sq(done, in)
	out := merge(done, out1, out2)

	var got []int
	for v := range out {
		got = append(got, v)
	}
	slices.Sort(got)

	want := []int{4, 9}
	if !slices.Equal(got, want) {
		t.Errorf("期望 %v，得到 %v", want, got)
	}
}
```

2. 运行 `go test ./tdd/pipeline` → **编译失败**：`undefined: gen`。
   这就是 RED——测试描述了你想要但还不存在的代码。
3. 新建 `pipeline.go`，照契约写出让测试通过的**最少代码**
   （先只实现 gen/sq/merge 三个，`workerPool` 留到行为 3）。
4. 再跑 `go test ./tdd/pipeline -race` → 绿，行为 1 完成。
   并发练习从第一轮起就带 `-race` 跑，别等最后才加。

---

## 二、任务单（边做边学）

每个行为 = 一轮完整的 RED → GREEN → REFACTOR，**先把测试写出来再实现**。

### 行为 1：整条管线跑通（测试代码已在第一节给出）

| 用例名 | 输入 | 期望 |
|---|---|---|
| 两数平方 | `gen(done, 2, 3)` → 两路 `sq` → `merge` | 收集结果排序后等于 `[4, 9]` |

**这一步用到的知识点：**

1. **扇出（fan-out）与扇入（fan-in）**：扇出 = 多个 goroutine 从**同一个** channel 读数据，把工作分摊下去（本测试里 `out1`、`out2` 两个 sq 都从 `in` 读）；扇入 = 把多个 channel 的数据合并到**同一个** channel（merge）。管线 = 生成器 → 若干阶段（扇出加工）→ 扇入收口。
2. **为什么断言前必须排序——channel 发送是单播**：`in` 里的 2 和 3，每个只会被**一个** sq 拿到（channel 的一个发送值只交付一个接收方）。哪个数走哪条支路、merge 先搬谁，由调度器决定，输出顺序不确定。断言并发结果的集合相等，标准做法就是 `slices.Sort` + `slices.Equal`——**对顺序敏感的断言在并发测试里必然抖动**。
3. **谁生产谁关闭**：channel 由**发送方**关闭，接收方永远不 close。关闭沿管线一级级传播：gen 发完后 `close(out)` → sq 的 `for range in` 读到关闭、退出循环、close 自己的 out → merge 的两个搬运 goroutine 各自读到关闭 → WaitGroup 归零 → `close(out)` → 测试里的 `for v := range out` 结束。**任何一级忘记关闭，下游的 range 就永远阻塞**——`go test` 卡死直到 `panic: test timed out`（机制实验会让你亲手踩一次）。
4. **每个发送点都要写 select + done 脚手架**：

```go
select {
case out <- v:   // 正常发送
case <-done:     // 被取消：立刻退出
	return
}
```

   没有它，接收方停止消费后发送方会永远阻塞在 `out <- v` 上——这就是 goroutine 泄漏的温床（行为 2 展开）。
5. **契约签名里的 `<-chan int` 是定向 channel**：返回只读 channel、参数只收只读 channel，调用方想往输出 channel 里写数据会直接**编译报错**。这是把"谁生产谁关闭"用类型系统在编译期焊死，回顾 [basic/goroutine_channel.go#L37](../../basic/goroutine_channel.go#L37)。

### 行为 2：close(done) 广播取消

管线运行中途关闭 done，全管线必须收拢退出。骨架如下，**补全 TODO 一行**：

```go
func TestPipeline_DoneBroadcast(t *testing.T) {
	done := make(chan struct{})
	out := merge(done, sq(done, gen(done, 1, 2, 3, 4, 5)))

	<-out // 先收到一个结果，证明管线已经转动

	// TODO：广播取消（一行代码）

	timeout := time.After(time.Second) // 看门狗
	for {
		select {
		case _, ok := <-out:
			if !ok {
				return // out 关闭 = 全管线已退出完毕
			}
		case <-timeout:
			t.Fatal("close(done) 后 1 秒内 out 未关闭：有 goroutine 没收到取消信号")
		}
	}
}
```

| 用例名 | 操作 | 期望 |
|---|---|---|
| 中途取消 | 收到第一个结果后 `close(done)` | out 最终关闭；收集循环在 1 秒看门狗内结束，不死锁 |

扩展断言（可选但推荐）：测试开头记 `before := runtime.NumGoroutine()`，结尾 `time.Sleep(100ms)` 后断言 `runtime.NumGoroutine() <= before`——数量回落证明没有泄漏。用 `>` 比较而不是 `!=`：testing 框架自己可能有后台 goroutine，直接判相等容易误伤。

**这一步用到的知识点：**

1. **close 是广播，发送是单播**——本练习最重要的一条机制。`close(ch)` 之后，**所有**正在阻塞等待接收的 goroutine、以及**将来**再执行 `<-ch` 的代码，全部立刻得到零值（`v, ok := <-ch` 中 ok 为 false）。channel 内部维护一个接收等待队列，close 会把队列里的接收方**一次性全部唤醒**——一处 close，全员通知。对比：向 channel **发送**一个值，无论多少接收方在等待，只有**一个**能拿到（单播）。done 模式能成立，靠的就是广播特性。
2. **为什么不能"向 done 发送一个值"来代替 close**：发送是单播，管线里 gen、sq、merge 搬运工共 N 个 goroutine 在等取消信号，只有一个能收到，其余 N-1 个照常阻塞——泄漏。要发 N 次就得知道 N，而 close 不需要知道。
3. **done 为什么是 `chan struct{}`**：接收方只关心"关了没有"这个**事件**，根本不读值。`struct{}` 占零字节，是"纯信号"的惯用类型；写 `chan bool` 会让人误以为 true/false 有语义区别——没有，close 后读到的永远是零值。
4. **超时看门狗是并发测试的标配**：要断言"某事最终会发生"时，必须给等待加上限。没有看门狗，实现出错时测试永远挂住（`go test` 默认 10 分钟才 panic 兜底），你会分不清是"还在跑"还是"死锁了"。`time.After` + select 两行就是标准写法。
5. **泄漏的检测手段**：`runtime.NumGoroutine()` 前后对比是单元测试里最轻量的办法；`go test` 超时 panic 时附带的 goroutine dump 里能看到每个泄漏 goroutine 阻塞在哪一行；生产环境用 `net/http/pprof` 的 goroutine profile。本练习先掌握第一种。

### 行为 3：worker pool

| 用例名 | 输入 | 期望 |
|---|---|---|
| 百任务四工人 | jobs 发送 1..100 后 `close(jobs)`，`workerPool(done, jobs, 4)` | 恰好收集到 100 个结果；排序后等于 1²..100²；`for range out` 正常结束（out 被关闭） |

骨架（收集与断言自己写）：

```go
done := make(chan struct{})
defer close(done)

jobs := make(chan int)
go func() {
	for i := 1; i <= 100; i++ {
		jobs <- i
	}
	close(jobs) // 任务发完，生产者负责关闭
}()

out := workerPool(done, jobs, 4)
// 收集全部结果 → len 断言 100 → 排序后与期望切片（循环生成 1²..100²）slices.Equal
```

**这一步用到的知识点：**

1. **worker pool 是扇出/扇入的特化**：固定 n 个消费者从同一个 jobs 读（扇出），结果写同一个 out（扇入）。与 merge 的区别：merge 推广到"任意多个输入 channel"，workerPool 是"固定规模的工人池 + 共享输出"。看清这个对应关系，workerPool 的实现没有任何新魔法。
2. **两条关闭链，别混**：jobs 由任务生产者关闭（谁生产谁关闭）→ 4 个 worker 的 `for range jobs` 各自读到关闭而退出 → `wg.Wait()` 返回 → workerPool 统一 `close(out)`。**绝不能在某个 worker 里 close(out)**——其余 worker 还在往 out 发送，向已关闭 channel 发送会 panic（close 三规则之一，练习 10 会系统深挖）。
3. **WaitGroup 的两个铁律**（回顾 [basic/goroutine.go](../../basic/goroutine.go)）：`wg.Add(n)` 必须在**启动 goroutine 之前**完成，否则 Wait 可能提前返回；每个 worker 里 `defer wg.Done()`——panic 也不会漏掉 Done，漏 Done 的 Wait 是死锁。
4. **`-race` 在本行为盯的是什么**：收集结果用的是 `for range out` 读 channel，没有共享变量。如果你的实现图省事用了全局 `count++` 之类共享计数，`-race` 立刻报 data race。Go 的格言是 **"不要通过共享内存来通信，而要通过通信来共享内存"**——channel 同时承担了传数据和同步两件事，这是本练习不加一把锁也能过 `-race` 的原因。

### 行为 4：限流模式（Ticker 节拍器）

| 用例名 | 设置 | 期望 |
|---|---|---|
| 限速生效 | 20ms 一拍的限速器，jobs 无限量供应，观察 100ms 窗口 | 处理次数 `0 < count <= 6`（理论约 5 拍，上限放宽一拍防调度抖动） |

骨架（断言自己写）：

```go
func TestRateLimit(t *testing.T) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	stop := make(chan struct{})
	defer close(stop)
	jobs := make(chan int)
	go func() { // 无限量任务供应者——它自己也要有退出路径
		for i := 0; ; i++ {
			select {
			case jobs <- i:
			case <-stop:
				return
			}
		}
	}()

	deadline := time.After(100 * time.Millisecond) // 观察窗口的看门狗
	count := 0
loop:
	for {
		select {
		case <-jobs:
			<-ticker.C // 每处理一个任务，先等一拍
			count++
		case <-deadline:
			break loop
		}
	}
	// 断言：0 < count <= 6
}
```

**这一步用到的知识点：**

1. **Ticker 与 time.After 的场景差异——周期 vs 一次性**：`time.NewTicker(d)` 每隔 d 往 `C` 送一个时间值，**周而复始**，直到 Stop——适合限速、心跳、定时轮询这类**重复节奏**；`time.After(d)` 只在 d 后送**一次**——适合超时、看门狗、deadline 这类**单次事件**。本测试两者同台：ticker 做限速（周期），deadline 做窗口截止（一次）。选错的典型症状：用 After 做周期任务就得每次重建，用 Ticker 做超时就会在超时后还白白滴答。
2. **`time.Tick` vs `time.NewTicker`**：`time.Tick(d)` 效果等价于 NewTicker，但**拿不到 Ticker 句柄、永远无法 Stop**。Go 1.23 之前这是硬伤——不 Stop 的 Ticker 连 GC 都不回收，泄漏到程序退出；Go 1.23 起 GC 能回收无引用的 ticker，泄漏问题已消除。但工程里仍推荐 `NewTicker` + `defer ticker.Stop()`：生命周期显式、意图清晰。测试里写 NewTicker，不写 `time.Tick`。
3. **select 套 for 时的 break 陷阱**：裸 `break` 只跳出 select、跳不出外层 for——上面骨架里的 `break loop` 用了标签语法，正是 [basic/break_continue_goto.go#L38](../../basic/break_continue_goto.go#L38) 的标签 break 在并发场景的实战。
4. **时间类断言的哲学：断上不断下，宁可宽松不可抖动**：第 5 拍（100ms 整）能否赶上 deadline 取决于调度，毫秒级抖动客观存在。所以上限放宽一拍（≤6），下限只要求 >0 证明确实在处理。把时间断言写成精确等值（== 5）是最典型的 flaky test 制造法。
5. **供应者 goroutine 的退出路径**：注意 jobs 供应者用 `select { case jobs <- i: case <-stop: }`——无限循环的 goroutine 同样需要被取消的手段，这本身就是 uber 条目的又一次演练。

### 机制实验（做完行为 1 后必做）

1. 把 `gen` 实现里的 `close(out)` 注释掉，重跑行为 1 → 测试死锁，最终 `panic: test timed out after ...`
2. 读 panic 输出里的 goroutine dump：能看到 sq、merge 的搬运 goroutine 还阻塞在 channel 接收上，**dump 里直接标着阻塞的行号**——这就是泄漏的现场
3. 改回来，确认重新变绿

原理：gen 不关闭 out，sq 的 `for range` 永远等下一个值，merge 的 WaitGroup 永远不归零，测试主 goroutine 的 `for range out` 永远阻塞——一行 close 的缺失，拖死整条管线。

---

## 三、知识点总结

### 管线模式速查

| 模式 | 一句话 | 本练习位置 |
|---|---|---|
| 生成器 gen | 把数据源流化的入口；发完自己关闭 | 行为 1 |
| 加工阶段 sq | 读一个 channel、写一个新 channel；in 关则 out 关 | 行为 1 |
| 扇出 | 多个消费者读同一个 channel，分摊工作 | 行为 1、3 |
| 扇入 merge | N 路搬 1 路；WaitGroup 等全部搬完再统一关闭 | 行为 1 |
| done 广播取消 | 一处 close，全管线所有 goroutine 同时收到退出信号 | 行为 2 |
| worker pool | 固定 n 个工人 + 共享输出 channel | 行为 3 |
| 限流 | Ticker 周期节拍控制速率；After 一次性做看门狗 | 行为 4 |

### channel 关闭规则速查

- 谁生产谁关闭；接收方不 close
- close 是**广播**：所有（含将来的）接收方立刻收到零值，`ok == false`
- 向已关闭 channel **发送** → panic；**重复 close** → panic
- 从已关闭 channel **接收** → 零值 + `ok == false`（不 panic）
- `for range` 遍历 channel 靠 close 结束（回顾 [basic/goroutine_channel.go#L125](../../basic/goroutine_channel.go#L125)）

### goroutine 生命周期速查（uber 条目）

**不要一劳永逸地使用 goroutine**——每启动一个 goroutine，都必须能回答"它什么时候退出"。
本练习每个 goroutine 的退出路径只有两类：**输入 channel 被关闭**（自然结束）或 **done 被关闭**（被取消）。
检测：单测里 `runtime.NumGoroutine()` 前后对比；超时 panic 的 goroutine dump；生产用 pprof。

### 与书目的对应

- 教程 ch09 后半（channel 模式：管线、扇出扇入、done 取消）/ ch10（goroutine 生命周期）
- uber 规范·goroutine 生命周期：不要一劳永逸地使用 goroutine——本练习每个函数都是它的正面示范
- Go 官方博客《Go Concurrency Patterns: Pipelines and cancellation》：gen/sq/merge 三件套的原始出处，做完本练习去读原文对照
- [basic/goroutine.go](../../basic/goroutine.go)：WaitGroup 与 defer Done；[basic/goroutine_channel.go](../../basic/goroutine_channel.go)：定向 channel、for range + close；[basic/time.go#L78](../../basic/time.go#L78)：Ticker 的第一次照面

---

## 四、验收标准

```bash
go test ./tdd/pipeline -v          # 全绿
go vet ./tdd/pipeline              # 无警告
go test ./tdd/pipeline -race       # 并发练习必须无数据竞争
go test ./tdd/pipeline -cover      # 核心逻辑 ≥80%
```

## 五、完成后自查（能口头回答才算过）

1. 扇出和扇入分别指什么？merge 里 WaitGroup 的作用是什么，`close(out)` 为什么必须等它？
2. "谁生产谁关闭"原则——为什么接收方不该 close channel？管线里关闭信号是怎么一级级传播的？
3. `close(done)` 为什么能同时通知管线所有阶段？换成"向 done 发送一个值"行不行，为什么？
4. done 为什么惯用 `chan struct{}` 而不是 `chan bool`？
5. 行为 1 的断言为什么要先排序？这背后对应 channel 发送的什么语义？
6. 说出 gen、sq、merge 各自 goroutine 的两条退出路径；哪个 goroutine 答不上"何时退出"就是什么？
7. Ticker 和 `time.After` 各自适合什么场景？为什么工程里用 `time.NewTicker` 而不是 `time.Tick`？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把练习 11 tdd/pipeline 从"待建"划掉（ch09 后半管线模式 + uber·goroutine 生命周期落地）。
