# 调度器系统文档

## 概述

本项目集成了基于 [gocron](https://github.com/go-co-op/gocron) 的任务调度系统，提供灵活的定时任务管理功能。调度器支持两种运行模式：独立服务模式和API集成模式。

## 特性

- 🕒 支持多种调度类型：间隔时间、Cron表达式、每日定时
- 🔧 配置驱动的任务管理
- 🚀 独立服务和API集成双模式
- 🛡️ 优雅启停和错误处理
- 📊 HTTP API控制接口
- 🔌 易于扩展的任务注册机制

## 架构设计

### 核心组件

```
internal/scheduler/
├── scheduler.go        # 调度器服务核心
├── job_registry.go    # 任务注册器
└── jobs/              # 任务实现目录
    └── hello_job.go   # 示例任务
```

### 组件说明

#### 1. SchedulerService (`scheduler.go`)
- 封装 gocron 调度器功能
- 提供任务添加、启动、停止接口
- 集成自定义日志适配器

#### 2. JobRegistry (`job_registry.go`)
- 负责任务注册和生命周期管理
- 实现工厂模式的任务创建
- 支持配置驱动的任务初始化

#### 3. Jobs (`jobs/`)
- 具体任务实现
- 遵循标准接口约定

## 配置说明

### 配置结构

```yaml
scheduler:
  enabled: true  # 是否启用调度器
  jobs:
    - name: "hello_job"           # 任务名称
      type: "duration"            # 调度类型：duration/cron/daily
      schedule: "30s"             # 调度规则
      enabled: true               # 是否启用此任务
    - name: "cleanup_job"
      type: "daily"
      schedule: "02:00"           # 每日02:00执行
      enabled: false              # 默认禁用
```

### 调度类型说明

| 类型 | 说明 | 示例 |
|------|------|------|
| `duration` | 间隔时间执行 | `30s`, `5m`, `1h` |
| `cron` | Cron表达式 | `0 */6 * * *` |
| `daily` | 每日定时 | `02:00`, `14:30` |

## 运行模式

### 1. 独立服务模式

独立运行调度器服务，专注于任务执行：

```bash
# 运行调度器服务
make scheduler-run

# 或直接运行
go run cmd/scheduler/main.go
```

特点：
- 轻量级，仅包含调度功能
- 适合生产环境的任务调度
- 支持优雅关闭

### 2. API集成模式

将调度器集成到API服务中，提供HTTP控制接口：

```bash
# 运行带调度器的API服务
make scheduler-api

# 或直接运行
go run cmd/api/main.go
```

特点：
- 提供REST API控制调度器
- 可通过HTTP接口管理任务
- 适合需要动态控制的场景

## API接口

当以API模式运行时，提供以下HTTP接口：

### 获取任务列表
```http
GET /api/v1/scheduler/jobs
```

响应示例：
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "job-uuid",
      "name": "hello_job",
      "next_run": "2024-01-01T10:00:30Z",
      "last_run": "2024-01-01T10:00:00Z"
    }
  ]
}
```

### 启动调度器
```http
POST /api/v1/scheduler/start
```

### 停止调度器
```http
POST /api/v1/scheduler/stop
```

## 任务开发指南

### 1. 创建新任务

在 `internal/scheduler/jobs/` 目录下创建新的任务文件：

```go
package jobs

import (
    "context"
    "go.uber.org/zap"
)

type MyJob struct {
    logger *zap.Logger
}

func NewMyJob(logger *zap.Logger) *MyJob {
    return &MyJob{
        logger: logger,
    }
}

func (j *MyJob) Execute(ctx context.Context) error {
    j.logger.Info("MyJob is executing")
    
    // 任务逻辑实现
    
    return nil
}

