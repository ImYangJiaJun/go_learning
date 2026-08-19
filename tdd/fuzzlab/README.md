# FuzzLab TDD —— 罗马数字：性质测试与模糊测试

目标：这是 TDD 轨道的第三站。被测功能经典而小巧（阿拉伯数字转罗马数字），
**注意力全部放在测试思想的两次升级上**：从"挑样例"到"提炼规则"（基于属性的测试），
从"人想用例"到"机器撞边界"（Go 1.18 原生模糊测试）。哨兵错误部分直接复用
[tdd/errhandling](../errhandling/README.md) 的结论，不再展开。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 核心功能

这个包对外只提供一个能力：

- **把阿拉伯数字 `n` 转成罗马数字字符串**；合法范围 **1~3999**，越界返回哨兵错误 `ErrOutOfRange`

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/fuzzlab` 就是它的运行方式，验收者是测试，不是人。

### 调用关系（谁在调用谁）

```text
测试代码 ──► ArabicToRoman(n)                         行为 1（表驱动）、行为 2（性质测试）
fuzz 引擎 ──► FuzzArabicToRoman ──► ArabicToRoman(n)  行为 3（模糊测试）
```

这个包没有接口也没有依赖注入：调用方只有测试代码和 fuzz 引擎。行为 3 的特殊之处
在第二行——**替你调用被测函数的是引擎**，输入由它变异生成，你只在 `f.Fuzz` 回调里写断言。

### 文件计划（共 3 个文件 + 1 个可能自动出现的目录，按编号顺序建）

最终目录长这样（`testdata/` 仅当 fuzz 撞出崩溃时才由引擎创建，可能始终不出现）：

```text
tdd/fuzzlab/
├── roman_test.go    # 行为 1 的表驱动测试 + 行为 2 的性质测试
├── roman.go         # 实现：转换函数 + 哨兵错误
├── fuzz_test.go     # 行为 3 的模糊测试
└── testdata/fuzz/FuzzArabicToRoman/…   # 崩溃语料，引擎自动写入，人不手写
```

| # | 文件 | 这个文件是干什么的 | 里面要写的符号 | 什么时候建 |
|---|---|---|---|---|
| 1 | `roman_test.go` | 行为 1 的表驱动测试和行为 2 的性质测试 | `TestArabicToRoman`、`TestArabicToRomanProperties` | **第 1 个建** |
| 2 | `roman.go` | 转换函数与越界哨兵错误 | `ArabicToRoman`、`ErrOutOfRange` | 测试编译报错时 |
| 3 | `fuzz_test.go` | 行为 3 的模糊测试 | `FuzzArabicToRoman` | 行为 1、2 全绿之后 |
| — | `testdata/fuzz/FuzzArabicToRoman/…` | 崩溃语料，**由 fuzz 引擎自动写入，人不手写** | 无（语料文件由引擎生成） | 仅当 fuzz 撞出崩溃时 |

对外要写的符号一共 2 个——1 个函数加 1 个哨兵错误，就是下面契约里的全部，一个不多一个不少。

### 接口契约（固定，按此实现，名字不要改）

完备性原则：**测试代码直接用到的每一个符号都在下面**——本练习只有一个实现文件，
契约就是这一组。你唯一需要自己实现的是函数体和它内部的算法数据
（如行为 1 给出的面值表）；如果写代码时发现要发明契约之外的对外符号，说明走偏了。

**写在 `roman.go`：**（需要 `import "errors"`）

```go
package fuzzlab

// ErrOutOfRange 是越界哨兵错误：n 不在 [1, 3999] 时由 ArabicToRoman 返回。
// 哨兵错误必须声明为包级变量，调用方用 errors.Is 识别，禁止比较错误文案。
var ErrOutOfRange = errors.New("number out of range")

// ArabicToRoman 把阿拉伯数字 n 转成罗马数字字符串。
// n 越界（< 1 或 > 3999）时返回空串和 ErrOutOfRange；合法时 err 为 nil。
func ArabicToRoman(n int) (string, error)
```

**契约核对清单**（写完代码后数一遍，应一个不少）：

- 0 个类型：本契约不定义类型，输入输出都是内置类型
- 1 个函数：`ArabicToRoman`
- 1 个哨兵错误：`ErrOutOfRange`

### 第一步：手把手起步（RED → GREEN）

1. 在 `tdd/fuzzlab/` 下新建 `roman_test.go`，写入行为 1 的第一个测试：

```go
package fuzzlab

