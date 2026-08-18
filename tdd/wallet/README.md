# Wallet TDD —— 指针接收者 + 自定义类型 + 哨兵错误

目标：这是 learn-go-with-tests "Pointers & errors" 章的本土化练习。场景故意简单
（比特币钱包：存、取、查），**注意力全部放在三个机制上**：方法为什么必须用指针接收者、
`type Bitcoin int` 自定义类型配 `String()` 的收益、哨兵错误的声明与识别。
前置练习：[tdd/errhandling](../errhandling/README.md)（行为 3 的 `errors.Is` 断言直接复用那里的成果）。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/wallet` 就是它的运行方式，验收者是测试，不是人。

这个包对外提供三个能力：

- **存款**：把一笔 BTC 存入钱包
- **查询余额**：返回当前余额
- **取款**：余额充足（含刚好取光）则扣款返回；余额不足则**一分不动**，返回可识别的"余额不足"错误

### 文件计划（共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `wallet_test.go` | 全部测试（含断言辅助函数） | **第 1 个建** |
| `wallet.go` | `Bitcoin`、`Wallet`、`ErrInsufficientFunds` 及全部方法 | 测试编译报错时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package wallet

// Bitcoin 自定义类型，底层是 int，单位"枚"。
// 注意不是别名：Bitcoin(10) 与 int(10) 在编译期就是两种东西，
// 编译器会挡住"把普通数字当钱存"这类错误。
type Bitcoin int

// String 让 Bitcoin 实现 fmt.Stringer 接口：
// 之后 %v/%s 打印（包括 t.Errorf 的失败输出）显示 "10 BTC" 而不是 "10"。
// 必须是值接收者；实现内部用 %d 格式化（原因见行为 2 知识点）。
func (b Bitcoin) String() string

// Wallet 比特币钱包。字段不导出——余额只能经 Deposit/Withdraw 改变，
// 调用方无法直接改字段绕开规则。
type Wallet struct {
	// 字段自行设计，不导出（小写开头）
}

// Deposit 存入 amount。必须指针接收者：改的是原钱包，不是副本。
func (w *Wallet) Deposit(amount Bitcoin)

// Balance 返回当前余额。契约统一用指针接收者（原因见行为 1 知识点）。
func (w *Wallet) Balance() Bitcoin

// Withdraw 取出 amount：
//   - 余额充足（含刚好取光）→ 扣款，返回 nil
//   - 余额不足 → 余额一分不动，返回 ErrInsufficientFunds
func (w *Wallet) Withdraw(amount Bitcoin) error

// 哨兵错误：包级变量，全进程唯一实例；调用方用 errors.Is 识别。
var ErrInsufficientFunds = errors.New("cannot withdraw, insufficient funds")
```

### 第一步：手把手起步（行为 1 的测试直接给你当模板）

1. 在 `tdd/wallet/` 下新建 `wallet_test.go`，写入：

```go
package wallet

import "testing"

func TestWallet(t *testing.T) {
	wallet := Wallet{}

	if got := wallet.Balance(); got != Bitcoin(0) {
		t.Errorf("新钱包余额期望 %v，得到 %v", Bitcoin(0), got)
	}

	wallet.Deposit(Bitcoin(10))

	if got := wallet.Balance(); got != Bitcoin(10) {
		t.Errorf("存入后余额期望 %v，得到 %v", Bitcoin(10), got)
	}
}
```

2. 运行 `go test ./tdd/wallet` → **编译失败**：`undefined: Wallet`。
   这就是 RED —— 测试描述了你想要但还不存在的代码。
3. 新建 `wallet.go`，写**最少**的代码让测试通过
   （本步只需 `Bitcoin`、`Wallet`、`Deposit`、`Balance`；`String()` 留到行为 2 再写，这是故意的）。
4. 再跑 `go test ./tdd/wallet` → 绿。第一轮 RED → GREEN 完成。

---

## 二、任务单（边做边学）

每个行为 = 一轮完整的 RED → GREEN → REFACTOR，**先把测试写出来再实现**。
行为 1 的测试已在第一节给出；从行为 2 起，测试由你自己写。

### 行为 1：新钱包余额为 0，存款后余额增加

