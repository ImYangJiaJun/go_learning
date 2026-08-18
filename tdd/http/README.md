# HttpLab TDD —— 内存用户仓库的 REST 服务

目标：通过 5 个行为切片，把教程 ch08 的应用层理论全部落地成代码——`net/http` 的
Handler 体系、`httptest` 的单元/集成两种测试层级、状态码语义、Go 1.22 路由写法、
`slog` 结构化日志。存储用 `map + sync.RWMutex`，是 [tdd/syncmap](../syncmap/README.md)
练习里 RWMutex 对照组的实战上岗。

> 本任务是**机制学习型**练习：接口契约已固定，不要花时间在 API 设计上。
> 用法：第一节看需求；第二节边做边学——每个行为下面附有这一步要用到的知识点讲解；
> 第三节是知识点总结，做完后对照自查。

---

## 一、需求规格

### 这个包要做什么

**没有 `main` 函数。** 本练习的产出物不是可执行程序，而是一个被测试验证的包——
`go test ./tdd/http` 就是它的运行方式，验收者是测试，不是人。

> ⚠️ **包名是 `httplab`，不是 `http`**：本练习要大量导入标准库 `net/http`，若包名
> 也叫 `http`，文件里 `http.Handler` 这类引用就会与本包名纠缠不清（本包名遮蔽标准库），
> 工具链和读者都分不清。目录名仍然是 `tdd/http`（`go test ./tdd/http` 按目录寻址，
> 不受包名影响），只有 package 声明写作 `httplab`。同理，将来 tdd/context 练习的包名
> 会用 `contextlab`。

这个包对外只提供一个能力：**内存用户仓库的 REST 服务**，四条行为：

1. **查用户**：`GET /users/{id}`——存在 → 200 + JSON 响应体；不存在 → 404
2. **建用户**：`POST /users`——合法 JSON → 201，随后能查到；非法 JSON → 400
3. **没实现的方法**（如 `DELETE /users/{id}`）→ 405
4. **每个请求**打一行 `slog` 结构化日志（method、path、status）

### 文件计划（共 4 个文件，分四次建）

| 文件 | 里面写什么 | 什么时候建 |
|---|---|---|
| `server_test.go` | 行为 1、2、3、5 的全部测试（单元级，Recorder） | **第 1 个建** |
| `server.go` | `User`、`NewServer`、路由与 handler、`map + sync.RWMutex` 存储 | 测试编译报错时 |
| `integration_test.go` | 行为 4 的集成测试（`httptest.NewServer` 起真实端口） | 行为 4 开始时 |
| `logging.go` | 行为 5 的日志中间件 | 行为 5 测试变红时 |

### 接口契约（固定，按此实现，名字不要改）

```go
package httplab

// User 是仓库里存的用户；json tag 决定 JSON 体里的字段名
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// NewServer 返回一个配好全部路由的 http.Handler——只装配，不监听端口。
// 测试可以直接调它的 ServeHTTP（单元级），也可以交给 httptest.NewServer
// 起真实端口（集成级）；将来真要上线，http.ListenAndServe 喂的也是它。
//
// 存储：map[int]User + sync.RWMutex（读锁查、写锁增），
// 即 syncmap 练习中 "RWMutex+map" 对照组的实战应用。
//
// 路由（Go 1.22 的 method+pattern 写法）与语义：
//
//	GET  /users/{id}   存在 → 200 + JSON 响应体，Content-Type: application/json
//	                   不存在 → 404
//	POST /users        合法 JSON → 201，随后 GET /users/{id} 能查到该用户
//	                   非法 JSON → 400
//	其他方法（如 DELETE /users/{id}）→ 405（由路由自动产生，行为 3 锁定它）
//
// 每个请求打一行 slog 结构化日志（method、path、status）；此条到行为 5 才有测试验收。
func NewServer() http.Handler
```

### 第一步：手把手起步（行为 1 的测试直接给你当模板）

1. 在 `tdd/http/` 下新建 `server_test.go`，写入：

```go
package httplab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// postUser 是测试的"布置"助手：通过 POST 接口预存一个用户。
// 在行为 2 之前，POST 本身还没有测试——这里只把它当准备工作用。
func postUser(t *testing.T, server http.Handler, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
}

func TestGetUser(t *testing.T) {
	server := NewServer()
	postUser(t, server, `{"id":1,"name":"Tom","age":20}`)

	t.Run("存在的用户返回200和JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("期望状态码 200，得到 %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("期望 Content-Type 是 application/json，得到 %q", ct)
		}
		var got User
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("响应体不是合法 JSON：%v", err)
		}
		if got.ID != 1 || got.Name != "Tom" || got.Age != 20 {
			t.Errorf("期望 Tom（id=1, age=20），得到 %+v", got)
		}
	})

	t.Run("不存在的用户返回404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("期望状态码 404，得到 %d", rec.Code)
		}
	})
}
```