import "testing"

func TestArabicToRoman(t *testing.T) {
	got, err := ArabicToRoman(1)
	if err != nil {
		t.Fatalf("合法输入 1 不应报错: %v", err)
	}
	if got != "I" {
		t.Errorf("期望 %q，得到 %q", "I", got)
	}
}
```

2. 运行 `go test ./tdd/fuzzlab` → **编译失败**：`undefined: ArabicToRoman`。
   这就是 RED——测试描述了你想要但还不存在的代码。
3. 新建 `roman.go`，先照抄接口契约，函数体写**最少**的代码让测试通过
   （哪怕先 `return "I", nil`——亲身体会"最小实现"，后面的用例会逼它长大）。
4. 再跑 `go test ./tdd/fuzzlab` → 绿。第一轮 RED → GREEN 完成。

注意这里用了 `t.Fatalf` 而不是 `t.Errorf`：`got` 只有在 `err == nil` 时才有意义，
前置条件不成立就必须立刻中断——这正是 testbasic 练过的 Fatal/Error 分工。

---

## 二、任务单（边做边学）

### 行为 1：表驱动 TDD 出转换算法

把第一步的测试重构成表驱动，覆盖下面全部用例。**每个用例都是一轮 RED → GREEN**：
加一行红一条，改一点实现变绿，实现是被用例一步步"逼"出来的。

| 用例名 | 输入 | 期望结果 | 期望错误 |
|---|---|---|---|
| 最小值 | `1` | `"I"` | `nil` |
| 重复累加 | `3` | `"III"` | `nil` |
| 减法形式四 | `4` | `"IV"` | `nil` |
| 减法形式九 | `9` | `"IX"` | `nil` |
| 混合 | `14` | `"XIV"` | `nil` |
| 减法形式四十 | `40` | `"XL"` | `nil` |
| 大数 | `2024` | `"MMXXIV"` | `nil` |
| 最大值 | `3999` | `"MMMCMXCIX"` | `nil` |
| 下界越界 | `0` | `""` | `ErrOutOfRange` |
| 上界越界 | `4000` | `""` | `ErrOutOfRange` |

行为 1 给出完整模板（这是本练习唯一一份完整测试代码，后续行为自己写）：

```go
package fuzzlab

import (
	"errors"
	"testing"
)

