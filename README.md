# Go 学习笔记总览

本仓库是 Go 语言学习项目，分四个区域：

| 目录 | 定位 | 运行方式 |
|---|---|---|
| `basic/` | 语法知识点实验区（每文件独立 `package main` + `func main`） | 单文件执行：`go run basic/xxx.go` |
| `tdd/` | `testing` 标准库工具与 TDD 行为驱动练习（单测/表驱动/Benchmark/Fuzz/Coverage） | `go test ./tdd/xxx` |
| `docs/` | 书单与易错点笔记 | — |

> 注意：`basic/` 所有文件同属 `package main` 且各有 `func main`，因此**不能** `go build ./basic/`，只能逐文件 `go run`。这是实验区的刻意取舍，不是错误。

---

## 一、basic/ 知识脉络（含精确跳转）

按推荐学习路径分为 10 个模块。链接可精确跳转到对应知识点的行号。

### 1. 语法基础

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [var.go](basic/var.go) | 变量声明：`var`、批量声明、短声明 `:=`（仅局部）、匿名变量 `_` | [批量声明](basic/var.go#L23) · [匿名变量](basic/var.go#L39) |
| [const.go](basic/const.go) | 常量必须赋值；批量声明沿用上值；`iota` 计数器 | [iota 与 `_` 跳过](basic/const.go#L37) · [一行多 iota 按列累加](basic/const.go#L48) |
| [print.go](basic/print.go) | `Println`/`Print`/`Printf` 区别；`%v`/`%+v`/`%#v`/`%T` | [格式化动词对比](basic/print.go#L25) |
| [operator.go](basic/operator.go) | `i++` 是语句不是表达式且无 `++i`；逻辑短路；位运算 | [位运算演示](basic/operator.go#L38) |
| [type_conv.go](basic/type_conv.go) | `fmt.Sprintf` 万能转 string；`strconv.Format*`/`Parse*` | [FormatFloat 参数详解](basic/type_conv.go#L25) |

### 2. 基本数据类型

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [int.go](basic/int.go) | int8~64/uint 取值范围；`unsafe.Sizeof`；**高低位转换静默溢出** | [溢出陷阱](basic/int.go#L32) |
| [float.go](basic/float.go) | float32/64；精度控制 `%.Nf`；**精度丢失**；`shopspring/decimal` 精确计算 | [精度丢失](basic/float.go#L35) · [decimal 用法](basic/float.go#L43) |
| [bool.go](basic/bool.go) | bool 零值 false；int 不能转 bool；不参与数值运算 | [注释要点](basic/bool.go#L4) |
| [string.go](basic/string.go) | 转义符；反引号多行串；`strings` 包 Split/Join/Contains/Index | [strings 常用函数](basic/string.go#L40) |
| [byte_rune.go](basic/byte_rune.go) | byte=uint8 / rune=UTF-8 字符；**按 byte 遍历汉字乱码**；`[]rune` 改字符串 | [乱码演示](basic/byte_rune.go#L32) · [`unsafe.Sizeof` 对 string 的坑](basic/byte_rune.go#L24) |

### 3. 流程控制

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [if.go](basic/if.go) | `{}` 不可省、`{` 必须同行；短声明限定作用域 | [短声明写法](basic/if.go#L20) |
| [for.go](basic/for.go) | for 三段式及执行顺序；Go 无 while；for range 遍历字符串 | [类 while 写法](basic/for.go#L30) |
| [switch.go](basic/switch.go) | 默认不穿透；case 多值；无表达式 switch；`fallthrough` 只穿透一次 | [fallthrough](basic/switch.go#L58) |
| [break_continue_goto.go](basic/break_continue_goto.go) | break/continue 带标签跳出外层循环；goto | [break 标签](basic/break_continue_goto.go#L38) |

### 4. 函数与错误处理

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [func.go](basic/func.go) | 参数简写；可变参数 `...int`（是切片、只能放最后）；多返回值；命名返回值+裸 return | [可变参数](basic/func.go#L15) · [命名返回值](basic/func.go#L37) |
| [func_type.go](basic/func_type.go) | 函数类型 `type calcType func(...)`；函数作参数/返回值；递归；**闭包**；匿名/自执行函数 | [闭包 adder](basic/func_type.go#L69) · [匿名函数](basic/func_type.go#L110) |
| [type_func.go](basic/type_func.go) | 自定义类型 `type MyInt int`；给自定义类型定义方法；非本地类型不能加方法 | [方法定义](basic/type_func.go#L8) |
| [defer.go](basic/defer.go) | defer 栈（后进先出）；**命名返回值才受 defer 影响**；defer 参数注册时立即求值 | [f1~f6 六连对比](basic/defer.go#L5) · [立即求值](basic/defer.go#L88) |
| [panic_recover.go](basic/panic_recover.go) | panic/recover 模式；recover 只在 defer 函数中直接调用才有效 | [实战模拟](basic/panic_recover.go#L36) |

### 5. 复合数据类型

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [array.go](basic/array.go) | 长度是类型一部分；四种初始化；**数组是值类型 vs 切片是引用类型**；多维数组 | [值类型验证](basic/array.go#L39) |
| [slice.go](basic/slice.go) | len/cap；`[:]` 左闭右开；`make`；**append 扩容策略**；`copy` 深拷贝；`append(s[:i], s[i+1:]...)` 删除 | [扩容与 size class](basic/slice.go#L60) · [Go1.18 扩容注释](basic/slice.go#L80) |
| [slice_sort.go](basic/slice_sort.go) | 选择/冒泡排序手写；`sort.Ints` 等；`sort.Reverse` 降序 | [sort 包](basic/slice_sort.go#L31) |
| [map.go](basic/map.go) | 无序引用类型、必须初始化；双返回值判 key；`delete`；**slice 套 map 需逐个 make**；key 排序 | [slice 套 map 的坑](basic/map.go#L62) · [map 排序手法](basic/map.go#L107) |
| [pointer.go](basic/pointer.go) | `&`/`*`；值传递 vs 指针传递；**new vs make 区别** | [new vs make](basic/pointer.go#L39) |

### 6. 结构体与方法

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [struct.go](basic/struct.go) | 5 种实例化方式；`new` 返回指针；`.` 自动解引用语法糖；结构体是值类型 | [5 种实例化](basic/struct.go#L12) |
| [struct_func.go](basic/struct_func.go) | 值接收者 vs 指针接收者；**修改字段必须用指针接收者** | [指针接收者](basic/struct_func.go#L16) |
| [struct_more.go](basic/struct_more.go) | 匿名字段；字段为 slice/map 默认 nil 须 make；嵌套嵌入 + **字段/方法提升**；模拟继承 | [字段提升](basic/struct_more.go#L103) · [继承模拟](basic/struct_more.go#L107) |
| [struct_json.go](basic/struct_json.go) | json tag；**私有字段不参与序列化**；`Marshal`/`Unmarshal`（传指针）；嵌套结构体 | [序列化](basic/struct_json.go#L37) |

### 7. 接口

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [interface.go](basic/interface.go) | 接口定义与实现；**值接收者 vs 指针接收者对赋值的影响**；接口作参数；类型断言 | [接收者对比](basic/interface.go#L17) · [类型断言](basic/interface.go#L41) |
| [interface_empty.go](basic/interface_empty.go) | 空接口 `interface{}` 表示任意类型；**`a.(type)` 只能用于 switch**；逗号-ok 断言 | [type switch](basic/interface_empty.go#L14) · [ok 断言](basic/interface_empty.go#L62) |
| [interfaces_struct.go](basic/interfaces_struct.go) | 一个类型实现多接口；接口嵌套；混合接收者导致只有指针满足接口 | [接口嵌套](basic/interfaces_struct.go#L13) |

### 8. 并发编程

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [goroutine.go](basic/goroutine.go) | `go` 关键字；`sync.WaitGroup`；**主协程退出全部中止**；协程内 panic 须自己 recover 且 `defer wg.Done()`；`GOMAXPROCS` | [推荐写法 testShow](basic/goroutine.go#L32) · [panic 处理](basic/goroutine.go#L40) |
| [goroutine_channel.go](basic/goroutine_channel.go) | channel 引用类型/FIFO；缓冲 channel；阻塞与死锁；**for range 前必须 close**；定向 channel `chan<-`/`<-chan`；`select`+`default`；素数筛并发模型 | [定向管道](basic/goroutine_channel.go#L37) · [select](basic/goroutine_channel.go#L179) · [素数筛](basic/goroutine_channel.go#L141) |
| [goroutine_lock.go](basic/goroutine_lock.go) | `sync.Mutex` 保护临界区；map 并发写 panic；`sync.RWMutex` 读共享写独占 | [RWMutex](basic/goroutine_lock.go#L47) |

### 9. 反射与泛型

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [reflect.go](basic/reflect.go) | `TypeOf`（Type/Name/Kind）；`ValueOf` 按 Kind 取值；**反射改值三要素：传指针 + Elem + CanSet** | [改值三要素](basic/reflect.go#L53) |
| [reflect_struct.go](basic/reflect_struct.go) | 结构体反射：Field/Tag/NumField；`Method(i).Type.NumIn()` 含接收者；传指针才见指针接收者方法 | [反射调方法](basic/reflect_struct.go#L84) |
| [generic.go](basic/generic.go) | 泛型函数 `[T comparable]`；自定义约束 `|` 联合 / **`~` 底层类型**；`cmp.Ordered`；泛型类型 `Stack[T]` | [约束对比](basic/generic.go#L37) · [泛型栈](basic/generic.go#L97) |

### 10. 工程化与其他

| 文件 | 知识点 | 关键跳转 |
|---|---|---|
| [file.go](basic/file.go) | 文件读写三条路线（`os` 分块 / `bufio` 缓冲 / `os.ReadFile` 一次性）；`io.Copy`；目录操作（MkdirAll/ReadDir/WalkDir） | [三种读法](basic/file.go#L98) · [bufio 按行读的 EOF 坑](basic/file.go#L126) · [目录操作](basic/file.go#L223) |
| [time.go](basic/time.go) | `time.Now`；**格式化模板 `2006/01/02 15:04:05`**；时间戳互转；`Duration`；Ticker | [格式化模板](basic/time.go#L22) · [Ticker](basic/time.go#L78) |
| [mod/](basic/mod/mod.go) | 跨包导入与别名；`go mod` 流程；**init 初始化顺序**（被依赖包先 init） | [init 顺序](basic/mod/mod.go#L28) |

---

## 二、高频易错点速查

这些是从笔记中提炼的"踩坑点"，建议优先掌握：

1. defer 只影响**命名返回值**，且参数注册时立即求值 → [defer.go f1~f6](basic/defer.go#L5)
2. 浮点数精度丢失，金钱计算用 `decimal` → [float.go#L35](basic/float.go#L35)
3. 高位 int 转低位**静默溢出**不报错 → [int.go#L32](basic/int.go#L32)
4. 字符串按 byte 遍历汉字乱码，用 range（rune）→ [byte_rune.go#L32](basic/byte_rune.go#L32)
5. slice `append` 扩容后**与原数组脱钩**，容量按 size class 对齐 → [slice.go#L60](basic/slice.go#L60)
6. map/slice 是引用类型，数组/结构体是值类型 → [map.go#L96](basic/map.go#L96)
7. slice 套 map，元素 map 是 nil 必须逐个 make → [map.go#L62](basic/map.go#L62)
8. 修改字段必须指针接收者；接口赋值时接收者类型决定值/指针能否赋 → [interface.go#L17](basic/interface.go#L17)
9. 类型分支 `x.(type)` **只能写在 switch 里** → [interface_empty.go#L14](basic/interface_empty.go#L14)
10. 小写私有字段不参与 JSON 序列化 → [struct_json.go#L12](basic/struct_json.go#L12)
11. 主协程退出，所有 goroutine 被中止 → [goroutine.go#L59](basic/goroutine.go#L59)
12. `for range` 遍历 channel 前必须 `close`，否则死锁 → [goroutine_channel.go#L113](basic/goroutine_channel.go#L113)
13. `make` 只用于 slice/map/channel，`new` 返回任意类型指针 → [pointer.go#L39](basic/pointer.go#L39)
14. 泛型约束没有 `~` 时自定义底层类型不满足 → [generic.go#L48](basic/generic.go#L48)
15. 反射改值：传指针 + `Elem()` + `CanSet()` → [reflect.go#L53](basic/reflect.go#L53)

---

## 三、对照《For-learning-Go-Tutorial》的覆盖检查

对照书单中的 [KeKe-Li/For-learning-Go-Tutorial](https://github.com/KeKe-Li/For-learning-Go-Tutorial) 章节目录逐章核对：

| 教程章节 | 覆盖 | 对应位置 / 缺口 |
|---|---|---|
| 例子入门 / 基本结构 | ✅ | 全部 basic 文件均可运行；[var.go](basic/var.go)、[const.go](basic/const.go)、[mod/](basic/mod/mod.go) |
| ch02 基本数据类型 | ✅ | int/float/bool/string/byte_rune/type_conv |
| ch03 复合数据类型 | ✅ | array/slice/map/pointer/struct 系列 |
| ch04 函数 | ✅ | func/func_type/defer/panic_recover |
| ch05 方法 | ✅ | struct_func/type_func |
| ch06 接口 | ✅ | interface 系列三文件 |
| ch07 反射 | ✅ | reflect/reflect_struct |
| ch09 Channel | ✅ | goroutine_channel |
| ch10 Goroutine 并发 | ✅ | goroutine + goroutine_lock |
| ch11 包详解 | ✅ | [mod/](basic/mod/mod.go)（含 init 顺序） |
| ch19 Sync.WaitGroup | ✅ | [goroutine.go#L15](basic/goroutine.go#L15) |
| ch16 排序算法 | ◐ 部分 | 只有选择/冒泡 + sort 包，缺快排/插排等及性能对比 |
| spec slice 注意事项 | ◐ 部分 | slice.go 覆盖大半；缺"函数内 append 扩容失效""共享底层数组互相污染"两个经典陷阱 |
| 编程规范（uber-go/guide） | ◐ 部分 | [books.md](docs/books.md) 已列书单，未逐条实践 |
| ch08 通信协议解析 | ❌ | 仅有 JSON 编解码（struct_json），无 HTTP/RPC；将由 `tdd/http` 补上 |
| **ch17 Go 程序测试** | ❌ **最大缺口** | 全仓库没有任何 `*_test.go` —— 正是 TDD 要求的核心 |
| **error 处理体系** | ❌ | 仅 panic_recover 用到 `errors.New`；缺 `fmt.Errorf %w` 包装、自定义错误类型、哨兵错误 |
| ch13 逃逸分析 | ❌ | 未学；用 `go build -gcflags="-m -m"` 观察 |
| ch18 Sync.Map | ❌ | 只学了 Mutex/RWMutex |
| context 包 | ❌ | 并发取消/超时控制，工程必备（教程并发章延伸） |
| ch20 抢占式调度 / GC | ❌ | 进阶原理，仅 [goroutine.go#L66](basic/goroutine.go#L66) 碰过 `GOMAXPROCS` |
| ch12 Grpc/Protobuf、ch14 Etcd、ch15 TiDB | ❌ | 工程应用层，建议基础扎实后按需学 |

### 遗漏清单（按优先级）

- **P0（已建立，需按目录逐题验收）**：`testing` 包全流程（TestXxx / 表驱动 / t.Run / Benchmark / Example）；error 体系（`errors.New`、`fmt.Errorf`+`%w`、自定义错误、`errors.Is/As`）
- **P1（一个月内补）**：context 包；Sync.Map；逃逸分析；slice 两个经典陷阱补充实验
- **P2（随 TDD 练习自然覆盖）**：HTTP/`net/http`/`httptest`（tdd/http）；JSON 已会
- **P3（进阶，缓学）**：Grpc/Protobuf、Etcd、TiDB、GC 与调度原理、uber 规范逐条过

### 笔记中已发现的疑义点

- 多文件混用内置 `println`（bool/int/array/for/switch.go）——应统一 `fmt.Println`，内置 println 不保证兼容性
- [print.go#L14](basic/print.go#L14) 注释"Print 输出多个中间没有空格"不准确：Print 只在**两个非字符串操作数之间**才加空格
- [goroutine_channel.go#L127](basic/goroutine_channel.go#L127) 在 channel 仍有余量时 close，容易误读为"close 清空数据"——close 只是不再接收发送，已缓冲数据仍可读完
- [struct_json.go#L37](basic/struct_json.go#L37) `json.Marshal` 忽略了 error（与 L63 规范写法不一致）；[type_conv.go#L45](basic/type_conv.go#L45) Parse 系列同样忽略错误——学 error 体系后应回来修正
- [goroutine_lock.go](basic/goroutine_lock.go) 中 `wgLock.Done()` 未用 defer，临界区 panic 会死锁（对比 [goroutine.go#L33](basic/goroutine.go#L33) 的推荐写法）

已修正的小问题：`goroutinue_lock.go` 文件名拼写 → `goroutine_lock.go`；defer.go 错别字"旧确定"；float.go 输出标签误写 `y`；tools_by 的 init 打印文件名不符。

---

## 四、目录划分评估与 TDD 驱动学习计划

### 目录划分结论：**合理，保持现状**

- `basic/` 保留为语法实验区（单文件 `go run`），**不要**为了 TDD 重写它——语法演示改成可测包是过度设计
- `tdd/` 同时承载 `testing` 标准库工具学习与行为驱动练习；仓库当前没有独立的 `testing/` 目录
- 2026-08 起 `tdd/` 已清空重建：练习按统一模板（需求规格 + 接口契约 + 用例表格 + 知识点映射）逐个创建，见下文路线
- 模板分两档：阶段 0（testbasic / errhandling / fuzzlab）带「手把手起步」与内联知识点；自阶段 1 起任务单不再给完整测试代码（只有行为目标 + 用例表 + 验收命令），知识点集中到第三节「知识点总结」，与任务单分离

### TDD 练习总目录（覆盖教程全部章节 + 现代工具链）

排序原则：先地基（测试与错误处理是后续一切练习的工具）→ 值语义 → 接口设计 → 并发 → 网络协议 → 算法性能。每个练习只引入一个新变量。
循环不变：RED → GREEN → REFACTOR。类型：机制学习型给固定接口契约，设计驱动型只给行为需求（见 [tdd/testbasic](tdd/testbasic/README.md) 开头说明）。
**🔺 = basic/ 未涉及的重点缺口，优先练；无标记 = 已学知识的 TDD 巩固。**
权威性核对：已与 TDD 领域权威教程 [learn-go-with-tests](https://github.com/quii/learn-go-with-tests) 及 [Go 1.25](https://go.dev/doc/go1.25) / [Go 1.26](https://go.dev/doc/go1.26) 官方 Release Notes 交叉核对，现代特性落点已标注在各行。

**阶段 0：TDD 地基**

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 1 | [tdd/testbasic](tdd/testbasic/README.md) | 机制 | `TestXxx` / 表驱动 / `t.Run` / Benchmark（`b.N` 与 Go 1.24 的 `b.Loop`）/ `-cover` | —（全新） | 教程 ch17；uber·表驱动测试 | ✅ 已建 |
| 2 | 🔺[tdd/errhandling](tdd/errhandling/README.md) | 机制 | 哨兵错误 / `%w` 包装 / 自定义错误 / `errors.Is/As/Join`；Go 1.26 泛型版 `errors.AsType` | panic_recover、interface_empty | uber·Errors 四节；EG·15 错误 | ✅ 已建 |
| 3 | 🔺[tdd/fuzzlab](tdd/fuzzlab/README.md) | 机制 | `FuzzXxx` 模糊测试：种子语料 / `-fuzz` 参数 / 崩溃语料；基于属性的测试思想（roman numerals kata） | —（全新） | learn-go-with-tests·property based tests；Go 1.18+ fuzzing | ✅ 已建 |

**阶段 1：值语义与复合类型**

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 4 | 🔺[tdd/slicelab](tdd/slicelab/README.md) | 机制 | 三段切片 `s[low:high:max]`、共享底层数组互污、append 扩容脱钩、函数内 append 失效、nil slice 可用 | slice、array | 教程 spec·03 全章；uber·nil 是有效 slice / 边界拷贝 | ✅ 已建 |
| 5 | [tdd/wallet](tdd/wallet/README.md) | 机制 | struct + 指针接收者 + error 综合 | struct、struct_func、pointer | EG·10 方法；uber·receiver 与接口 | ✅ 已建 |
| 6 | [tdd/dictionary](tdd/dictionary/README.md) | 机制 | map CRUD + 哨兵错误 + interface 抽象 | map、interface | 教程 ch03/ch06；EG·11 接口 | ✅ 已建 |

**阶段 2：接口与设计**

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 7 | [tdd/clock](tdd/clock/README.md) | 设计 | 接口设计 + 依赖注入 + fake 时钟；🔺对照 Go 1.25 `testing/synctest` 气泡（不等待、不抖动的并发时间测试） | interface 系列、time | EG·11；uber·避免可变全局变量；learn-go-with-tests·Time + synctest | ✅ 已建 |
| 8 | 🔺[tdd/sortiface](tdd/sortiface/README.md) | 机制 | `sort.Interface` 三方法（Len/Less/Swap）、自定义类型排序、`sort.Reverse`、`sort.Stable` 与稳定性 | interface、slice_sort | 教程 ch16 前半 | ✅ 已建 |
| 9 | [tdd/genericlab](tdd/genericlab/README.md) | 机制 | 泛型工具函数（Map/Filter/Reduce）；🔺`slices`/`maps`/`cmp` 标准库（Go 1.21）、range-over-func 迭代器（Go 1.23） | generic | learn-go-with-tests·generics 两章 | ✅ 已建 |

**阶段 3：并发**（全部要求 `-race` 通过）

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 10 | [tdd/concurrency](tdd/concurrency/README.md) | 机制 | close 三规则（重复 close / 向已关闭发送 panic、已关闭读出零值）+ ok-idiom、无缓冲同步语义、`time.After` 超时、select+break+label、WaitGroup 陷阱（Add 位置、计数为负）与 Go 1.25 `WaitGroup.Go` 🔺 | goroutine、goroutine_channel、goroutine_lock | 教程 ch09 前半 / ch19 深化；EG·14 | ✅ 已建 |
| 11 | 🔺[tdd/pipeline](tdd/pipeline/README.md) | 机制 | done channel 广播退出、扇出/扇入、pipeline 三阶段、worker pool、限流模式（`time.Tick`） | goroutine_channel | 教程 ch09 后半 / ch10；uber·goroutine 生命周期 | ✅ 已建 |
| 12 | 🔺[tdd/context](tdd/context/README.md) | 机制 | `WithCancel` / `WithTimeout` / `WithDeadline` / `WithValue`、取消传播树、测试中超时控制（可对照 synctest） | goroutine_channel 的 select | 教程 ch09 延伸（工程必备） | ✅ 已建 |
| 13 | 🔺[tdd/syncmap](tdd/syncmap/README.md) | 机制 | `sync.Map` 的 Load/Store/Delete/Range、read/dirty 双结构与延迟删除思想、对比 RWMutex+map 与 `sync/atomic` 计数器 | goroutine_lock | 教程 ch18 全章；uber·零值 Mutex / atomic | ✅ 已建 |

**阶段 4：网络与协议**

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 14 | 🔺[tdd/http](tdd/http/README.md) | 机制 | `net/http` handler、`httptest`、JSON 断言、状态码与 GET/POST 语义落地 | struct_json、interface | 教程 ch08 应用层；EG·16 Web 服务器 | ✅ 已建 |
| 15 | 🔺[tdd/tcpudp](tdd/tcpudp/README.md) | 机制 | `net` 包：TCP echo 与 UDP echo 的测试写法（`127.0.0.1:0` 随机端口），亲身体会 TCP/UDP 区别 | file（io 读写概念） | 教程 ch08 传输层 | ✅ 已建 |

**阶段 5：算法与性能**

| # | 练习 | 类型 | 主知识点 | 巩固的 basic 笔记 | 对应书目 | 状态 |
|---|---|---|---|---|---|---|
| 16 | 🔺[tdd/sortbench](tdd/sortbench/README.md) | 机制 | 插入 / 归并 / 快排 / 堆排序的 TDD 实现、Benchmark 对比（可用 `b.Loop`）、`b.StopTimer/StartTimer`、顺带验证 uber 性能条（指定容量、strconv vs fmt） | slice_sort | 教程 ch16 后半；uber·性能章 | ✅ 已建 |

> 全部 16 个练习任务单已创建完毕（含 3 个 basic/ 观察实验程序）。已在 basic 掌握的知识点（反射、文件、time 等）不单独设练习，会在后续练习中自然复用。
> 现代工具链落点汇总：`b.Loop`→练习 1/16；`errors.AsType`→练习 2；`testing/synctest`→练习 7/12；`WaitGroup.Go`→练习 10；`slices`/`maps`/`cmp`、range-over-func→练习 9；`log/slog`（Go 1.21 结构化日志，Go 1.26 增 `MultiHandler`）→ 随练习 14 顺带使用。表格中的“✅ 已建”表示任务单/目录已建立，不代表学习者已经完成实现或验收。

**教程章节覆盖核对**：ch01~07、ch11、ch19 已由 basic 覆盖，由巩固练习承载；ch08→练习 14/15、ch09→练习 10/11、ch10→练习 11+观察实验、ch13/16/17/18/20、spec·02/03 均有练习或实验落点；“有落点”不等于“已完成验证”，需要以对应目录的实际源码和测试结果为准；ch12/14/15 缓学（见文末）。learn-go-with-tests 的 "Build an application" 篇（HTTP server 迭代 / JSON 路由 / IO 持久化 / 命令行结构 / WebSockets / 验收测试）属项目实战，当前仍未覆盖。

**不适合 TDD 的部分（配套安排）**

观察实验（放 `basic/`）：

- 🔺`basic/escape.go`：逃逸分析（教程 ch13）——`go build -gcflags="-m -l"` 观察五类逃逸：取地址返回、interface 装箱、闭包引用、大对象、动态大小
- 🔺`basic/sched_trace.go`：调度与抢占观察（教程 ch10/ch20）——`GODEBUG=schedtrace=1000` 看 GPM 状态；`runtime/trace` + `go tool trace` 观察 `GOMAXPROCS(1)` 下纯计算 goroutine 的异步抢占（go1.14 特性）
- 🔺`basic/gc_observe.go`：GC 观察（教程 spec·02）——`GODEBUG=gctrace=1`、`runtime.ReadMemStats`、`runtime.GC()` 主动触发、`GOGC` 调参对比；注意 Go 1.26 默认启用 Green Tea GC（按内存 span 扫描，开销降约 40%），观察时与旧版行为对比
- 顺手修正：`struct_json.go` / `type_conv.go` 忽略 error 的写法（做完练习 2 后回头改）

概念阅读（无法 TDD 的理论，配合对应阶段读）：

| 时机 | 阅读内容 |
|---|---|
| 阶段 3 开始前 | 教程 ch10：进程/线程/协程、并发 vs 并行、GPM 模型与工作窃取；ch20：sysmon、协作式抢占标记、异步抢占信号 |
| 阶段 4 开始前 | 教程 ch08 网络理论：OSI 七层、IP/ICMP/ARP、TCP 报文与三次握手四次挥手、TIME_WAIT、TCP vs UDP 六区别、端口与 socket、HTTP 报文结构 / 状态码族 / GET vs POST / 版本差异 / Cookie 与 Session / Web 缓存 / HTTPS |
| 阶段 5 开始前 | 教程 spec·02：GC 四算法（引用计数/标记清除/复制/分代）、三色标记流程、写屏障、STW、GOGC 调参；ch16：排序稳定性概念 |
| 全程（规范类） | uber·规范章（包名 / import 分组 / 减少嵌套 / 缩小作用域等）；EG·02 格式化 / 03 注释 / 04 命名；uber·Errors 四节随练习 2、表驱动测试随练习 1、其余条目随对应阶段 |

缓学（工程应用层，需外部服务环境，完成全部阶段后按需）：教程 ch12 Grpc/Protobuf、ch14 Etcd、ch15 TiDB。

> 书目缩写：教程 = [For-learning-Go-Tutorial](https://github.com/KeKe-Li/For-learning-Go-Tutorial)；uber = [uber_go_guide_cn](https://github.com/xxjwxc/uber_go_guide_cn)；EG = [Effective Go 中英双语版](https://github.com/bingohuang/effective-go-zh-en)

### 每个练习的验收标准

```bash
go test ./tdd/xxx -v          # 全绿
go vet ./tdd/xxx              # 无警告
go test ./tdd/xxx -race       # 并发练习必须无数据竞争
go test ./tdd/xxx -cover      # 核心逻辑覆盖（目标 ≥80%，不盲目追 100%）
```

完成一个练习后，回到本文档第一节，把对应知识点的链接再点一遍自查——能用测试描述清楚的知识，才算真正掌握。
