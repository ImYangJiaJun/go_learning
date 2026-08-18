# Sortiface TDD —— sort.Interface：把排行榜交给标准库排序

目标：场景是应用下载排行榜（教程 ch16 的例子）。被测逻辑很轻——就三个方法的事；
**注意力全部放在 `sort.Interface` 这个能力契约上**：标准库的排序算法完全不知道你的
类型是什么，只依赖 `Len`/`Less`/`Swap` 三个方法。顺带看清稳定性、`Reverse` 装饰器，
以及 Go 1.21 泛型时代的新写法。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/sortiface` 就是它的运行方式，验收者是测试，不是人。

这个包对外提供的能力：

- 一个 `App` 结构体：应用 ID + 下载量
- `ByDownloads`：让 `[]App` 能按**下载量降序**被 `sort.Sort` 排序（排行榜主视角）
- `ByID`：让同一份数据能按 **ID 升序**排序（第二视角）

> 包名取 `sortiface` 而不是 `sort`：本练习要大量导入标准库 `sort`，包名若同名，
> 读代码时 `sort.Sort` 到底指谁都分不清。下文凡 `import "sort"` 处，标识符 `sort`
> 一律指标准库。

### 文件计划（共 2 个文件）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `app_test.go` | 全部测试 | **第 1 个建** |
| `app.go` | `App` 结构体 + `ByDownloads`/`ByID` 共六个方法 | 测试编译报错时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package sortiface

// App 是排行榜里的一个应用。两个字段都是 int，因此 App 可比较（支持 == / !=）
type App struct {
	ID        int // 应用唯一标识
	Downloads int // 累计下载量
}

// ByDownloads 让 []App 具备"按下载量降序"的排序能力。
// 实现 sort.Interface：Less(i, j) 返回 true 表示第 i 个元素应排在第 j 个前面
type ByDownloads []App

func (a ByDownloads) Len() int           // 固定套路：return len(a)
func (a ByDownloads) Less(i, j int) bool // 本类型唯一有语义的方法：下载量大的在前（降序）
func (a ByDownloads) Swap(i, j int)      // 固定套路：a[i], a[j] = a[j], a[i]

// ByID 让 []App 具备"按 ID 升序"的排序能力，同样实现 sort.Interface
type ByID []App

func (a ByID) Len() int           // 固定套路
func (a ByID) Less(i, j int) bool // ID 小的在前（升序）
func (a ByID) Swap(i, j int)      // 固定套路
```

### 第一步：手把手起步（行为 1 的 RED → GREEN）

1. 新建 `tdd/sortiface/` 目录，**先只建** `app_test.go`，原样粘贴：

```go
package sortiface

import (
	"sort"
	"testing"
)

func TestSortByDownloads(t *testing.T) {
	apps := []App{
		{ID: 3, Downloads: 500},
		{ID: 1, Downloads: 1200},
		{ID: 2, Downloads: 800},
	}
	sort.Sort(ByDownloads(apps))
	want := []App{
		{ID: 1, Downloads: 1200},
		{ID: 2, Downloads: 800},
		{ID: 3, Downloads: 500},
	}
	for i := range want {
		if apps[i] != want[i] {
			t.Errorf("第 %d 位：期望 %+v，得到 %+v", i, want[i], apps[i])
		}
	}
}
```

2. 运行 `go test ./tdd/sortiface` → **编译失败**：`undefined: App`、`undefined: ByDownloads`。
   编译失败同样算 RED——测试描述了你想要但还不存在的代码。
3. 新建 `app.go`，写**最少**的代码：只写 `App` 结构体和 `ByDownloads` 的三个方法
   （`ByID` 现在不写——留到行为 2，让测试再驱动一次）。
4. 再跑 `go test ./tdd/sortiface` → 绿。第一轮 RED → GREEN 完成。

---

## 二、任务单（边做边学）

### 行为 1：用 sort.Sort 排出下载量降序

测试代码已在第一节贴出并跑绿。对照用例表确认你理解了每一行：

