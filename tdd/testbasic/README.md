# TestBasic TDD —— 第一个练习：学会 testing 工具

目标：这是 TDD 轨道的第一站。被测函数故意简单（重复字符串），**注意力全部放在
testing 工具本身**：`TestXxx`、表驱动、`t.Run` 子测试、Benchmark、覆盖率。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/testbasic` 就是它的运行方式，验收者是测试，不是人。

这个包对外只提供一个能力：

- **把字符串 `s` 重复 `n` 次返回**；`n <= 0` 时返回空字符串

### 文件计划（共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `repeat_test.go` | 全部测试 + 基准测试 | **第 1 个建** |
| `repeat.go` | `Repeat` 函数（唯一要实现的函数） | 测试编译报错时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package testbasic

// Repeat 将 s 重复 n 次返回；n <= 0 时返回空字符串
func Repeat(s string, n int) string
```

---

## 二、任务单（边做边学）

### 行为 1：跑通第一个测试（RED → GREEN）

1. 在 `tdd/testbasic/` 下新建 `repeat_test.go`，写入：

```go
package testbasic

import "testing"

func TestRepeat(t *testing.T) {
	got := Repeat("a", 3)
	want := "aaa"
	if got != want {
		t.Errorf("期望 %q，得到 %q", want, got)
	}
}
```

2. 运行 `go test ./tdd/testbasic` → **编译失败**：`undefined: Repeat`。
3. 新建 `repeat.go`，写**最少**的代码让测试通过
   （哪怕先 `return "aaa"` 再泛化成循环——亲身体会"最小实现"）。
4. 再跑 `go test ./tdd/testbasic` → 绿。第一轮 RED → GREEN 完成。

**这一步用到的知识点：**

1. **`_test.go` 文件约定**：`go test` 只认以 `_test.go` 结尾的文件；`go build` 打包时会完全忽略它们，所以测试代码永远不会进入可执行文件。测试文件必须和被测代码**同目录**；本练习与被测代码**同包**（`package testbasic`），因此可以直接调用未导出（小写开头）的函数。
2. **`TestXxx` 命名规则**：函数名必须是 `Test` + 大写字母开头的单词——`TestRepeat` 会被执行，`Testrepeat`（小写 r）会被静默忽略。参数固定为 `*testing.T`，无返回值。`go test` 没有注册表，就是靠这套命名约定找到所有测试的。
3. **`t.Errorf`**：报告"这条测试失败"，但测试函数继续往下执行。`%q` 是给字符串加引号的格式化动词——空串 `""` 和 `" "` 在输出里一眼能分清。
4. **RED 的第一课**：测试失败有两种形态——断言失败和**编译失败**。`undefined: Repeat` 就是编译失败，同样算 RED：测试描述了你想要但还不存在的代码，这就是"测试驱动"的字面意思。

### 行为 2：表驱动 + 子测试（REFACTOR 测试代码）

把行为 1 的测试重构成表驱动，覆盖下面全部用例。**实现 `repeat.go` 一行不动**：

| 用例名 | 输入 | 期望 |
|---|---|---|
| 重复三次 | `Repeat("a", 3)` | `"aaa"` |
| 重复两次 | `Repeat("ab", 2)` | `"abab"` |
| 零次得空串 | `Repeat("x", 0)` | `""` |
| 空串重复 | `Repeat("", 5)` | `""` |

跑 `go test ./tdd/testbasic -v` 观察树形输出；再跑 `go test ./tdd/testbasic -run 'TestRepeat/零次得空串'` 只执行一个子测试。

**这一步用到的知识点：**

1. **表驱动测试**：把用例组织成"结构体切片 + for 循环"，是 Go 社区最主流的测试组织方式。对比复制 4 个 Test 函数：加用例只需加一行，断言逻辑只有一份。骨架如下，**按用例表补齐其余 3 行**：

```go
func TestRepeat(t *testing.T) {
	cases := []struct {
		name  string
		input string
		times int
		want  string
	}{
		{name: "重复三次", input: "a", times: 3, want: "aaa"},
		// 按用例表补齐其余 3 行
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Repeat(c.input, c.times)
			if got != c.want {
				t.Errorf("期望 %q，得到 %q", c.want, got)
			}
		})
	}
}
```