| 用例 | 操作序列 | 期望 |
|---|---|---|
| 新钱包 | `Wallet{}` | `Balance() == Bitcoin(0)` |
| 存一笔 | `Deposit(Bitcoin(10))` | `Balance() == Bitcoin(10)` |

两条断言都已包含在第一节的模板里。做完下面的机制实验再往下走。

**这一步用到的知识点：**

1. **指针接收者的原理**：方法只是"第一个参数是接收者的普通函数"。`func (w Wallet) Deposit(...)` 里的 `w` 是整个结构体的**副本**——调用时拷贝一份，改的是副本的字段，函数返回后副本销毁，原钱包分毫未动，钱"存丢了"。`func (w *Wallet) Deposit(...)` 里的 `w` 是原钱包的地址，顺着指针改到的才是真钱包。对照你在 [basic/struct_func.go#L15](../../basic/struct_func.go#L15) 写过的注释"要修改结构体内部的值必须使用指针"——本练习就是把这句话变成肌肉记忆。
2. **调用时的自动取地址**：`wallet.Deposit(Bitcoin(10))` 不需要写成 `(&wallet).Deposit(...)`——只要 `wallet` 是可寻址的变量，编译器自动取地址（[basic/struct_func.go#L36](../../basic/struct_func.go#L36) 演示过）。反过来，`Wallet{}.Deposit(Bitcoin(10))` 直接编译失败：临时值没有地址可取。这也解释了为什么测试里要先 `wallet := Wallet{}` 再调用，而不是一条链写完。
3. **值接收者还是指针接收者（uber·receiver）**：要修改字段 → 必须指针；只读但结构体较大 → 指针（避免整份拷贝）；只读且小（基本类型包装、小 struct）→ 值也可以；含 `sync.Mutex` 等不可拷贝字段 → 必须指针（拷贝锁是未定义行为）。最后一条团队规则：**同一个类型的方法，接收者风格要统一**——所以只读的 `Balance()` 在契约里也是指针接收者，与 `Deposit`/`Withdraw` 保持一致，不让调用方猜。
4. **自定义类型 `type Bitcoin int`**：不是类型别名（别名写作 `type Bitcoin = int`，那只是同一个类型的另一个名字）。`Bitcoin` 是以 `int` 为底层类型的**全新类型**：`int` 值不能隐式当 `Bitcoin` 用，必须显式写 `Bitcoin(10)`。收益是编译器替你守边界——"10 岁""10 件""10 BTC"在代码里永远不会混；代价是金额字面量都要包一层，测试里能直接看到这个写法。
5. **RED 的复习**：`undefined: Wallet` 是编译失败形态的 RED（testbasic 行为 1 已学）。本练习每个新行为都会再经历一次——行为 2 是 `undefined: Withdraw`，习惯它。

### 行为 2：取款成功扣减余额；Stringer 让失败输出说人话

| 用例 | 操作序列 | 期望 |
|---|---|---|
| 先存后取 | `Deposit(Bitcoin(10))` → `Withdraw(Bitcoin(5))` | `err == nil`，且 `Balance() == Bitcoin(5)` |

步骤：自己写测试（参考第一节模板）→ 跑 `go test ./tdd/wallet` 看到 `undefined: Withdraw`（RED）→
写最小实现到绿 → **REFACTOR**：把两个测试里重复的余额断言提取成辅助函数，骨架：

```go
func assertBalance(t *testing.T, wallet Wallet, want Bitcoin) {
	t.Helper()
	if got := wallet.Balance(); got != want {
		t.Errorf("期望 %v，得到 %v", want, got)
	}
}
```

**这一步用到的知识点：**

1. **`fmt.Stringer` 接口**：定义只有一行——`String() string`。`fmt` 包遇到 `%v`/`%s`/`%q`/`%x`/`%X` 这些字符串类动词时，先检查值是否实现了 `String()`，实现了就调用它再打印结果。**亲手做这个实验**：行为 1 的实现还没有 `String()`——把断言里的 `want` 故意改错跑一次，失败输出是纯数字 `期望 20，得到 10`；按契约补上 `String()`（`fmt.Sprintf("%d BTC", b)`）再跑，输出变成 `期望 20 BTC，得到 10 BTC`。失败信息的说服力就此不同——实验完把 `want` 改回正确值。
2. **`String()` 里的无限递归陷阱**：实现里若写 `fmt.Sprintf("%s BTC", b)`，`%s` 又会回头调 `b.String()`——无限递归直到栈溢出。`%d` 之所以安全：`fmt` 只在字符串类动词上查 Stringer，`%d` 走整数格式化路径，不再回调。`go vet` 能静态抓到这个递归（printf 检查）——验收标准里的 `go vet` 不是摆设。
3. **`String` 必须是值接收者**：契约写的是 `func (b Bitcoin) String()`。若写成指针接收者，只有 `*Bitcoin` 满足 Stringer；而 `t.Errorf` 传参时 `Bitcoin` 值被拷贝打包进 `any`，接口里的值**不可寻址**，`fmt` 取不到它的指针方法——输出悄悄退化回纯数字。这就是"接收者决定谁满足接口"（[basic/interface.go](../../basic/interface.go)，errhandling 行为 3 踩过同一条规则）。
4. **`t.Helper()`**：辅助函数开头调它之后，断言失败报告的行号指向**调用方**而不是辅助函数内部——多个用例挂了你才知道是谁挂了。不加它，所有失败都指向 `assertBalance` 里同一行，排查时要多绕一层。

### 行为 3：余额不足时取款失败，且余额一分不动

| 用例 | 操作序列 | 期望 |
|---|---|---|
| 余额不足 | `Deposit(Bitcoin(10))` → `Withdraw(Bitcoin(100))` | ① `errors.Is(err, ErrInsufficientFunds)` 为 true<br>② `Balance() == Bitcoin(10)`（一分没动） |
| 刚好取光 | `Deposit(Bitcoin(10))` → `Withdraw(Bitcoin(10))` | `err == nil`，且 `Balance() == Bitcoin(0)` |

"刚好取光"是边界用例：它逼你把实现里的比较写成 `amount > balance` 才拒绝，而不是 `>=`——少它，边界写反也能全绿。

断言辅助函数骨架（错误断言也值得复用）：

```go
func assertError(t *testing.T, err error, want error) {
	t.Helper()
	if err == nil {
		t.Fatal("期望返回错误，却没有返回")
	}
	if !errors.Is(err, want) {
		t.Errorf("期望错误 %v，得到 %v", want, err)
	}
}
```

**这一步用到的知识点：**

1. **哨兵错误**：`var ErrInsufficientFunds = errors.New(...)` 声明为包级变量，全进程唯一实例。"余额不足"是**需要调用方识别**的错误类别（调用方可能据此提示充值），所以必须可用代码判定，而不是只在文案里提一句。复习 [tdd/errhandling](../errhandling/README.md) 的铁律：禁止 `err.Error() == "cannot withdraw..."` 字符串比较——文案一改判断就崩。
2. **`Err` 前缀命名（uber·错误命名）**：错误**变量**以 `Err` 开头（`ErrInsufficientFunds`），错误**类型**以 `Error` 结尾（如 errhandling 的 `ValidationError`）。一眼看到名字就知道"这是个错误值、该用 `errors.Is/As` 判"。
3. **`errors.Is` 而不是 `==`**：本例 `Withdraw` 返回的就是哨兵本身，`==` 也能通过；但一律写 `errors.Is(err, ErrInsufficientFunds)`——将来错误被 `%w` 包装后 `==` 立即失效，`Is` 沿链仍然命中（errhandling 行为 2 的机制实验）。习惯在简单场景养成，复杂场景才不翻车。
4. **"余额不变"是业务不变量，必须进测试**：`Withdraw` 失败时余额一分不动，是这条行为的另一半规格。只断言 `err` 的话，"先扣钱、再发现余额不足、然后报错"的错误实现也能通过。测试的职责就是把这类不变量钉死。
5. **`assertError` 里为什么用 `t.Fatal`**：期望的错误压根没返回时，后面的 `errors.Is` 检查已经没有意义，立即终止、报错信息更准确（复习 testbasic 机制实验：`t.Fatal` 内部是 `runtime.Goexit`，终止当前测试函数，defer 仍执行）。

### 机制实验（行为 1 完成后必做）

1. **值接收者丢钱实验**：把 `Deposit` 的接收者从 `*Wallet` 改成 `Wallet`，重跑测试——存款断言失败，余额仍是 `0`：改的是副本，钱存丢了。改回指针接收者，确认重新变绿。
2. **String 递归实验**（行为 2 完成后做）：把 `String()` 里的 `%d` 改成 `%s`，先跑 `go vet ./tdd/wallet`——vet 直接报递归调用；再 `go test` 看栈溢出 panic。改回 `%d`，确认 vet 和测试都恢复干净。

---

## 三、知识点总结

### 方法接收者决策表（uber·receiver）

| 场景 | 接收者 | 原因 |
|---|---|---|
| 要修改字段 | 指针 | 值接收者改的是副本，修改丢失 |
| 只读、结构体较大 | 指针 | 避免每次调用整份拷贝 |
| 只读、小结构体 / 基本类型包装 | 值也可以 | 拷贝便宜，语义更安心 |
| 含 `sync.Mutex` 等不可拷贝字段 | 必须指针 | 拷贝锁是未定义行为 |
| 同一个类型 | 风格统一 | `Balance` 只读也用指针，与 `Deposit`/`Withdraw` 一致 |

另记一条方法集规则：值的方法集只含**值接收者**方法；指针的方法集含值 + 指针两种。
接口满足性因此不同——`String` 用值接收者实现，则 `Bitcoin` 和 `*Bitcoin` 都满足 Stringer。

### `fmt.Stringer` 速查

- 接口定义：`String() string`；`%v %s %q %x %X` 打印时自动调用
- `%d` 不查 Stringer → `String()` 内部格式化自己必须用 `%d`；用 `%s` 无限递归（`go vet` 可静态发现）
- 指针接收者实现 → 只有指针满足接口；打包进 `any` 的值调不到，输出悄悄退化

### 哨兵错误速查

- 声明：包级变量 + `errors.New`，全进程唯一实例
- 命名：变量 `Err` 前缀、类型 `Error` 后缀（uber·错误命名）
- 判断：一律 `errors.Is`；禁止字符串比较；被 `%w` 包装后 `Is` 仍命中、`==` 失效
- 返回错误的函数，业务不变量（如"失败则余额不变"）同样要写进测试

### 与书目的对应

- Effective Go·10 方法（指针 vs 值接收者）——行为 1 全部落地
- uber 规范·receiver（接收者选择）与接口相关条目——行为 1、2；uber·错误命名（`Err` 前缀）——行为 3
- [learn-go-with-tests·pointers-and-errors](https://github.com/quii/learn-go-with-tests/tree/main/pointers-and-errors)：本练习的原始出处
- 仓库内复用：[basic/struct_func.go](../../basic/struct_func.go)（指针接收者）、[basic/pointer.go](../../basic/pointer.go)、[basic/interface.go](../../basic/interface.go)（方法集与接口满足）、[tdd/errhandling](../errhandling/README.md)（`errors.Is`）

---

## 四、验收标准

```bash
go test ./tdd/wallet -v        # 全绿
go vet ./tdd/wallet            # 无警告（顺带能抓住 String 递归这类错误）
go test ./tdd/wallet -cover    # 每条路径都有用例，必须 100%
```

## 五、完成后自查（能口头回答才算过）

1. 为什么 `Deposit`/`Withdraw` 必须用指针接收者？改成值接收者会发生什么、为什么？
2. `wallet.Deposit(Bitcoin(10))` 调用指针方法时编译器帮你做了什么？为什么 `Wallet{}.Deposit(...)` 编译失败？
3. `type Bitcoin int` 和 `type Bitcoin = int` 有什么区别？自定义类型帮你挡住了哪类错误？
4. 实现 `String()` 前后，`t.Errorf` 的失败输出有什么变化？`fmt` 包是怎么找到 `String` 方法的？
5. 在 `String()` 里用 `%s` 打印接收者会发生什么？为什么 `%d` 是安全的？
6. 哨兵错误为什么要声明为包级变量、以 `Err` 前缀命名？为什么断言用 `errors.Is` 而不是 `==`？
7. 行为 3 为什么必须同时断言"余额不变"？少这条断言会放过什么样的错误实现？

全部答清后，回到 [根 README](../../README.md) 的 TDD 练习总目录（阶段 1·值语义与复合类型），
把 #5 `tdd/wallet` 划掉。
