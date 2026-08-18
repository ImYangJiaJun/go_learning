# Clock TDD —— 设计驱动：让接口从测试里长出来（倒计时）

目标：TDD 轨道第一个**设计驱动型**练习。场景一句话说完——倒计时输出 3、2、1、Go!，
每步隔 1 秒。但"测试不许真等 4 秒"这条约束会逼你做出本练习真正的产出：一个睡眠抽象。
注意力放在：**接口是怎么被测试逼出来的**、注入与全局变量替换的取舍，最后用 Go 1.25 的
`testing/synctest` 把 DI 方案整个对照重写一遍。

> 本任务是**设计驱动型**练习：不固定契约，接口从测试中长出来。第一节只给行为需求——
> 函数名、参数个数、接口长什么样，全部由你在 RED → GREEN → REFACTOR 循环里决定。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/clock` 就是它的运行方式，验收者是测试，不是人。

这个包要提供的能力只有一条行为主线：

- **倒计时**：依次输出 `3`、`2`、`1`、`Go!`（每个一行），每步之间等待 1 秒
- 输出必须写到调用方指定的地方，不能焊死在标准输出——否则测试无法捕获
  （这条会逼出第一个设计决策）
- **测试不允许真的等 4 秒**——慢测试在这里被视为设计缺陷，不是测试写得不好
  （这条会逼出第二个设计决策）

### 文件计划（建议，共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `countdown_test.go` | 全部测试 + Spy + synctest 对照测试 | **第 1 个建** |
| `countdown.go` | 倒计时函数、睡眠抽象、真实睡眠实现（名字都由你定） | 测试编译报错时 |

### 接口契约（本练习不设）

设计驱动型不给固定契约：上一站 testbasic 的签名是发下来的，这一站要你自己从测试里长出来。
硬约束只有三条：

1. package 名是 `clock`
2. 满足上面的全部行为需求
3. `go test ./tdd/clock` 全绿

其余——函数叫什么、收几个参数、接口几个方法——都是你的设计作品。做完后对照
learn-go-with-tests 的 Mocking 章，看你长出来的设计和作者的重合多少、差在哪。

### 第一步：手把手起步（行为 1 的第一个测试）

1. 在 `tdd/clock/` 下新建 `countdown_test.go`，写入（可直接粘贴）：

```go
package clock

import (
	"bytes"
	"testing"
)

func TestCountdown(t *testing.T) {
	buffer := &bytes.Buffer{}

	Countdown(buffer)

	want := `3
2
1
Go!
`
	got := buffer.String()

	if got != want {
		t.Errorf("期望输出 %q，得到 %q", want, got)
	}
}
```

2. 运行 `go test ./tdd/clock` → **编译失败**：`undefined: Countdown`。
   这就是 RED——测试描述了你想要但还不存在的代码（与 testbasic 行为 1 同一条纪律）。
3. 新建 `countdown.go`，写**最少**的代码让测试通过。先别管"等 1 秒"，那是下一个行为的事。
   最小实现的轮廓：`for i := 3; i > 0; i--` 循环里 `fmt.Fprintln(out, i)`，
   循环结束后再 `fmt.Fprintln(out, "Go!")`——参数类型写 `io.Writer`
   （想想为什么不是 `*bytes.Buffer`，行为 1 的知识点里对答案）。
4. 再跑 `go test ./tdd/clock` → 绿。第一轮 RED → GREEN 完成，
   你也被迫做了第一个设计决策：`Countdown` 接收一个 `io.Writer`。

---

## 二、任务单（边做边学）

### 行为 1：输出内容正确（已在「第一步」跑通）

| 用例名 | 输入 | 期望断言 |
|---|---|---|
| 倒数输出 | 空 `bytes.Buffer` | `buffer.String()` 恰为 `"3\n2\n1\nGo!\n"` |

行为 1 没有新用例要补——「第一步」的测试就是它的全部。真正的工作从下一个行为开始。

**这一步用到的知识点：**

1. **`bytes.Buffer`：测试里捕获输出的标准工具**。它实现了 `io.Writer`，内部维护一个
   可增长的字节切片：`Write` 把数据追加进去，`String()` 把累计内容取出来。注意必须传指针
   （`&bytes.Buffer{}`）：`Write` 是指针接收者方法（要改内部状态），值拷贝会让写入的内容
   落到副本上、原 buffer 永远为空——这是本练习最容易踩的第一个坑。
2. **`io.Writer`：一个方法的接口**。定义就一行 `Write([]byte) (int, error)`，却是标准库里
   被实现最多的接口之一——文件、网络连接、buffer、压缩流都是 Writer。`fmt.Fprintln(w, x)`
   等价于"先格式化，再调 `w.Write`"。生产传 `os.Stdout`，测试传 `&bytes.Buffer{}`，
   同一份代码两种环境——这是"面向接口编程"的最小完整范例，也是本练习 DI 思想的第一块砖。
3. **参数为什么写 `io.Writer` 而不是 `*bytes.Buffer`**：参数类型越具体，调用方能传的东西
   越少。写 `*bytes.Buffer`，`os.Stdout` 就传不进来，程序永远无法真正打印。原则一句话：
   **函数参数要接口，返回值要具体类型**（accept interfaces, return structs）。
4. **反引号多行字符串**：`want` 用反引号写成字面多行，所见即所得，末尾自带一个真实换行——
   正好对上 `Fprintln` 每行结尾的 `\n`。易错点：换双引号就得手写 `"3\n2\n1\nGo!\n"`，
   既难读又容易漏掉最后一个 `\n`（包括 `Go!` 在内的每一行都有换行）。

### 行为 2：不许真等——把"睡觉"抽象出来（RED → GREEN）

现在加上约束：每步之间等 1 秒。直接在实现里写 `time.Sleep(1 * time.Second)`，测试就要
真跑 3 秒多——**测试慢是设计的警报**：它说明"倒计时"和一个它并不关心的细节（怎么睡）焊死了。

把测试改成：注入一个 Spy 代替真实睡眠，断言"睡了几次"。

| 用例名 | 输入 | 期望断言 |
|---|---|---|
| 每步之间睡一次 | Spy 注入倒计时 | Spy 的调用次数 == 3（写完 3、2、1 后各睡一次，`Go!` 之后不睡） |

Spy 骨架（名字自己定，这是你的设计）：

```go
type SpySleeper struct {
	Calls int
}

