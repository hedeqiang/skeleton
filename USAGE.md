# 🚀 Go Skeleton Framework Usage Guide

## 📋 概述

Go Skeleton 是一个基于 Go 语言的简洁 Web 应用框架，遵循 Clean Architecture 模式，使用 Gin 框架构建。框架设计遵循 Go 的 "Less is More" 哲学，避免过度抽象。

## 🏗️ 架构特点

- **简洁架构**: Handler → Service → Repository → Database
- **依赖注入**: 使用 Google Wire 进行编译时依赖注入
- **统一错误处理**: 标准化的错误类型和 HTTP 状态映射
- **配置管理**: 基于 Viper 的灵活配置系统
- **数据库支持**: GORM 多数据源支持
- **中间件**: 请求 ID、CORS、日志、恢复等常用中间件

## 🚀 快速开始

### 1. 项目结构

```
cmd/
├── api/          # HTTP 服务入口点
├── consumer/     # 消息队列消费者
└── scheduler/    # 任务调度器

internal/
├── app/          # 应用程序容器
├── config/       # 配置管理
├── handler/      # HTTP 处理器
├── service/      # 业务逻辑层
├── repository/   # 数据访问层
├── model/        # 数据模型
├── router/       # 路由定义
├── middleware/   # 中间件
├── messaging/    # 消息队列处理
├── scheduler/    # 任务调度
└── wire/         # 依赖注入配置

pkg/              # 公共包
├── database/     # 数据库连接管理
├── redis/        # Redis 客户端
├── mq/           # RabbitMQ 工具
├── logger/       # 日志工具
├── errors/       # 错误处理
├── validator/    # 输入验证
└── idgen/        # ID 生成器
```

### 2. 创建应用入口点

```go
// cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/hedeqiang/skeleton/internal/wire"
    "go.uber.org/zap"
)

func main() {
    // 使用 Wire 创建应用实例
    application, err := wire.InitializeApplication()
    if err != nil {
        log.Fatalf("Failed to create application: %v", err)
    }

    // 创建信号处理通道
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // 启动应用
    go func() {
        if err := application.Run(); err != nil {
            application.Logger().Error("Application failed to run", zap.Error(err))
        }
    }()

    // 等待退出信号
    sig := <-quit
    application.Logger().Info("Received signal, shutting down...", zap.String("signal", sig.String()))

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := application.Stop(ctx); err != nil {
        application.Logger().Error("Error during application shutdown", zap.Error(err))
    }

    application.Logger().Info("Application shut down gracefully")
}
```

### 3. 配置管理

```go
// configs/config.dev.yaml
app:
  host: "0.0.0.0"
  port: 8080
  env: "development"

databases:
  default:
    type: "mysql"
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: "1h"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"

jwt:
  secret: "your-secret-key"
  expires_in: "24h"
```

### 4. 数据模型

```go
// internal/model/user.go
package model

import "time"

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex;not null"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null"`
    Password  string    `json:"-" gorm:"not null"`
    Status    int       `json:"status" gorm:"default:1"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
    Username string  `json:"username,omitempty"`
    Email    string  `json:"email,omitempty"`
    Status   *int    `json:"status,omitempty"`
}

type UserResponse struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 5. 数据仓储

```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "github.com/hedeqiang/skeleton/internal/model"
    "gorm.io/gorm"
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id uint) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
    *BaseRepository
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    return r.BaseRepository.Create(ctx, user)
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    err := r.BaseRepository.FindByID(ctx, &user, id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    err := r.BaseRepository.FindOne(ctx, &user, "username = ?", username)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
    return r.BaseRepository.Exists(ctx, &model.User{}, "username = ?", username)
}

// ... 其他方法实现
```

### 6. 业务服务

```go
// internal/service/user_service.go
package service

import (
    "context"
    "github.com/hedeqiang/skeleton/internal/model"
    "github.com/hedeqiang/skeleton/internal/repository"
    "github.com/hedeqiang/skeleton/pkg/errors"
    stdErrors "errors"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type UserService interface {
    CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
    GetUser(ctx context.Context, id uint) (*model.UserResponse, error)
    UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
    DeleteUser(ctx context.Context, id uint) error
    ListUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error)
    Login(ctx context.Context, username, password string) (*model.UserResponse, error)
}

type userService struct {
    userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
    return &userService{
        userRepo: userRepo,
    }
}

