# Go Skeleton Framework 使用指南

## 概述

Go Skeleton 是一个现代化的Go应用程序开发框架，提供了统一的基础设施和最佳实践，帮助开发者快速构建高质量、可维护的Go应用程序。

## 核心特性

- ✅ **统一的错误处理系统** - 标准化的错误类型和HTTP状态码映射
- ✅ **配置验证系统** - 灵活的配置验证框架
- ✅ **服务管理** - 统一的服务生命周期管理
- ✅ **数据库层优化** - 连接池管理和基础仓储模式
- ✅ **应用框架** - 优雅的应用启动和关闭

## 快速开始

### 1. 项目结构

```
your-project/
├── cmd/
│   └── server/
│       └── main.go          # 应用程序入口
├── internal/
│   ├── config/
│   │   └── config.go        # 配置加载
│   ├── handlers/
│   │   └── user_handler.go # HTTP处理器
│   ├── repositories/
│   │   └── user_repo.go    # 数据仓储
│   └── services/
│       └── user_service.go # 业务服务
├── pkg/
│   ├── app/                 # 应用框架
│   ├── config/              # 配置管理
│   ├── database/            # 数据库层
│   ├── errors/              # 错误处理
│   ├── logger/              # 日志系统
│   └── service/             # 服务管理
├── config/
│   └── config.yaml          # 配置文件
├── go.mod
└── go.sum
```

### 2. 主应用程序 (main.go)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/hedeqiang/skeleton/internal/config"
    "github.com/hedeqiang/skeleton/internal/handlers"
    "github.com/hedeqiang/skeleton/internal/repositories"
    "github.com/hedeqiang/skeleton/internal/services"
    "github.com/hedeqiang/skeleton/pkg/app"
    "github.com/hedeqiang/skeleton/pkg/database"
    "github.com/hedeqiang/skeleton/pkg/service"
    "go.uber.org/zap"
    "gorm.io/gorm/logger"
)

type App struct {
    *app.BaseApp
    server         *http.Server
    serviceManager *service.ServiceManager
    dbManager      *database.Database
}

func NewApp(cfg *config.Config, logger *zap.Logger) *App {
    baseApp := app.NewBaseApp("your-app", "1.0.0", logger, cfg)
    return &App{
        BaseApp: baseApp,
    }
}

func (a *App) Start() error {
    if err := a.BaseApp.Start(); err != nil {
        return err
    }

    // 初始化服务管理器
    a.serviceManager = service.NewServiceManager(a.GetLogger())

    // 初始化数据库
    if err := a.initDatabase(); err != nil {
        return fmt.Errorf("failed to initialize database: %w", err)
    }

    // 初始化服务
    if err := a.initServices(); err != nil {
        return fmt.Errorf("failed to initialize services: %w", err)
    }

    // 初始化HTTP服务器
    if err := a.initHTTPServer(); err != nil {
        return fmt.Errorf("failed to initialize HTTP server: %w", err)
    }

    // 启动所有服务
    if err := a.serviceManager.StartAll(a.Context()); err != nil {
        return fmt.Errorf("failed to start services: %w", err)
    }

    return nil
}

func (a *App) Stop() error {
    // 停止HTTP服务器
    if a.server != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        a.server.Shutdown(ctx)
    }

    // 停止所有服务
    if a.serviceManager != nil {
        a.serviceManager.StopAll(a.Context())
    }

    // 关闭数据库连接
    if a.dbManager != nil {
        a.dbManager.Close()
    }

    return a.BaseApp.Stop()
}

func (a *App) initDatabase() error {
    dbConfig := a.GetConfig().GetDatabase().Primary
    
    dbConfigWrapper := &database.DBConfig{
        Driver:            dbConfig.Type,
        DSN:               dbConfig.DSN,
        MaxOpenConns:      dbConfig.MaxOpenConns,
        MaxIdleConns:      dbConfig.MaxIdleConns,
        ConnMaxLifetime:   30 * time.Minute,
        ConnMaxIdleTime:   5 * time.Minute,
        SlowThreshold:     200 * time.Millisecond,
        LoggerLevel:       logger.Warn,
        DisableColor:      false,
        IgnoreRecordNotFound: true,
    }

    var err error
    a.dbManager, err = database.NewDatabase(dbConfigWrapper)
    return err
}

