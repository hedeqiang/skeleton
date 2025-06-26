# 多数据源使用指南

## 概述

本项目支持同时连接和使用多个数据库，每个数据源可以是不同类型的数据库（MySQL、PostgreSQL等）。**不指定数据源时自动使用默认数据源。**

## 🚀 快速开始

### 1. 配置多个数据源

在现有的配置文件中添加多个数据源：

```yaml
# configs/config.dev.yaml
databases:
  # 主业务数据库（默认数据源，必须命名为 primary）
  primary:
    type: "postgres"
    dsn: "host=localhost port=5432 user=postgres password=123456 dbname=main_db sslmode=disable"
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: "1h"

  # 用户系统数据库
  user_db:
    type: "mysql"
    dsn: "root:password@tcp(127.0.0.1:3306)/user_system?charset=utf8mb4&parseTime=True&loc=Local"
    max_open_conns: 50
    max_idle_conns: 5
    conn_max_lifetime: "30m"

  # 日志数据库
  log_db:
    type: "postgres"
    dsn: "host=localhost port=5432 user=postgres password=123456 dbname=logs sslmode=disable"
    max_open_conns: 30
    max_idle_conns: 5
    conn_max_lifetime: "2h"

  # 报表数据库
  report_db:
    type: "mysql"
    dsn: "root:password@tcp(127.0.0.1:3306)/reports?charset=utf8mb4&parseTime=True&loc=Local"
    max_open_conns: 20
    max_idle_conns: 3
    conn_max_lifetime: "1h"
```

### 2. 使用多数据源的三种方式

#### 方式一：使用默认数据源（推荐）

**当Repository不指定数据源时，自动使用 `primary` 数据源。**

```go
// 使用默认数据源的Repository（现有代码无需修改）
type userRepository struct {
    db *gorm.DB  // 这将是primary数据源
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error  // 使用primary数据源
}
```

#### 方式二：Repository中指定数据源

```go
// 使用特定数据源的Repository
type logRepository struct {
    db *gorm.DB
}

func NewLogRepository(dataSources map[string]*gorm.DB) LogRepository {
    // 优先使用log_db，不存在则回退到primary
    if logDB, exists := dataSources["log_db"]; exists {
        return &logRepository{db: logDB}
    }
    return &logRepository{db: dataSources["primary"]}
}

func (r *logRepository) CreateLog(ctx context.Context, log *Log) error {
    return r.db.WithContext(ctx).Create(log).Error  // 使用log_db数据源
}
```

#### 方式三：Repository内部动态选择

```go
// 多数据源Repository
type multiRepository struct {
    dataSources map[string]*gorm.DB
    defaultDB   *gorm.DB
}

func NewMultiRepository(dataSources map[string]*gorm.DB, defaultDB *gorm.DB) *multiRepository {
    return &multiRepository{
        dataSources: dataSources,
        defaultDB:   defaultDB,
    }
}

// 获取指定数据源，不存在则返回默认数据源
func (r *multiRepository) getDB(name string) *gorm.DB {
    if db, exists := r.dataSources[name]; exists {
        return db
    }
    return r.defaultDB
}

// 根据业务需要选择不同数据源
func (r *multiRepository) CreateUser(ctx context.Context, user *User) error {
    return r.getDB("user_db").WithContext(ctx).Create(user).Error
}

func (r *multiRepository) CreateLog(ctx context.Context, log *Log) error {
    return r.getDB("log_db").WithContext(ctx).Create(log).Error
}

func (r *multiRepository) CreateOrder(ctx context.Context, order *Order) error {
    return r.defaultDB.WithContext(ctx).Create(order).Error  // 使用默认数据源
}
```

### 3. Wire依赖注入配置

```go
// internal/wire/providers.go

// 为特定数据源创建Provider
func ProvideLogRepository(dataSources map[string]*gorm.DB) LogRepository {
    return repository.NewLogRepository(dataSources)
}

func ProvideMultiRepository(dataSources map[string]*gorm.DB, defaultDB *gorm.DB) *repository.MultiRepository {
    return repository.NewMultiRepository(dataSources, defaultDB)
}

// 更新Repository Set
var RepositorySet = wire.NewSet(
    repository.NewUserRepository,     // 使用默认数据源
    ProvideLogRepository,            // 使用log_db数据源
    ProvideMultiRepository,          // 使用多数据源
)
```