func (s *userService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
    // 检查用户名是否已存在
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to check username")
    }
    if exists {
        return nil, errors.ErrUserExists
    }

    // 检查邮箱是否已存在
    exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to check email")
    }
    if exists {
        return nil, errors.ErrUserExists
    }

    // 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to hash password")
    }

    // 创建用户
    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Status:   1,
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to create user")
    }

    return s.toUserResponse(user), nil
}

// ... 其他方法实现
```

### 7. HTTP 处理器

```go
// internal/handler/v1/user_handler.go
package v1

import (
    "github.com/gin-gonic/gin"
    "github.com/hedeqiang/skeleton/internal/model"
    "github.com/hedeqiang/skeleton/internal/service"
    "github.com/hedeqiang/skeleton/pkg/response"
    "net/http"
)

type UserHandler struct {
    userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
    return &UserHandler{
        userService: userService,
    }
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "用户信息"
// @Success 201 {object} response.Response{data=model.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req model.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request parameters", err)
        return
    }

    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to create user", err)
        return
    }

    response.Success(c, http.StatusCreated, "User created successfully", user)
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 根据ID获取用户信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=model.UserResponse}
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    // 转换ID类型并调用服务
    // ...
}

// ... 其他处理器方法
```

### 8. 路由配置

```go
// internal/router/api/v1/user.go
package v1

import (
    "github.com/gin-gonic/gin"
    "github.com/hedeqiang/skeleton/internal/handler/v1"
)

func SetupUserRoutes(router *gin.RouterGroup, userHandler *handler.UserHandler) {
    users := router.Group("/users")
    {
        users.POST("", userHandler.CreateUser)
        users.GET("/:id", userHandler.GetUser)
        users.PUT("/:id", userHandler.UpdateUser)
        users.DELETE("/:id", userHandler.DeleteUser)
        users.GET("", userHandler.ListUsers)
    }
}
```

### 9. 依赖注入配置

```go
// internal/wire/providers.go
package wire

import (
    "github.com/google/wire"
    "github.com/hedeqiang/skeleton/internal/app"
    "github.com/hedeqiang/skeleton/internal/config"
    "github.com/hedeqiang/skeleton/internal/handler/v1"
    "github.com/hedeqiang/skeleton/internal/repository"
    "github.com/hedeqiang/skeleton/internal/service"
    // ... 其他导入
)

var RepositorySet = wire.NewSet(
    repository.NewBaseRepository,
    repository.NewUserRepository,
    // ... 其他仓储
)

var ServiceSet = wire.NewSet(
    service.NewUserService,
    // ... 其他服务
)

var HandlerSet = wire.NewSet(
    v1.NewUserHandler,
    // ... 其他处理器
)

var ApplicationSet = wire.NewSet(
    app.NewApp,
    RepositorySet,
    ServiceSet,
    HandlerSet,
    // ... 其他依赖
)
```

## 🔧 错误处理

框架提供了统一的错误处理系统：

```go
// 错误类型
errors.ErrValidation      // 参数验证错误
errors.ErrNotFound        // 资源不存在
errors.ErrUnauthorized    // 未授权
errors.ErrForbidden       // 禁止访问
errors.ErrConflict        // 资源冲突
errors.ErrInternal        // 内部错误
errors.ErrDatabase        // 数据库错误
errors.ErrExternal        // 外部服务错误

// 使用示例
if exists {
    return nil, errors.ErrUserExists
}

if err != nil {
    return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to create user")
}
```

## 🚀 部署和运行

### 开发环境

```bash
# 安装依赖
make deps

# 生成依赖注入代码
make wire

# 构建应用
make build

# 运行应用
make run

# 使用 Docker
make up
```

### 生产环境

```bash
# 构建生产镜像
make docker-build

# 运行生产容器
make docker-prod
```

## 📝 最佳实践

1. **保持简洁**: 避免过度抽象，遵循 Go 的哲学
2. **错误处理**: 使用统一的错误类型和处理方式
3. **依赖注入**: 使用 Wire 进行编译时依赖注入
4. **配置管理**: 使用环境变量覆盖配置文件
5. **数据库操作**: 使用 GORM 和事务确保数据一致性
6. **API 设计**: 遵循 RESTful 设计原则
7. **中间件**: 合理使用中间件处理横切关注点

## 📚 相关文档

- [消息队列架构](docs/MESSAGE_QUEUE.md)
- [多数据源配置](docs/MULTI_DATASOURCE.md)
- [路由架构](docs/ROUTER_ARCHITECTURE.md)
- [任务调度器](docs/SCHEDULER.md)
- [ID 生成器](docs/SONYFLAKE_ID_GENERATOR.md)
- [Wire 依赖注入](docs/WIRE_ARCHITECTURE.md)