2. 运行 `go test ./tdd/http` → **编译失败**：`undefined: NewServer`（还有 `undefined: User`）。
   这就是 RED —— 测试描述了你想要但还不存在的代码。
3. 新建 `server.go`，照契约写出让测试通过的**最少代码**。提示：`http.NewServeMux`、
   `mux.HandleFunc("GET /users/{id}", h)`、`r.PathValue("id")` 配 `strconv.Atoi`、
   `json.NewEncoder(w).Encode(u)`。注意：POST 此刻只是 arrange 工具，存进 map 即可，
   它的状态码对不对还没有人管——行为 2 会来收拾它。
4. 再跑 `go test ./tdd/http` → 全绿，行为 1 完成。

---

## 二、任务单（边做边学）

每个行为 = 一轮完整的 RED → GREEN → REFACTOR，**先把测试写出来再实现**。

### 行为 1：按 id 查用户（测试代码已在第一节给出）

| 用例名 | 输入 | 期望 |
|---|---|---|
| 存在的用户 | 先 POST `{"id":1,"name":"Tom","age":20}`，再 `GET /users/1` | ① 状态码 200<br>② `Content-Type` 是 `application/json`<br>③ body 能解出 `User{ID:1, Name:"Tom", Age:20}` |
| 不存在的用户 | `GET /users/99` | 状态码 404 |

**这一步用到的知识点：**

1. **`http.Handler`：一个方法的接口撑起整个 net/http**。签名就是
   `ServeHTTP(w http.ResponseWriter, r *http.Request)`，你的服务、路由器、中间件全是它。
   `NewServer` 返回 Handler 而不是"启动服务"，测试拿到接口值后直接
   `server.ServeHTTP(rec, req)`——一次普通的接口方法调用，没有网络、没有端口。
   这正是 [basic/interface.go](../../basic/interface.go) 学的"接口作参数与返回值"的实战：
   依赖接口而非具体实现，测试才能注入假的 `ResponseWriter`。
2. **`http.HandlerFunc`：函数变接口的适配器**。它是一个函数类型
   `type HandlerFunc func(ResponseWriter, *Request)`，但给自己定义了 `ServeHTTP` 方法
   （方法体就是调用自己），于是普通函数一转型就满足了 Handler。
   `mux.HandleFunc("GET /users/{id}", h)` 内部就是 `mux.Handle(..., HandlerFunc(h))`。
   这是标准库最经典的"函数类型 + 方法"技巧，对照 [basic/type_func.go](../../basic/type_func.go)。
3. **`httptest.NewRecorder`：内存里的假 ResponseWriter**。真实服务器里 ResponseWriter
   背后是一条 TCP 连接，写给它就是写给客户端；`ResponseRecorder` 背后是 `bytes.Buffer`，
   handler 写下的状态码、响应头、响应体全被记录成可断言的字段：`Code`、`Header()`、`Body`。
   易错点：handler 若一次都没写，`rec.Code` 是 **0** 而不是 200（真实服务器会隐式发 200，
   Recorder 不会替你补）；`rec.Result()` 会把 0 归一成 200。
4. **状态码即语义**：永远写 `http.StatusOK` 这类常量，不写裸数字。教程 ch08 的状态码族
   理论在这里落地：2xx 成功（200 有实体、201 创建了资源）、4xx 是客户端的错（400 请求本身
   坏、404 资源不存在、405 方法不支持）、5xx 是服务端的错。断言状态码，就是断言"服务把
   错误归对了类"。
5. **Go 1.22 的路由写法**：`mux.HandleFunc("GET /users/{id}", h)`——模式 = 方法 + 空格 +
   路径；`{id}` 是路径通配符，handler 里用 `r.PathValue("id")` 取出（是字符串，要
   `strconv.Atoi`）。老写法只能在 handler 里手动切 `r.URL.Path`、手动 `switch r.Method`；
   1.22 起这两件事 ServeMux 全包，而且"路径匹配但方法不匹配"会**自动回 405**——行为 3
   会白捡这个行为。
6. **`t.Helper()`**：在测试助手函数第一行调用，断言失败时 go test 报告的行号指向
   **调用处**而不是助手内部。没有它，所有失败都报在 `postUser` 里，排查时要多跳一层。