func (s *SpySleeper) Sleep() {
	s.Calls++
}
```

**这一步用到的知识点：**

1. **依赖注入（DI）**：`Countdown` 需要"睡一下"的能力，但不自己 `time.Sleep`，而是由调用方
   把"会睡觉的东西"传进来。生产传真睡的，测试传装睡的。DI 的全部要义就这一句：
   **需要的协作方从参数进来，不在函数体内写死**。
2. **接口最小化：只抽象你需要的动作**。`Countdown` 只需要一个动作——"睡"。所以它要求的
   接口只该有一个方法 `Sleep()`。反面教材：抽象一个 `Clock` 接口带上
   `Now/After/Sleep/NewTicker` 一排方法——用不到的方法就是每个实现者的负担，Spy 也得跟着
   实现一堆空方法。Go 接口哲学：接口越小越好，单方法接口是常态（`io.Reader`/`io.Writer`/
   `fmt.Stringer` 全是单方法）。
3. **接口定义在消费方**：Go 接口是隐式实现的，所以接口该写在**使用它的包**里（本练习就是
   `clock` 包），而不是实现方的包里。"实现方声明接口"是 Java/C# 的习惯，搬过来会造成
   不必要的包依赖。另外注意：`time.Sleep` 的签名是 `Sleep(time.Duration)`，但你的接口方法
   可以不要参数、内部固定睡 1 秒——接口形状按消费方需求定，不必照搬被包装的函数。
4. **签名演进也是 TDD 的一部分**：这一步你要把 `Countdown(buffer)` 改成
   `Countdown(buffer, sleeper)`——已有测试编译不过，这同样算 RED。改完测试里做三件事：
   构造 Spy、把 Spy 传进去、断言 `spy.Calls == 3`。
5. **为什么 Spy 用指针接收者**：`Sleep` 要修改 `Calls` 字段——与
   [basic/struct_func.go](../../basic/struct_func.go) 的规则一致：改字段必须指针接收者。
   测试里传 `spy`（`&SpySleeper{}` 造出来的），断言时读 `spy.Calls`。

### 行为 3：顺序也要对——先写 3 再睡，再写 2……（GREEN 加固）

只断言"睡了 3 次"有漏洞：先连睡 3 次、再一口气输出全部内容的错误实现，也能过行为 2。
真正的需求是**交替**：写 3 → 睡 → 写 2 → 睡 → 写 1 → 睡 → 写 Go!。

| 用例名 | 输入 | 期望断言 |
|---|---|---|
| 写睡交替 | Spy 同时记录"写"和"睡"两类操作 | 操作序列恰为 `[写3, 睡, 写2, 睡, 写1, 睡, 写Go!]` |

思路骨架：Spy 需要截获"写"——让它不仅是 Sleeper，还兼任 Writer。测试里把**同一个 Spy**
同时作为输出目标和 Sleeper 传进 `Countdown`：

```go
type CountdownOperationsSpy struct {
	Calls []string
}

