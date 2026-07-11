# basic-cases-go

记录 Go 基础语法、并发题，以及常用框架的最小示例。每个示例尽量保持独立，方便直接运行或改着练习。

## 目录

- `syntax/`：基础类型、流程控制、函数、集合、并发原语、定时任务和类型系统。
- `algorithm/`：协程交替打印、按序打印、生产消费和固定大小工作池。
- `examples/gin/`：Gin 路由与中间件。
- `examples/gorm/`：GORM + SQLite 的增删改查。

## 使用

```bash
go test ./...
go run ./algorithm/alternate-print
go run ./examples/gin
go run ./examples/gorm
```

GORM 示例会在当前目录创建 `test.db`。它只是本地练习数据，已经被 `.gitignore` 忽略。
