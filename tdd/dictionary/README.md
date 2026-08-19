# Dictionary TDD —— 在 map 类型上定义方法：字典 CRUD

目标：用 TDD 实现一个基于 map 的单词字典（查/增/改/删四个方法）。
表面上练 CRUD，**核心领悟点是接收者语义**：map 是引用类型，值接收者的方法也能改内容——
与 wallet 练习里 struct 必须指针接收者形成鲜明对照。顺带巩固 errhandling 学过的哨兵错误。
本练习是 learn-go-with-tests maps 章的落地。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求规格（接口契约固定，照此实现）；第二节是纯任务单——只给行为目标、用例表和验收命令，测试代码全部自己写；第三节是知识点讲解，做之前通读或卡壳时查阅，做完后对照自查。

---

## 一、需求规格

### 核心功能

实现一个基于 map 的**单词字典包**，对外提供四个能力：

- **查**：词存在 → 返回释义；不存在 → 返回调用方能用代码识别的"未找到"
- **增**：新词写入成功；词已存在 → 报错且**不覆盖**旧释义
- **改**：词存在 → 覆盖释义；不存在 → 报错，不能悄悄退化成"增"
- **删**：删除单词；删不存在的词不算错误

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/dictionary` 就是它的运行方式，验收者是测试，不是人。

为什么是这个需求：CRUD 只是载体——真正要练的是**接收者语义**
（map 是引用类型，值接收者的方法也能改内容），四个方法正好覆盖"读、条件写、写、删"四种与底层表的交互方式。

### 调用关系（谁在调用谁）

```text
测试代码 ──► Dictionary.Search(word)       查
测试代码 ──► Dictionary.Add(word, def)     增
测试代码 ──► Dictionary.Update(word, def)  改
测试代码 ──► Dictionary.Delete(word)       删
```

没有接口、没有依赖注入：四个方法直接挂在 map 类型上，测试自己造 `Dictionary{...}` 字面量直接调用——
与 errhandling 的"服务 → 接口 → 替身"三层结构对照，这是"在命名类型上直接定义行为"的最小形态。

### 文件计划（共 2 个文件，按编号顺序建）

最终目录长这样：

```text
tdd/dictionary/
├── dictionary_test.go   # 行为 1~4 的全部测试 + 断言辅助函数
└── dictionary.go        # Dictionary 类型 + 3 个哨兵错误 + 4 个方法（本练习的全部实现）
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `dictionary_test.go` | 行为 1~4 的全部测试，以及行为 4 重构抽出的断言辅助函数 | `TestSearch`、`TestAdd`、`TestUpdate`、`TestDelete`、`assertError`、`assertDefinition` | **第 1 个建**（辅助函数在行为 4 重构时加入） |
| 2 | `dictionary.go` | 字典类型、哨兵错误和四个方法，本练习的全部实现 | `Dictionary`、`ErrNotFound`、`ErrWordExists`、`ErrWordDoesNotExist`、`Dictionary.Search`、`Dictionary.Add`、`Dictionary.Update`、`Dictionary.Delete` | 测试编译报错时 |

四个 `TestXxx` 的测试代码全部自己写（命名沿用 `TestXxx`）；
`assertError`/`assertDefinition` 在行为 4 的重构环节从重复断言里抽出（写法见第三节「行为 4」知识点）。

要写的代码 = 1 个类型 + 3 个错误变量 + 4 个方法，就是下面契约里的全部，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**你要写的每一个类型、每一个签名都在下面**（本练习只有一个实现文件，契约就这一组）。
你唯一需要自己实现的是函数体；如果写代码时发现要发明契约之外的类型或函数，说明走偏了。

**写在 `dictionary.go`：**（需要 `import "errors"`）

