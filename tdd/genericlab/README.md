# GenericLab TDD —— 泛型三件套与现代标准库

目标：泛型语法在 [basic/generic.go](../../basic/generic.go) 已经见过一遍，本练习把它放进
TDD 流程里真正用熟。先手写函数式三件套 Map / Filter / Reduce 吃透**类型参数机制**，
再认识两个现代事实：Go 1.21 的 `slices` / `maps` / `cmp` 标准库让多数手写工具函数不再必要；
Go 1.23 的 range-over-func 让自定义迭代器成为一等公民。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求规格（接口契约固定，照此实现）；第二节是纯任务单——只给行为目标、用例表和验收命令，测试代码全部自己写；第三节是知识点讲解，做之前通读或卡壳时查阅，做完后对照自查。

---

## 一、需求规格

### 核心功能

实现一个最小的**泛型工具函数包**，它对外提供四个能力：

- **Map**：把切片的每个元素映射成另一个值（允许变成另一种类型），返回新切片
- **Filter**：按谓词筛出元素，保持原相对顺序，返回新切片
- **Reduce**：从初始值出发，把整个切片折叠成一个值（结果类型可与元素类型不同）
- **All**：返回切片的 range-over-func 迭代器（Go 1.23，🔺行为 5 才做）

**没有 `main` 函数。** 产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/genericlab` 就是它的运行方式，验收者是测试，不是人。

### 调用关系（谁在调用谁）

```text
测试代码 ──► Map / Filter / Reduce（generic.go）                  行为 1–3
测试代码 ──► slices / maps 标准库函数（对照实验，不写实现）        行为 4
测试代码 ──► All（iterator.go）──► 可喂给 slices.Collect 等       行为 5
```

四个能力全是包级函数：没有结构体、没有接口、没有构造器，测试直接调用——
本练习的复杂度全在类型参数机制上，不在 API 组织上。

### 文件计划（共 5 个文件，按编号顺序建）

最终目录长这样：

```text
tdd/genericlab/
├── generic_test.go    # 行为 1–3：Map / Filter / Reduce 的全部测试
├── generic.go         # 行为 1–3：三件套实现（本练习的核心）
├── stdlib_test.go     # 行为 4：标准库对照实验（只有测试，没有对应实现文件）
├── iterator_test.go   # 行为 5：All 的全部测试
└── iterator.go        # 行为 5：All 的实现
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `generic_test.go` | Map / Filter / Reduce 的全部测试 | `TestMap`（平方、R≠T 两个用例）、`TestFilter`、`TestReduce`（覆盖行为 1–3，均按用例表自拟，可表驱动） | **第 1 个建** |
| 2 | `generic.go` | 三件套实现 | `Map`、`Filter`、`Reduce` | 测试编译报错时 |
| 3 | `stdlib_test.go` | 行为 4 的标准库对照实验（只写测试，没有实现文件） | 6 个对照实验测试（名字自拟，如 `TestSlicesEqual`） | 行为 4 |
| 4 | `iterator_test.go` | All 的全部测试 | 完整遍历 / 空切片 / string / 提前停止 / 对接标准库用例（覆盖行为 5，均按用例表自拟） | 行为 5 |
| 5 | `iterator.go` | All 函数 | `All` | 行为 5 测试编译报错时 |

要写的函数一共 4 个，就是下面契约里的全部，一个不多一个不少；
本练习不定义任何类型、常量或哨兵错误。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**你要写的每一个签名都在下面**，按文件分组。
你唯一需要自己实现的是函数体；如果写代码时发现要发明契约之外的函数或类型，说明走偏了。

**写在 `generic.go`：**（无需 import）

```go
package genericlab

// Map 对 s 的每个元素应用 f，按原顺序返回结果组成的新切片；
// 输入为空（含 nil）时返回非 nil 的空切片
func Map[T, R any](s []T, f func(T) R) []R

// Filter 返回 s 中所有让 pred 返回 true 的元素，保持原相对顺序；
// 没有元素满足（含输入为空）时返回非 nil 的空切片
func Filter[T any](s []T, pred func(T) bool) []T

// Reduce 以 init 为初始累积值，从左到右用 f(累积值, 元素) 依次折叠整个切片；
// s 为空时原样返回 init。R 允许与 T 不同类型（如 []int 折叠成 string）
func Reduce[T, R any](s []T, init R, f func(R, T) R) R
```

**写在 `iterator.go`：**（契约按原始函数类型写，无需 import；等价的 `iter.Seq[T]` 写法在行为 5 讲解）