func (a *App) initServices() error {
    // 创建用户仓储
    userRepo := repositories.NewUserRepository(a.dbManager.DB())
    
    // 创建用户服务
    userService := services.NewUserService(userRepo, a.GetLogger())
    
    // 添加到服务管理器
    a.serviceManager.AddService(userService)
    
    return nil
}

func (a *App) initHTTPServer() error {
    mux := http.NewServeMux()
    
    // 注册路由
    mux.HandleFunc("/health", a.healthHandler)
    
    a.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", a.GetConfig().GetApp().Port),
        Handler: mux,
    }
    
    go func() {
        if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            a.GetLogger().Error("HTTP server failed", zap.Error(err))
        }
    }()
    
    return nil
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
    health := a.Health()
    
    // 检查数据库健康状态
    if a.dbManager != nil {
        if err := a.dbManager.Health(r.Context()); err != nil {
            health["database"] = "unhealthy"
        } else {
            health["database"] = "healthy"
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status": "ok", "health": %v}`, health)
}

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 初始化日志器
    logger := zap.NewExample()
    defer logger.Sync()
    
    // 创建应用程序
    app := NewApp(cfg, logger)
    
    // 设置信号处理
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // 启动应用程序
    go func() {
        if err := app.Run(); err != nil {
            logger.Fatal("Application failed", zap.Error(err))
        }
    }()
    
    // 等待信号
    sig := <-sigChan
    logger.Info("Received signal, shutting down...", zap.String("signal", sig.String()))
    
    // 优雅关闭
    if err := app.GracefulShutdown(30 * time.Second); err != nil {
        logger.Error("Graceful shutdown failed", zap.Error(err))
        os.Exit(1)
    }
    
    logger.Info("Application shutdown completed")
}
```

### 3. 配置文件 (config/config.yaml)

```yaml
app:
  name: "your-app"
  env: "development"
  host: "0.0.0.0"
  port: 8080

logger:
  level: "info"
  encoding: "json"
  output_path: ["stdout"]

database:
  enabled: true
  primary:
    type: "mysql"
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: "30m"
    conn_max_idle_time: "5m"
    slow_threshold: "200ms"
    logger_level: "warn"
    disable_color: false

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  enabled: false

jwt:
  secret: "your-secret-key-here"
  expire_duration: "24h"
  refresh_duration: "168h"
  enabled: true
```

### 4. 配置加载 (internal/config/config.go)

```go
package config

import (
    "fmt"
    "os"

    "github.com/hedeqiang/skeleton/pkg/config"
    "gopkg.in/yaml.v3"
)

type Config struct {
    App      config.AppConfig      `yaml:"app"`
    Logger   config.LoggerConfig   `yaml:"logger"`
    Database config.DatabaseConfig `yaml:"database"`
    Redis    config.RedisConfig    `yaml:"redis"`
    JWT      config.JWTConfig      `yaml:"jwt"`
}

func LoadConfig(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    if err := validateConfig(&cfg); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return &cfg, nil
}

func validateConfig(cfg *Config) error {
    validator := config.NewConfigValidation()

    // 验证应用配置
    validator.Required("app.name", cfg.App.Name)
    validator.Required("app.env", cfg.App.Env)
    validator.OneOf("app.env", cfg.App.Env, []interface{}{"development", "staging", "production"})
    validator.Required("app.host", cfg.App.Host)
    validator.Min("app.port", cfg.App.Port, 1)
    validator.Max("app.port", cfg.App.Port, 65535)

    // 验证日志配置
    validator.Required("logger.level", cfg.Logger.Level)
    validator.OneOf("logger.level", cfg.Logger.Level, []interface{}{"debug", "info", "warn", "error"})

    // 验证数据库配置
    if cfg.Database.Enabled {
        validator.Required("database.primary.type", cfg.Database.Primary.Type)
        validator.OneOf("database.primary.type", cfg.Database.Primary.Type, []interface{}{"mysql", "postgres"})
        validator.Required("database.primary.dsn", cfg.Database.Primary.DSN)
    }

    if errors := validator.Validate(); len(errors) > 0 {
        return fmt.Errorf("%s", errors[0].Error())
    }

    return nil
}

