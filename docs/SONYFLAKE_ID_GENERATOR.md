# Sonyflake ID生成器使用文档

## 概述

本项目集成了Sony的[sonyflake](https://github.com/sony/sonyflake)雪花算法ID生成器，用于生成分布式唯一ID。Sonyflake相比Twitter的Snowflake有以下优势：

- **更长的生命周期**：174年 vs 69年
- **更多分布式机器支持**：2^16 vs 2^10
- **更灵活的配置**：支持自定义位数分配和时间单位

## ID结构

默认情况下，Sonyflake ID由以下部分组成：
```
39位时间戳（10毫秒为单位）+ 8位序列号 + 16位机器ID = 63位唯一ID
```

## 基本使用

### 1. 默认配置使用

```go
// 在服务中注入ID生成器
type UserService struct {
    idGenerator idgen.IDGenerator
    // ... 其他依赖
}

func NewUserService(idGenerator idgen.IDGenerator) *UserService {
    return &UserService{
        idGenerator: idGenerator,
    }
}

// 生成唯一ID
func (s *UserService) CreateUser(name string) (*User, error) {
    // 生成用户ID
    userID, err := s.idGenerator.NextID()
    if err != nil {
        return nil, fmt.Errorf("failed to generate user ID: %w", err)
    }

    user := &User{
        ID:   userID,
        Name: name,
    }
    
    // ... 保存到数据库
    return user, nil
}

// 生成字符串形式的ID
func (s *UserService) GenerateOrderNo() (string, error) {
    orderID, err := s.idGenerator.NextIDString()
    if err != nil {
        return "", fmt.Errorf("failed to generate order ID: %w", err)
    }
    
    return fmt.Sprintf("ORDER_%s", orderID), nil
}
```

### 2. 直接使用（不通过依赖注入）

```go
package main

import (
    "fmt"
    "github.com/hedeqiang/skeleton/pkg/idgen"
)

func main() {
    // 创建默认ID生成器
    generator, err := idgen.NewSonyflakeGenerator()
    if err != nil {
        panic(err)
    }

    // 生成ID
    id, err := generator.NextID()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated ID: %d\n", id)
    
    // 生成字符串ID
    idStr, err := generator.NextIDString()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated ID String: %s\n", idStr)
}
```

## 配置选项

### 1. 配置文件配置

在`config.yaml`中添加ID生成器配置（可选）：

```yaml
# ID生成器配置（可选，不配置则使用默认值）
id_generator:
  start_time: "2024-01-01T00:00:00Z"  # 起始时间
  machine_id: 1001                     # 机器ID（0表示自动获取）
  bits_sequence: 8                     # 序列号位数
  bits_machine_id: 16                  # 机器ID位数
  time_unit: "10ms"                    # 时间单位
```

### 2. 代码配置

```go
// 创建自定义配置
config := idgen.Config{
    StartTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    MachineID:     1001, // 指定机器ID，0表示自动获取
    BitsSequence:  8,    // 8位序列号
    BitsMachineID: 16,   // 16位机器ID
    TimeUnit:      10 * time.Millisecond, // 10毫秒时间单位
}

// 使用自定义配置创建生成器
generator, err := idgen.NewSonyflakeGeneratorWithConfig(config)
if err != nil {
    panic(err)
}
```

### 3. 默认配置

如果不提供配置，系统使用以下默认值：

```go
Config{
    StartTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    MachineID:     0, // 自动获取（基于本地IP地址）
    BitsSequence:  8, // 8位序列号
    BitsMachineID: 16, // 16位机器ID
    TimeUnit:      10 * time.Millisecond, // 10毫秒时间单位
}
```

## 机器ID获取策略

### 1. 自动获取（默认）
- 基于本地网络接口的IPv4地址
- 使用IP地址的低16位作为机器ID
- 适用于大多数场景

### 2. 手动指定
- 在配置中指定`machine_id`
- 确保在分布式环境中每台机器的ID唯一
- 适用于Docker、Kubernetes等容器化环境

### 3. AWS环境
如果在AWS环境中，可以使用sonyflake提供的AWS工具：

```go
import "github.com/sony/sonyflake/v2/awsutil"

// 在AWS EC2中获取机器ID
machineID, err := awsutil.AmazonEC2MachineID()
if err != nil {
    // 处理错误
}

config := idgen.Config{
    MachineID: machineID,
    // ... 其他配置
}
```

## 性能特点

- **生成速度**：单实例每10毫秒最多生成256个ID（2^8）
- **并发安全**：线程安全，支持并发调用
- **内存占用**：极小，无状态存储
- **高可用**：无外部依赖，本地生成

## 使用场景

### 1. 数据库主键
```go
type User struct {
    ID       int64  `gorm:"primaryKey"`
    Username string
    Email    string
}

func (s *UserService) CreateUser(username, email string) error {
    userID, err := s.idGenerator.NextID()
    if err != nil {
        return err
    }
    
    user := &User{
        ID:       userID,
        Username: username,
        Email:    email,
    }
    
    return s.userRepo.Create(user)
}
```

### 2. 订单号生成
```go
func (s *OrderService) GenerateOrderNumber() (string, error) {
    id, err := s.idGenerator.NextID()
    if err != nil {
        return "", err
    }
    
    // 格式化为订单号
    return fmt.Sprintf("ORD%d", id), nil
}
```

### 3. 分布式追踪ID
```go
func (s *TraceService) GenerateTraceID() (string, error) {
    traceID, err := s.idGenerator.NextIDString()
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("trace-%s", traceID), nil
}
```

## 注意事项

1. **时钟回拨**：Sonyflake会检测时钟回拨并返回错误，确保ID的单调递增性
2. **机器ID唯一性**：在分布式环境中确保每台机器的ID唯一
3. **生成速率限制**：单实例每10毫秒最多256个ID，高并发场景可考虑多实例
4. **生命周期**：默认配置下可使用174年（从2024年开始）

## 故障排除

### 1. 机器ID获取失败
```
Error: failed to get machine ID: no valid IP address found
```
**解决方案**：手动指定机器ID或检查网络配置

### 2. 时钟回拨错误
```
Error: clock moved backwards
```
**解决方案**：等待系统时钟恢复正常或重启应用

### 3. ID生成失败
```
Error: failed to generate ID
```
**解决方案**：检查系统时间和机器ID配置

## 最佳实践

1. **单例使用**：每个应用实例只创建一个ID生成器实例
2. **错误处理**：始终检查ID生成的错误返回
3. **配置管理**：在配置文件中统一管理ID生成器参数
4. **监控告警**：监控ID生成失败率和性能指标
5. **测试验证**：在部署前测试ID的唯一性和生成性能 