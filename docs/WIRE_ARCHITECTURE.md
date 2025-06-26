# Wire 依赖注入架构

## 🎯 **架构概述**

本项目使用 [Google Wire](https://github.com/google/wire) 进行依赖注入管理，这是 Google 推荐的企业级 Go 项目依赖注入解决方案。

## 🏗️ **架构层次**

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
├─────────────────────────────────────────────────────────────┤
│                     Handler Layer                           │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                           │
├─────────────────────────────────────────────────────────────┤
│                     Repository Layer                        │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                      │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │   Database  │    Redis    │  RabbitMQ   │   Logger    │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 📁 **Wire 文件结构**

```
internal/wire/
├── providers.go     # 提供者定义和集合
├── wire.go          # Wire 注入器定义
└── wire_gen.go      # Wire 自动生成的代码
```

## 🔧 **核心组件**

### 1. **提供者集合 (Provider Sets)**

```go
// 基础设施层
var InfrastructureSet = wire.NewSet(
    config.LoadConfig,
    ProvideLoggerConfig,
    ProvideDatabasesConfig,
    ProvideRedisConfig,
    ProvideRabbitMQConfig,
    logger.New,
    database.NewDatabases,
    ProvideMainDatabase,
    redispkg.NewRedis,
    mq.NewRabbitMQ,
)

// Repository 层
var RepositorySet = wire.NewSet(
    repository.NewUserRepository,
)

// Service 层
var ServiceSet = wire.NewSet(
    service.NewUserService,
)

// Handler 层
var HandlerSet = wire.NewSet(
    v1.NewUserHandler,
)

// App 层
var AppSet = wire.NewSet(
    app.NewApp,
)
```

### 2. **依赖注入器 (Injector)**

```go
//go:build wireinject
// +build wireinject

func InitializeApplication() (*app.App, error) {
    wire.Build(AllSet)
    return &app.App{}, nil
}
```

### 3. **App 结构体（直接包含所有依赖）**

```go
type App struct {
    // HTTP 服务
    Engine *gin.Engine
    Server *http.Server

    // 基础设施依赖
    Logger      *zap.Logger
    Config      *config.Config
    DataSources map[string]*gorm.DB
    MainDB      *gorm.DB
    Redis       *redis.Client
    RabbitMQ    *amqp.Connection

    // 业务层依赖
    UserHandler *v1.UserHandler
}
```

## 🚀 **使用方式**

### 1. **生成 Wire 代码**

```bash
# 生成依赖注入代码
make wire

# 或者直接使用 wire 命令
cd internal/wire && wire
```

### 2. **在应用中使用**

```go
// 初始化应用（包含所有依赖）
app, err := wire.InitializeApplication()
if err != nil {
    return fmt.Errorf("failed to initialize application: %w", err)
}

// 直接使用依赖
app.Logger.Info("Application started")
app.Redis.Set(ctx, "key", "value", time.Hour)
```

## 🆚 **架构对比**

| 方案 | 结构 | 优势 | 劣势 |
|------|------|------|------|
| **Dependencies 包装** | App → Dependencies → 各种依赖 | 依赖分组清晰 | 多一层间接访问 |
| **直接依赖 (当前)** | App → 各种依赖 | 访问简洁直观 | App 结构体较大 |

## 🎯 **当前架构优势**

### 1. **访问简洁**
```go
app.Logger.Info("message")
app.Redis.Set(ctx, "key", "value", time.Hour)

```

### 2. **代码更清晰**
- 减少了一层间接访问
- App 结构体直接反映所有依赖
- 更符合 Go 的简洁哲学

### 3. **Wire 配置更简单**
```go
// 直接提供 App
var AppSet = wire.NewSet(
    app.NewApp,
)

// 不需要额外的 Dependencies 包装函数
```

## 📝 **最佳实践**

### 1. **App 构造函数**

```go
func NewApp(
    logger *zap.Logger,
    config *config.Config,
    dataSources map[string]*gorm.DB,
    mainDB *gorm.DB,
    redis *redis.Client,
    rabbitMQ *amqp.Connection,
    userHandler *v1.UserHandler,
) *App {
    // 构造逻辑
    return &App{
        Logger:      logger,
        Config:      config,
        DataSources: dataSources,
        // ... 其他依赖
    }
}
```

### 2. **依赖分组**

```go
type App struct {
    // HTTP 服务
    Engine *gin.Engine
    Server *http.Server

    // 基础设施依赖 (按类型分组)
    Logger      *zap.Logger
    Config      *config.Config
    DataSources map[string]*gorm.DB
    MainDB      *gorm.DB
    Redis       *redis.Client
    RabbitMQ    *amqp.Connection

    // 业务层依赖 (按层次分组)
    UserHandler *v1.UserHandler
    // OrderHandler *v1.OrderHandler  // 未来扩展
}
```

### 3. **接口使用**

```go
// 在 App 中使用接口而不是具体类型
type App struct {
    UserService interfaces.UserService  // 接口
    UserRepo    interfaces.UserRepository  // 接口
}
```

## 🔄 **开发工作流**

1. **添加新依赖**
   ```go
   // 1. 在 App 结构体中添加字段
   type App struct {
       // ... 现有字段
       OrderHandler *v1.OrderHandler // 新增
   }
   
   // 2. 在构造函数中添加参数
   func NewApp(
       // ... 现有参数
       orderHandler *v1.OrderHandler, // 新增
   ) *App {
       return &App{
           // ... 现有字段
           OrderHandler: orderHandler, // 新增
       }
   }
   
   // 3. 在相应的 ProviderSet 中添加构造函数
   var HandlerSet = wire.NewSet(
       v1.NewUserHandler,
       v1.NewOrderHandler, // 新增
   )
   ```

2. **重新生成代码**
   ```bash
   make wire
   ```

3. **测试编译**
   ```bash
   make build
   ```

## 🧪 **测试**

### 1. **单元测试**

```go
func TestUserService(t *testing.T) {
    // 直接创建依赖
    mockRepo := &MockUserRepository{}
    logger := zaptest.NewLogger(t)
    
    service := service.NewUserService(mockRepo, logger)
    
    // 测试逻辑...
}
```

### 2. **集成测试**

```go
func TestIntegration(t *testing.T) {
    // 使用 Wire 创建完整应用
    app, err := wire.InitializeApplication()
    require.NoError(t, err)
    defer app.Stop()
    
    // 使用真实的依赖进行集成测试
    // 可以直接访问 app.UserHandler, app.Redis 等
}
```

## 🚨 **注意事项**

1. **App 结构体大小**：随着项目增长，App 结构体会变大，但这是可接受的
2. **依赖分组**：通过注释和字段顺序来组织依赖
3. **接口使用**：优先使用接口类型而不是具体实现
4. **构造函数参数**：参数较多时要注意顺序和可读性

## 🔗 **相关资源**

- [Google Wire 官方文档](https://github.com/google/wire)
- [Wire 用户指南](https://github.com/google/wire/blob/main/docs/guide.md)
- [Go 项目布局标准](https://github.com/golang-standards/project-layout) 