### 行为 2：POST 创建用户，且创建后真的能查到

| 用例名 | 输入 | 期望 |
|---|---|---|
| 合法 JSON 创建成功 | POST /users，body `{"id":2,"name":"Jerry","age":18}` | ① 状态码 201<br>② 随后 GET /users/2 → 200 且解出 `User{2, "Jerry", 18}`（同一个 server，数据真的存下了） |
| 请求体是非法 JSON | POST /users，body `{"id":` | 状态码 400 |
| 请求体为空（可选） | POST /users，body 传 nil | 状态码 400（Decode 返回 `io.EOF`，同样是 error） |

**这一步用到的知识点：**

1. **这次的 RED 是断言失败，不是编译失败**：行为 1 的最小实现里 POST 的状态码没人约束过
   （多半是隐式 200），新测试一跑，期望 201 得到 200——红。回忆
   [tdd/testbasic](../testbasic/README.md) 行为 1：测试失败有两种形态，现在两种你都见过了。
2. **`w.WriteHeader` 只能调一次，且必须先于写 body**：`w.Write` 会隐式补一个
   `WriteHeader(200)`，之后再调 WriteHeader 无效（运行时警告 superfluous WriteHeader call）。
   正确顺序：Set 响应头 → `WriteHeader(http.StatusCreated)` → 写 body。
3. **Decode 的 error 第一次必须处理**：`json.NewDecoder(r.Body).Decode(&u)` 遇到非法 JSON
   返回 error——这就是 400 的实现来源：检查 error，翻译成状态码，正是
   [tdd/errhandling](../errhandling/README.md) 的"error 是值"在处理流程中的应用。
   顺带复习：Decode 必须传指针（[basic/struct_json.go](../../basic/struct_json.go)）。
4. **map + sync.RWMutex**：读路径 `RLock/RUnlock`，写路径 `Lock/Unlock`。为什么现在就要锁？
   HTTP 服务器**每个请求一个 goroutine**——行为 4 起真服务后，两个请求并发读写 map 会直接
   panic（[basic/goroutine_lock.go](../../basic/goroutine_lock.go) 演示过并发写 map 崩溃）。
   RWMutex 读共享写独占，适合"读多写少"——这正是 [tdd/syncmap](../syncmap/README.md)
   练习里 RWMutex+map 对照组的实战上岗。
5. **201 的完整语义（可选加分）**：教程 ch08 说 201 宜带 `Location: /users/2` 响应头指明
   新资源地址。契约没要求，加上只要一行，体会"状态码 + 头部"协同表达语义。

### 行为 3：没实现的方法返回 405

| 用例名 | 输入 | 期望 |
|---|---|---|
| DELETE 用户 | `DELETE /users/1` | 状态码 405；可选：响应头 `Allow` 包含 `GET` |
| PUT 用户（可选） | `PUT /users/1` | 状态码 405 |

**这一步用到的知识点：**

1. **这个测试可能一写就是绿的**——只要行为 1 按契约用了 method+pattern 路由，ServeMux
   在"路径匹配、方法不匹配"时会自动回 405 并带上 `Allow: GET` 响应头。一行实现没写测试
   就过了，RED 去哪了？这是本练习的重要一课：**测试的另一个角色是锁定行为、防回归**——
   它守护的是"明天有人把路由改回老写法时，405 不丢"。
2. **405 vs 404 的语义分野**（教程 ch08）：404 说"这个路径下没有这个东西"；405 说"路径
   认识，但这个方法不行"。`DELETE /users/1` 返回 405 意味着"删除用户"是个没被实现但语义
   合理的操作——将来要实现它，注册 `"DELETE /users/{id}"` 即可，这个测试就是先行坐标。
3. **如果没绿**：说明路由写成了老风格（只匹配路径，方法没区分）。趁机把 HandleFunc 的
   第一个参数改成 `"GET /users/{id}"` 形式——有测试保护，重构不慌。

### 行为 4：集成级对照——httptest.NewServer 起真实服务

| 用例名 | 操作 | 期望 |
|---|---|---|
| 真实端口全流程 | `httptest.NewServer(NewServer())` 起服务；`http.Post` 创建 id=3 的用户；再 `http.Get(ts.URL + "/users/3")` | POST 返回 201；GET 返回 200 且 body 解出该用户 |

测试骨架（在 `integration_test.go` 里补齐）：

