# TcpUdp TDD —— socket 落地：TCP 与 UDP 的 echo 服务器

目标：这是 TDD 轨道阶段 4 的练习（对应根 README 练习 15）。被测对象是两个 echo
服务器启动函数，**注意力全部放在网络服务器怎么测**上：`127.0.0.1:0` 随机端口、
TCP 三段模型 vs UDP 无连接模型、字节流 vs 数据报的边界差异、并发服务器的 `-race` 验证。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求规格（接口契约固定，照此实现）；第二节是纯任务单——只给行为目标、
> 用例表和验收命令，测试代码全部自己写；第三节是知识点讲解，做之前通读或卡壳时查阅，做完后对照自查。

---

## 一、需求规格

### 核心功能

实现两个 **echo 服务器启动函数**，这个包对外只提供这两个能力：

- **启动一个 TCP echo 服务器**：接受连接后，把从连接里读到的数据原样写回
- **启动一个 UDP echo 服务器**：把收到的每个数据报原样写回发送方

两个函数都返回真实监听地址和一个关闭函数——测试可以"启动 → 通信 → 关闭"全程自助，
不需要人去指定端口、启动进程。

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/tcpudp` 就是它的运行方式，验收者是测试，不是人。

服务器内部必然有 goroutine（accept 循环 + 每条连接的处理循环），所以本练习
**全部测试要求 `-race` 通过**。

### 调用关系（谁在调用谁）

```text
测试代码 ──► StartTCPEcho() ──► addr ──► net.Dial("tcp", addr) ⇄ TCP 服务器
测试代码 ──► StartUDPEcho() ──► addr ──► WriteToUDP / ReadFromUDP ⇄ UDP 服务器
测试代码 ──► shutdown() ──► 关闭监听器、释放端口（之后 Dial 该地址被拒绝）
```

服务器内部的 accept 循环和 goroutine 对测试完全不可见——测试只通过 `addr` 和
`shutdown` 两个返回值与服务器交互，这正是"只验证外部行为、不检查实现细节"的练法。

### 文件计划（共 2 个文件，按编号顺序建）

最终目录长这样：

```text
tdd/tcpudp/
├── echo_test.go    # 全部测试（TCP 往返 / UDP 往返 / 关闭语义）
└── echo.go         # 两个 echo 服务器启动函数（本练习仅有的两个要实现函数）
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `echo_test.go` | 全部测试：TCP 往返、UDP 往返、关闭语义 | `TestTCPEcho`、`TestUDPEcho`、`TestTCPEchoShutdown` | **第 1 个建**（行为 2、3 往里追加测试） |
| 2 | `echo.go` | 两个 echo 服务器的启动函数 | `StartTCPEcho`、`StartUDPEcho` | 测试编译报错时 |

要写的函数一共 2 个，就是下面契约里的全部，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**你要写的每一个签名都在下面**；本练习不定义任何类型、常量或哨兵错误——
地址就是 `string`，关闭动作就是 `func()`，错误直接透传标准库的 `error`。
你唯一需要自己实现的是函数体（accept 循环、goroutine 都是函数体内部的事）；
如果写代码时发现要发明契约之外的类型或函数，说明走偏了。

**写在 `echo.go`：**（需要 `import "net"` 和 `"io"`）