// 实现config.Config接口
func (c *Config) GetApp() config.AppConfig { return c.App }
func (c *Config) GetLogger() config.LoggerConfig { return c.Logger }
func (c *Config) GetDatabase() config.DatabaseConfig { return c.Database }
func (c *Config) GetRedis() config.RedisConfig { return c.Redis }
func (c *Config) GetJWT() config.JWTConfig { return c.JWT }
func (c *Config) GetRabbitMQ() config.RabbitMQConfig { return config.RabbitMQConfig{} }
func (c *Config) GetTrace() config.TraceConfig { return config.TraceConfig{} }
func (c *Config) GetScheduler() config.SchedulerConfig { return config.SchedulerConfig{} }
```

### 5. 数据仓储 (internal/repositories/user_repo.go)

```go
package repositories

import (
    "context"
    "time"

    "github.com/hedeqiang/skeleton/pkg/database"
    "gorm.io/gorm"
)

type User struct {
    ID        int64     `gorm:"primaryKey" json:"id"`
    Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
    Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
    Password  string    `gorm:"size:255" json:"-"`
    FirstName string    `gorm:"size:50" json:"first_name"`
    LastName  string    `gorm:"size:50" json:"last_name"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

type userRepository struct {
    *database.BaseRepository
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        BaseRepository: database.NewBaseRepository(db),
    }
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
    return r.BaseRepository.Create(ctx, user)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    var user User
    if err := r.BaseRepository.FindByID(ctx, &user, id); err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
    var user User
    if err := r.BaseRepository.FindOne(ctx, &user, "username = ?", username); err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *User) error {
    return r.BaseRepository.Update(ctx, user)
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
    user := &User{ID: id}
    return r.BaseRepository.Delete(ctx, user)
}
```

### 6. 业务服务 (internal/services/user_service.go)

```go
package services

import (
    "context"

    "github.com/hedeqiang/skeleton/internal/repositories"
    "github.com/hedeqiang/skeleton/pkg/errors"
    "github.com/hedeqiang/skeleton/pkg/service"
    "go.uber.org/zap"
    "golang.org/x/crypto/bcrypt"
)

type UserService interface {
    CreateUser(ctx context.Context, username, email, password string) (*repositories.User, error)
    GetUserByID(ctx context.Context, id int64) (*repositories.User, error)
    GetUserByUsername(ctx context.Context, username string) (*repositories.User, error)
}

type userService struct {
    *service.BaseService
    userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository, logger *zap.Logger) UserService {
    return &userService{
        BaseService: service.NewBaseService("UserService", logger),
        userRepo:    userRepo,
    }
}

func (s *userService) CreateUser(ctx context.Context, username, email, password string) (*repositories.User, error) {
    // 检查用户名是否已存在
    exists, err := s.userRepo.ExistsByUsername(ctx, username)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to check username existence")
    }
    if exists {
        return nil, errors.New(errors.ErrorTypeValidation, "username already exists")
    }

    // 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to hash password")
    }

    // 创建用户
    user := &repositories.User{
        Username:  username,
        Email:     email,
        Password:  string(hashedPassword),
        FirstName: "",
        LastName:  "",
        IsActive:  true,
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to create user")
    }

    s.GetLogger().Info("User created successfully", zap.Int64("user_id", user.ID))
    return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*repositories.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to get user")
    }
    return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*repositories.User, error) {
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to get user")
    }
    return user, nil
}
```

### 7. HTTP处理器 (internal/handlers/user_handler.go)

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/hedeqiang/skeleton/internal/services"
    "github.com/hedeqiang/skeleton/pkg/errors"
    "go.uber.org/zap"
)

type UserHandler struct {
    userService services.UserService
    logger      *zap.Logger
}