```go
func TestIntegration(t *testing.T) {
	ts := httptest.NewServer(NewServer())
	defer ts.Close()
	// 用 http.Post / http.Get 对 ts.URL 发真实请求，断言同用例表
}
```

**这一步用到的知识点：**

1. **Recorder 与 NewServer 是两种测试层级**：行为 1~3 是**单元级**——`ServeHTTP` 是进程内
   函数调用，没有网络、没有端口、没有并发；本行为是**集成级**——`httptest.NewServer` 真的在
   `127.0.0.1` 的随机端口监听，请求真的走 TCP + HTTP 协议栈，服务器真的一个请求一个
   goroutine。同一份 `NewServer()` 两种喂法——这就是契约"返回 Handler 而不是启动服务"的
   设计红利。
2. **取舍：速度 vs 真实度**。单元级快、毫秒级跑几百个、能精确构造边界输入；集成级慢，但
   覆盖"JSON 序列化 → 网络传输 → 路由分发 → 并发安全"全链路，能抓到单元级抓不到的 bug
   （比如忘了加锁）。工程实践是金字塔：大量单元级 + 少量集成级冒烟。
3. **三个使用细节**：`ts.URL` 是形如 `http://127.0.0.1:52341` 的基地址；`defer ts.Close()`
   释放端口；客户端侧 `resp.Body` 必须 `defer resp.Body.Close()`——HTTP 客户端的连接复用
   依赖 body 被读完并关闭，忘关是真实服务里最常见的资源泄漏之一。
4. **`-race` 在这里才真正上岗**：单元级测试是单 goroutine 顺序调用，忘了给 map 加锁也不会
   暴露；集成级是真并发服务器，请求交错执行，数据竞争立刻现形。可选加分：用
   `sync.WaitGroup` 并发发 50 个 POST 再跑 `-race`，亲眼见锁的价值。

### 行为 5：每个请求打一行 slog 结构化日志

| 用例名 | 操作 | 期望 |
|---|---|---|
| 正常请求被记录 | 把默认 logger 换成写 `bytes.Buffer` 的 TextHandler，GET /users/1 | 日志一行，含 `method=GET`、`path=/users/1`、`status=200` |
| 404 也被记录 | GET /users/99 | 日志含 `status=404`（中间件看到的是真实状态码） |

测试骨架（在 `server_test.go` 里补齐）：

```go
func TestRequestLog(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })
	// 发请求，然后用 strings.Contains(buf.String(), "method=GET") 等断言
}
```

**这一步用到的知识点：**

1. **slog 是结构化日志**：`slog.Info("request", "method", r.Method, "path", r.URL.Path,
   "status", status)`——一条消息 + 若干键值对。对比 `fmt.Printf` 的纯文本：键值对能被日志
   系统直接索引，"查所有 status=500 的请求"是一次过滤而不是写正则。TextHandler 输出 `k=v`
   适合开发；`slog.NewJSONHandler` 输出 JSON 适合生产采集。
2. **中间件模式：`func(http.Handler) http.Handler`**。日志、鉴权、限流这类"每个请求都要做"
   的横切逻辑，都写成包一层 Handler 的函数：进去前记时间、调内层、出来后记状态。这是
   Handler 接口组合性的体现，也是 Go Web 生态的标准扩展点（对照
   [basic/func_type.go](../../basic/func_type.go) 的函数作参数与返回值）。
3. **捕获状态码要包装 ResponseWriter**：`http.ResponseWriter` 接口只有写的方法，没有
   "读出刚才的状态码"。标准手法是自定义 struct **嵌入** `http.ResponseWriter`、重写
   `WriteHeader` 记下 code 再透传；默认状态记 200（handler 只写 body 不调 WriteHeader 时
   就是 200）。嵌入 + 方法重写正是 [basic/struct_more.go](../../basic/struct_more.go) 的手法。
4. **`slog.SetDefault` 改的是全局状态**：测试里换成写 buffer 的 logger 才能断言输出。两条
   纪律：用 `t.Cleanup` 恢复原 logger；这个测试**不能加 `t.Parallel()`**（全局替换会互相踩）。
   生产代码更干净的做法是把 `*slog.Logger` 注入依赖它的组件——本练习为了不破坏
   `NewServer()` 的固定契约选择了全局替换，体会这两种取舍。
5. **Go 1.26 的 `slog.NewMultiHandler`**：一个 Handler 扇出到多个 Handler——文本打到
   stdout 给人看、JSON 同时写文件给采集系统，一次日志两路输出，不必自己写包装。知道有它
   即可，本练习用不上。

---

## 三、知识点总结

### Handler 体系速查