```go
package dictionary

// Dictionary 单词字典：键是单词，值是释义
type Dictionary map[string]string

// 哨兵错误（包级变量）：调用方一律用 errors.Is 识别，禁止比较错误文案
var ErrNotFound = errors.New("could not find the word you were looking for")
var ErrWordExists = errors.New("cannot add word because it already exists")
var ErrWordDoesNotExist = errors.New("cannot update word because it does not exist")

// 四个方法全部是值接收者——这正是本练习的核心领悟点：
// map 是引用类型，值接收者拷贝的只是引用头，改的是同一张底层表

// Search 查词：存在 → (释义, nil)；不存在 → ("", ErrNotFound)
func (d Dictionary) Search(word string) (string, error)

// Add 加词：词已存在 → ErrWordExists，且不覆盖旧释义；否则写入，返回 nil
func (d Dictionary) Add(word, definition string) error

// Update 改词：词存在 → 覆盖释义，返回 nil；不存在 → ErrWordDoesNotExist
func (d Dictionary) Update(word, definition string) error

// Delete 删词：直接删除；删不存在的词是 no-op，不算错误（所以没有返回值）
func (d Dictionary) Delete(word string)
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 1 个类型：`Dictionary`
- 4 个方法：`Dictionary.Search`、`Dictionary.Add`、`Dictionary.Update`、`Dictionary.Delete`
- 3 个哨兵错误：`ErrNotFound`、`ErrWordExists`、`ErrWordDoesNotExist`

---

## 二、任务单

每个行为 = 一轮完整的 RED → GREEN → REFACTOR，**先把测试写出来再实现**：
测试代码全部自己写——先写测试，编译失败即 RED，再写最少实现变绿。

### 行为 1：Search 查词，存在返回释义，不存在返回 ErrNotFound

新建 `dictionary_test.go`，按下面用例表写 `TestSearch`（两个子测试共享一份只读字典即可）；
跑 `go test ./tdd/dictionary` → **编译失败**：`undefined: Dictionary`——这就是 RED，
测试描述了你想要但还不存在的代码。再新建 `dictionary.go`，照契约写出让测试通过的
**最少代码**，再跑一遍全绿，行为 1 完成。

| 用例 | 初始字典 | 调用 | 期望（逐条断言） |
|---|---|---|---|
| 查存在的词 | `Dictionary{"test": "this is just a test"}` | `Search("test")` | ① err 是 nil<br>② 返回值是 `"this is just a test"` |
| 查不存在的词 | 同上 | `Search("unknown")` | `errors.Is(err, ErrNotFound)` |

### 行为 2：Add 新词成功，重复词报错且不覆盖旧释义

先写 `TestAdd` 的两个用例（每个用例新建字典），跑出 `undefined: Add` 的编译失败即 RED；
再回 `dictionary.go` 写最少实现——一个用例驱动一行实现，先让"添加新词"变绿，再补重复检查。

| 用例 | 初始字典 | 调用 | 期望（逐条断言） |
|---|---|---|---|
| 添加新词 | 空字典 `Dictionary{}` | `Add("test", "this is just a test")` | ① 返回的 err 是 nil<br>② 再 `Search("test")` 得到 `"this is just a test"` 且无错误 |
| 重复添加 | `Dictionary{"test": "old"}` | `Add("test", "new")` | ① `errors.Is(err, ErrWordExists)`<br>② 再 `Search("test")` 仍是 `"old"`——**没被覆盖** |

### 行为 3：Update 存在的词生效，不存在的词报错

先写 `TestUpdate` 的两个用例（RED），再写最少实现变绿；第二个用例的断言 ② 不能省——
它防的是哪种错误实现，见第三节「行为 3」知识点。

| 用例 | 初始字典 | 调用 | 期望（逐条断言） |
|---|---|---|---|
| 更新存在的词 | `Dictionary{"test": "old"}` | `Update("test", "new")` | ① 返回的 err 是 nil<br>② 再 `Search("test")` 得到 `"new"` |
| 更新不存在的词 | 空字典 `Dictionary{}` | `Update("test", "new")` | ① `errors.Is(err, ErrWordDoesNotExist)`<br>② 再 `Search("test")` 仍返回 `ErrNotFound`——Update 不能悄悄变成 Add |

### 机制实验：nil map 写入 panic（做完行为 3 必做）

[basic/map.go](../../basic/map.go) 开头就写着"必须初始化才能使用"，L65 还演示过 slice 套 map 要逐个 make——现在亲手踩一次这个伏笔：

1. 在测试里临时写 `var d Dictionary`（只声明、不初始化），调 `d.Add("a", "b")` → panic：`assignment to entry in nil map`
2. 改成 `d := make(Dictionary)`（或字面量 `Dictionary{}`）→ 通过
3. 再试读：`var d Dictionary; d.Search("a")` → **不 panic**，返回零值和 `ErrNotFound`——读 nil map 是合法的，只有写会 panic
4. 删掉实验代码，确认测试重新全绿

原理：nil map 的底层哈希表指针是 nil，查表时 Go 直接返回零值；写入时却必须真有桶可写，运行时发现没有表，只能 panic。这也是"每个用例新建 Dictionary"要用字面量或 `make`、而不是 `var d Dictionary` 的原因。

### 行为 4：Delete 之后 Search 返回 ErrNotFound

先写 `TestDelete` 的两个用例（RED），再写最少实现变绿；然后做一轮针对测试代码本身的
REFACTOR——四个行为写下来，"判断 `errors.Is`"和"Search 后比对释义"各出现了好几次，
把这两类断言抽成测试文件底部的辅助函数（`assertError`/`assertDefinition`），替换各处断言，
全程保持绿。辅助函数怎么写、`t.Helper()` 起什么作用，见第三节「行为 4」知识点。

| 用例 | 初始字典 | 操作序列 | 期望 |
|---|---|---|---|
| 删除后查不到 | `Dictionary{"test": "this is just a test"}` | `Delete("test")`，再 `Search("test")` | `errors.Is(err, ErrNotFound)` |
| 删不存在的词 | 空字典 `Dictionary{}` | `Delete("ghost")`，再 `Search("ghost")` | 不 panic；Search 返回 `ErrNotFound` |

---

## 三、知识点总结

### 行为 1：Search 查词

1. **`Search` 的主体就是一次逗号-ok 查询**：`ok` 为 false 时返回 `("", ErrNotFound)`。
2. **查词分支用 `t.Fatalf` 而不是 `t.Errorf`**：err 非 nil 时 `got` 没有意义，后续断言必须中止（testbasic 机制实验的结论）。
3. **判断哨兵错误用 `errors.Is`**：字符串比较是禁止的（errhandling 的规矩）。

### 行为 2：Add 新词成功，重复词报错且不覆盖旧释义

1. **在 map 类型上定义方法（本练习核心领悟点）**。[basic/type_func.go](../../basic/type_func.go) 学过"自定义类型可以定义方法"；`type Dictionary map[string]string` 之后，`func (d Dictionary) Add(...)` 的接收者 `d` 是一次**值拷贝**——但 map 是引用类型，拷走的只是"引用头"（指向底层哈希表的指针），所以 `d[word] = definition` 写的是调用方那张表。[basic/map.go#L96](../../basic/map.go#L96) 演示过同一现象：`userinfo5 := userinfo4` 之后两边互相可见。
2. **与 struct 的对照（wallet 练习预告）**。struct 是值类型，值接收者改的是副本、调用方不可见，所以 wallet 练习的 `Deposit` 必须用指针接收者 `*Wallet`——[basic/struct_func.go#L16](../../basic/struct_func.go#L16) 的"修改字段必须用指针接收者"说的正是 struct 的情况。**口诀：别问"要不要改"，要问"拷贝之后还共享底层数据吗"——map/slice/channel 共享，值接收者就能改内容；struct/数组/基本类型不共享，必须上指针。** 注意界限：值接收者能改 map 的"内容"，却不能替调用方把 nil map 初始化——引用头本身是拷走的，方法里 `d = make(...)` 只改了本地副本。这正是机制实验要踩的坑。
3. **逗号-ok 判存在**：`_, exists := d[word]`（[basic/map.go#L52](../../basic/map.go#L52)）。不要用 `d[word] == ""` 当"不存在"——空串也可能是合法释义，零值不代表缺失。
4. **测试之间不共享状态**。行为 1 的两个子测试共享一个字典没事，因为都只读；从行为 2 开始出现**写操作**，共享可变状态就会让"先跑的用例污染后跑的"——失败时你分不清是实现错了，还是数据被上一个用例改过。每个用例新建 `Dictionary{...}` 只要一行，这是测试确定性最便宜的投资。
5. **小步节奏**：先只写 `d[word] = definition` 让"添加新词"变绿；再跑，看"重复添加"变红；最后补 `if exists` 检查。一个用例驱动一行实现，别一次写全。

### 行为 3：Update 存在的词生效，不存在的词报错

1. **Update 和 Add 是同一副骨架**：都是"逗号-ok 查存在性 → 按分支决定写不写"，只是判断方向相反。CRUD 四个操作的本质是同一件事：先回答"key 在不在"，再行动。写实现时体会这种对称，但**不要**为了消除重复提前抽象——两个分支各三行，直接写最清楚。
2. **第二个用例在防什么**：如果实现漏掉存在性检查、直接 `d[word] = definition`，Update 就退化成了 Add——断言 ② 专门抓这个回归。写每个断言时多问一句"删掉它，错误实现能蒙混过关吗"——不能，才说明这个断言有存在的价值。

### 行为 4：Delete 之后 Search 返回 ErrNotFound

1. **`delete` 内建函数**：`delete(m, k)` 删除 key；key 不存在是 **no-op，不报错**（[basic/map.go#L59](../../basic/map.go#L59)），对 nil map 也安全——所以 `Delete` 方法无错可报，契约里它没有返回值。
2. **断言重复了，就抽辅助函数**：四个行为写下来，"判断 `errors.Is`"和"Search 后比对释义"各出现了好几次。抽成测试文件底部的辅助函数，把各处断言替换掉。参考写法：

```go
func assertError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("期望错误 %v，得到 %v", want, got)
	}
}