2. **`t.Run` 子测试**：`t.Run(名字, func(t *testing.T) {...})` 把一组断言变成命名子测试。`-v` 输出里显示为 `TestRepeat/重复三次` 的树形层级，哪个用例挂了一目了然。
3. **`-run` 过滤**：参数是正则表达式；用斜杠路径定位子测试，如 `-run 'TestRepeat/零次得空串'`。注意：子测试名里的空格会被替换成下划线（"重复 三次" 会变成 `重复_三次`），过滤时要用替换后的名字。
4. **本步的循环是 RED → GREEN → REFACTOR 里的 REFACTOR**：重构对象是测试代码本身，全程测试保持绿——体会"有测试保护，才敢重构"。

### 行为 3：基准测试与 Example

在 `repeat_test.go` 里加 `BenchmarkRepeat`，跑：

```bash
go test ./tdd/testbasic -bench=. -benchmem
```

同时补充一个可执行文档示例 `ExampleRepeat`：它必须打印 `aaa`，并使用 `// Output:`
作为验收标准。普通 `go test` 会执行带 `// Output:` 的 Example；没有该注释的 Example
只参与编译，不会比较输出。

**这一步用到的知识点：**

1. **`BenchmarkXxx` 签名**：`func BenchmarkRepeat(b *testing.B)`，`Benchmark` + 大写字母开头。**`go test` 默认不跑 benchmark**，必须加 `-bench`——这是它和 `TestXxx` 最关键的区别。
2. **`b.N`**：框架会自动调整 `b.N`，直到单次耗时测量稳定。你的职责只有一个——循环体里只放被测代码：

```go
func BenchmarkRepeat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Repeat("a", 3)
	}
}
```

3. **读懂输出**：

```text
BenchmarkRepeat-8    50000000    25.3 ns/op    16 B/op    1 allocs/op
```

依次是：测试名-CPU核数、循环执行次数、每次操作耗时（ns/op）、每次操作分配的字节数（B/op）、每次操作的堆分配次数（allocs/op）。后两列由 `-benchmem` 开启；`allocs/op` 是 Go 性能优化的重点观察对象——每次堆分配都会增加 GC 压力（与 [basic/goroutine.go](../../basic/goroutine.go) 的并发一起构成性能课的两块基石）。

4. `-bench=.` 的 `.` 是正则，匹配所有 benchmark；也可以写 `-bench=BenchmarkRepeat` 精确匹配。
5. **Go 1.24+ 的现代写法 `b.Loop()`**：`for b.Loop() { Repeat("a", 3) }` 可以替代 `b.N` 循环——更简洁，且能防止编译器把被测调用优化掉。本练习先用 `b.N` 理解机制，之后写 benchmark 一律用 `b.Loop()`。

### 行为 4：用覆盖率找出漏掉的用例

```bash
go test ./tdd/testbasic -cover
```

如果实现里 `n <= 0` 分支没被任何用例走到，覆盖率就不到 100%。补一个 `n = -1` 的用例进表驱动，再跑到 **100%**。

**这一步用到的知识点：**

1. **`-cover`**：报告测试执行到了被测代码百分之多少的**语句**（语句覆盖）。它不是质量分，是找漏网用例的工具。
2. **想看具体哪行没覆盖**：

```bash
go test ./tdd/testbasic -coverprofile=cover.out
go tool cover -html=cover.out   # 浏览器打开，红色 = 未覆盖
```

用完删掉 `cover.out`，不要提交进仓库。

3. **覆盖率的态度**：本练习逻辑只有两条路径（n>0 / n≤0），必须 100%；以后遇到复杂代码，不盲目追 100%——为凑数字写的测试没有行为意义。

### 机制实验：`t.Error` 和 `t.Fatal` 的区别

1. 把行为 1 测试里的 `want` 故意改成 `"aaaa"`，跑一遍，看 `t.Errorf` 的失败输出
2. 在断言后面加一行 `t.Log("断言之后还能执行到我")` 再跑——Log 出现了
3. 把 `t.Errorf` 换成 `t.Fatalf` 再跑——Log 消失了
4. 把 `want` 改回正确值、删掉 `t.Log`，确认测试重新变绿

