# base - 基础工具库

一个高性能、易用的 Go 基础工具库，提供统一的错误处理和日志系统。

## 项目概述

**base** 是一个 Go 语言基础工具库，专注于解决团队开发中的通用问题。项目包含三个核心模块：

- **errors**: 统一的错误处理接口和错误码体系
- **log**: 高性能结构化日志系统
- **utils**: 通用工具函数集合

### 主要特性

- ✅ **高性能**: 基于 zap 的结构化日志，性能优异
- ✅ **易用性**: 简洁的 API 设计，开箱即用
- ✅ **可扩展**: 支持多种输出方式和配置选项
- ✅ **标准兼容**: 支持标准库 slog 接口
- ✅ **热更新**: 配置文件支持热更新
- ✅ **生产级**: 日志轮转、采样、管道分发等生产级特性

## 技术栈

| 技术栈 | 版本 | 用途 |
|--------|------|------|
| Go | 1.22 | 编程语言 |
| zap | v1.27.0 | 高性能日志库 |
| viper | v1.19.0 | 配置管理 |
| fsnotify | v1.7.0 | 文件监听 |
| protobuf | v1.34.2 | Protobuf 编解码 |
| lumberjack | v2.2.1 | 日志文件轮转 |

## 快速开始

### 安装

```bash
go get github.com/coffeehc/base
```

### 错误处理

```go
import "github.com/coffeehc/base/errors"

// 创建业务错误
err := errors.MessageError("用户不存在")
err := errors.NotFountError("用户ID: 123")

// 创建系统错误
err := errors.SystemError("数据库连接失败")

// 包装错误
originalErr := someFunction()
err := errors.WrappedSystemError(originalErr)

// 判断错误类型
if errors.IsSystemError(err) {
    // 处理系统错误
}

// 获取错误码
code := err.GetCode()
```

### 日志系统

```go
import "github.com/coffeehc/base/log"

// 基本日志
log.Info("用户登录", zap.String("user", "admin"))
log.Error("数据库错误", zap.Error(err))

// 设置日志级别
log.SetLevel("debug")

// 使用 slog 标准库
logger := log.GetSlog()
logger.Info("slog 日志", "key", "value")

// 日志管道
writeChan := make(chan []byte)
id := log.RegisterAccept(writeChan)
// ... 使用 writeChan 发送日志
log.UnRegisterAccept(id)
```

### 工具函数

```go
import "github.com/coffeehc/base/utils"

// 获取本地 IP
ip, err := utils.GetLocalIP()

// 规范化服务地址
addr, err := utils.WarpServiceAddr("localhost:8080")
// 自动补充本地 IP
addr, err := utils.WarpServiceAddr(":8080")
```

## 模块文档

### errors 模块

#### Error 接口

```go
type Error interface {
    error
    GetCode() int64              // 获取错误码
    GetFields(fields ...zap.Field) []zap.Field  // 获取日志字段
    GetFieldsWithCause(fields ...zap.Field) []zap.Field  // 获取包含原因的字段
    FormatRPCError() string      // 格式化为 RPC 错误格式
    Is(Error) bool               // 判断是否为指定错误
    ToError() error              // 转换为 error 接口
}
```

#### 错误码体系

**系统级别错误** (0x10000000 - 0x100FFFF):
- `ErrorSystem`: 系统级别错误
- `ErrorSystemInternal`: 内部错误
- `ErrorSystemDB`: 数据库错误
- `ErrorSystemRedis`: Redis 错误
- `ErrorSystemRPC`: RPC 错误
- `ErrorSystemNet`: 网络错误

**业务级别错误** (0x20000000 - 0x200FFFF):
- `ErrorMessage`: 业务级别错误
- `ErrorMessageNotFount`: 未找到错误

#### 错误构建函数

```go
// 基础错误
err := errors.BuildError(code, message)

// 业务错误
err := errors.MessageError(message)
err := errors.NotFountError(message)

// 系统错误
err := errors.SystemError(message)
err := errors.SystemDBError(message)
err := errors.SystemRedisError(message)
err := errors.SystemRPCError(message)
err := errors.SystemNetError(message)

// 错误包装
err := errors.WrappedError(err)
err := errors.WrappedSystemError(err)
err := errors.WrappedMessageError(err)
```

### log 模块

#### 日志配置

```go
type Config struct {
    Level         string        // 日志级别: debug, info, warn, error, panic, dpanic, fatal
    FileConfig    FileLogConfig // 文件日志配置
    EnableConsole bool          // 是否启用控制台输出
    EnableColor   bool          // 是否启用颜色
    EnableSampler bool          // 是否启用采样
}

type FileLogConfig struct {
    FileName   string // 日志文件路径
    Enable     bool   // 是否启用文件日志
    Maxsize    int    // 单个文件最大大小(MB)
    MaxBackups int    // 保留的备份文件数
    MaxAge     int    // 保留天数
    Compress   bool   // 是否压缩备份
}
```

#### 默认配置

```go
{
    Level:         "info",
    FileConfig: FileLogConfig{
        FileName:   "./logs/service.log",
        Enable:     true,
        Maxsize:    100,
        MaxBackups: 5,
        MaxAge:     7,
        Compress:   true,
    },
    EnableConsole: true,
    EnableColor:   true,
    EnableSampler: false,
}
```

#### 日志级别

- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息
- `panic`: panic 信息
- `dpanic`: 仅在开发环境 panic
- `fatal`: 致命错误，退出程序

#### API 参考

```go
// 获取服务实例
service := log.GetService()

// 获取日志记录器
logger := log.GetLogger()

// 初始化日志系统
log.InitLogger()

// 重置日志字段
log.ResetLogger(fields ...zap.Field)

// 设置日志级别
log.SetLevel(level string)

// 打印日志
log.SendLog(level zapcore.Level, msg string, fields ...zap.Field)

// 注册日志接收通道
id := log.RegisterAccept(writeChan chan<- []byte)

// 注销日志接收通道
log.UnRegisterAccept(id)

// 获取 slog 兼容的日志器
slogLogger := log.GetSlog()

// 加载配置
service.LoadConfig()

// 设置日志级别
service.SetLevel(level string)

// 打印日志到指定 Writer
service.PrintLog(write io.Writer)

// 注册日志管道接收器
id := service.RegisterAccept(logWrite chan<- []byte)

// 注销日志管道接收器
service.UnRegisterAccept(id)
```

### utils 模块

#### 工具函数

```go
// 获取本地 IP 地址
// 通过环境变量 NET_INTERFACE 指定网络接口
func GetLocalIP() (string, error)

// 规范化服务地址
// 自动补充本地 IP，验证 TCP 地址格式
func WarpServiceAddr(addr string) (string, error)
```

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        base 库                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   errors     │  │     log      │  │    utils     │      │
│  │              │  │              │  │              │      │
│  │ • Error      │  │ • Service    │  │ • GetLocalIP │      │
│  │   接口       │  │ • Logger     │  │ • WarpAddr   │      │
│  │ • 错误码     │  │ • Config     │  │              │      │
│  │ • 错误包装   │  │ • slog 适配  │  │              │      │
│  │              │  │ • Protobuf   │  │              │      │
│  │              │  │ • 管道分发   │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 设计理念

1. **单一职责**: 每个模块职责明确，errors 处理错误，log 处理日志，utils 提供工具
2. **接口优先**: 提供清晰的接口定义，便于扩展和测试
3. **性能优先**: 使用对象池、buffer 复用等技术优化性能
4. **易用性**: 简洁的 API 设计，降低使用门槛
5. **标准兼容**: 支持标准库 slog 接口，便于迁移

## 最佳实践

### 错误处理

1. **使用错误码**: 统一管理错误码，便于错误分类和处理
2. **包装错误**: 使用 `WrappedError` 包装原始错误，保留错误链
3. **记录日志**: 在捕获错误时记录日志，包含上下文信息
4. **判断错误**: 使用 `IsXxxError` 判断错误类型，而不是比较错误消息

```go
// ❌ 不推荐
if err.Error() == "用户不存在" {
    // 处理
}

// ✅ 推荐
if errors.IsMessageError(err) {
    // 处理
}
```

### 日志使用

1. **使用结构化日志**: 使用 zap.Field 传递结构化数据
2. **合理设置级别**: 生产环境使用 info/warn，开发环境使用 debug
3. **避免敏感信息**: 不要记录密码、token 等敏感信息
4. **使用日志管道**: 集中收集和转发日志

```go
// ❌ 不推荐
log.Info("用户登录", "password", "123456")

// ✅ 推荐
log.Info("用户登录", zap.String("user", "admin"), zap.String("ip", ip))
```

### 配置管理

1. **使用配置文件**: 通过 viper 管理配置，支持热更新
2. **合理设置轮转**: 根据业务量设置合适的日志大小和保留天数
3. **启用采样**: 高频日志场景启用采样减少 I/O

## 常见问题

### Q: 如何自定义日志格式？

A: 当前使用 zap 默认格式，如需自定义，可以扩展 `impl.go` 中的 `NewCore` 方法。

### Q: 日志文件如何配置？

A: 通过 `FileLogConfig` 配置日志文件路径、大小、备份等参数。使用 lumberjack 实现日志轮转。

### Q: 如何添加新的错误码？

A: 在 `errorcode.go` 中添加新的错误码常量，遵循现有命名规范。

### Q: 日志管道如何使用？

A: 创建 `chan []byte`，调用 `RegisterAccept` 注册，日志会自动写入所有注册的通道。

## 贡献指南

### 开发环境

```bash
# 克隆项目
git clone https://github.com/coffeehc/base.git
cd base

# 安装依赖
go mod download

# 运行测试
go test ./...

# 代码检查
go vet ./...
gofmt -s -w .
```

### 代码规范

- 遵循 Go 官方代码规范
- 函数和变量使用驼峰命名
- 添加必要的注释
- 保持函数简短，单一职责

### 提交规范

```
feat: 添加新功能
fix: 修复 bug
docs: 更新文档
refactor: 重构代码
test: 添加测试
chore: 构建/工具变更
```

### 测试要求

- 每个新功能必须添加单元测试
- 测试覆盖率不低于 80%
- 重要功能需要集成测试

## 许可证

Apache License 2.0 - 详见 [LICENSE](LICENSE) 文件

## 联系方式

- 项目地址: https://github.com/coffeehc/base
- 问题反馈: https://github.com/coffeehc/base/issues

## 更新日志

### v1.0.0 (当前版本)

- 初始版本发布
- 提供 errors、log、utils 三个核心模块
- 支持高性能结构化日志
- 支持错误码体系和错误包装
- 支持日志热更新和管道分发