func TestArabicToRoman(t *testing.T) {
	cases := []struct {
		name    string
		arabic  int
		want    string
		wantErr error
	}{
		{name: "最小值", arabic: 1, want: "I", wantErr: nil},
		{name: "重复累加", arabic: 3, want: "III", wantErr: nil},
		{name: "减法形式四", arabic: 4, want: "IV", wantErr: nil},
		{name: "减法形式九", arabic: 9, want: "IX", wantErr: nil},
		{name: "混合", arabic: 14, want: "XIV", wantErr: nil},
		{name: "减法形式四十", arabic: 40, want: "XL", wantErr: nil},
		{name: "大数", arabic: 2024, want: "MMXXIV", wantErr: nil},
		{name: "最大值", arabic: 3999, want: "MMMCMXCIX", wantErr: nil},
		{name: "下界越界", arabic: 0, want: "", wantErr: ErrOutOfRange},
		{name: "上界越界", arabic: 4000, want: "", wantErr: ErrOutOfRange},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ArabicToRoman(c.arabic)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("错误断言失败：期望 %v，得到 %v", c.wantErr, err)
			}
			if got != c.want {
				t.Errorf("期望 %q，得到 %q", c.want, got)
			}
		})
	}
}
```

**这一步用到的知识点：**

1. **一张表同时覆盖成功与失败**：`errors.Is(err, c.wantErr)` 对成功用例传入 `wantErr: nil`，
   而 `errors.Is(err, nil)` 在语义上等价于 `err == nil`——`Is` 的实现里 target 为 nil 时
   退化为直接判等。所以成功和失败用例共用同一句断言，不用在循环里写 if 分支。
2. **哨兵错误的断言方式**：失败用例不断言错误文案，而是用 `errors.Is` 识别哨兵变量——
   这是 errhandling 练过的规则：文案可能改，哨兵变量的身份不变。
3. **罗马数字算法的核心技巧**：不要为 4、9、40、90、400、900 写 if——把减法形式
   （`IV`=4、`IX`=9、`XL`=40……）也当作"面值"，和正常面值一起排进**从大到小**的表里：

```go
var romanNumerals = []struct {
	value  int
	symbol string
}{
	{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
	{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
	{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
}
```

   转换就变成了**贪心找零**：从最大面额开始，够减就拼符号再减，直到 n 归零——
   和人民币找零完全同构。面值表不从大到小排，贪心就失效（会先拼出一堆小符号）。
4. **TDD 的"三角法"**：先用例 1、3 逼出循环拼接 `I`；用例 4、9 逼出减法特判；
   用例 40、90、400、900 再逼你把特判泛化成查表。每一步实现都不超前于测试——
   这就是"测试驱动设计"的字面意思：设计的形状由用例序列决定。
5. **字符串拼接用 `strings.Builder`**：循环里 `result += symbol` 每次都产生新字符串；
   `strings.Builder` 内部维护字节缓冲区，`WriteString` 零拷贝，`String()` 一次成型。
   本练习最多拼 15 个字符，性能无所谓——用这个包是为了养成习惯。

### 行为 2：基于属性的测试——从样例到规则

行为 1 的 8 个合法用例全是"人挑的样例"。现在换一种问法：**对任意合法 n，什么永远成立？**
不写具体期望值，只断言规则。测试骨架如下，断言自己补：

```go
func TestArabicToRomanProperties(t *testing.T) {
	for n := 1; n <= 3999; n++ {
		roman, err := ArabicToRoman(n)
		// 性质 0：合法输入永不报错（err 必须先 Fatal 掉，否则 roman 不可用）
		// 性质 1：roman 只由 "IVXLCDM" 中的字符组成
		// 性质 2：同一字符连续出现不超过 3 次
	}
}
```

| 性质 | 形式化断言 | 它能抓住什么反例 |
|---|---|---|
| 合法输入永不报错 | `err == nil` 对 1~3999 全部成立 | 实现漏处理某段区间（如 900~999 panic 或报错） |
| 字符集封闭 | 结果只含 `I V X L C D M` | 实现误输出别的字符（如小写、多余符号） |
| 最多三连 | 不存在 4 个相同字符连续出现 | `4 → "IIII"`、`900 → "DCCCC"` 这类没用减法形式的实现 |

跑 `go test ./tdd/fuzzlab -run TestArabicToRomanProperties -v` 确认全绿。

**这一步用到的知识点：**

1. **性质测试 vs 示例测试的思维差异**：示例测试断言"**这个**输入 → **那个**输出"，
   用例是人挑的代表；性质测试断言"**任意**输入 → 总成立的规则"，规则是从领域定义提炼的
   不变量。前者是规约的实例，后者就是规约本身。罗马数字的定义里本来就写着
   "符号集 {I,V,X,L,C,D,M}"和"同一符号不连用超过三次"——性质不是发明的，是需求里本来就有的。
2. **为什么这里敢穷举**：性质测试的经典做法是随机生成大量输入（quickcheck 风格），
   但本练习定义域只有 3999 个值——**"任意"可以直接翻译成"全部"**，穷举比随机更彻底、
   更快、还确定性可复现。定义域小的时候，这是性质测试的最强形态。
3. **性质之间的互补**：性质 1 抓不到 `4 → "IIII"`（字符都合法），性质 2 能抓到；
   但两条性质合起来也证明不了 `14` 恰好是 `"XIV"` 而不是 `"VIV"`——**具体映射关系只有
   示例测试能钉死**。所以行为 2 是行为 1 的增强，不是替代：样例钉死映射，性质守住规则。
4. **遍历字符串的粒度**：罗马符号都是 ASCII，按 byte 遍历和按 rune 遍历结果一样；
   但写 `for _, r := range roman`（rune）是更安全的默认——遇到多字节字符时 byte 循环
   会把一个字符拆成几段，导致"连续重复"判断出错。判断字符集用 `strings.ContainsRune`
   或一个 `switch r { case 'I','V','X','L','C','D','M': }` 均可。
5. **失败信息要带输入**：性质测试断言之广，失败时第一件事是问"哪个 n 挂的"——
   `t.Errorf("n=%d: 性质X被违反，得到 %q", n, roman)`，没有输入值的失败报告等于没报告。

### 行为 3：`FuzzXxx` 模糊测试——让机器替你撞边界

新建 `fuzz_test.go`，骨架如下，断言自己补：

```go
package fuzzlab

import (
	"errors"
	"testing"
)

func FuzzArabicToRoman(f *testing.F) {
	// 种子语料：人给的"重点嫌疑值"
	f.Add(1)    // 下界
	f.Add(4)    // 减法形式代表
	f.Add(3999) // 上界
	f.Fuzz(func(t *testing.T, n int) {
		roman, err := ArabicToRoman(n)
		if n < 1 || n > 3999 {
			// 越界输入：必须返回 ErrOutOfRange（errors.Is 断言）
		} else {
			// 合法输入：err 为 nil，且满足行为 2 的两条性质
		}
	})
}
```

| 种子 | 为什么选它 |
|---|---|
| `f.Add(1)` | 下界边界，越界判断的第一道关 |
| `f.Add(4)` | 减法形式的代表，最容易写错的分支 |
| `f.Add(3999)` | 上界边界，面值表耗尽处 |

运行（一次只 fuzz 一个目标，给它 15 秒预算）：

```bash
go test -fuzz=FuzzArabicToRoman -fuzztime=15s ./tdd/fuzzlab
```

实现正确时输出类似 `fuzz: elapsed: 15s, execs: 800000 ... PASS`；
若撞出崩溃，崩溃输入会自动落盘到 `testdata/fuzz/FuzzArabicToRoman/<哈希>` 文件。

**这一步用到的知识点：**

1. **Go 1.18 的 fuzzing 是覆盖率引导的随机测试，不是瞎随机**：引擎每生成一个输入，
   都会观察它让被测代码走到了哪些分支；**能覆盖新路径的输入会被收进语料库继续变异**，
   走老路的直接丢弃。这套反馈机制（源自 AFL/libFuzzer）让它能在海量输入空间里
   高效摸到你没想到的边界组合——比如本练习的 `-1`、`0`、`4000`、`math.MinInt`。
2. **`FuzzXxx` 的结构三段式**：`func FuzzXxx(f *testing.F)`（同样是大写开头的命名约定）；
   `f.Add(...)` 注册种子语料，参数类型必须与 `f.Fuzz` 回调里 `t` 之后的参数一一对应
   （本练习回调是 `func(t *testing.T, n int)`，所以种子是 int）；
   `f.Fuzz` 注册 fuzz 目标函数——它就是被反复执行的那个测试体。
3. **不加 `-fuzz` 时 `FuzzXxx` 也参与普通 `go test`**：此时引擎只用种子语料各跑一遍，
   相当于把种子当成普通测试用例。这意味着 fuzz 文件提交进仓库后，CI 不加 `-fuzz`
   也能守住种子——**种子和崩溃语料都是回归测试资产**。
4. **`-fuzztime=15s` 的作用**：fuzzing 是不收敛的持续过程，没有"跑完了"的概念，
   不加时间预算它会一直跑到你按 Ctrl-C。`-fuzztime` 就是给它的时间预算——
   本地练手 15s 足够（本练习一秒能执行几十万次）；CI 里常用 `-fuzztime=30s` 或 `1m`，
   专门的 fuzz 流水线才按小时计。
5. **崩溃语料的落盘与回归**：引擎撞出崩溃时，会把**最小化的**触发输入写入
   `testdata/fuzz/FuzzArabicToRoman/` 下的文件——`testdata` 是 Go 工具链认定的
   测试附属数据目录，`go test` 会自动把里面的语料加入种子集。也就是说：
   **崩溃一次，从此每次 `go test`（不带 `-fuzz`）都会重放它**，这个 bug 不可能悄悄复发。
6. **fuzz 目标里为什么不能断言具体期望值**：输入是机器随机生成的，你事先不知道
   它会喂什么，自然写不出 `want`。所以 fuzz 断言的素材只有两类——**不变量/性质**
   （本练习复用行为 2 的两条性质 + 越界必须返回哨兵错误）和**粗粒度健康检查**
   （不 panic、不超时）。这逼你把"什么永远成立"想清楚，性质测试的思维方式在这里变现。
7. **fuzz 与表驱动的分工**：表驱动**固化已知**——人认为重要的样例，可读、可评审、
   有文档价值；fuzz **探索未知**——机器生成人想不到的输入组合。两者不是竞争关系：
   正确姿势是表驱动钉死行为契约，fuzz 在其上扫边界。fuzz 撞出的有价值的崩溃，
   修好后还应回头补一行进表驱动——让"未知"沉淀为"已知"。

---

## 三、知识点总结

### 三种测试思维速查

| | 示例测试（表驱动） | 性质测试 | 模糊测试 |
|---|---|---|---|
| 断言对象 | 具体输入 → 具体输出 | 任意输入 → 规则 | 随机输入 → 不变量 |
| 用例来源 | 人挑选代表 | 人提炼规则，机器枚举/生成 | 覆盖率引导的引擎变异生成 |
| 失败意味着 | 这个样例算错了 | 违反了规约 | 撞到了没人想到的边界 |
| 回答不了 | 没挑到的输入对不对 | 具体映射是否精确 | 输出是否符合预期 |
| 本练习 | 行为 1 | 行为 2 | 行为 3 |

一句话：**样例钉死映射，性质守住规则，fuzz 探索边界**——三者互补，缺一不可。

### Fuzz 机制速查

| 机制 | 说明 |
|---|---|
| `func FuzzXxx(f *testing.F)` | fuzz 入口，`Fuzz` + 大写字母开头 |
| `f.Add(种子...)` | 注册种子语料，类型与 `f.Fuzz` 回调参数一一对应 |
| `f.Fuzz(func(t *testing.T, n int){...})` | fuzz 目标，被引擎反复执行 |
| `go test`（不带 `-fuzz`） | 只把种子语料当普通测试各跑一遍 |
| `go test -fuzz=FuzzXxx -fuzztime=15s` | 进入 fuzzing 模式，跑满时间预算 |
| `testdata/fuzz/FuzzXxx/<哈希>` | 崩溃语料自动落盘点，此后成为永久回归用例 |
| 覆盖率引导 | 能覆盖新路径的输入被保留变异，走老路的丢弃 |

### 罗马数字面值表速查

`M=1000 CM=900 D=500 CD=400 C=100 XC=90 L=50 XL=40 X=10 IX=9 V=5 IV=4 I=1`——
从大到小贪心找零；减法形式不当规则处理，当**面值**处理。

### 与书目的对应

- **learn-go-with-tests《Intro to property based tests》**：本练习的母题——
  roman numerals kata。行为 1 的表驱动推演、行为 2 从样例到性质的思维跃迁都出自该章。
  该章后续还有 `RomanToArabic` 反向转换与"转回去等于原数"的往返性质（round-trip property），
  可作为本练习的扩展项。
- **Go 1.18 Release Notes 与官方文档 [go.dev/doc/fuzz](https://go.dev/doc/fuzz)**：
  fuzzing 进入 Go 标准工具链的起点；`f.Add`/`f.Fuzz`/`-fuzztime`/崩溃语料落盘的官方出处。
- 哨兵错误与 `errors.Is` 的规则回顾：[tdd/errhandling](../errhandling/README.md)。

---

## 四、验收标准

```bash
go test ./tdd/fuzzlab -v                                       # 全绿
go vet ./tdd/fuzzlab                                           # 无警告
go test ./tdd/fuzzlab -cover                                   # 合法/越界两条路径，必须 100%
go test -fuzz=FuzzArabicToRoman -fuzztime=15s ./tdd/fuzzlab    # 跑满 15 秒无崩溃
```

## 五、完成后自查（能口头回答才算过）

1. 罗马数字的减法形式（IV/IX/XL…）在算法里是怎么处理的？面值表为什么必须从大到小排序？
2. 为什么 `errors.Is(err, c.wantErr)` 能让一张表同时覆盖成功和失败用例？
3. 性质测试和示例测试的思维差异是什么？本练习为什么敢用穷举 1~3999 代替"任意"？
4. 为什么两条性质合起来也替代不了表驱动？样例和性质各自抓住什么、漏掉什么？
5. Go 的 fuzzing 是"纯随机"吗？"覆盖率引导"具体引导的是什么？
6. 不加 `-fuzz` 时 `go test` 怎么对待 `FuzzXxx`？`-fuzztime` 解决什么问题？
7. fuzz 撞出的崩溃输入存到哪去了？为什么说它从此自动变成回归测试？

全部答清后，回到 [根 README](../../README.md#四目录划分评估与-tdd-驱动学习计划)，
把 TDD 练习总目录阶段 0 第 3 行 `tdd/fuzzlab` 的状态从「待建」改成「✅ 已建」；
testbasic 在覆盖检查里给 ch17 留下的"还剩 Fuzz 未学"的尾巴，至此也可以补齐了。
