# ErrHandling TDD —— error 处理体系

目标：通过 4 个行为切片，用 TDD 补齐 Go error 体系的完整闭环。
这是全仓库 P0 缺口（教程 ch17 前置依赖），是后续 `wallet`、`dictionary` 等练习的地基。

> 本任务是**机制学习型**练习：接口契约已固定，你的注意力全部放在 error 机制本身。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/errhandling` 就是它的运行方式，验收者是测试，不是人。

这个包对外提供两个能力：

1. **按 id 查找用户**，三种结果调用方都能用代码区分：
   - 找到 → 返回用户
   - 未找到 → 返回可识别的"未找到"错误
   - 底层存储故障 → 返回带用户 id 上下文的故障错误
2. **校验用户信息**：
   - 单个校验 → 返回指明"哪个字段错了"的错误
   - 批量校验 → 一次返回全部违规，而不是遇到第一个就停

### 文件计划（共 5 个文件，分三次建）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `service_test.go` | 行为 1、2 的测试 | **第 1 个建** |
| `store.go` | `User`、`Store` 接口、`MapStore`、`NewMapStore`、哨兵错误 | 测试编译报错时 |
| `service.go` | `UserService`、`NewUserService`、`FindUser` | 同上 |
| `validate_test.go` | 行为 3、4 的测试 | 行为 3 开始时 |
| `validate.go` | `ValidationError`、`Validate`、`ValidateAll` | 测试编译报错时 |

要写的函数一共 7 个，就是下面契约里的全部，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

```go
package errhandling

type User struct {
    ID   int
    Name string
    Age  int
}

// 哨兵错误（包级变量）
var ErrUserNotFound = errors.New("user not found")
var ErrStorage      = errors.New("storage failure")
var ErrEmptyName    = errors.New("empty name")
var ErrInvalidAge   = errors.New("invalid age")

// 自定义错误类型（行为 3 用）
type ValidationError struct {
    Field string
    Msg   string
}
func (e *ValidationError) Error() string // 指针接收者，格式自定

// 存储层抽象：让"底层故障"在测试中可确定性地触发
type Store interface {
    Get(id int) (User, error)
}

// map 实现：正常存取；id 不存在返回 (User{}, ErrUserNotFound)
func NewMapStore(users map[int]User) MapStore

// 服务层
func NewUserService(s Store) UserService

// 语义：
// - 底层返回 ErrUserNotFound → 原样返回（不包装）
// - 底层返回其他错误 → fmt.Errorf("find user %d: %w", id, err) 包装后返回
// - 找到 → 返回 User 和 nil
func (s UserService) FindUser(id int) (User, error)

// 只校验 Age：Age < 0 → 返回 &ValidationError{Field: "age", ...}；否则 nil
func Validate(u User) error

// 校验 Name 和 Age，收集全部违规（不是遇到第一个就返回）：
// - Name == "" → ErrEmptyName；Age < 0 → ErrInvalidAge
// - 有违规 → errors.Join 聚合返回；全合法 → nil
func ValidateAll(u User) error
```

### 第一步：手把手起步（行为 1 的测试直接给你当模板）

1. 在 `tdd/errhandling/` 下新建 `service_test.go`，写入：

```go
package errhandling

import (
	"errors"
	"testing"
)

func TestFindUser_Found(t *testing.T) {
	store := NewMapStore(map[int]User{1: {ID: 1, Name: "Tom", Age: 20}})
	svc := NewUserService(store)

	got, err := svc.FindUser(1)

	if err != nil {
		t.Fatalf("期望没有错误，得到 %v", err)
	}
	if got.ID != 1 || got.Name != "Tom" {
		t.Errorf("期望找到 Tom，得到 %+v", got)
	}
}

func TestFindUser_NotFound(t *testing.T) {
	store := NewMapStore(map[int]User{})
	svc := NewUserService(store)

	_, err := svc.FindUser(99)

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("期望 ErrUserNotFound，得到 %v", err)
	}
}
```

2. 运行 `go test ./tdd/errhandling` → **编译失败**：`undefined: NewMapStore`。
   这就是 RED —— 测试描述了你想要但还不存在的代码。
3. 新建 `store.go` 和 `service.go`，照契约写出让测试通过的**最少代码**。
4. 再跑 `go test ./tdd/errhandling` → 全绿，行为 1 完成。

---

## 二、任务单（边做边学）

每个行为 = 一轮完整的 RED → GREEN → REFACTOR，**先把测试写出来再实现**。

### 行为 2：底层故障被包装后仍可识别，且上下文不丢

先在 `service_test.go` 里加测试替身：

```go
type stubStore struct{ err error }