| 用例名 | 输入 apps（花括号内是 {ID, Downloads}） | 排序后期望 |
|---|---|---|
| 三个应用按下载量降序 | `[{3,500} {1,1200} {2,800}]` | `[{1,1200} {2,800} {3,500}]` |

**这一步用到的知识点：**

1. **`sort.Interface` 是能力契约**：标准库 `func Sort(data Interface)` 的入参是接口——
   排序算法从头到尾只调用 `Len`/`Less`/`Swap` 三个方法，**完全不知道你的类型是
   `[]App`**。定义在 `sort` 包源码里就三行：`Len() int`、`Less(i, j int) bool`、
   `Swap(i, j int)`。这是 [basic/interface.go](../../basic/interface.go) 里"接口作参数"
   在标准库的真实落地：面向能力编程，而不是面向类型编程。Go 接口**隐式满足**——你不用
   声明"ByDownloads 实现了 sort.Interface"，方法集够了，传参时编译器自然放行。
2. **为什么接收者是值类型 `(a ByDownloads)` 而不是指针**：[basic/struct_func.go](../../basic/struct_func.go)
   说过"修改字段必须用指针接收者"，但那里改的是结构体本身。切片是引用类型——slice
   header 里存着指向底层数组的指针，`Swap` 改的是**底层数组**，不是 header，所以值接收者
   照样生效。易错点：哪天方法里要 `append`（会改 header 的 len/cap），才轮到指针接收者。
3. **`ByDownloads(apps)` 是类型转换不是复制**：`type ByDownloads []App` 与 `[]App` 底层类型
   相同，转换零开销，只是给同一个切片"换了个带方法的名字"。为什么要这个壳？因为方法只能
   定义在**命名类型**上——`[]App` 是复合类型字面量，没法挂方法（[basic/type_func.go](../../basic/type_func.go)：
   非本地类型不能加方法）。
4. **三个方法里只有 `Less` 因类型而异**：`Len` 永远 `return len(a)`，`Swap` 永远
   `a[i], a[j] = a[j], a[i]`，纯样板；真正表达"排行榜语义"的只有 `Less` 一行。
   易错点：`Less(i, j)` 的语义是"**第 i 个是否应排在第 j 个前面**"，不是"i 是否小于 j"——
   按下载量降序就要写 `a[i].Downloads > a[j].Downloads`。写反了测试立刻红，这正是测试的价值。
5. **`sort.Sort` 是原地排序（in-place）**：不返回新切片，直接改传入切片的底层数组。
   副作用意识：如果这份 `apps` 的底层数组还被别的切片共享，排序会"污染"对方
   （共享底层数组互污问题，后续 tdd/slicelab 会专门练）。
6. **`%+v` 与结构体可比较**：测试里 `apps[i] != want[i]` 能直接比，是因为 `App` 的字段
   全是可比较类型（int），结构体整体就可比较；`%+v` 打印结构体时带字段名
   （`{ID:1 Downloads:1200}`），失败输出一眼可读（[basic/print.go](../../basic/print.go)）。

### 行为 2：sort.Reverse 反向 + ByID 第二套规则

先写测试（骨架如下，**按用例表自己补全**），跑出 `undefined: ByID` 的 RED，
再回 `app.go` 补 `ByID` 的三个方法：

```go
func TestReverseAndByID(t *testing.T) {
	// 子测试 1：sort.Sort(sort.Reverse(ByDownloads(apps))) → 下载量升序
	// 子测试 2：sort.Sort(ByID(apps)) → ID 升序
	// 子测试 3：sort.Sort(sort.Reverse(ByID(apps))) → ID 降序
}
```

| 用例名 | 输入 apps | 操作 | 排序后期望 |
|---|---|---|---|
| Reverse 得下载量升序 | `[{1,1200} {3,500} {2,800}]` | `sort.Sort(sort.Reverse(ByDownloads(apps)))` | `[{3,500} {2,800} {1,1200}]` |
| ByID 升序 | `[{3,500} {1,1200} {2,800}]` | `sort.Sort(ByID(apps))` | `[{1,1200} {2,800} {3,500}]` |
| Reverse(ByID) 得降序 | `[{3,500} {1,1200} {2,800}]` | `sort.Sort(sort.Reverse(ByID(apps)))` | `[{3,500} {2,800} {1,1200}]` |

