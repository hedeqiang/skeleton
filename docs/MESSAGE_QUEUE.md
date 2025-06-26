# RabbitMQ 消息队列使用指南

这是一个完整的 RabbitMQ 消息队列使用指南，包含架构设计、Hello World 示例和扩展指导。

## 📋 目录

- [架构概述](#架构概述)
- [快速开始](#快速开始)
- [Hello World 示例](#hello-world-示例)
- [消息处理器架构](#消息处理器架构)
- [配置管理](#配置管理)
- [扩展指南](#扩展指南)
- [最佳实践](#最佳实践)
- [故障排除](#故障排除)

## 🏗️ 架构概述

### 整体架构

```
HTTP API → Service → Producer → RabbitMQ → Consumer → Processor → Business Logic
```

### 核心组件

1. **Producer (生产者)** - 负责发布消息到 RabbitMQ
2. **Consumer (消费者)** - 负责监听队列并分发消息
3. **Processor (处理器)** - 负责处理具体业务逻辑
4. **Registry (注册器)** - 负责管理处理器注册和路由

### 目录结构

```
internal/messaging/
├── message_processor.go           # 处理器接口定义
├── consumer/message_consumer.go   # 消费服务
└── processors/                    # 具体处理器
    └── hello_processor.go         # Hello 消息处理器

pkg/mq/
└── rabbitmq.go                   # RabbitMQ 客户端封装
```

## 🚀 快速开始

### 1. 启动 RabbitMQ

```bash
# 使用 Docker 启动 RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# 或者使用本地安装
brew install rabbitmq
brew services start rabbitmq
```

### 2. 启动消费者

```bash
# 启动消息消费者
make mq-consumer

# 或者直接运行
go run cmd/consumer/main.go
```

### 3. 启动 API 服务

```bash
# 启动 API 服务
make mq-api

# 或者直接运行
go run cmd/api/main.go
```

### 4. 发布消息

```bash
# 测试消息发布
make test-mq-api

# 或者手动发送
curl -X POST http://localhost:8080/api/v1/messages/hello/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, World!", "sender": "test-user"}'
```

## 👋 Hello World 示例

### API 端点

**发布消息**: `POST /api/v1/messages/hello/publish`

**请求格式**:
```json
{
  "content": "Hello, World!",
  "sender": "user123"
}
```

**响应格式**:
```json
{
  "code": 200,
  "message": "Hello消息发布成功",
  "data": {
    "message_id": "msg-1703123456789"
  }
}
```

### 消息格式

发布到队列的消息格式：
```json
{
  "message_id": "msg-1703123456789",
  "message_type": "hello",
  "payload": {
    "content": "Hello, World!",
    "sender": "user123",
    "timestamp": 1703123456
  },
  "timestamp": 1703123456
}
```

### 消息处理流程

1. **API 接收** - HTTP 请求到 HelloHandler
2. **业务处理** - HelloService 构建消息
3. **消息发布** - Producer 发布到 RabbitMQ
4. **消息路由** - 通过 `hello.exchange` 路由到 `hello.queue`
5. **消息消费** - Consumer 监听队列
6. **消息处理** - HelloProcessor 处理业务逻辑
7. **消息确认** - 处理完成后确认消息

## 🔧 消息处理器架构

### 核心接口

```go
// MessageProcessor 消息处理器接口
type MessageProcessor interface {
    ProcessMessage(ctx context.Context, msg BusinessMessage, app *app.App) error
    GetSupportedMessageType() string
}

// BusinessMessage 业务消息接口
type BusinessMessage interface {
    GetMessageID() string
    GetMessageType() string
    GetTimestamp() int64
}
```

### 消息封装

```go
// MessageEnvelope 消息信封
type MessageEnvelope struct {
    MessageID   string          `json:"message_id"`
    MessageType string          `json:"message_type"`
    Payload     json.RawMessage `json:"payload"`
    Timestamp   int64           `json:"timestamp"`
}

// UnmarshalPayload 解析载荷
func (e *MessageEnvelope) UnmarshalPayload(v interface{}) error {
    return json.Unmarshal(e.Payload, v)
}
```

### Hello 处理器实现

```go
// HelloProcessor Hello 消息处理器
type HelloProcessor struct {
    logger *zap.Logger
}

func NewHelloProcessor(logger *zap.Logger) *HelloProcessor {
    return &HelloProcessor{logger: logger}
}

func (p *HelloProcessor) GetSupportedMessageType() string {
    return "hello"
}

func (p *HelloProcessor) ProcessMessage(ctx context.Context, msg BusinessMessage, app *app.App) error {
    // 类型断言获取消息封装
    envelope, ok := msg.(*MessageEnvelope)
    if !ok {
        return fmt.Errorf("invalid message type")
    }

    // 解析 Hello 事件
    var event HelloEvent
    if err := envelope.UnmarshalPayload(&event); err != nil {
        return fmt.Errorf("failed to unmarshal hello event: %w", err)
    }

    // 处理业务逻辑
    return p.handleHelloMessage(ctx, &event, app)
}
```

### 处理器注册

```go
// ProcessorRegistry 处理器注册表
type ProcessorRegistry struct {
    processors map[string]MessageProcessor
    logger     *zap.Logger
}

func (r *ProcessorRegistry) RegisterProcessor(processor MessageProcessor) {
    messageType := processor.GetSupportedMessageType()
    r.processors[messageType] = processor
    r.logger.Info("Registered message processor", zap.String("message_type", messageType))
}

func (r *ProcessorRegistry) GetProcessor(messageType string) (MessageProcessor, bool) {
    processor, exists := r.processors[messageType]
    return processor, exists
}
```

## ⚙️ 配置管理

### 队列配置

在 `configs/config.dev.yaml` 中统一管理队列和交换机配置：

```yaml
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
  exchanges:
    - name: "hello.exchange"
      type: "direct"
      durable: true
      auto_delete: false
  queues:
    - name: "hello.queue"
      durable: true
      auto_delete: false
      exclusive: false
      exchange: "hello.exchange"
      routing_keys: ["hello"]
```

### 配置结构

```go
// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
    URL       string           `yaml:"url"`
    Exchanges []ExchangeConfig `yaml:"exchanges"`
    Queues    []QueueConfig    `yaml:"queues"`
}

// ExchangeConfig 交换机配置
type ExchangeConfig struct {
    Name       string `yaml:"name"`
    Type       string `yaml:"type"`
    Durable    bool   `yaml:"durable"`
    AutoDelete bool   `yaml:"auto_delete"`
}

// QueueConfig 队列配置
type QueueConfig struct {
    Name        string   `yaml:"name"`
    Durable     bool     `yaml:"durable"`
    AutoDelete  bool     `yaml:"auto_delete"`
    Exclusive   bool     `yaml:"exclusive"`
    Exchange    string   `yaml:"exchange"`
    RoutingKeys []string `yaml:"routing_keys"`
}
```

### 自动配置

```go
// SetupInfrastructureFromConfig 根据配置自动创建基础设施
func SetupInfrastructureFromConfig(conn *amqp.Connection, config *config.RabbitMQConfig) error {
    ch, err := conn.Channel()
    if err != nil {
        return fmt.Errorf("failed to open channel: %w", err)
    }
    defer ch.Close()

    // 创建交换机
    for _, exchange := range config.Exchanges {
        err := ch.ExchangeDeclare(
            exchange.Name,
            exchange.Type,
            exchange.Durable,
            exchange.AutoDelete,
            false, // internal
            false, // no-wait
            nil,   // arguments
        )
        if err != nil {
            return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
        }
    }

    // 创建队列并绑定
    for _, queue := range config.Queues {
        _, err := ch.QueueDeclare(
            queue.Name,
            queue.Durable,
            queue.AutoDelete,
            queue.Exclusive,
            false, // no-wait
            nil,   // arguments
        )
        if err != nil {
            return fmt.Errorf("failed to declare queue %s: %w", queue.Name, err)
        }

        // 绑定队列到交换机
        for _, routingKey := range queue.RoutingKeys {
            err := ch.QueueBind(
                queue.Name,
                routingKey,
                queue.Exchange,
                false, // no-wait
                nil,   // arguments
            )
            if err != nil {
                return fmt.Errorf("failed to bind queue %s: %w", queue.Name, err)
            }
        }
    }

    return nil
}
```

## 🛠️ 扩展指南

### 添加新的消息类型

#### 1. 定义消息结构

```go
// internal/model/product.go
type ProductEvent struct {
    ProductID   string `json:"product_id"`
    ProductName string `json:"product_name"`
    Action      string `json:"action"` // created, updated, deleted
    Timestamp   int64  `json:"timestamp"`
}
```

#### 2. 创建消息处理器

```go
// internal/messaging/processors/product_processor.go
type ProductProcessor struct {
    logger *zap.Logger
}

func NewProductProcessor(logger *zap.Logger) *ProductProcessor {
    return &ProductProcessor{logger: logger}
}

func (p *ProductProcessor) GetSupportedMessageType() string {
    return "product"
}

func (p *ProductProcessor) ProcessMessage(ctx context.Context, msg messaging.BusinessMessage, app *app.App) error {
    envelope, ok := msg.(*messaging.MessageEnvelope)
    if !ok {
        return fmt.Errorf("invalid message type")
    }

    var event model.ProductEvent
    if err := envelope.UnmarshalPayload(&event); err != nil {
        return fmt.Errorf("failed to unmarshal product event: %w", err)
    }

    return p.handleProductEvent(ctx, &event, app)
}

func (p *ProductProcessor) handleProductEvent(ctx context.Context, event *model.ProductEvent, app *app.App) error {
    p.logger.Info("Processing product event",
        zap.String("message_id", event.ProductID),
        zap.String("action", event.Action),
    )

    // 处理具体的产品事件逻辑
    switch event.Action {
    case "created":
        return p.handleProductCreated(ctx, event, app)
    case "updated":
        return p.handleProductUpdated(ctx, event, app)
    case "deleted":
        return p.handleProductDeleted(ctx, event, app)
    default:
        return fmt.Errorf("unknown product action: %s", event.Action)
    }
}
```

#### 3. 注册处理器

在 `internal/messaging/consumer/message_consumer.go` 中注册：

```go
func (s *MessageConsumerService) registerEventProcessors() {
    // 注册 Hello 处理器
    s.processorRegistry.RegisterProcessor(
        processors.NewHelloProcessor(s.logger),
    )
    
    // 注册 Product 处理器
    s.processorRegistry.RegisterProcessor(
        processors.NewProductProcessor(s.logger),
    )
}
```

#### 4. 配置队列

在 `configs/config.dev.yaml` 中添加：

```yaml
rabbitmq:
  exchanges:
    - name: "product.exchange"
      type: "direct"
      durable: true
      auto_delete: false
  queues:
    - name: "product.queue"
      durable: true
      auto_delete: false
      exclusive: false
      exchange: "product.exchange"
      routing_keys: ["product"]
```

#### 5. 添加发布 API

```go
// internal/service/product_service.go
func (s *productService) PublishProductEvent(ctx context.Context, event *model.ProductEvent) error {
    messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())
    
    message := struct {
        MessageID   string             `json:"message_id"`
        MessageType string             `json:"message_type"`
        Payload     *model.ProductEvent `json:"payload"`
        Timestamp   int64              `json:"timestamp"`
    }{
        MessageID:   messageID,
        MessageType: "product",
        Payload:     event,
        Timestamp:   time.Now().Unix(),
    }

    body, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    amqpMsg := amqp.Publishing{
        ContentType:  "application/json",
        Body:         body,
        DeliveryMode: amqp.Persistent,
        MessageId:    messageID,
        Timestamp:    time.Now(),
    }

    return s.mqProducer.Publish(ctx, "product.exchange", "product", amqpMsg)
}
```

### 消息处理器动态注册

```go
// GetRegisteredProcessorTypes 动态获取已注册的处理器类型
func (s *MessageConsumerService) GetRegisteredProcessorTypes() []string {
    s.processorRegistry.mu.RLock()
    defer s.processorRegistry.mu.RUnlock()
    
    types := make([]string, 0, len(s.processorRegistry.processors))
    for messageType := range s.processorRegistry.processors {
        types = append(types, messageType)
    }
    
    sort.Strings(types) // 保证顺序一致性
    return types
}
```

## 📝 最佳实践

### 1. 消息持久化

```go
// 重要消息设置持久化
amqpMsg := amqp.Publishing{
    ContentType:  "application/json",
    Body:         body,
    DeliveryMode: amqp.Persistent, // 持久化消息
    MessageId:    messageID,
    Timestamp:    time.Now(),
}
```

### 2. 手动确认

```go
// 消费者使用手动确认
msgs, err := ch.Consume(
    queueName,
    "",    // consumer
    false, // auto-ack = false，使用手动确认
    false, // exclusive
    false, // no-local
    false, // no-wait
    nil,   // args
)

// 处理完成后手动确认
if err := processor.ProcessMessage(ctx, envelope, s.app); err != nil {
    s.logger.Error("Failed to process message", zap.Error(err))
    d.Nack(false, false) // 拒绝消息，不重新入队
} else {
    d.Ack(false) // 确认消息
}
```

### 3. 错误处理和重试

```go
func (p *HelloProcessor) ProcessMessage(ctx context.Context, msg BusinessMessage, app *app.App) error {
    const maxRetries = 3
    
    for i := 0; i < maxRetries; i++ {
        err := p.processWithRetry(ctx, msg, app)
        if err == nil {
            return nil
        }
        
        if i < maxRetries-1 {
            p.logger.Warn("Processing failed, retrying",
                zap.Error(err),
                zap.Int("attempt", i+1),
                zap.Int("max_retries", maxRetries),
            )
            time.Sleep(time.Duration(i+1) * time.Second)
        } else {
            p.logger.Error("Processing failed after max retries", zap.Error(err))
            return err
        }
    }
    
    return nil
}
```

### 4. 监控和指标

```go
// 添加处理指标
func (p *HelloProcessor) ProcessMessage(ctx context.Context, msg BusinessMessage, app *app.App) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        p.logger.Info("Message processed",
            zap.String("message_type", "hello"),
            zap.Duration("duration", duration),
        )
    }()
    
    return p.handleHelloMessage(ctx, msg, app)
}
```

### 5. 配置管理

```go
// 使用环境变量覆盖配置
type RabbitMQConfig struct {
    URL string `yaml:"url" env:"RABBITMQ_URL"`
}

// 在启动时验证配置
func validateConfig(config *RabbitMQConfig) error {
    if config.URL == "" {
        return errors.New("RabbitMQ URL is required")
    }
    
    if len(config.Queues) == 0 {
        return errors.New("at least one queue configuration is required")
    }
    
    return nil
}
```

## 🔍 故障排除

### 常见问题

#### 1. 连接失败
**症状**: 应用启动时报错 "dial tcp connection refused"
**解决方案**:
- 检查 RabbitMQ 服务是否启动: `docker ps` 或 `brew services list`
- 检查端口是否被占用: `lsof -i :5672`
- 验证连接字符串格式: `amqp://user:pass@host:port/`

#### 2. 消息不消费
**症状**: 消息发送成功但消费者没有处理
**解决方案**:
- 检查队列绑定: 访问 RabbitMQ 管理界面
- 验证路由键配置: 确保发布和绑定的路由键一致
- 检查消费者日志: 查看是否有错误信息

#### 3. 消息丢失
**症状**: 消息发送后丢失，队列中没有消息
**解决方案**:
- 设置消息持久化: `DeliveryMode: amqp.Persistent`
- 确保队列持久化: `Durable: true`
- 使用事务或发布确认机制

#### 4. 内存泄漏
**症状**: 应用内存持续增长
**解决方案**:
- 检查连接和通道是否正确关闭
- 使用连接池管理连接
- 监控 goroutine 数量

### 调试工具

#### RabbitMQ 管理界面
- 访问: http://localhost:15672
- 用户名/密码: guest/guest
- 功能:
  - 查看队列状态和消息堆积
  - 监控连接和通道状态
  - 手动发送和接收消息
  - 查看交换机绑定关系

#### 日志分析
```bash
# 查看消费者日志
make mq-consumer

# 查看特定级别日志
grep "ERROR" logs/consumer.log

# 实时监控日志
tail -f logs/consumer.log | grep "Processing message"
```

#### 性能监控
```go
// 添加性能指标
func (s *MessageConsumerService) StartConsumingWithMetrics(queueName string) error {
    // 消息计数器
    messageCounter := 0
    
    // 处理消息
    for d := range msgs {
        messageCounter++
        start := time.Now()
        
        // 处理逻辑...
        
        s.logger.Info("Message metrics",
            zap.Int("total_processed", messageCounter),
            zap.Duration("processing_time", time.Since(start)),
        )
    }
}
```

### 生产环境配置建议

```yaml
# configs/config.prod.yaml
rabbitmq:
  url: "${RABBITMQ_URL}"
  connection_timeout: 30s
  heartbeat: 60s
  
  # 连接池配置
  max_connections: 10
  max_channels_per_connection: 100
  
  # 重试配置
  retry_delay: 5s
  max_retries: 3
  
  # 队列配置
  queues:
    - name: "hello.queue"
      durable: true
      auto_delete: false
      exclusive: false
      exchange: "hello.exchange"
      routing_keys: ["hello"]
      # 生产环境特定配置
      arguments:
        x-message-ttl: 86400000  # 消息TTL 24小时
        x-max-length: 10000      # 队列最大长度
        x-dead-letter-exchange: "dlx.exchange"  # 死信交换机
```

---

## 📚 相关文档

- [Wire 依赖注入架构](WIRE_ARCHITECTURE.md)
- [API 接口文档](../README.md#API接口)
- [配置文件说明](../configs/config.dev.yaml)

**快速开始**: `make mq-consumer` → `make mq-api` → `make test-mq-api` 