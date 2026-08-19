# Context TDD —— 慢存储查询：学会取消与超时

目标：TDD 轨道的并发进阶站。场景是一个**慢存储查询**：存储可能要几百毫秒才返回，而调用方随时可能反悔。注意力全部放在 `context` 包的四个构造函数上——`WithCancel` / `WithTimeout` / `WithDeadline` / `WithValue`——以及"取消沿树传播"和"`select` 等结果还是等取消"这两个标准动作。

前置：需要会 goroutine 和 `select`（见 [basic/goroutine_channel.go](../../basic/goroutine_channel.go)）；与 `tdd/pipeline` 练习的 done channel 手写法对照见第三节。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求规格（接口契约固定，照此实现）；第二节是纯任务单——只给行为目标、用例表和验收命令，测试代码全部自己写；第三节是知识点讲解，做之前通读或卡壳时查阅，做完后对照自查。

---

## 一、需求规格

### 核心功能

实现一个最小的**慢存储查询包**，它对外只提供一个能力：

- **`Fetch(ctx, store, id)`**：从慢存储查询 `id` 对应的数据；存储慢可以等——但只要 `ctx` 被取消或超时，立即放弃等待，返回 `ctx.Err()`

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/context` 就是它的运行方式，验收者是测试，不是人。

### 调用关系（谁在调用谁）

```text
测试代码 ──► Fetch(ctx, store, id) ──► Store.Get(id)（慢存储，放进后台 goroutine 等）
```

`Fetch` 内部用 `select` 同时等"结果 channel"和"ctx.Done()"，谁先就绪听谁的。
`Store` 是接口：测试里注入的是"固定睡 delay 再返回"的 `slowStore` fake——
它故意不响应取消，"存储慢"才能在测试里确定性重现，取消动作全由 `Fetch` 负责。

### 包名说明：目录叫 `context`，包名必须叫 `contextlab`

本练习每一行代码都要 `import "context"`。如果自己的包也叫 `context`，包名就会**遮蔽标准库**——写 `context.Background()` 时编译器无法区分你指的是哪个。所以包名定为 `contextlab`。

**Go 允许目录名与包名不一致**：导入路径看目录（`.../tdd/context`），代码里引用看包名（`contextlab.Fetch`），`go test ./tdd/context` 不受影响。同理，以后涉及 `net/http` 的练习目录叫 `tdd/http`、包名会叫 `httplab`。

### 文件计划（共 2 个文件，按编号顺序建）

最终目录长这样：

```text
tdd/context/
├── fetch_test.go    # 全部测试 + 测试用的慢存储 fake
└── fetch.go         # Store 接口 + Fetch 函数（本练习唯一要实现的文件）
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `fetch_test.go` | 全部测试 + 慢存储 fake | `slowStore`、`slowStore.Get`、`TestFetch_存储慢于取消时返回Canceled`、`TestCancel_父取消整棵子树`、`TestFetch_存储慢于超时返回DeadlineExceeded`、`requestIDKey`、`TestFetch_超时_气泡版`（可选）；行为 3、4 的测试名字自定 | **第 1 个建** |
| 2 | `fetch.go` | 慢存储抽象 + 本练习唯一要实现的 `Fetch` | `Store`、`Fetch` | 测试编译报错时 |

契约符号一共 2 个（`Store` 接口和 `Fetch` 函数），就是下面契约里的全部，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**你要写的每一个符号都在下面**。本练习只有一个实现文件，契约就这一组——
你唯一需要自己实现的是 `Fetch` 的函数体；测试侧的 `slowStore`、`requestIDKey` 按任务单的行为要求自己写，不属于契约。

**写在 `fetch.go`：**（需要 `import "context"`）