**这一步用到的知识点：**

1. **`sort.Reverse` 是装饰器（包装器）模式**：`func Reverse(data Interface) Interface`
   返回一个内部类型 `reverse{data}`——它自己的 `Len`/`Swap` 直接转发给被包装者，
   只有 `Less(i, j)` 调的是 `被包装者.Less(j, i)`（下标对调）。**不复制数据、不重写
   算法，只反转比较结果**。这是接口组合的经典手法：新行为 = 旧接口 + 一层包装。
2. **降序的两种写法对比**：`ByDownloads.Less` 里写 `>`（排序规则内建在类型里）vs
   `Less` 写 `<` 再外包 `Reverse`（内建升序 + 外部反转）。结果相同，可读性不同：
   类型名叫 `ByDownloads`、业务场景就是排行榜，降序内建在类型里更贴合语义；`Reverse`
   更适合"偶尔需要反着看一眼"的临时需求。
3. **同一数据、多种排序规则 = 多个命名类型各实现一遍三方法**：`ByDownloads` 和 `ByID`
   的 `Len`/`Swap` 一字不差，只有 `Less` 不同——样板代码开始显得多余了。记住这种
   "重复感"，行为 4 会看到现代写法怎么消灭它。

### 行为 3：sort.Stable 与排序稳定性

骨架（断言语句自己写）：

```go
func TestStableKeepsOriginalOrder(t *testing.T) {
	apps := []App{
		{ID: 1, Downloads: 100},
		{ID: 2, Downloads: 300},
		{ID: 3, Downloads: 100},
	}
	sort.Stable(ByDownloads(apps))
	// 断言：ID 2 排第一；关键断言——ID 1 必须仍在 ID 3 前面
}
```

| 用例名 | 输入 apps | 操作 | 排序后期望 |
|---|---|---|---|
| 同下载量保持原先后顺序 | `[{1,100} {2,300} {3,100}]` | `sort.Stable(ByDownloads(apps))` | `[{2,300} {1,100} {3,100}]`：ID 1 在 ID 3 前 |

**这一步用到的知识点：**

1. **稳定性的定义**：相等元素在排序后**保持原来的相对顺序**，就是稳定排序。注意"相等"
   由 `Less` 定义而不是 `==`：`Less(i,j)` 和 `Less(j,i)` 都返回 false 时，i、j 视为相等。
   对 `ByDownloads` 来说，`{1,100}` 和 `{3,100}` 就是"相等"元素——尽管 ID 不同、`==`
   比较是 false。
2. **`sort.Sort` 为什么不保证稳定**：Go 1.19 起 `sort.Sort` 用 pdqsort（模式击败快速排序），
   为速度牺牲了稳定性；`sort.Stable` 用插入排序 + 归并的混合策略，保证稳定但常数开销更大。
   接口一模一样，差别只在算法属性的承诺上——**选哪个是语义问题，不是口味问题**。
3. **测试纪律：只断言契约承诺的事**。本测试只断言 `Stable` 保证稳定；**不要**加一条
   "用 `sort.Sort` 排完顺序一定乱"的反向断言——"不保证稳定"不等于"保证不稳定"。对小切片
   或近乎有序的输入，pdqsort 内部走插入排序分支，结果可能恰好是稳定的，断言"一定乱"
   就是 flaky test（时红时绿）。测试只能锁定文档承诺的行为。
4. **业务视角的稳定性**：排行榜里两个应用下载量相同，先录入的排前面，靠 `Stable` 能做到。
   但工程上更推荐**显式二级排序键**：`Less` 里写"下载量不同比下载量，相同再比 ID"，这样
   连不稳定的 `sort.Sort` 都能得到唯一确定的顺序——把确定性握在自己手里，而不是依赖算法属性。
   （可作为本练习的扩展用例自己加一条。）