func assertDefinition(t *testing.T, d Dictionary, word, want string) {
	t.Helper()
	got, err := d.Search(word)
	if err != nil {
		t.Fatalf("期望查到 %q，得到错误 %v", word, err)
	}
	if got != want {
		t.Errorf("期望 %q，得到 %q", want, got)
	}
}
```

3. **`t.Helper()`**：标记"我是辅助函数"，断言失败时报错行号指向**调用方**那一行，而不是辅助函数内部。没有它，所有失败都显示在辅助函数里，定位用例要靠猜。替换断言的过程全程保持绿——这是 RED → GREEN → **REFACTOR** 里的重构，对象是测试代码本身。

### 接收者怎么选：一张对照表（本练习核心领悟点）

| 底层类型 | 值接收者拷贝后 | 值接收者能改"内容"吗 | 结论 |
|---|---|---|---|
| map / channel | 仍共享底层数据（拷的是引用头） | ✅ 能 | 增删改方法用值接收者即可 |
| slice | 仍共享底层数组 | ⚠️ 改元素能；`append` 扩容可能换头，对调用方失效 | 细节见 tdd/slicelab |
| struct / 数组 / 基本类型 | 整份拷贝，互不可见 | ❌ 不能 | 改字段必须指针接收者（wallet 练习） |

一句话：**决定因素不是"要不要改"，而是"拷贝之后还共不共享底层数据"。**
再记界限：值接收者改不了"引用头本身"——方法内 `d = make(...)` 救不了调用方的 nil map。

### map CRUD 惯用法速查

| 操作 | 写法 | 易错点 |
|---|---|---|
| 查 | `v, ok := m[k]` | `ok` 才是存在性；`v == ""` 不能当"不存在" |
| 增 / 改 | `m[k] = v` | **写 nil map 会 panic**；读 nil map 合法、返回零值 |
| 删 | `delete(m, k)` | key 不存在是 no-op；对 nil map 也安全 |
| 初始化 | `make(map[K]V)` 或字面量 `map[K]V{...}` | `var m map[K]V` 得到 nil map，只能读不能写 |

### 哨兵错误速查（errhandling 的复习）

声明为包级变量 `var ErrXxx = errors.New(...)`；判断用 `errors.Is(err, ErrXxx)`；
**禁止** `err.Error() == "..."` 字符串比较。细节回 [tdd/errhandling](../errhandling/README.md)。

### 测试组织速查

- 每个用例新建被测对象，不共享可变状态（只读共享可以，写操作必须隔离）
- 同类断言重复出现 → 抽辅助函数 + `t.Helper()`，失败行号才会指向调用方
- 后续步骤依赖前置条件（err 为 nil 才能用返回值）用 `t.Fatal`，独立断言用 `t.Error`

### 与书目的对应

- 教程 ch03 复合数据类型（map 部分）——本练习把 [basic/map.go](../../basic/map.go) 的知识点全部变成了可运行的测试
- 教程 ch06 接口 / EG·11 接口——在命名类型上定义方法，是"类型满足接口"的前提；`Dictionary` 的四个值接收者方法同时属于 `Dictionary` 和 `*Dictionary` 的方法集（对照 [basic/interfaces_struct.go](../../basic/interfaces_struct.go) 的混合接收者讨论），这决定了它未来能被什么样的接口抽象
- learn-go-with-tests·maps 章——本练习的出处
- [tdd/errhandling](../errhandling/README.md)——哨兵错误地基；[tdd/testbasic](../testbasic/README.md)——表驱动与 `t.Run` 地基

---

## 四、验收标准

```bash
go test ./tdd/dictionary -v        # 全绿
go vet ./tdd/dictionary            # 无警告
go test ./tdd/dictionary -cover    # 四个方法分支有限，必须 100%
```

## 五、完成后自查（能口头回答才算过）

1. `Add`/`Update`/`Delete` 用值接收者为什么能生效？判断口诀是什么？wallet 练习的 `Deposit` 改余额为什么必须指针接收者？
2. `var d Dictionary; d["a"] = "b"` 会发生什么？`d["a"]` 读取又会发生什么？
3. 判断"词已存在"为什么用逗号-ok，而不是 `d[word] == ""`？
4. "重复 Add 不覆盖旧释义"这个断言如果删掉，什么样的错误实现会漏网？
5. 为什么每个用例都新建 Dictionary？行为 1 两个子测试共享字典为什么又没事？
6. `Delete` 为什么不需要返回 error？`delete` 一个不存在的 key 会怎样？
7. 辅助函数里的 `t.Helper()` 改变了失败输出的什么？

全部答清后，回到 [根 README 的 TDD 练习总目录](../../README.md#四目录划分评估与-tdd-驱动学习计划)，
把 #6 tdd/dictionary 从"待建"划掉——ch03 map 与 ch06 接口的知识点已用测试巩固。