```go
package contextlab

import "context"

// Store 是慢存储的抽象：Get 可能耗时几百毫秒才返回。
// 实现方不负责响应取消——那是 Fetch 的事。
type Store interface {
	Get(id int) (string, error)
}

// Fetch 从慢存储查询 id 对应的数据。
// 必须尊重 ctx：内部用 select 同时等"存储结果"和"ctx.Done()"——
// 存储先返回，则返回存储结果；ctx 先 Done（取消/超时），则立即放弃等待，
// 返回 "", ctx.Err()（context.Canceled 或 context.DeadlineExceeded）。
func Fetch(ctx context.Context, s Store, id int) (string, error)
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 1 个接口类型：`Store`（含方法 `Get(id int) (string, error)`）
- 1 个函数：`Fetch`
- 0 个自有哨兵错误：取消/超时直接用标准库的 `context.Canceled`、`context.DeadlineExceeded`，不要自己造

---

## 二、任务单

### 行为 1：WithCancel —— 父一取消，整棵子树一起死

新建 `fetch_test.go`，先写一个"固定睡 delay 再返回 `item-<id>`"的慢存储 fake（比如 `slowStore`），然后自己写下面两个用例的测试。先写测试——编译失败（`undefined: Fetch`）即 RED；再新建 `fetch.go` 抄入契约里的 `Store` 接口，写**最少**的 `Fetch` 让测试变绿——哪怕先 `return "", ctx.Err()`，行为 2 的"正常返回"用例会逼你写出真实现。

| 用例名 | 操作 | 期望 |
|---|---|---|
| 存储慢于取消 | `slowStore{200ms}` + 已 cancel 的 ctx | `errors.Is(err, context.Canceled)` 为真 |
| 父取消整棵子树 | parent→child→grandchild 三层，只 cancel parent | 三层 `Done()` 全部关闭；`Err()` 全为 `context.Canceled` |

用例 2 的断言技巧：`select` 里加 `default` 分支可以"看一眼"某个 Done channel 此刻是否已关闭（原理见第三节行为 1 知识点 3）。

顺手实验：在 `cancel()` 前后各加一行 `t.Log("Err:", parent.Err())`，跑
`go test ./tdd/context -v -run 'TestCancel_父取消整棵子树'`，亲眼看 `Err()` 从 `<nil>` 变成 `context.Canceled`。

### 行为 2：WithTimeout —— 等得起，但只等 50ms

| 用例名 | 输入 | 期望 |
|---|---|---|
| 存储慢于超时 | `WithTimeout(bg, 50ms)` + `slowStore{200ms}` | `errors.Is(err, context.DeadlineExceeded)` 为真 |
| 存储快于超时 | `WithTimeout(bg, 200ms)` + `slowStore{10ms}`，id=42 | 返回 `"item-42"`，err 为 nil |

第二个用例会戳穿行为 1 里"最少实现"的假把式——现在必须写真正的 `select` 实现了：内部起一个 goroutine 调 `s.Get(id)`，结果走 channel 送回来，`select` 同时等结果 channel 和 `ctx.Done()`，谁先就绪听谁的。实现骨架、以及"结果 channel 为什么要缓冲 1"这个最大的易错点，见第三节行为 2 知识点。

测试完全自己写：`context.WithTimeout(context.Background(), 50*time.Millisecond)` 建 ctx，紧跟 `defer cancel()`（原因见第三节行为 2 知识点 4），断言用 `errors.Is`。

### 行为 3：WithDeadline —— 绝对时刻，语义对照

| 用例名 | 输入 | 期望 |
|---|---|---|
| 截止时间已过 | `WithDeadline(bg, now.Add(-time.Second))` + `slowStore{10ms}` | **立即**返回 `DeadlineExceeded`（哪怕存储只要 10ms） |
| 截止时间将到 | `WithDeadline(bg, now.Add(50ms))` + `slowStore{200ms}` | `DeadlineExceeded` |
| 对照：未到期且存储快 | `WithDeadline(bg, now.Add(200ms))` + `slowStore{10ms}` | 正常返回 |

测试完全自己写（用例结构同行为 2）。

### 行为 4：WithValue —— 给请求挂元数据

先在 `fetch_test.go` 里定义 key——用非导出的空结构体类型（形如 `type requestIDKey struct{}`），为什么必须这样见第三节行为 4 知识点 2——然后自己写测试：

| 用例名 | 操作 | 期望 |
|---|---|---|
| 存后能取 | `ctx = WithValue(bg, requestIDKey{}, "req-123")` | `ctx.Value(requestIDKey{})` == `"req-123"` |
| 没存过 | `Background()` 直接取 | `Value` 返回 `nil` |
| 子覆盖父 | parent 存 `"p"`，`child := WithValue(parent, requestIDKey{}, "c")` | child 取到 `"c"`；parent 取到仍是 `"p"` |

### 行为 5（可选进阶）：testing/synctest 气泡 —— 零等待重写行为 2

行为 2 的测试真实花了 50ms。如果超时是 5 分钟呢？用 `testing/synctest` 气泡把这个测试重写一遍（写法见第三节行为 5 知识点），跑一下，和行为 2 对比真实耗时。

---

## 三、知识点总结

### 行为 1：WithCancel —— 父一取消，整棵子树一起死

1. **`context.Context` 接口四方法**——整个包就围绕这个小接口：

```go
type Context interface {
	Deadline() (time.Time, bool) // 截止时刻；ok=false 表示没有截止
	Done() <-chan struct{}       // 取消/超时后被关闭的 channel
	Err() error                  // Done 未关时是 nil；关后是 Canceled 或 DeadlineExceeded
	Value(key any) any           // 沿树向上查请求级元数据
}
```

注意 `Done()` 返回的是**只读** channel（`<-chan`）：拿到 ctx 的一方只能听，不能关——能关的只有持有 cancel 函数的那一方。

2. **取消传播树**：`context.Background()` 是根；`WithCancel(parent)` 把自己注册到父节点名下，形成一棵树。`cancel()` 做两件事：close 自己的 Done channel、级联取消所有子孙。为什么 close 能"通知所有人"——**close 一个 channel 会让所有 `<-done` 的接收方同时被唤醒**（各自读出零值），这是 channel 的广播语义，和 `tdd/pipeline` 练习里 `close(done)` 是同一招，只不过 context 把树的注册和级联替你写好了。
3. **`select` + `default` 变成"看一眼"**：`select` 里加 `default` 分支就不阻塞——Done 关了走第一个 case，没关走 default。测试里用它断言"此刻已经关闭"；真实代码里很少需要。
4. **两个哨兵错误**：`context.Canceled`（主动取消）和 `context.DeadlineExceeded`（超时/到点），都声明在 context 包里。断言用 `errors.Is`：目前它们是值比较，但 `errors.Is` 对未来被 `%w` 包装过的错误同样成立，是工程惯例（包装话题见 `tdd/errhandling` 练习）。
5. **谁创建谁 cancel**：`WithCancel` 返回的 cancel 函数**必须被调用**，惯例是紧跟一句 `defer cancel()`。不调用的后果：子节点一直挂在父的注册表里，直到父被取消才释放——这就是 context 泄漏。`go vet` 的 lostcancel 检查专门抓这个，所以测试代码里也不能偷懒——每个 `WithCancel` 后面都要有 `defer cancel()` 兜底。

### 行为 2：WithTimeout —— 等得起，但只等 50ms

1. **`WithTimeout` 就是语法糖**：标准库里它就一行——`return WithDeadline(parent, time.Now().Add(timeout))`。相对时间人看着方便，机器内部统一用绝对时刻。
2. **Fetch 的灵魂是 select 二选一**：Fetch 没办法让慢存储变快，只能在等待时竖起两只耳朵——一只听结果 channel，一只听 `ctx.Done()`，谁先说话听谁的。取消后测试**不用真的等 200ms**：Fetch 在 50ms 超时时就返回了，存储的 200ms 睡眠在后台 goroutine 里自己睡完。
3. **结果 channel 通常使用 `make(chan result, 1)`（缓冲 1）**：如果无缓冲，Fetch 因超时先 `return` 后，后台 goroutine 醒来执行 `ch <- ...` 可能永远阻塞。缓冲 1 给结果发送留一个位置，使 worker 能完成发送并退出。但这只解决“结果发送方”的阻塞；如果 `Store.Get` 自身不支持 context 并永久阻塞，底层调用仍无法被 Fetch 强制取消。

综合 2、3 两点，`Fetch` 的实现骨架长这样（逐行读懂，再自己敲进 `fetch.go`）：

```go
func Fetch(ctx context.Context, s Store, id int) (string, error) {
	type result struct {
		val string
		err error
	}
	ch := make(chan result, 1) // 缓冲 1：Fetch 先返回时，worker 仍能写完退出，不泄漏
	go func() {
		v, err := s.Get(id)
		ch <- result{v, err}
	}()

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
```

4. **为什么 `defer cancel()` 在超时的 ctx 上也不能省**：`WithTimeout` 内部有定时器；提前返回时不 cancel，定时器要等到期才释放。损失虽小，但"谁创建谁 cancel"是一刀切的纪律——不给自己留"这次要不要 cancel"的判断题。
5. **对照 `time.After`**：`select` + `time.After(d)` 也能做超时，但定时器无法被取消（提前返回时它照样活到到期），超时错误也得自己造。context 把"取消 + 超时 + 错误语义"打包成标准件——这正是它存在的意义。

### 行为 3：WithDeadline —— 绝对时刻，语义对照

1. **绝对 vs 相对**：`WithDeadline` 收 `time.Time` 绝对时刻，适合"整个请求必须在 15:00:00 前完成"这种**跨函数共享**的截止——每个环节拿到的都是同一个终点，不用各自换算剩余时间。日常代码里 `WithTimeout` 更常用。
2. **`Deadline()` 方法读回截止时刻**：`deadline, ok := ctx.Deadline()`；`ok == false` 表示这个 ctx 没有截止。服务端可以用它判断"还剩多少时间，这活值不值得开始干"。
3. **已过期的 deadline 创建即 Done**：不用等，`Err()` 立刻就是 `DeadlineExceeded`。任务单里"截止时间已过"的用例验证了这一点——Fetch 还没等存储，select 里 `ctx.Done()` 已经是就绪状态。

### 行为 4：WithValue —— 给请求挂元数据

1. **Value 沿树向上查**：`child.Value(key)` 自己这层没有就问 parent，一路问到根，全没有返回 `nil`。所以"子覆盖父"只是遮蔽，父那份原样还在。
2. **key 为什么必须用非导出的自定义类型**：如果包 A 和包 B 都用字符串 `"id"` 作 key，后写的覆盖先写的，而且编译器一声不吭——内置类型 key 的命名空间是全局的。包私有类型**别的包根本无法构造它的值**，物理上杜绝冲突。静态检查工具 staticcheck 的 SA1029 规则专门警告"用内置类型作 key"。key 的写法示例：

```go
// key 用非导出、空结构体类型：类型本身就是包私有的命名空间
type requestIDKey struct{}
```

3. **Value 里该放什么**：只放**请求级元数据**——请求 id、trace id、认证 token 这类"跟着请求走、横切所有层"的东西。官方文档的原话是：只放跨越进程和 API 边界的请求级数据，**不要**用它给函数传可选参数。业务参数走显式函数签名——藏在 Value 里的依赖，IDE 跳转不过去、编译器查不出来、写测试时要靠猜。
4. **WithValue 不返回 cancel**：它只是树上一个挂值的节点，没有自己的取消机制，Done 语义完全随父。所以行为 1 的"整棵子树"里，它照样跟着 Done。

### 行为 5（可选进阶）：testing/synctest 气泡

1. **气泡里是假时钟**：`synctest.Test(t, f)` 在"气泡"里运行 f——气泡内的 `time.Sleep`、`time.Timer`、`context` 的超时全部走假时钟。气泡发现所有 goroutine 都阻塞时，把时钟**直接快进**到最近的定时器：200ms 的睡眠、50ms 的超时，真实耗时约等于 0。
2. **确定性是副产品**：不抖。真实时间的测试在负载高的 CI 机器上偶尔会抖动（50ms 的定时器 51ms 才触发），气泡版永远精确。
3. **版本与限制**：Go 1.24 是实验特性（要 `GOEXPERIMENT=synctest`），Go 1.25 转正，本仓库 go 1.26 直接用。真实网络/文件 I/O 不在气泡管辖内；气泡外启动的 goroutine 气泡看不见。

用法示例——把行为 2 的测试整个包进 `synctest.Test` 即可：

```go
import "testing/synctest"

func TestFetch_超时_气泡版(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := Fetch(ctx, slowStore{delay: 200 * time.Millisecond}, 1)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("期望 context.DeadlineExceeded，得到 %v", err)
		}
	})
}
```

### 四个构造函数速查

| 函数 | 什么情况下 Done | 手动 cancel | 典型场景 |
|---|---|---|---|
| `WithCancel(parent)` | 只有手动 cancel | 就是靠它 | 用户取消、出错叫停整棵树 |
| `WithTimeout(parent, d)` | 相对时长到点 | 有效（提前叫停） | 给单次操作限时，最常用 |
| `WithDeadline(parent, t)` | 绝对时刻到点 | 有效 | 跨函数共享同一个终点 |
| `WithValue(parent, k, v)` | 随父，自己没有取消机制 | 不返回 cancel | 挂请求 id / trace id |

### `ctx.Err()` 三态

| Done channel | `Err()` 返回值 |
|---|---|
| 未关闭 | `nil` |
| 因主动 cancel 关闭 | `context.Canceled` |
| 因超时/到点关闭 | `context.DeadlineExceeded` |

### 使用纪律速查（四条）

1. **谁创建谁 cancel**：`defer cancel()` 紧跟创建，无条件。
2. **ctx 当第一个参数显式传，不存进 struct**：`func Fetch(ctx context.Context, ...)`。
3. **不把 ctx 当参数袋**：Value 只放请求级元数据；key 用非导出自定义类型。
4. **等待时永远多竖一只耳朵**：`select` 里结果 channel 与 `ctx.Done()` 并列；发给"可能比接收方活得久"的 goroutine 的 channel 记得缓冲。

### 与 pipeline 练习的 done channel 对照

| `tdd/pipeline` 的手写法 | context 的标准件 |
|---|---|
| `close(done)` 广播退出 | `cancel()` 关闭 `Done()` channel |
| `select { case <-done: ... }` | `select { case <-ctx.Done(): ... }` |
| 每层 goroutine 各自监听、手写传播 | 树形注册，父取消自动级联整棵子树 |
| 没有超时概念 | `WithTimeout` / `WithDeadline` 内建 |
| 没有元数据 | `WithValue` |

一句话：**context = 标准化的 done channel + 超时 + 请求级元数据**。下次想自己 `make(chan struct{})` 传取消信号时，先用 context。

### 与书目的对应

- **教程 ch09 延伸**：教程本身没有 context 章；本练习就是根 README 遗漏清单里"context 包（并发取消/超时控制，工程必备）"那一行。它把 [basic/goroutine_channel.go](../../basic/goroutine_channel.go) 里手写的 `select` 取消模式标准化成了工程通用件。
- **learn-go-with-tests 的 Context 章**：本练习的 `Store` 接口取材自它的 Store/Server 例子（讲"取消正在进行的操作"），可对照阅读。

---

## 四、验收标准

```bash
go test ./tdd/context -v        # 全绿
go vet ./tdd/context            # 无警告（lostcancel：每个 WithXxx 的 cancel 都被调用）
go test ./tdd/context -race     # 并发练习必须无数据竞争
go test ./tdd/context -cover    # Fetch 的 select 两个分支（结果 / Done）都要被走到
```

## 五、完成后自查（能口头回答才算过）

1. `context.Context` 接口有哪四个方法？`ctx.Err()` 的三种取值分别是什么意思？
2. 父 ctx 被 cancel 后，子和孙会怎样？底层靠什么机制实现"一传一大片"？
3. 为什么必须"谁创建谁 cancel"？漏掉 `cancel()` 泄漏的到底是什么？
4. `WithTimeout(d)` 和 `WithDeadline(now.Add(d))` 是什么关系？什么场景该用绝对时刻？
5. `WithValue` 的 key 为什么必须用非导出的自定义类型？Value 里该放什么、不该放什么？
6. `Fetch` 里的结果 channel 为什么要 `make(chan result, 1)`？用无缓冲会怎样？
7. 对照 pipeline 练习的 done channel，context 标准化了哪三件事？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把"context 包"从 ❌ 改成 ✅。