### 行为 4：现代写法对照 —— sort.Slice 与 slices 泛型版

不再定义新类型，直接在测试里写匿名比较函数。骨架（需要新增 `import "slices"` 和
`"cmp"`，断言自己补）：

```go
func TestModernSort(t *testing.T) {
	// 子测试 1：sort.Slice 按下载量降序 + sort.SliceIsSorted 校验
	// 子测试 2：slices.SortFunc + cmp.Compare 按下载量降序
	// 子测试 3：slices.Sort([]int{3, 1, 2}) → [1, 2, 3]
}
```

| 用例名 | 输入 | 操作 | 期望 |
|---|---|---|---|
| sort.Slice 降序 | `[{3,500} {1,1200} {2,800}]` | `sort.Slice(apps, func(i, j int) bool { ... })` | `[{1,1200} {2,800} {3,500}]` |
| 已序校验 | 上一步排好的切片 | `sort.SliceIsSorted(apps, 同一个 less)` | `true` |
| slices.SortFunc 降序 | `[{3,500} {1,1200} {2,800}]` | `slices.SortFunc(apps, func(a, b App) int { return cmp.Compare(b.Downloads, a.Downloads) })` | `[{1,1200} {2,800} {3,500}]` |
| slices.Sort 升序 | `[]int{3, 1, 2}` | `slices.Sort(nums)` | `[1, 2, 3]` |

**这一步用到的知识点：**

1. **`sort.Slice`（Go 1.8）**：不用定义命名类型，传一个匿名 `less func(i, j int) bool`
   就行——行为 2 里的样板代码（`ByID` 的三个方法）被压缩成一个闭包。代价有两个：
   内部用**反射**（`reflect.ValueOf` + `Swapper`）操作任意切片，比接口写法慢；
   闭包参数是**下标 i、j** 不是元素本身，写起来要 `apps[i].Downloads > apps[j].Downloads`，
   且切片换了闭包就废。
2. **`slices` 包（Go 1.21，泛型）是新代码的推荐写法**：`slices.Sort(x []E)` 要求元素
   满足 `cmp.Ordered`（int、string 等内置有序类型，见 [basic/generic.go](../../basic/generic.go)
   的约束一节）；`slices.SortFunc(x, cmp)` 自定义比较。泛型在编译期实例化，**零反射开销**，
   性能与手写 `sort.Interface` 相当。
3. **比较函数的三态约定变了**：`Less` 返回 bool（是否在前），`slices.SortFunc` 的 `cmp`
   返回 **int 三态**——负数 = a 在前，0 = 相等，正数 = b 在前（和 C 的 `qsort` 回调、
   `strings.Compare` 一脉相承）。`cmp.Compare` 就是内置的三态比较器，写降序只要交换
   实参：`cmp.Compare(b.Downloads, a.Downloads)`。
4. **稳定性对应关系**：`sort.Slice` / `slices.Sort` / `slices.SortFunc` 都不保证稳定；
   稳定版分别是 `sort.SliceStable` / `slices.SortStableFunc`（`slices.Sort` 没有稳定版
   对应物，因为内置类型的元素"相等即相同"，稳定与否观测不到差别）。
5. **历史定位一句话**：`sort.Interface`（Go 1）→ `sort.Slice`（Go 1.8，反射换便利）→
   `slices`（Go 1.21，泛型终结样板）。新代码直接用 `slices`；但读老代码、写需要
   `Reverse` 包装的场景，`sort.Interface` 仍是必备知识——本练习四个行为全覆盖，就是为了
   两代代码都读得懂。

---

## 三、知识点总结

### `sort` 包函数速查