```go
// All 返回 s 的迭代器：按下标顺序产出每个元素；
// 消费方在 yield 中返回 false 时立即停止遍历，且之后不得再调用 yield
func All[T any](s []T) func(yield func(T) bool)
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 0 个类型：本练习不定义任何结构体 / 接口
- 4 个函数：`Map`、`Filter`、`Reduce`、`All`
- 0 个哨兵错误 / 常量

---

## 二、任务单

### 行为 1：Map —— 类型参数与类型推断

先写 `TestMap` 的测试，再写最少实现让它变绿——先写测试，编译失败（`undefined: Map`）即 RED。
用例覆盖下面两条（写成两个 Test 函数或表驱动，二选一；切片判等先用 `reflect.DeepEqual`，
行为 4 再统一换成 `slices.Equal`）：

| 用例名 | 输入 | 期望 |
|---|---|---|
| 平方映射 | `Map([]int{1, 2, 3}, func(x int) int { return x * x })` | `[]int{1, 4, 9}` |
| int 映射成 string | `Map([]int{1, 2}, func(x int) string { return strconv.Itoa(x * 10) })` | `[]string{"10", "20"}` |

### 行为 2：Filter —— 闭包做谓词

按用例表自己写 `TestFilter`（结构仿照行为 1），然后实现 `Filter` 到变绿：

| 用例名 | 输入 | 期望 |
|---|---|---|
| 筛出偶数 | `Filter([]int{1, 2, 3, 4, 5, 6}, func(x int) bool { return x%2 == 0 })` | `[]int{2, 4, 6}` |
| 全部不满足 | `Filter([]int{1, 3, 5}, func(x int) bool { return x%2 == 0 })` | `[]int{}`（非 nil） |
| 空切片输入 | `Filter([]int{}, func(x int) bool { return x%2 == 0 })` | `[]int{}`（非 nil） |
| 换类型：按长度筛字符串 | `Filter([]string{"go", "java", "c"}, func(s string) bool { return len(s) > 1 })` | `[]string{"go", "java"}` |

### 行为 3：Reduce —— R 与 T 可以不同

按用例表自己写 `TestReduce`，然后实现 `Reduce` 到变绿：

| 用例名 | 输入 | 期望 |
|---|---|---|
| 求和 | `Reduce([]int{1, 2, 3, 4}, 0, func(acc, x int) int { return acc + x })` | `10` |
| 空切片返回 init | `Reduce([]int{}, 5, func(acc, x int) int { return acc + x })` | `5` |
| 求积（init 是单位元 1） | `Reduce([]int{2, 3, 4}, 1, func(acc, x int) int { return acc * x })` | `24` |
| 🔺R≠T：int 切片拼成 string | `Reduce([]int{1, 2, 3}, "", func(acc string, x int) string { return acc + strconv.Itoa(x) })` | `"123"` |

### 行为 4：🔺标准库替代 —— slices / maps（Go 1.21+）

这一步**没有要实现的新函数**：新建 `stdlib_test.go`，把下面的对照实验写成测试跑绿，
并把行为 1–3 测试里的 `reflect.DeepEqual` 全部换成 `slices.Equal`：

| 用例名 | 输入 | 期望 |
|---|---|---|
| slices.Equal 断言相等 | `slices.Equal([]int{1, 4, 9}, []int{1, 4, 9})` | `true` |
| slices.Equal 对 nil 宽容 | `slices.Equal(nil, []int{})` | `true`（DeepEqual 在这里是 false） |
| slices.Sort 原地排序 | `a := []int{3, 1, 2}; slices.Sort(a)` | `a` 本身变成 `[]int{1, 2, 3}` |
| slices.Concat 拼接 | `slices.Concat([]int{1, 2}, []int{3}, []int{4, 5})` | `[]int{1, 2, 3, 4, 5}`，且三个入参都不被修改 |
| maps.Keys 收集并排序 | `m := map[string]int{"b": 2, "a": 1}`；`slices.Sorted(maps.Keys(m))` | `[]string{"a", "b"}` |
| maps.Values 收集并排序 | `slices.Sorted(maps.Values(m))` | `[]int{1, 2}` |

### 行为 5：🔺range-over-func 迭代器（Go 1.23）

`All` 的契约见第一节「接口契约」（同样固定，名字不要改；契约按原始函数类型写，
等价的 `iter.Seq[T]` 写法见第三节行为 5 的知识点）。

新建 `iterator_test.go`，按用例表自己写测试——先写测试，编译失败即 RED，再写最少实现变绿
（"提前 break"是验证协议的关键用例：遍历到第 2 个元素就 break，统计循环体执行次数）：

| 用例名 | 输入 | 期望 |
|---|---|---|
| 完整遍历 | for range 收集 `All([]int{1, 2, 3})` | `[]int{1, 2, 3}` |
| 空切片 | for range 收集 `All([]int{})` | `[]int{}`，yield 一次都没被调用 |
| 换类型：string | for range 收集 `All([]string{"a", "b"})` | `[]string{"a", "b"}` |
| 提前 break | 遍历到第 2 个元素就 break，统计循环体执行次数 | 恰好 2 次（迭代器没有多产出） |
| 对接标准库 | `slices.Collect(All([]int{1, 2, 3}))` | `[]int{1, 2, 3}` |

---

## 三、知识点总结

### 行为 1：Map —— 类型参数与类型推断

1. **类型参数**：`Map[T, R any]` 方括号里声明的 T、R 就是类型参数——它们不是具体类型，
   是"占位符"，调用时才被换成真类型。函数体里可以把 T/R 当普通类型用（声明变量、append、传参）。
2. **类型推断**：调用 `Map([]int{1,2,3}, func(x int) int {...})` 时，编译器按函数实参从左到右推断——
   由 `s` 得到 `T = int`，再由 `f` 的签名得到 `R = int`。所以大多数调用不用显式写类型；
   也可以显式写成 `Map[int, int](...)`，推断不出来（如实参类型有歧义）时才需要。
3. **实例化发生在编译期**：把类型参数换成具体类型的动作叫实例化，编译器为不同的类型组合生成对应版本的代码。
   **泛型不是运行时反射**——类型错误（比如 f 的签名对不上）在编译期就被拦下，这正是它优于
   `interface{}` + 类型断言旧写法的地方。
4. **`any` 就是 `interface{}`**：Go 1.18 起 `any` 是 `interface{}` 的别名。作为约束它表示"零承诺"——
   函数体对 T 值只能赋值、传参、装进容器，不能比较、不能运算。Map 只需要"搬运"值，所以 any 足够。
5. **易错点：空结果的 nil 问题**。实现时如果写 `var result []R`，空输入会返回 nil 切片，而
   `reflect.DeepEqual(nil切片, 空切片)` 是 **false**！契约统一要求用 `make([]R, 0, len(s))`
   返回非 nil 空切片——既避免调用方踩坑，也让断言写法一致（行为 4 会看到 `slices.Equal` 对此更宽容）。
6. **切片不能 `==`**：切片只能和 nil 比较，两个切片判等必须逐元素——`reflect.DeepEqual`、
   手写循环，或 Go 1.21 的 `slices.Equal`（行为 4 统一替换）。

### 行为 2：Filter —— 闭包做谓词

1. **类型参数按需声明**：Filter 只有一个 T——泛型不是越多越好，每个类型参数都得在签名里真正用到才有意义。
2. **闭包捕获**：`pred` 是闭包，可以捕获测试里的局部变量，如
   `threshold := 3; Filter(s, func(x int) bool { return x > threshold })`——
   不用为每个阈值单独定义函数，这是函数式风格在 Go 里的常态用法。
3. **保持契约一致**：与 Map 一样，空结果返回非 nil 空切片、保持原相对顺序——
   工具函数的可预期性比省一次分配重要。
4. **易错点：不要复用入参的底层数组**。原地过滤（`result := s[:0]` 往里写）会覆盖原切片后续内容
   （共享底层数组互污，正是 [tdd/slicelab](../slicelab/README.md) 要练的陷阱）——工具函数一律新分配。

### 行为 3：Reduce —— R 与 T 可以不同

1. **R 由 init 推断**：`Reduce[T, R any](s []T, init R, f func(R, T) R)` 里 R 的推断来源是
   `init` 实参而不是 `s`——这正是 R 与 T 能不同类型的原因：`[]int` 可以折叠成 `string`。
   注意 f 的签名是 `func(R, T) R`：累积值在左、元素在右、返回新的累积值。
2. **为什么必须传 init，而不是库内默认零值**：折叠空切片要返回 init——如果库里偷偷用零值，
   求积会得到 0 而不是 1。"单位元"（加法 0、乘法 1、拼接 ""）只有调用方知道，显式传参让签名更诚实。
3. **any vs comparable vs cmp.Ordered**（回顾 [basic/generic.go](../../basic/generic.go)）——
   经验法则：**函数体里需要用什么操作，约束就写什么**；能用 any 就别收紧，约束越宽能复用的类型越多：

   | 约束 | 承诺的操作 | 典型用途 | 出处 |
   |---|---|---|---|
   | `any` | 只能赋值/传参 | 本练习三件套——只搬运值 | — |
   | `comparable` | `==` / `!=` | 查找元素、判重、map key | basic/generic.go 的 `indexOf` |
   | `cmp.Ordered` | `<` `>` 等大小比较 | max / min / 排序 | basic/generic.go 的 `maxValue` |

   顺带复习 `~`：不带 `~` 的约束只认精确类型（`MyintGen` 不满足 `StrictNumber`），带 `~` 认底层类型，
   对比实验见 basic/generic.go#L37。
4. **Go 目前不允许方法有自己的类型参数（了解即可）**：`func (s Stack[T]) Map[R any](...)` 编译报错
   `method must have no type parameters`——方法的类型参数只能来自接收者的泛型类型本身。
   想给泛型类型加"映射方法"，绕法是写成包级函数：`func MapStack[T, R any](s Stack[T], f func(T) R) Stack[R]`。
   不用深究原因，记住绕法即可。

### 行为 4：🔺标准库替代 —— slices / maps（Go 1.21+）

1. **标准库已有就别重复造轮子**：Go 1.21 新增 `slices` / `maps` / `cmp` 三个泛型包。注意官方刻意
   **没有**提供泛型 Map/Filter/Reduce——Go 团队评估后认为普通 for 循环已经足够直白，
   泛型链式调用在 Go 的语法里反而别扭。所以本练习手写三件套是**练手理解原理**；生产代码的优先级是：
   能用 slices/maps 现成函数就用 → 否则写 for 循环 → 最后才考虑自己抽象。
2. **slices.Equal 取代 reflect.DeepEqual**：逐元素 `==` 比较，编译期做类型检查（DeepEqual 是运行时反射，
   类型错了编译不拦）；且 **nil 切片与空切片视为相等**——行为 1 里 DeepEqual 的 nil 坑在这里不存在。
3. **易错点：slices.Sort 是原地排序**——直接修改入参的底层数组，不返回新切片。想保留原切片，
   先 `slices.Clone` 一份再排；或者用 Go 1.23 的 `slices.Sorted(迭代器)`，它返回新切片。
4. **slices.Concat**：把任意多个切片拼成一个新切片，内部一次性算好总长度只分配一次，
   比自己循环 append 高效，且入参保证不被修改。
5. **maps.Keys/Values 在 Go 1.23 改了签名**：返回 `iter.Seq[K]` 迭代器而**不是切片**——
   不为临时需求分配内存。要断言就配合 `slices.Sorted(maps.Keys(m))`（收集+排序一步完成）；
   **map 遍历顺序随机，断言前必须排序**。这个 iter.Seq 到底是什么——正是行为 5 的主角。

### 行为 5：🔺range-over-func 迭代器（Go 1.23）

1. **迭代器协议**：`func(yield func(T) bool)` 就是 Go 1.23 定义的标准迭代器形状。
   `for v := range seq` 会被编译器翻译成 `seq(func(v T) bool { 循环体; return true })`——
   **你的循环体被包装成 yield 回调**，由迭代器在自己的循环里反复调用，把值"推"给你。
2. **yield 返回 false = 停止遍历**：`break` 翻译成 `return false`，`continue` 翻译成 `return true`。
   所以实现必须写成 `if !yield(v) { return }`——收到 false 立刻返回；协议规定此后不得再调用 yield
   （再调会直接 panic）。"提前 break"用例验证的正是这一点。
3. **更地道的写法 `iter.Seq[T]`**：标准库 `iter` 包定义了一个命名函数类型
   `type Seq[V any] = func(yield func(V) bool)`，把返回类型写成 `iter.Seq[T]` 与契约完全等价；
   双值版 `iter.Seq2[K, V]` 产出键值对（`maps.All` 就是 Seq2）。契约按原始函数类型写，
   是为了先看清协议本身长什么样。
4. **推模型 vs 拉模型**：range-over-func 是"推"（生产者在函数内驱动节奏，纯函数调用、零并发开销）；
   channel 遍历是"拉"（消费者主动取值，需要 goroutine 配合）。标准库选择推模型做迭代器协议，
   就是因为它轻量且不引入并发。
5. **协议统一的红利**：行为 4 的 `maps.Keys` 返回 iter.Seq，`slices.Collect` / `slices.Sorted` 消费
   iter.Seq——你手写的 `All` 同样可以直接喂给它们（`slices.Collect(All(s))`）。一套协议，全库互通。

### 约束速查

| 约束 | 承诺的操作 | 典型场景 | 回顾 |
|---|---|---|---|
| `any`（= `interface{}`） | 只能赋值、传参 | 只搬运值的工具函数（Map/Filter/Reduce/All） | 本练习 |
| `comparable` | `==` / `!=` | 查找、判重、map key | [basic/generic.go](../../basic/generic.go) `indexOf` |
| `cmp.Ordered` | `<` `>` `<=` `>=` | max/min、排序 | basic/generic.go `maxValue` |
| 自定义 `\|` + `~` 联合 | 由联合中类型的共性决定 | 数值泛型（`Number`） | basic/generic.go#L37 |

经验法则：函数体需要什么操作就写什么约束；能用 any 就别收紧。

### 类型推断与实例化速查

- 类型实参按函数实参从左到右推断；`Reduce` 的 R 来自 `init`，不是 `s`
- 推断失败（如实参类型有歧义）才显式写 `Map[int, string](...)`
- 实例化在编译期完成：无反射开销，类型错误编译期拦截
- 类型参数必须在签名里用到才有意义；方法不能有自己的类型参数（绕法：包级泛型函数）

### 标准库速查（Go 1.21 / 1.23）

| 函数 | 版本 | 作用 | 易错点 |
|---|---|---|---|
| `slices.Equal` | 1.21 | 切片相等断言 | nil 与空切片视为相等（DeepEqual 则否） |
| `slices.Sort` | 1.21 | 排序 | **原地修改入参**；要保留原切片先 `slices.Clone` |
| `slices.Concat` | 1.21 | 拼接多切片为新切片 | 入参不被修改 |
| `maps.Keys` / `maps.Values` | 1.21 加入、**1.23 改签名** | 取出全部键/值 | 返回 `iter.Seq` 迭代器，不是切片 |
| `slices.Collect` | 1.23 | 迭代器 → 切片 | map 来源顺序随机，断言前用 `slices.Sorted` |
| `slices.Sorted` | 1.23 | 迭代器 → 排序后的新切片 | 不碰原数据 |
| `cmp.Ordered` | 1.21 | 可比较大小约束 | 内部是带 `~` 的类型联合 |

### 迭代器协议一句话

`func(yield func(T) bool)`；实现里 `if !yield(v) { return }`；
for range 的 `break` / `continue` 就是 yield 的 `false` / `true`；
`iter.Seq[T]` 是它的标准别名。

### 与书目的对应

- learn-go-with-tests「Generics」章：泛型类型（Stack）与泛型函数、类型推断 → 本练习行为 1–2 的语法基础
- learn-go-with-tests「Revisiting arrays and slices with generics」章：手写 Reduce 折叠交易记录算余额 → 行为 3 的直接对应
- [Go 1.21 Release Notes](https://go.dev/doc/go1.21)：新增 `slices` / `maps` / `cmp` 包 → 行为 4
- [Go 1.23 Release Notes](https://go.dev/doc/go1.23)：range-over-func、`iter` 包、`maps` 函数改返回迭代器 → 行为 5

---

## 四、验收标准

```bash
go test ./tdd/genericlab -v          # 全绿
go vet ./tdd/genericlab              # 无警告
go test ./tdd/genericlab -cover      # 三件套与 All 的每个分支（含空输入、提前停止）都有用例
```

## 五、完成后自查（能口头回答才算过）

1. `Map([]int{1,2,3}, func(x int) int {...})` 的 T 和 R 分别从哪个实参推断出来？什么情况必须显式写 `Map[int, int](...)`？
2. any / comparable / cmp.Ordered 三个约束各自承诺什么操作？为什么 Map/Filter/Reduce 都选 any？
3. 切片为什么不能直接用 `==` 断言？`slices.Equal` 和 `reflect.DeepEqual` 对 nil 与空切片的判定有何不同？
4. `Reduce` 的 R 由哪个实参推断？为什么要求显式传 init，而不是在库里默认零值？
5. `slices.Sort` 和 `slices.Sorted` 对原数据的影响有何不同？想保留原切片怎么办？
6. Go 1.23 起 `maps.Keys` 返回什么类型？怎么把它变成可断言的有序切片？
7. `for v := range seq` 里 `break` 在迭代器协议里对应什么？迭代器实现为什么必须写 `if !yield(v) { return }`？

全部答清后，回到 [根 README](../../README.md#四目录划分评估与-tdd-驱动学习计划)，
把 TDD 练习总目录第 9 行 `tdd/genericlab` 的状态从「待建」改成「✅ 已建」。