## 🎯 核心概念

### 默认数据源机制
- **必须命名为 `primary`**：这是约定，不可更改
- **自动注入**：Wire会自动将 `primary` 注入为 `*gorm.DB` 类型
- **兜底机制**：当不指定数据源时使用此连接

### 数据源选择策略
1. **明确指定**：`dataSources["specific_db"]`
2. **回退机制**：指定数据源不存在时使用 `primary`
3. **动态选择**：Repository内部根据业务逻辑选择

## 📝 使用场景

### 按业务模块分库
```yaml
databases:
  primary:      # 主业务（订单、商品等）
  user_db:      # 用户系统
  payment_db:   # 支付系统
  log_db:       # 系统日志
```

### 按数据类型分库  
```yaml
databases:
  primary:      # 核心业务数据
  analytics_db: # 分析统计数据
  cache_db:     # 缓存和临时数据
  archive_db:   # 归档历史数据
```

### 按访问频率分库
```yaml
databases:
  primary:      # 高频读写
  report_db:    # 低频读取  
  backup_db:    # 备份存储
```

## ⚙️ 配置建议

### 连接池参数建议

| 业务类型 | max_open_conns | max_idle_conns | conn_max_lifetime |
|---------|---------------|---------------|-------------------|
| 主业务数据库 | 100 | 20 | 1h |
| 日志数据库 | 30 | 5 | 2h |
| 报表数据库 | 50 | 10 | 30m |
| 临时数据库 | 15 | 3 | 30m |

### 支持的数据库类型

| 数据库 | type 值 | 驱动包 |
|-------|---------|-------|
| MySQL | `mysql` | gorm.io/driver/mysql |
| PostgreSQL | `postgres` | gorm.io/driver/postgres |

## 💡 最佳实践

### 1. 数据源命名规范
- 使用有意义的名称：`user_db`, `log_db`, `report_db`
- 主数据源必须命名为 `primary`
- 避免使用数字编号：`db1`, `db2`

### 2. 优雅降级处理
```go
func (r *repository) getDB(name string) *gorm.DB {
    if db, exists := r.dataSources[name]; exists {
        return db
    }
    // 回退到默认数据源
    return r.defaultDB
}
```

### 3. 错误处理
```go
func NewSpecificRepository(dataSources map[string]*gorm.DB) Repository {
    db, exists := dataSources["specific_db"]
    if !exists {
        // 记录警告并使用默认数据源
        log.Warn("specific_db not found, using primary database")
        db = dataSources["primary"]
    }
    return &repository{db: db}
}
```

### 4. 事务处理注意事项
```go
// 跨数据源不支持分布式事务，需要业务层面处理
func (s *service) CreateUserWithLog(ctx context.Context, user *User) error {
    // 1. 先创建用户
    if err := s.userRepo.Create(ctx, user); err != nil {
        return err
    }
    
    // 2. 记录日志（如果失败，不影响主业务）
    if err := s.logRepo.CreateLog(ctx, log); err != nil {
        s.logger.Error("Failed to create log", zap.Error(err))
        // 不返回错误，避免影响主业务
    }
    
    return nil
}
```

## ⚠️ 注意事项

1. **跨数据源事务**：不同数据源之间不支持分布式事务
2. **主数据源必须存在**：必须配置名为 `primary` 的数据源
3. **优雅降级**：当指定数据源不存在时，应该有回退机制
4. **性能监控**：定期监控各数据源的连接使用情况

## 🔧 示例完整配置

```yaml
# configs/config.dev.yaml
databases:
  # 默认主数据库（必须）
  primary:
    type: "postgres"  
    dsn: "host=localhost port=5432 user=postgres password=123456 dbname=main_db sslmode=disable"
    max_open_conns: 100
    max_idle_conns: 20
    conn_max_lifetime: "1h"

  # 业务数据库
  business_db:
    type: "mysql"
    dsn: "root:password@tcp(127.0.0.1:3306)/business?charset=utf8mb4&parseTime=True&loc=Local"
    max_open_conns: 80
    max_idle_conns: 15
    conn_max_lifetime: "45m"
```

通过以上配置，你就可以灵活地在应用中使用多个数据源，同时保持代码的简洁性。**记住：不指定数据源时会自动使用 `primary` 数据源！** 