| 函数 | 作用 | 稳定？ |
|---|---|---|
| `sort.Sort(data Interface)` | 原地排序 | ❌（pdqsort） |
| `sort.Stable(data Interface)` | 原地排序 | ✅ |
| `sort.Reverse(data Interface) Interface` | 包装出反序视图（不复制数据） | 随被包装者 |
| `sort.IsSorted(data Interface) bool` | 校验是否已有序 | — |
| `sort.Slice(x, less)` / `sort.SliceStable` / `sort.SliceIsSorted` | 免定义类型的反射版 | 仅 `SliceStable` 稳定 |
| `sort.Ints` / `sort.Strings` / `sort.Float64s` | 内置类型快捷排序（见 [basic/slice_sort.go](../../basic/slice_sort.go)） | ❌ |

### `slices` / `cmp`（Go 1.21+）速查

| 函数 | 作用 |
|---|---|
| `slices.Sort(x []E)` | 元素为 `cmp.Ordered` 时的升序排序（不稳定） |
| `slices.SortFunc(x, cmp func(a, b E) int)` | 自定义三态比较排序（不稳定） |
| `slices.SortStableFunc(x, cmp)` | 稳定版 |
| `slices.IsSorted(x)` / `slices.IsSortedFunc(x, cmp)` | 已序校验 |
| `cmp.Compare[T cmp.Ordered](x, y T) int` | 内置三态比较器；交换实参即得降序 |

### 三代写法对比

| | `sort.Interface`（行为 1~3） | `sort.Slice`（行为 4） | `slices.SortFunc`（行为 4） |
|---|---|---|---|
| 引入版本 | Go 1 | Go 1.8 | Go 1.21 |
| 排序规则写在哪 | 命名类型的 `Less` 方法 | 匿名闭包（拿下标） | 匿名函数（拿元素，返回三态 int） |
| 样板量 | 大（每套规则三个方法） | 小 | 小 |
| 实现机制 | 接口动态派发 | 反射 | 泛型编译期实例化 |
| 适用 | 读老代码、需 `Reverse`/复用规则 | 过渡方案 | **新代码首选** |

### 稳定性一句话

稳定与否看**算法承诺**不看类型：相等元素（`Less` 双向都 false）保持原相对顺序 = 稳定；
`sort.Sort` 不承诺稳定，但"不承诺"≠"一定乱"，测试永远只断言承诺过的行为。

### 与书目的对应

- **教程 ch16 前半**（sort 包与自定义排序）：本练习全覆盖
- 教程 ch16 后半（快排/插排等算法手写与性能对比）：留给 tdd/sortbench
- 呼应已学：[basic/interface.go](../../basic/interface.go)（接口作参数→真实落地）、
  [basic/type_func.go](../../basic/type_func.go)（命名类型才能挂方法）、
  [basic/struct_func.go](../../basic/struct_func.go)（值/指针接收者之辨）、
  [basic/generic.go](../../basic/generic.go)（`cmp.Ordered` 约束）

---

## 四、验收标准

```bash
go test ./tdd/sortiface -v       # 全绿
go vet ./tdd/sortiface           # 无警告
go test ./tdd/sortiface -cover   # 六个方法 + 各排序分支都应走到，目标 100%
```

## 五、完成后自查（能口头回答才算过）

1. `sort.Interface` 有哪三个方法？为什么标准库的排序算法不需要知道你的元素类型？
2. `Less(i, j)` 的准确语义是什么？按下载量降序，`Less` 里该写 `>` 还是 `<`？
3. 为什么 `Len`/`Less`/`Swap` 用值接收者就能让排序生效？什么情况下才需要指针接收者？
4. `sort.Sort` 和 `sort.Stable` 的区别是什么？为什么测试不能断言 `sort.Sort`"一定不稳定"？
5. `sort.Reverse` 是怎么实现的？为什么不复制数据就能得到反序？
6. `sort.Slice` 和 `slices.SortFunc` 各用什么机制（反射/泛型）？新代码该选哪个？
7. "下载量相同再按 ID 升序"的二级排序键，`Less` 该怎么写？

全部答清后，回到 [根 README](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
ch16"排序算法"一行的前半（sort 包自定义排序）已补上；整行仍保持 ◐——快排/插排与性能对比
由 tdd/sortbench 完成后才能划成 ✅。
