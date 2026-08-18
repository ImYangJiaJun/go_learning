# TDD 实战

这里不追求覆盖所有 Go 语法，而是通过小需求训练 TDD 的完整循环。

## RED → GREEN → REFACTOR

每个练习都遵循：

1. 先写一个描述行为的测试。
2. 运行 `go test`，确认测试因为正确原因失败（RED）。
3. 写满足当前测试的最少代码（GREEN）。
4. 重构实现和测试，同时保持测试通过（REFACTOR）。
5. 再增加下一个行为，重复循环。

## 练习顺序

- `hello/`：最小 TDD 循环
- `integers/`：函数、条件和边界
- `wallet/`：struct、method、error
- `dictionary/`：map、error、interface
- `clock/`：dependency injection 与时间测试
- `concurrency/`：goroutine、channel、select
- `http/`：HTTP handler、JSON 和集成测试

## 与 `basic/` 的关系

`basic/` 保留为 Go 语言知识点实验区。不要为了 TDD 重写已有文件；当一个知识点已经掌握后，在 `tdd/` 中用它解决一个行为驱动的小问题。