func (j *MyJob) Name() string {
    return "my_job"
}
```

### 2. 注册任务

在 `job_registry.go` 的 `registerDefaultJobs` 方法中添加新任务：

```go
func (r *JobRegistry) registerDefaultJobs() {
    // 现有任务...
    
    r.registeredJobs["my_job"] = func(logger *zap.Logger) Job {
        return NewMyJob(logger)
    }
}
```

### 3. 添加配置

在 `configs/config.dev.yaml` 中添加任务配置：

```yaml
scheduler:
  jobs:
    - name: "my_job"
      type: "cron"
      schedule: "0 */2 * * *"  # 每2小时执行
      enabled: true
```

### 4. 任务接口规范

所有任务必须实现 `Job` 接口：

```go
type Job interface {
    Execute(ctx context.Context) error
    Name() string
}
```

## 部署建议

### 生产环境部署

1. **独立部署调度器服务**
   ```bash
   # 构建
   make scheduler-build
   
   # 部署
   ./bin/scheduler
   ```

2. **配置文件管理**
   - 使用环境变量区分配置
   - 敏感信息使用环境变量注入
   - 任务配置支持热更新

3. **监控和日志**
   - 集成结构化日志
   - 任务执行状态监控
   - 错误告警机制

### Docker 部署

```dockerfile
# Dockerfile示例
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o scheduler cmd/scheduler/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/scheduler .
COPY --from=builder /app/configs ./configs
CMD ["./scheduler"]
```

## 常用命令

```bash
# 开发相关
make scheduler-run           # 运行独立调度器
make scheduler-api           # 运行API模式
make scheduler-build         # 构建调度器服务

# 测试相关
make test-scheduler-api      # 测试API接口
go test ./internal/scheduler/...  # 单元测试

# 构建相关
make build                   # 构建所有服务
make clean                   # 清理构建文件
```

## 故障排查

### 常见问题

1. **任务不执行**
   - 检查配置文件中 `enabled` 字段
   - 验证调度表达式格式
   - 查看日志中的错误信息

2. **调度器启动失败**
   - 检查依赖注入配置
   - 验证配置文件格式
   - 确认端口占用情况

3. **API接口404**
   - 确认路由注册
   - 检查中间件配置
   - 验证URL路径

### 调试技巧

1. **启用详细日志**
   ```yaml
   log:
     level: "debug"
   ```

2. **单独测试任务**
   ```go
   job := jobs.NewHelloJob(logger)
   err := job.Execute(context.Background())
   ```

3. **检查调度器状态**
   ```bash
   curl http://localhost:8080/api/v1/scheduler/jobs
   ```

## 扩展开发

### 添加新的调度类型

在 `job_registry.go` 中扩展 `createJobDefinition` 方法：

```go
case "weekly":
    // 实现周调度逻辑
case "monthly":
    // 实现月调度逻辑
```

### 集成外部服务

任务中可以注入各种服务依赖：

```go
type ServiceJob struct {
    userService *service.UserService
    database    *gorm.DB
    redis       *redis.Client
}
```

### 任务持久化

可以扩展任务状态持久化：

```go
type JobExecution struct {
    ID        uint      `gorm:"primaryKey"`
    JobName   string    `gorm:"index"`
    Status    string    
    StartTime time.Time
    EndTime   time.Time
    Error     string
}
```

## 最佳实践

1. **任务设计原则**
   - 保持任务幂等性
   - 合理设置超时时间
   - 避免长时间阻塞操作

2. **错误处理**
   - 记录详细错误信息
   - 实现重试机制
   - 设置告警通知

3. **性能优化**
   - 避免任务重叠执行
   - 合理分配资源
   - 监控任务执行时间

4. **安全考虑**
   - 任务权限控制
   - 敏感数据保护
   - API访问控制

## 版本历史

- v1.0.0: 基础调度器功能
- v1.1.0: 添加HTTP API支持
- v1.2.0: 支持配置驱动的任务管理

## 参考资料

- [gocron官方文档](https://pkg.go.dev/github.com/go-co-op/gocron/v2)
- [Cron表达式格式](https://en.wikipedia.org/wiki/Cron)
- [Go Context使用指南](https://golang.org/pkg/context/) 