原理：`t.Fatal` 内部调用 `runtime.Goexit`，立即终止当前测试函数所在的 goroutine——所以** defer 语句仍会执行**，但后续普通语句不再执行。结论：后面代码依赖前置条件时（比如 err 为 nil 才能用返回值）用 `t.Fatal`，多个独立断言用 `t.Error`。

---

## 三、知识点总结

### `Test` / `Benchmark` / `Example` 三种函数的区别

`go test` 靠命名前缀识别三种入口，它们的签名、执行时机、成败判断完全不同：

| | `TestXxx(t *testing.T)` | `BenchmarkXxx(b *testing.B)` | `ExampleXxx()` |
|---|---|---|---|
| 用途 | 正确性断言 | 性能测量 | 可执行的文档示例 |
| `go test` 默认执行？ | ✅ 执行 | ❌ 必须加 `-bench` | 有 `// Output:` 时执行；否则只编译 |
| 怎么判断成败 | 代码里调 `t.Error`/`t.Fatal` | 不判成败，输出耗时数据 | 比较实际 stdout 和 `// Output:` 注释 |
| 本练习出现在 | 行为 1、2、4 | 行为 3 | 未含（见下） |

`ExampleXxx` 长这样——它是写给包的**使用者**看的示例，会出现在 godoc 文档里：

```go
func ExampleRepeat() {
	fmt.Println(Repeat("a", 3))
	// Output: aaa
}
```

有 `// Output:` 注释才会被当作测试执行并比对输出；没有则只参与编译。本练习要求完成带输出注释的版本。

### 命名与文件约定速查

- 文件：`*_test.go`，与被测代码同目录；`go build` 完全忽略
- 函数：`Test`/`Benchmark`/`Example` 前缀，后接**大写字母**开头的名字
- 一句话：**测试就是一个普通函数 + 一套命名约定**，`go test` 没有魔法

### `go test` 常用参数速查

| 参数 | 作用 |
|---|---|
| `-v` | 显示每个测试的名字和结果（含子测试树） |
| `-run 正则` | 只跑匹配的测试；斜杠路径定位子测试 |
| `-cover` | 打印语句覆盖率 |
| `-coverprofile=f.out` + `go tool cover -html=f.out` | 看具体哪行没覆盖 |
| `-bench 正则` | 跑基准测试 |
| `-benchmem` | 基准报告附带内存分配指标 |
| `-race` | 数据竞争检测（后续并发练习用） |

### 断言工具速查

| 方法 | 失败后 | 适用场景 |
|---|---|---|
| `t.Error` / `t.Errorf` | 继续执行 | 多个独立断言 |
| `t.Fatal` / `t.Fatalf` | 立即中断当前测试（defer 仍执行） | 后续步骤依赖前置条件 |
| `t.Log` / `t.Logf` | 不是断言 | 只在失败或 `-v` 时显示的辅助信息 |

### 表驱动模式一句话

用例 = 结构体切片里的一行；断言逻辑只有一份；加用例 = 加一行。

---

## 四、验收标准

```bash
go test ./tdd/testbasic -v                      # 全绿
go vet ./tdd/testbasic                          # 无警告
go test ./tdd/testbasic -cover                  # 本练习逻辑简单，必须 100%
go test ./tdd/testbasic -bench=. -benchmem      # 能跑出基准数据
go test ./tdd/testbasic -run ExampleRepeat       # 执行文档示例
```

## 五、完成后自查（能口头回答才算过）

1. `_test.go` 文件会被 `go build` 打进可执行文件吗？为什么？
2. `TestXxx` 的命名规则是什么？`Testrepeat` 会被 `go test` 执行吗？
3. `Test`/`Benchmark`/`Example` 三种函数，签名、默认是否执行、成败判断各有什么不同？
4. `t.Error` 和 `t.Fatal` 的区别？什么场景必须用 `t.Fatal`？
5. 表驱动测试相比"复制粘贴多个 Test 函数"，好处是什么？
6. `-run 'TestRepeat/零次得空串'` 里的斜杠是什么语法？子测试名含空格会怎样？
7. Benchmark 报告里 `ns/op`、`B/op`、`allocs/op` 分别表示什么？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把 ch17"Go 程序测试"从 ❌ 改成 ◐（还剩 Fuzz 未学，后续练习补）。