| 概念 | 是什么 | 本练习在哪用到 |
|---|---|---|
| `http.Handler` | 只有 `ServeHTTP(w, r)` 一个方法的接口，net/http 的地基 | `NewServer` 的返回类型 |
| `http.HandlerFunc` | 函数类型适配器，让普通函数满足 Handler | `mux.HandleFunc` 内部 |
| `http.ServeMux` | 路由器：按 method+pattern 分发，方法不匹配自动 405 | `NewServer` 内部 |
| 中间件 | `func(Handler) Handler`，包装出日志/鉴权等横切逻辑 | 行为 5 |
| `http.ListenAndServe` | 真上线才用的监听入口 | 本练习不需要——测试即验收 |

### httptest 两种测试层级速查

| | 单元级：`NewRecorder` + `NewRequest` | 集成级：`httptest.NewServer` |
|---|---|---|
| 调用方式 | 直接 `server.ServeHTTP(rec, req)` | `http.Get(ts.URL + ...)` 走真实网络 |
| 网络 / 端口 / 并发 | 无 / 无 / 单 goroutine | TCP / 随机端口 / 每请求一个 goroutine |
| 速度 | 极快，可跑几百个 | 较慢，只做冒烟 |
| 能抓到的 bug | 逻辑、状态码、序列化 | 加上：并发竞争、协议层问题 |
| 断言对象 | `rec.Code` / `rec.Header()` / `rec.Body` | `resp.StatusCode` / `resp.Header` / `resp.Body`（必须 Close） |

### 状态码族速查（呼应教程 ch08）

| 码 | 语义 | 本练习场景 |
|---|---|---|
| 200 OK | 成功且有实体 | GET 命中 |
| 201 Created | 成功且创建了资源（宜带 Location 头） | POST 成功 |
| 400 Bad Request | 请求本身坏（如 JSON 解析失败） | 非法 body |
| 404 Not Found | 路径认识，但资源不存在 | id 不存在 |
| 405 Method Not Allowed | 路径认识，但方法不支持（带 Allow 头） | DELETE /users/1 |
| 500 Internal Server Error | 服务端自己出错 | 未用到 |

### 与书目的对应

- 教程 ch08 应用层：HTTP 报文结构（请求行 / 头 / 体）、状态码族、GET vs POST 语义——
  本练习全部落地成代码；ch08 的传输层（TCP/UDP）留给 [tdd/tcpudp](../tcpudp/README.md)
- Effective Go·16 Web 服务器：Handler 接口设计与 httptest 的用法
- [basic/struct_json.go](../../basic/struct_json.go) 的 JSON 编解码、
  [basic/goroutine_lock.go](../../basic/goroutine_lock.go) 的 RWMutex 在此实战复用；
  [tdd/syncmap](../syncmap/README.md) 的 RWMutex+map 对照组在此上岗
- learn-go-with-tests 的 "Build an application" 篇（HTTP server 迭代 / JSON 路由）
  与本练习同源，可顺读深化

---

## 四、验收标准

```bash
go test ./tdd/http -v          # 全绿
go vet ./tdd/http              # 无警告
go test ./tdd/http -race       # 行为 4 起真服务后是真实并发，必须无数据竞争
go test ./tdd/http -cover      # 核心逻辑 ≥80%
```

## 五、完成后自查（能口头回答才算过）

1. `http.Handler` 接口的签名是什么？`HandlerFunc` 凭什么能把普通函数变成 Handler？
2. `ResponseRecorder` 记录了哪三样东西？为什么不启动服务器就能测试 handler？
3. handler 一次都没写时 `rec.Code` 是几？和真实服务器的行为差在哪？
4. 200 / 201 / 400 / 404 / 405 各自什么语义？`DELETE /users/1` 回 405 而不是 404，说明了什么？
5. `mux.HandleFunc("GET /users/{id}", h)` 相比老写法多送了你什么？路径参数怎么取？
6. Recorder 与 NewServer 两种测试层级的区别和取舍？为什么 `-race` 要有 NewServer 才真正有意义？
7. `slog.Info("request", "method", m)` 比 `fmt.Printf` 强在哪？Go 1.26 的 `slog.NewMultiHandler` 解决什么问题？

全部答清后，回到 [根 README 遗漏清单](../../README.md#三对照for-learning-go-tutorial的覆盖检查)，
把 ch08"通信协议解析"从 ❌ 改成 ◐（应用层 HTTP 已落地；传输层 TCP/UDP 待
[tdd/tcpudp](../tcpudp/README.md) 补上）。