func NewUserHandler(userService services.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.handleError(w, errors.Wrap(err, errors.ErrorTypeValidation, "invalid request body"))
        return
    }

    user, err := h.userService.CreateUser(r.Context(), req.Username, req.Email, req.Password)
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.success(w, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 解析用户ID从URL参数
    // 这里简化处理，实际应该从URL路径或查询参数获取
    
    user, err := h.userService.GetUserByID(r.Context(), 1) // 示例ID
    if err != nil {
        h.handleError(w, err)
        return
    }

    h.success(w, user)
}

func (h *UserHandler) success(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    data,
    })
}

func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
    h.logger.Error("Request failed", zap.Error(err))

    if appErr, ok := err.(*errors.AppError); ok {
        statusCode := errors.GetHTTPStatus(appErr.Type)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   appErr.Message,
            "code":    string(appErr.Type),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": false,
        "error":   "Internal server error",
        "code":    "internal_error",
    })
}
```

## 错误处理

框架提供了统一的错误处理系统：

### 创建错误

```go
// 创建新的应用错误
err := errors.New(errors.ErrorTypeValidation, "invalid input")

// 包装现有错误
err := errors.Wrap(existingErr, errors.ErrorTypeDatabase, "database operation failed")

// 添加详情
err := errors.New(errors.ErrorTypeValidation, "invalid input").WithDetails("field 'username' is required")
```

### 错误类型

```go
const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeNotFound     ErrorType = "not_found"
    ErrorTypeUnauthorized ErrorType = "unauthorized"
    ErrorTypeForbidden    ErrorType = "forbidden"
    ErrorTypeConflict     ErrorType = "conflict"
    ErrorTypeInternal     ErrorType = "internal"
    ErrorTypeDatabase     ErrorType = "database"
    ErrorTypeExternal     ErrorType = "external"
)
```

### 检查错误类型

```go
if errors.IsNotFoundError(err) {
    // 处理未找到错误
}

if errors.IsValidationError(err) {
    // 处理验证错误
}
```

## 配置验证

框架提供了灵活的配置验证系统：

```go
validator := config.NewConfigValidation()

// 基础验证
validator.Required("app.name", cfg.App.Name)
validator.Min("app.port", cfg.App.Port, 1)
validator.Max("app.port", cfg.App.Port, 65535)

// 字符串验证
validator.MinLength("app.name", cfg.App.Name, 3)
validator.MaxLength("app.name", cfg.App.Name, 50)

// 枚举验证
validator.OneOf("app.env", cfg.App.Env, []interface{}{"development", "staging", "production"})

// URL和邮箱验证
validator.URL("app.url", cfg.App.URL)
validator.Email("user.email", cfg.User.Email)

// 端口验证
validator.Port("app.port", cfg.App.Port)

// 时间验证
validator.Duration("session.timeout", cfg.Session.Timeout)

// 获取验证结果
if errors := validator.Validate(); len(errors) > 0 {
    // 处理验证错误
}
```

## 数据库操作

框架提供了优化的数据库操作接口：

### 基础CRUD操作

```go
// 创建记录
err := repo.Create(ctx, &user)

// 更新记录
err := repo.Update(ctx, &user)

// 删除记录
err := repo.Delete(ctx, &user)

// 根据ID查找
err := repo.FindByID(ctx, &user, id)

// 条件查找
err := repo.FindOne(ctx, &user, "username = ?", username)

// 查找多个
err := repo.FindMany(ctx, &users, "is_active = ?", true)

// 统计数量
count, err := repo.Count(ctx, &User{}, "is_active = ?", true)

// 检查存在
exists, err := repo.Exists(ctx, &User{}, "username = ?", username)
```

### 事务操作

```go
// 执行事务
err := db.Transaction(ctx, func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    return nil
})
```

### 数据库健康检查

```go
err := db.Health(ctx)
if err != nil {
    // 数据库连接有问题
}
```

## 服务管理

框架提供了统一的服务管理：

### 创建服务

```go
type MyService struct {
    *service.BaseService
    // 添加服务特定的字段
}

func NewService(logger *zap.Logger) *MyService {
    return &MyService{
        BaseService: service.NewBaseService("MyService", logger),
    }
}