func (s *CountdownOperationsSpy) Sleep() {
	s.Calls = append(s.Calls, sleep)
}

func (s *CountdownOperationsSpy) Write(p []byte) (n int, err error) {
	s.Calls = append(s.Calls, write)
	return len(p), nil
}

const (
	write = "write"
	sleep = "sleep"
)
```

断言：期望序列是 `[write, sleep, write, sleep, write, sleep, write]`——4 次写、3 次睡，
严格交替。测试代码自己写，写法和行为 2 的断言同构。

**这一步用到的知识点：**

1. **一个类型实现多个接口**：Spy 同时满足 `io.Writer`（有 `Write`）和你的 `Sleeper`
   （有 `Sleep`），于是一个对象占两个参数位。Go 接口小，这种"身兼数职"毫无成本——这正是
   接口最小化的红利（对照 [basic/interfaces_struct.go](../../basic/interfaces_struct.go)）。
2. **stub / spy / mock / fake 的区别**（测试替身的四个档次）：
   - **stub（桩）**：只提供固定答案，没有智能——`Sleep()` 什么都不做就是一个 stub
   - **spy（间谍）**：记录"被怎么调用了"（次数/参数/顺序），测试结尾对记录做断言——
     本练习的用法
   - **mock（模拟）**：自带期望、自我验证——调用前告诉它"你将被按某顺序调 7 次"，违例它
     自己让测试失败。功能最强也最啰嗦；Go 社区风格更偏向 spy + 手写断言，直白、无框架依赖
   - **fake（伪实现）**：简化但**真能工作**的实现，如内存版数据库。行为 4 的 synctest
     假时钟，本质就是标准库送你的一个时间维度的 fake
3. **顺序断言是行为断言的完整形态**：输出内容对 + 次数对 + 顺序对，"倒计时"这个行为才算
   被钉死。以后测交互密集的代码（先开后关、先写日志再返回），都套这个模式。
4. **易错点——切片比较**：`Calls` 是 `[]string`，Go 里切片不能用 `==` 比（编译错误）。
   要么逐元素循环比，要么用 Go 1.21+ 标准库的 `slices.Equal(spy.Calls, want)`——这是
   `slices` 包的第一次实战，记住它。

### REFACTOR：补一个真的 Sleeper

测试全绿、行为钉死后，轮到生产侧：在 `countdown.go` 里实现一个**真睡 1 秒**的 Sleeper
（比如 `type DefaultSleeper struct{}`，`Sleep` 里写 `time.Sleep(1 * time.Second)`）。

问题立刻出现：**没有 main 函数，怎么验证真睡版？** 直接写测试就要真等 3 秒——回到了
行为 2 要消灭的慢测试。先别妥协，把这个问题揣着，行为 4 用 synctest 正面解决它。

**这一步用到的知识点：**

1. **uber 规范·避免可变全局变量**：有人不用 DI，用包级变量替换——`var sleep = time.Sleep`，
   测试里 `sleep = spy.Sleep`，测完 defer 恢复。uber-go/guide 明确反对这种模式，三个硬伤：
   - **测试间泄漏**：忘了恢复，后面的测试全部拿到假的
   - **无法并行**：两个并行测试同时替换同一个全局变量，互相踩踏（`-race` 直接报警）
   - **隐式依赖**：读 `Countdown` 源码看不出它依赖一个可替换的全局变量，依赖关系藏在
     包级作用域里
   注入把这三条全解决：依赖显式出现在签名上、每个测试持有自己的 spy、天然可并行。
2. **接口 + 注入 vs 全局变量替换，是两条设计路线的分叉**：前者是 Go 主流（`http.Server`
   的 `Handler`、`http.Client` 的 `Transport` 全是注入）；后者只剩一个合法场景——兼容
   无法改签名的旧代码。新代码一律注入。
3. **REFACTOR 阶段的纪律**：加真实实现时，**已有测试一行不改**且必须保持绿——它们保护的
   是行为，不是某个具体实现。新实现暂时没有直接测试是可接受的：它的正确性由行为 4 补上。

### 行为 4（进阶对照）：用 Go 1.25 的 `testing/synctest` 重写

`testing/synctest` 给出一个完全不同的答案：**不改设计，直接让时间变假**。
真实 `time.Sleep` 版的倒计时，在气泡里瞬间测完，还能断言中间时刻的输出。

| 气泡内时刻 | 期望 buffer 内容 |
|---|---|
| 倒计时 goroutine 启动并阻塞后 | `"3\n"` |
| 假时钟前进 1 秒后 | `"3\n2\n"` |
| 再前进 1 秒 | `"3\n2\n1\n"` |
| 再前进 1 秒 | `"3\n2\n1\nGo!\n"`（倒计时结束；全程真实耗时 ≈ 0） |

骨架（断言辅助函数 `assertOutput(t, buffer, want)` 自己写）：

```go
func TestCountdownWithRealSleeper(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		buffer := &bytes.Buffer{}

		go Countdown(buffer, DefaultSleeper{})

		synctest.Wait()             // 等倒计时 goroutine 睡到第一个 Sleep（此时 3 已写出）
		assertOutput(t, buffer, "3\n")

		time.Sleep(1 * time.Second) // 测试里的 Sleep 不真睡，是在开假时钟的阀门
		synctest.Wait()
		assertOutput(t, buffer, "3\n2\n")

		// ……按上表补完 +2s、+3s 两个时刻
	})
}
```

**这一步用到的知识点：**

1. **气泡（bubble）**：`synctest.Test(t, func(t *testing.T) {...})` 把函数放进一个隔离世界
   运行；函数里 `go` 出去的 goroutine 都属于这个气泡。气泡有自己的**假时钟**，起点固定为
   2000-01-01 00:00 UTC——泡内每次跑 `time.Now()` 都从它起算，所以 synctest 测试是
   完全确定的。
2. **durably blocked（持久阻塞）= 时间前进的条件**：气泡内**所有** goroutine 都阻塞在
   "只能被同气泡伙伴或假时钟解除"的操作上时，假时钟瞬间快进到下一个能唤醒 goroutine 的时刻。
   算持久阻塞的：泡内 channel 的收发、case 全是泡内 channel 的 `select`、`sync.Cond.Wait`、
   泡内 `Add` 过的 `WaitGroup.Wait`、`time.Sleep`。**不算**的：`sync.Mutex` 加锁、网络/文件
   I/O、系统调用（这些可能被泡外事件解除）。全员卡在非持久阻塞上 → 直接 panic 报死锁。
   推论：气泡里不能真用网络，要用 `net.Pipe()` 这类进程内 fake。
3. **测试函数自己也是气泡成员**：测试里写 `time.Sleep(1 * time.Second)` 不是真睡，是主动把
   假时钟往前拧 1 秒。上表"前进 1 秒后"读作：测试的 Sleep 返回时，气泡时钟恰好走过 1 秒，
   倒计时 goroutine 的定时器已触发、`2` 已写出。
4. **`synctest.Wait()`**：阻塞到"气泡里除自己外的 goroutine 全部持久阻塞"。没有它，
   `go Countdown(...)` 之后立刻断言，可能撞上倒计时 goroutine 还没来得及写的调度间隙——
   `Wait` 把"等它跑完手头工作"变成显式、确定的动作。这就是 synctest 版测试**不抖动**
   （不 flaky）的原因，也是骨架里每次推进时间后都补一个 `Wait` 再断言的道理。
5. **气泡的纪律**：`synctest.Test` 返回前会等泡内 goroutine 全部退场；有 goroutine 泄漏
   （永不退出）就报死锁 panic——顺带成了免费的 goroutine 泄漏检测。泡内禁止调
   `t.Run` / `t.Parallel` / `t.Deadline`（想用子测试会编译/运行报错，断言辅助函数足矣）。
   版本注意：Go 1.24 里它是实验特性（`GOEXPERIMENT=synctest`，API 叫 `synctest.Run`），
   **Go 1.25 起转正**，API 定为 `synctest.Test(t, func(t *testing.T) {...})`——本仓库
   go 1.26 可直接 `import "testing/synctest"`。
6. **DI 方案 vs synctest 方案的取舍**：

| | Sleeper 注入（行为 2/3） | synctest（行为 4） |
|---|---|---|
| 生产代码 | 签名多一个参数，设计被测试塑形 | **零侵入**，直接测真实 `time.Sleep` 版本 |
| 能断言什么 | 调用次数/顺序；没有时间概念，断言不了"第 2 秒时输出了几行" | 中间时刻状态、相对时序，行为覆盖更真 |
| 依赖范围 | 任何协作方都能换（时间、数据库、网络……） | 只解决**时间**；外部 I/O 依赖仍得靠接口/DI |
| 版本要求 | 任意 Go 版本 | Go 1.25+ |
| 心智成本 | 设计一个小接口，人人都懂 | 要理解气泡与 durably blocked 语义 |

一句话取舍：**纯时间类逻辑**（超时、退避、定时任务）优先 synctest——测得更真还不动设计；
**时间之外的外部依赖**（要替换行为的那些）仍归 DI。两者不是替代关系，是分工。本练习是
刻意的"小场景"，好让两条路线能正面对比。

---

## 三、知识点总结

### 测试替身速查

| 替身 | 干什么 | 本练习落点 |
|---|---|---|
| stub | 提供固定答案，无智能 | `Sleep()` 什么都不做的版本 |
| spy | 记录调用（次数/参数/顺序），测试断言记录 | 行为 2、3 的两个 Spy |
| mock | 自带期望、自我验证，违例即失败 | 未用——Go 社区偏好 spy + 手写断言 |
| fake | 简化但真实可用的实现 | 行为 4 的假时钟（标准库内置的 fake） |

### 依赖注入 vs 可变全局变量（uber 规范）

| | 注入（推荐） | 包级变量替换（`var sleep = time.Sleep`） |
|---|---|---|
| 依赖可见性 | 显式写在签名上 | 隐式，藏在包级作用域 |
| 测试隔离 | 每个测试持有自己的 spy | 忘了恢复就泄漏给后续测试 |
| 并行测试 | 天然安全 | 互相踩踏，`-race` 报警 |

### synctest 速查

| 概念 | 一句话 |
|---|---|
| 气泡 | `synctest.Test` 圈出的隔离世界，泡内 goroutine 共享假时钟 |
| 假时钟 | 起点固定 2000-01-01 00:00 UTC；全员持久阻塞时瞬间快进到下一个定时器 |
| durably blocked | 阻塞在只能被同气泡伙伴/假时钟解除的操作上（泡内 channel、Cond、WaitGroup、Sleep）；Mutex 加锁、I/O、系统调用**不算** |
| `synctest.Wait()` | 等泡内其他 goroutine 全部持久阻塞——断言前的确定性同步点 |
| 泄漏检测 | 泡内 goroutine 不退场 → `Test` 报死锁 panic |
| 泡内禁区 | `t.Run` / `t.Parallel` / `t.Deadline` 不可调用 |
| 版本 | Go 1.24 实验（`synctest.Run` + GOEXPERIMENT）；Go 1.25 转正（`synctest.Test`） |

### 设计原则一句话

- 参数要接口、返回值要具体类型（accept interfaces, return structs）
- 接口只抽象你需要的动作，定义在消费方包里
- 测试慢/测试难写，先怀疑设计，再怀疑测试

### 与书目的对应

- Effective Go·11 接口：单方法接口、接口定义在消费方——本练习全部落地
- uber 规范·避免可变全局变量：REFACTOR 环节的对比对象
- learn-go-with-tests·Mocking 章：行为 1~3 的原型，做完对照你的设计和作者的差异
- learn-go-with-tests·Revisiting time with synctest 章：行为 4 的原型
- 前置知识复用：[basic/interface.go](../../basic/interface.go)（接口实现与接收者）、
  [basic/struct_func.go](../../basic/struct_func.go)（指针接收者改字段）

---

## 四、验收标准

```bash
go test ./tdd/clock -v        # 全绿（含 synctest 版，瞬间跑完）
go vet ./tdd/clock            # 无警告
go test ./tdd/clock -race     # 行为 4 涉及 goroutine，必须无数据竞争
go test ./tdd/clock -cover    # 核心逻辑覆盖
```

## 五、完成后自查（能口头回答才算过）

1. 为什么"测试要等 4 秒"是设计问题而不是测试问题？它逼出了什么决策？
2. `Countdown` 的输出参数为什么是 `io.Writer` 而不是 `*bytes.Buffer`？对应哪条设计原则？
3. 接口最小化是什么意思？为什么 Sleeper 只有一个方法、且定义在 `clock` 包里？
4. stub / spy / mock / fake 四者的区别？本练习的 `CountdownOperationsSpy` 是哪一类，synctest 的假时钟最接近哪一类？
5. uber 规范为什么反对用包级变量替换做测试替身？注入方案解决了哪三个问题？
6. synctest 气泡里的时间什么时候前进？什么叫 durably blocked？`sync.Mutex` 加锁算不算？`synctest.Wait()` 解决什么抖动问题？
7. DI 注入 Sleeper 和 synctest 两条路线各自适合什么场景？为什么 synctest 替代不了 DI？

全部答清后，回到 [根 README](../../README.md#三对照for-learning-go-tutorial的覆盖检查)：
把 TDD 练习总目录里练习 7（tdd/clock）的状态划掉，并在"编程规范（uber-go/guide）"一行
记下进展——"避免可变全局变量"条目已实践。
