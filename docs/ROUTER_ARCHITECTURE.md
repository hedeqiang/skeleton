# 🚦 路由架构文档

## 📋 概述

本项目采用分层路由架构设计，按业务模块和版本分类组织路由，提供清晰的结构和良好的可扩展性。

## 🏗️ 目录结构

```
internal/router/
├── router.go              # 主路由入口
├── system/                # 系统级路由
│   └── health.go         # 健康检查路由
└── api/                  # API 路由
    ├── api.go            # API 路由入口
    └── v1/               # v1 版本 API
        ├── v1.go         # v1 路由注册
        ├── user.go       # 用户相关路由
        ├── message.go    # 消息队列路由
        └── scheduler.go  # 调度器路由
```

## 🔗 路由层级

### 1. 主路由 (router.go)
负责：
- 设置 Gin 引擎
- 注册中间件
- 分发到各个子路由模块

```go
func SetupRouter(logger *zap.Logger, handlers *Handlers) *gin.Engine {
    r := gin.New()
    setupMiddleware(r, logger)                    // 中间件
    system.RegisterSystemRoutes(r, logger)       // 系统路由
    api.RegisterAPIRoutes(r, handlers)           // API 路由
    return r
}
```

### 2. 系统路由 (system/)
负责系统级功能：
- `/health` - 健康检查
- `/ready` - 就绪检查  
- `/ping` - 存活检查

### 3. API 路由 (api/)
负责业务 API：
- `/api/v1/*` - v1 版本 API
- 未来可扩展：`/api/v2/*` - v2 版本 API

### 4. V1 API 路由 (api/v1/)
按业务模块分类：
- **用户模块** (`user.go`)
  - `/api/v1/users/*` - 用户 CRUD
  - `/api/v1/auth/*` - 认证相关
- **消息模块** (`message.go`) 
  - `/api/v1/messages/*` - 消息队列
  - `/api/v1/hello/*` - Hello 消息（兼容性）
- **调度器模块** (`scheduler.go`)
  - `/api/v1/scheduler/*` - 计划任务管理

## 📍 路由映射

### 系统路由
| 路径 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/ready` | GET | 就绪检查 |
| `/ping` | GET | 存活检查 |

### 用户路由
| 路径 | 方法 | 描述 |
|------|------|------|
| `/api/v1/users` | POST | 创建用户 |
| `/api/v1/users/:id` | GET | 获取用户信息 |
| `/api/v1/users/:id` | PUT | 更新用户信息 |
| `/api/v1/users/:id` | DELETE | 删除用户 |
| `/api/v1/users` | GET | 获取用户列表 |
| `/api/v1/auth/login` | POST | 用户登录 |

### 消息路由
| 路径 | 方法 | 描述 |
|------|------|------|
| `/api/v1/messages/hello/publish` | POST | 发布 Hello 消息 |
| `/api/v1/hello/publish` | POST | 发布 Hello 消息（兼容性） |

### 调度器路由
| 路径 | 方法 | 描述 |
|------|------|------|
| `/api/v1/scheduler/jobs` | GET | 获取任务列表 |
| `/api/v1/scheduler/start` | POST | 启动调度器 |
| `/api/v1/scheduler/stop` | POST | 停止调度器 |

## 🔧 扩展指南

### 1. 添加新的业务模块

创建新的路由文件：
```go
// internal/router/api/v1/order.go
package v1

import (
    "github.com/gin-gonic/gin"
    handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
)

func RegisterOrderRoutes(group *gin.RouterGroup, orderHandler *handlers.OrderHandler) {
    orders := group.Group("/orders")
    {
        orders.POST("", orderHandler.CreateOrder)
        orders.GET("/:id", orderHandler.GetOrder)
        // ...
    }
}
```

在 `v1.go` 中注册：
```go
// 订单路由
if handlers.OrderHandler != nil {
    RegisterOrderRoutes(v1Group, handlers.OrderHandler)
}
```

### 2. 添加新的 API 版本

创建新版本目录：
```
internal/router/api/v2/
├── v2.go
├── user.go
└── ...
```

在 `api.go` 中注册：
```go
// 注册 v2 版本的 API
v2.RegisterV2Routes(api, handlers)
```

### 3. 添加新的系统路由

在 `system/` 目录下创建新文件：
```go
// internal/router/system/metrics.go
func RegisterMetricsRoutes(router *gin.Engine, logger *zap.Logger) {
    router.GET("/metrics", prometheusHandler())
}
```

在 `health.go` 的 `RegisterSystemRoutes` 中调用。

## 🎯 设计原则

### 1. 单一职责
- 每个文件只负责一个业务模块的路由
- 系统路由与业务路由分离

### 2. 版本隔离
- 不同 API 版本独立目录
- 便于版本管理和向后兼容

### 3. 分层清晰
```
主路由 → 系统路由/API路由 → 版本路由 → 业务模块路由
```

### 4. 易于扩展
- 新增模块只需创建对应文件
- 注册方式统一
- 依赖注入灵活

### 5. 向后兼容
- 保留旧路由路径
- 渐进式迁移

## 🔍 最佳实践

### 1. 路由命名
- 使用 RESTful 风格
- 路径清晰表达资源关系
- 动词使用 HTTP 方法表达

### 2. 分组策略
- 按业务领域分组
- 合理使用中间件
- 避免过深嵌套

### 3. 错误处理
- 统一的错误响应格式
- 适当的 HTTP 状态码
- 详细的错误信息

### 4. 文档维护
- 及时更新路由映射表
- 记录破坏性变更
- 提供使用示例

## 🚀 快速测试

### 测试系统路由
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/ping
```

### 测试 API 路由
```bash
# 用户 API
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com"}'

# 消息 API
curl -X POST http://localhost:8080/api/v1/hello/publish \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello World","sender":"test"}'

# 调度器 API
curl http://localhost:8080/api/v1/scheduler/jobs
```

这种分层路由架构为项目提供了清晰的结构，便于维护和扩展。 