func (s *MyService) Start(ctx context.Context) error {
    s.GetLogger().Info("Starting my service")
    // 初始化服务
    return nil
}

func (s *MyService) Stop(ctx context.Context) error {
    s.GetLogger().Info("Stopping my service")
    // 清理资源
    return nil
}

func (s *MyService) Health(ctx context.Context) error {
    // 健康检查
    return nil
}
```

### 服务管理器

```go
// 创建服务管理器
serviceManager := service.NewServiceManager(logger)

// 添加服务
serviceManager.AddService(myService1)
serviceManager.AddService(myService2)

// 启动所有服务
err := serviceManager.StartAll(ctx)

// 停止所有服务
err := serviceManager.StopAll(ctx)

// 健康检查
err := serviceManager.HealthAll(ctx)
```

## 应用生命周期

框架提供了标准的应用生命周期管理：

### 启动应用

```go
app := NewApp(cfg, logger)

// 启动应用
if err := app.Start(); err != nil {
    log.Fatal("Failed to start app:", err)
}

// 运行应用（阻塞直到停止）
if err := app.Run(); err != nil {
    log.Fatal("App failed:", err)
}
```

### 优雅关闭

```go
// 优雅关闭
err := app.GracefulShutdown(30 * time.Second)
if err != nil {
    log.Error("Graceful shutdown failed:", err)
}
```

### 健康检查

```go
health := app.Health()
fmt.Printf("App health: %+v\n", health)
```

## 最佳实践

### 1. 错误处理

- 始终使用 `pkg/errors` 包处理错误
- 提供有意义的错误消息
- 使用适当的错误类型
- 在服务层包装错误

### 2. 配置管理

- 使用配置验证确保配置正确性
- 将敏感信息存储在环境变量中
- 为不同环境提供不同的配置文件

### 3. 数据库操作

- 使用事务确保数据一致性
- 实现适当的索引
- 使用连接池优化性能
- 定期进行健康检查

### 4. 服务设计

- 保持服务简单和专注
- 实现适当的接口
- 提供完整的生命周期管理
- 实现健康检查

### 5. 应用架构

- 使用分层架构
- 保持依赖注入
- 实现优雅关闭
- 提供监控和日志

## 监控和运维

### 健康检查

框架提供了健康检查功能：

```go
// 应用健康检查
health := app.Health()

// 数据库健康检查
err := db.Health(ctx)

// 服务健康检查
err := serviceManager.HealthAll(ctx)
```

### 日志记录

使用结构化日志记录重要信息：

```go
logger.Info("User created", 
    zap.Int64("user_id", user.ID),
    zap.String("username", user.Username),
    zap.String("ip", clientIP),
)
```

### 指标监控

数据库连接池状态：

```go
stats := db.Stats()
fmt.Printf("Database stats: %+v\n", stats)
```

## 部署

### 环境变量

使用环境变量覆盖配置：

```bash
export APP_ENV=production
export DATABASE_DSN="user:password@tcp(prod-db:3306)/dbname"
export JWT_SECRET="your-production-secret"
```

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./server"]
```

### 配置文件

生产环境配置示例：

```yaml
app:
  name: "your-app"
  env: "production"
  host: "0.0.0.0"
  port: 8080

logger:
  level: "info"
  encoding: "json"
  output_path: ["stdout", "/var/log/app.log"]

database:
  enabled: true
  primary:
    type: "mysql"
    dsn: "${DATABASE_DSN}"
    max_open_conns: 100
    max_idle_conns: 20
    conn_max_lifetime: "1h"
    conn_max_idle_time: "10m"
    slow_threshold: "100ms"
    logger_level: "error"
    disable_color: true
```

## 总结

Go Skeleton 框架提供了：

1. **统一的错误处理** - 标准化的错误类型和处理机制
2. **灵活的配置管理** - 支持验证和环境变量
3. **完整的服务管理** - 统一的生命周期管理
4. **优化的数据库层** - 连接池和基础仓储模式
5. **标准的应用架构** - 优雅启动和关闭

使用这个框架可以快速构建高质量、可维护的Go应用程序。