```go
package tcpudp

// StartTCPEcho 在本机回环地址的随机空闲端口上启动一个 TCP echo 服务器：
// 接受连接后，把从连接中读到的数据原样写回（echo）。
// addr 是服务器的真实监听地址（形如 "127.0.0.1:54321"），供测试 Dial 使用；
// shutdown 关闭监听器、释放端口（之后再 Dial 该地址会被拒绝）；启动失败时 err 非 nil。
func StartTCPEcho() (addr string, shutdown func(), err error)

// StartUDPEcho 语义同上，但基于 UDP：
// 把收到的每个数据报原样写回发送方，一个 socket 服务所有客户端。
func StartUDPEcho() (addr string, shutdown func(), err error)
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 2 个函数：`StartTCPEcho`、`StartUDPEcho`
- 0 个类型、0 个哨兵错误：全部使用标准库类型，不要自己发明

---

## 二、任务单

### 行为 1：TCP 往返——写读相等，连接可复用

在 `echo_test.go` 里写 `TestTCPEcho`，覆盖下面两条用例。**测试自己写**：
先写测试，编译失败（`undefined: StartTCPEcho`）即 RED，再写最少实现变绿。

| 用例名 | 输入 | 期望 |
|---|---|---|
| 第一次往返 | `conn.Write("hello")` 后 `conn.Read` | 读回 `"hello"` |
| 连接复用第二次往返 | **同一条 conn** 再 `Write("world")` 后 `Read` | 读回 `"world"` |

### 行为 2：UDP 往返——无连接，每次操作都带地址

在 `echo_test.go` 里新增 `TestUDPEcho`，覆盖下面用例。**测试自己写**：
先写测试，编译失败即 RED，再写最少实现变绿。

| 用例名 | 输入 | 期望 |
|---|---|---|
| 数据报往返 | `WriteToUDP("ping", 服务器地址)` 后 `ReadFromUDP` | 读回 `"ping"`，且来源地址 == 服务器地址 |
| 第二个数据报 | 同一 socket `WriteToUDP("pong", 服务器地址)` 后 `ReadFromUDP` | 读回 `"pong"` |

UDP 不需要 Dial：客户端自己 `net.ListenUDP` 拿一个随机端口的 socket；
发送前先用 `net.ResolveUDPAddr` 把服务器地址字符串解析成 `*net.UDPAddr`。
实现 `StartUDPEcho` 的提示：`net.ListenUDP("udp", ...)` 直接得到 `*net.UDPConn`，
起一个 goroutine 循环 `ReadFromUDP` → 按来源地址 `WriteToUDP` 写回即可，没有 accept。

### 行为 3：关闭语义——shutdown 后连接被拒绝

新增 `TestTCPEchoShutdown`，覆盖下面用例。**测试自己写**：
先写测试，编译失败即 RED，再写最少实现变绿。

| 用例名 | 操作 | 期望 |
|---|---|---|
| 关闭后拒绝连接 | `StartTCPEcho()` → `shutdown()` → `net.Dial("tcp", addr)` | Dial 返回的 `err` 非 nil |

注意：这个测试里不能 `defer shutdown()`——要主动调它，才能验证关闭语义。

---

## 三、知识点总结

### 行为 1：TCP 往返——写读相等，连接可复用

1. **为什么测试用 `127.0.0.1:0`**：端口号写 0 是告诉内核"从临时端口池里分配一个当前空闲的端口"，
   监听成功后通过 `ln.Addr()`（本契约里就是返回的 `addr`）拿到真实地址。对比写死 `8080`：
   本机已有程序占用、或 CI 上两个测试并行抢同一个端口时，bind 直接失败，测试变成"看运气通过"
   （flaky）。`127.0.0.1` 是回环地址：流量不离开本机、不触发防火墙、不依赖真实网络环境——
   任何机器上跑测试行为都一致。**测试里写死端口是网络测试最常见的坑，记住 `:0` 这个手法。**
2. **`net.Conn` 就是 `io.ReadWriter`**：`net.Dial` 返回的 `conn` 和 [basic/file.go](../../basic/file.go)
   里 `os.Open` 返回的 `*os.File` 实现了同一组接口——`io.Reader` / `io.Writer`。所以 file.go 学过的
   读写语义原样适用：`Read(buf)` 返回实际读到的字节数 `n`，有效数据是 `buf[:n]`，不是整个 buf。
   服务器端能一行 `io.Copy(conn, conn)` 完成 echo，靠的也是这套抽象——`io.Copy` 不关心两端是文件
   还是网络连接。这是接口章（[basic/interface.go](../../basic/interface.go)）在标准库里最大规模的落地：
   **文件、网络连接、内存 buffer，同一套读写代码。**
3. **`t.Fatalf` 和 `t.Errorf` 的分工**：Dial、Write、Read 失败时后续步骤必然无法继续（连接都没有，
   读什么？），用 `t.Fatalf` 立即中断——它内部调 `runtime.Goexit` 终止当前测试 goroutine，
   **defer 仍会执行**，所以已注册的 `shutdown()` / `conn.Close()` 不会泄漏。两条内容断言互相独立，
   用 `t.Errorf`：第一条挂了第二条照样跑，一次看到全部失败。
4. **`defer Close` 防泄漏 + `SetReadDeadline` 防挂死**：每个测试都启动真实服务器、建立真实连接，
   不 Close 就泄漏文件描述符和端口，测试跑多了系统资源会被耗尽；defer 在测试函数返回时执行，
   中途 Fatal 也拦不住它。`SetReadDeadline` 解决另一个问题：如果服务器实现有 bug 不回数据，
   没有超时的 `Read` 会**永远阻塞**，测试挂死而不是失败——挂了你看不到任何断言输出。
   网络测试的铁律：**宁可超时失败，不可无限等待。**

### 行为 2：UDP 往返——无连接，每次操作都带地址

1. **listen / dial / accept 三段模型（TCP）**：socket 编程的经典三段，分工明确——
   `net.Listen` 创建的 Listener 是"前台接待"，本身不传数据，只负责等连接（三次握手由内核协议栈
   在后台完成，应用代码看不见）；`net.Dial` 由客户端发起握手，返回一条已建立的连接；
   `Accept()` 从已完成握手的队列里取出一条连接，返回的 `net.Conn` 才是传数据的"包间"。
   一条 TCP 连接 = 一对 Conn（客户端一个、服务器一个），各自独立读写——所以服务器必须给
   每条连接各开一个 goroutine，否则一条慢连接会堵死所有后续连接。
2. **UDP 无连接模型**：`net.ListenUDP` 直接得到 `*net.UDPConn`——没有 Listener 和 Conn 的分离，
   **没有 accept**。一个 UDP socket 同时接待所有客户端：`ReadFromUDP` 一次读一个数据报，
   顺带告诉你"谁发的"；`WriteToUDP` 每次发送都要带上目标地址。"无连接"的字面意思就在这里：
   socket 不绑定到某个对端，**每次操作都带地址**。对比记忆：

   | | TCP | UDP |
   |---|---|---|
   | 服务器启动 | `net.Listen` → 循环 `Accept()` | `net.ListenUDP` 一步搞定 |
   | 客户端 | `net.Dial`（建立连接） | `net.ListenUDP`（只开个 socket，不连谁） |
   | 收发 | `conn.Read` / `conn.Write`（地址在连接里） | `ReadFromUDP` / `WriteToUDP`（地址在每次调用里） |
   | 一条连接 vs 一对多 | 每客户端一条独立管道 | 一个 socket 服务所有客户端 |

   （`net.DialUDP` 也存在——"connected UDP"，之后可以不带地址用 `Read`/`Write`——但本练习故意不用它，
   就是要亲手体会"每次操作都带地址"的原生 UDP 模型。）
3. **TCP 字节流 vs UDP 数据报的边界差异**：这是本练习最要体会的一点。TCP 是**字节流**——
   它只保证字节按序到达，不保留"几次写"的边界：写两次 `"hello"`，对端可能一次 `Read` 读到
   `"hellohello"`（粘包），也可能分三次读到（半包）。所以 TCP 上的应用层必须自己定义消息边界
   （长度前缀、分隔符——这正是教程 ch08"通信协议解析"的主题）。UDP 是**数据报**——
   一次 `WriteToUDP` = 一个数据报，对端一次 `ReadFromUDP` 完整读出它，**写一次读一次严格对应**。
   注意：行为 1 的 TCP 测试"写一次读一次"恰好能过，靠的是本机回环 + 数据小 + 立即读，
   **不是 TCP 的保证**——别把巧合当契约。
4. **UDP 易错点——缓冲区截断**：`ReadFromUDP` 的 buf 小于数据报时，**多余部分被静默丢弃**，
   不报错（`n` 就是 buf 长度）。UDP 数据报理论上限约 64KB，本练习 1024 字节足够；以后处理
   不可信输入时，要么 buf 给足，要么按协议先读长度。

### 行为 3：关闭语义——shutdown 后连接被拒绝

1. **关闭语义 = 资源释放的可观测证明**：`shutdown()` 内部调用 `listener.Close()`，内核随即释放端口；
   此后再 `Dial` 这个地址，内核直接回 RST 包，Dial 立即返回"连接被拒绝"错误。测试用"Dial 失败"
   这个**外部可观测行为**断言"端口已释放"这个内部状态——这正是 TDD 的核心思想：不检查实现细节，
   只验证行为。
2. **不要字符串匹配错误文案**：Windows 上这个错误是 `connectex: ... refused`，Linux 上是
   `connection refused`——文案因系统而异，断言 `err != nil` 即可。以后学了 error 体系
   （[tdd/errhandling](../errhandling/README.md)）可以用 `errors.Is` 做跨平台的精确判断，本练习判非 nil 就够。
3. **`shutdown` 是一个闭包**：契约把"清理动作"打包成函数值返回，它捕获了内部的 listener 变量——
   调用方不需要知道服务器内部有什么资源，拿起来就调。这是 [basic/func_type.go](../../basic/func_type.go)
   闭包在工程里的典型用法：**资源和它的清理函数一起出厂**。
4. **UDP 为什么没有这条用例**：UDP 无连接，没有"拒绝连接"的语义——shutdown 后往已关闭的端口发
   数据报，只是没人接收，客户端 `Read` 一直等到超时。"连接被拒绝"是 TCP 面向连接模型独有的行为，
   对比之下更能体会两个模型的差异。

### `net` 包 API 速查

| 函数 / 方法 | 作用 |
|---|---|
| `net.Listen("tcp", addr)` | TCP 监听；addr 端口写 0 时内核分配空闲端口 |
| `net.Dial("tcp", addr)` | 建立 TCP 连接，返回 `net.Conn`（实现了 `io.ReadWriter`） |
| `Listener.Accept()` | 取出一条已就绪连接；监听器 Close 后 Accept 返回错误（退出 accept 循环的信号） |
| `ln.Addr().String()` | 拿真实监听地址（`"127.0.0.1:54321"` 形式），配合端口 0 使用 |
| `net.ListenUDP("udp", addr)` | 创建 UDP socket，无 accept，一个 socket 服务所有客户端 |
| `net.ResolveUDPAddr("udp", s)` | `"host:port"` 字符串 → `*net.UDPAddr` |
| `ReadFromUDP` / `WriteToUDP` | 无连接收发，每次操作都带地址；读缓冲区过小会静默截断 |
| `conn.SetReadDeadline(t)` | 读超时：防止服务器 bug 导致测试永远挂死 |

### TCP vs UDP 一句话

TCP：listen/dial/accept 三段、每连接一条管道、字节流无消息边界、有序可靠；
UDP：一个 socket 收发所有人、每次操作带地址、数据报边界严格对应、不保证送达。

### 与已有笔记的呼应

- `Read` 返回 `n`、有效数据是 `buf[:n]`、`io.Copy` 流式搬运 → [basic/file.go](../../basic/file.go)（io 接口一处学、处处用）
- `shutdown` 闭包捕获资源变量 → [basic/func_type.go](../../basic/func_type.go)
- `defer Close` 防泄漏 → [basic/defer.go](../../basic/defer.go) + file.go 的 `closeFile`
- accept 循环 + 每连接一个 goroutine → [basic/goroutine.go](../../basic/goroutine.go)，`-race` 守护

### 与书目的对应

- **教程 ch08 传输层**：socket 模型的落地。教程里讲的"TCP 三次握手""TCP vs UDP 六区别"，
  在本练习变成可运行的代码——listen/dial/accept 对应握手之后的连接建立，数据报边界对应
  "UDP 面向报文"，shutdown 后 RST 拒连对应"四次挥手"的收尾。
- 概念阅读清单见根 README 第四节：阶段 4 开始前读教程 ch08 网络理论（OSI 七层、TCP 报文、
  TIME_WAIT 等），与本练习的动手体验互相印证。

---

## 四、验收标准

```bash
go test ./tdd/tcpudp -v       # 全绿
go vet ./tdd/tcpudp           # 无警告
go test ./tdd/tcpudp -race    # 服务器内部跑 goroutine，必须无数据竞争
go test ./tdd/tcpudp -cover   # 核心逻辑覆盖（目标 ≥80%，不盲目追 100%）
```

## 五、完成后自查（能口头回答才算过）

1. 为什么测试服务器监听 `127.0.0.1:0` 而不是写死 `8080`？真实端口号从哪拿到？
2. TCP 的 listen / dial / accept 三段各做什么？为什么 UDP 没有 accept？
3. TCP 字节流和 UDP 数据报在"消息边界"上有什么区别？行为 1 里 TCP"写一次读一次"能过，是 TCP 的保证吗？
4. `net.Conn` 和 `*os.File` 有什么共同之处？这体现了 Go 接口设计的什么思想？
5. 为什么网络测试要加 `SetReadDeadline`？不加会发生什么？
6. `shutdown()` 之后 Dial 为什么失败？这个断言证明了什么？
7. 为什么本练习必须带 `-race` 跑？服务器内部有哪些并发单元？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把 ch08"通信协议解析"从 ❌ 改成 ◐（传输层 socket 已由本练习落地，HTTP 应用层待 [tdd/http](../http/README.md) 补上）。
