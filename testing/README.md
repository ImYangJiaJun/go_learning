# Go Testing 学习

这一目录专门学习 Go 标准库 `testing`，不改动 `basic/` 中已有的知识点实验。

## 学习顺序

1. 基础单元测试：`_test.go`、`TestXxx`、`testing.T`
2. 表驱动测试：同一行为覆盖多组输入
3. 子测试：`t.Run`
4. 错误与边界条件
5. Benchmark：`BenchmarkXxx`
6. Fuzzing：`FuzzXxx`
7. Coverage：`go test -cover ./...`

## 核心循环

```text
写测试 → go test → FAIL → 最小实现 → PASS → 重构
```

`testing/` 负责学习测试工具本身；真正的 TDD 练习放在 `tdd/`。