func (s stubStore) Get(id int) (User, error) { return User{}, s.err }
```

| 用例 | stub 返回 | 调用 | 期望（三条都要断言） |
|---|---|---|---|
| 底层故障 | `ErrStorage` | `FindUser(99)` | ① `errors.Is(err, ErrStorage) == true`（穿透包装）<br>② `err.Error()` 包含 `"99"`（上下文不丢）<br>③ `errors.Is(err, ErrUserNotFound) == false`（能区分故障类型） |

**这一步用到的知识点：**

1. **`%w` 包装**：`fmt.Errorf("...: %w", err)` 把原始错误挂到一条**错误链**上，新错误既有新文案，又不丢原始身份。`errors.Is` 会沿着链逐层比较，所以包装后依然认得出 `ErrStorage`。
2. **`%w` vs `%v`**：`%v` 只把错误的字符串拼进去，原始错误就此"消失"——`errors.Is` 再也找不到它。这就是用例②③成立而换成 `%v` 后断言①失败的原因。
3. **错误链的名字由来**：`fmt.Errorf` 返回的错误内部存着原始错误的引用，`errors.Is/As` 通过 `Unwrap() error` 方法一层层往下拆。你可以给包装后的错误调 `errors.Unwrap` 亲手拆一层看看。
4. **易错点**：禁止用 `err.Error() == "storage failure"` 做字符串比较——字符串是给人看的，不是给代码判断的；文案一改判断就崩。

### 行为 3：调用方能取出错误里的具体字段

| 用例 | 调用 | 期望 |
|---|---|---|
| 年龄非法 | `Validate(User{Name:"Tom", Age:-1})` | `var ve *ValidationError; errors.As(err, &ve) == true`，且 `ve.Field == "age"` |
| 合法 | `Validate(User{Name:"Tom", Age:0})` | `err == nil` |
| 非校验错误 | 对 `ErrUserNotFound` 调用 `errors.As(err, &ve)` | 返回 `false` |

**这一步用到的知识点：**

1. **`errors.As` 与类型断言的关系**：`errors.As(err, &ve)` 等价于"沿错误链逐层做类型断言"——它的底层机制就是你在 [basic/interface_empty.go#L62](../../basic/interface_empty.go#L62) 学过的逗号-ok 断言，只是多了沿链遍历。
2. **第二个参数必须是指针的指针**：`&ve` 的类型是 `**ValidationError`——`As` 要通过它把找到的错误"写回"你的变量。传 `&ValidationError{}` 会在运行时 panic，这是本练习最值得亲手踩一次的坑。
3. **指针接收者 `Error()` 的含义**：只有 `*ValidationError` 实现了 error 接口（值类型没有），所以 `Validate` 必须返回 `&ValidationError{...}`，`As` 的目标也必须是 `*ValidationError`。这条规则与 [basic/interface.go#L17](../../basic/interface.go#L17) 的"接收者决定谁满足接口"完全一致。
4. **Go 1.26 新写法 `errors.AsType[T]`**：泛型版 `As`，`ve, ok := errors.AsType[*ValidationError](err)`，不用声明临时变量、也不会因传错参数 panic。学完 `As` 后用它改写一遍本行为的测试，体会取舍。

### 行为 4：一次返回多个违规，调用方能逐个识别

| 用例 | 调用 | 期望 |
|---|---|---|
| 两个字段都非法 | `ValidateAll(User{Name:"", Age:-1})` | `errors.Is(err, ErrEmptyName)` 和 `errors.Is(err, ErrInvalidAge)` **都**为 true |
| 只有一个非法 | `ValidateAll(User{Name:"", Age:1})` | 只有 `ErrEmptyName` 命中 |
| 全合法 | `ValidateAll(User{Name:"Tom", Age:20})` | `err == nil` |

**这一步用到的知识点：**

1. **`errors.Join`（Go 1.20+）**：把多个错误聚合成一棵树。`Is/As` 对这棵树做**深度遍历**——所以两个哨兵都能命中。`Join` 返回的错误实现了 `Unwrap() []error`（注意是切片，不是单个错误），这是它与 `fmt.Errorf %w` 链的结构差异。
2. **过滤 nil**：`errors.Join(nil, nil)` 返回 nil。实现 `ValidateAll` 时先收集非 nil 的违规，再决定返回什么——这正好是"全合法返回 nil"用例的自然写法。
3. **两种风格并存是故意的**：行为 3 用自定义类型（调用方要结构化字段），行为 4 用哨兵聚合（调用方只需识别类别）。uber 规范的决策表就是这个判断：需要匹配？→ 静态文案用哨兵变量、动态文案用自定义类型；不需要匹配？→ `errors.New`/`fmt.Errorf` 即可。

### 机制实验（做完行为 2 后必做）

把 `FindUser` 实现里的 `%w` 改成 `%v` 重跑测试，观察断言 ① 失败——亲身体会两者区别，然后改回来。

---

## 三、知识点总结

### 第一性原理：error 是值，不是异常

Go 没有 try/catch，`error` 只是一个**普通接口值**（`Error() string`）。
"是值"这一个事实，推导出全部知识点：

1. 值可以被**比较** → 哨兵错误 + `errors.Is`（行为 1）
2. 值在层层返回时需要**附加上下文**又不丢原始身份 → `%w` 包装链（行为 2）
3. 调用方有时需要值里面的**结构化信息** → 自定义错误类型 + `errors.As`（行为 3）
4. 多个值可以同时返回 → `errors.Join` 错误树（行为 4）
5. panic/recover 只用于"程序员犯了错、无法恢复"的场景，不是错误处理手段 → [basic/panic_recover.go](../../basic/panic_recover.go)

### error 体系速查

| 需求 | 工具 | 注意 |
|---|---|---|
| 造一个静态错误 | `errors.New` | 要匹配就声明为包级变量（哨兵） |
| 造带动态文案的错误 | `fmt.Errorf`（无 %w） | 调用方无法匹配 |
| 附加上下文且保留匹配能力 | `fmt.Errorf` + `%w` | 可多次 `%w`（Go 1.20+） |
| 沿链按值匹配 | `errors.Is` | 穿透 `%w` 链和 `Join` 树 |
| 沿链按类型提取 | `errors.As` / Go 1.26 `errors.AsType[T]` | `As` 要传指针的指针 |
| 聚合多个错误 | `errors.Join` | 过滤 nil；`Unwrap() []error` |
| 判断"是不是某种错" | 哨兵 + `Is` | 禁止字符串比较 |

### 与书目的对应

- uber 规范·Errors 四节（错误类型决策表 / 错误包装 / 错误命名 `Err` 前缀与 `Error` 后缀 / 一次处理错误）——本练习全部落地
- Effective Go·15 错误
- [basic/panic_recover.go](../../basic/panic_recover.go)：panic 与 error 的边界

---

## 四、验收标准

```bash
go test ./tdd/errhandling -v        # 全绿
go vet ./tdd/errhandling            # 无警告
go test ./tdd/errhandling -cover    # 核心逻辑 ≥80%
```

## 五、完成后自查（能口头回答才算过）

1. `%w` 和 `%v` 格式化错误时有什么区别？（用机制实验的结果回答）
2. 为什么哨兵错误要声明为包级变量而不是每次 `errors.New` 一个新的？
3. `errors.Is` 和 `==` 在包装错误场景下有什么不同？
4. 什么场景该用 panic 而不是返回 error？（对照 [basic/panic_recover.go](../../basic/panic_recover.go)）
5. 行为 3 为什么用自定义错误类型、行为 4 为什么用哨兵聚合？两种风格各自适合什么调用方？
6. `ValidationError.Error()` 用指针接收者意味着什么？`errors.As` 的目标类型因此应该写什么？
7. `errors.Join` 的错误树和 `%w` 的错误链，结构差异是什么（提示：`Unwrap` 的签名）？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，把"error 处理体系"从 P0 划掉。
