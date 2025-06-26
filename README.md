# Skeleton - 企业级 Go Web 应用模板

这是一个功能完整的企业级 Go Web 应用模板，集成了现代 Go 开发的最佳实践和常用中间件。项目采用分层架构设计，具有良好的可扩展性和可维护性。

## 🚀 项目特性

### 核心技术栈
- **Web 框架**: Gin
- **数据库**: GORM (支持 MySQL、PostgreSQL)
- **缓存**: Redis
- **消息队列**: RabbitMQ
- **配置管理**: Viper
- **日志**: Zap
- **参数验证**: Validator
- **依赖注入**: Wire

### 架构特点
- **依赖注入**: 完全的 DI 模式，使用 Wire 进行代码生成
- **分层架构**: Handler -> Service -> Repository 清晰分层
- **中间件栈**: Recovery、CORS、RequestID、Logger 等完整支持
- **优雅启停**: 完整的生命周期管理
- **统一响应**: 标准化的 API 返回格式
- **配置化管理**: 多环境配置支持
- **消息队列**: 生产者和消费者分离架构

## 📁 项目结构

```
skeleton/
├── cmd/                          # 应用程序入口
│   ├── api/                     # API 服务
│   └── consumer/                # 消息消费者服务
├── configs/                     # 配置文件
│   └── config.dev.yaml         # 开发环境配置
├── internal/                    # 内部代码
│   ├── app/                    # 应用容器
│   ├── config/                 # 配置管理
│   ├── handler/v1/             # HTTP 处理器
│   ├── messaging/              # 消息处理
│   │   ├── consumer/           # 消息消费者
│   │   └── processors/         # 消息处理器
│   ├── middleware/             # 中间件
│   ├── model/                  # 数据模型
│   ├── repository/             # 数据访问层
│   ├── router/                 # 路由配置
│   ├── service/                # 业务逻辑层
│   └── wire/                   # 依赖注入
├── pkg/                        # 公共包
│   ├── database/              # 数据库连接
│   ├── logger/                # 日志工具
│   ├── mq/                    # 消息队列
│   ├── redis/                 # Redis 客户端
│   └── response/              # 响应工具
├── docs/                       # 文档
├── scripts/                    # 脚本文件
│   ├── migrate/               # 数据库迁移
│   └── seed/                  # 种子数据
├── Makefile                    # 构建工具
└── README.md                   # 项目文档
```

## 🛠️ 快速开始

### 1. 环境要求
- Go 1.21+
- MySQL 8.0+ 或 PostgreSQL 13+
- Redis 6.0+
- RabbitMQ 3.8+

### 2. 安装依赖
```bash
git clone https://github.com/hedeqiang/skeleton.git
cd skeleton
go mod tidy
```

### 3. 配置文件
项目使用 `configs/config.dev.yaml` 配置文件，根据需要修改：

```yaml
# 数据库配置
databases:
  default:
    type: "mysql"
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

# Redis 配置
redis:
  addr: "localhost:6379"
  password: ""
  db: 0

# RabbitMQ 配置
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
```

### 4. 数据库迁移
```bash
make db-migrate
```

### 5. 启动服务
```bash
# 启动 API 服务
make run-api

# 启动消费者服务 (可选)
make run-consumer
```

### 6. 访问服务
- API 服务: http://localhost:8080
- 健康检查: http://localhost:8080/ping

## 📊 核心功能模块

### 🌐 Web API
- RESTful API 设计
- 统一的错误处理和响应格式
- 参数验证和数据绑定
- 中间件支持 (CORS、日志、恢复等)

### 🗄️ 数据库
- GORM ORM 支持
- 多数据源配置
- 自动迁移和种子数据
- 连接池管理

### 📨 消息队列
- RabbitMQ 集成
- 生产者和消费者分离
- 配置化队列管理
- 消息处理器模式

> 详细使用说明请参考: [消息队列文档](docs/MESSAGE_QUEUE.md)

### 🔧 Redis 缓存
- Redis 客户端封装
- 连接管理
- 支持各种数据类型操作

### 📝 日志系统
- 结构化日志记录 (Zap)
- 请求追踪 (Request ID)
- 多级别日志输出
- JSON 格式支持

### ⚙️ 配置管理
- 多环境配置支持
- 环境变量覆盖
- 实时配置重载
- 类型安全的配置绑定

## 🔌 API 接口

### 用户管理
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users/:id` - 获取用户信息
- `PUT /api/v1/users/:id` - 更新用户信息
- `DELETE /api/v1/users/:id` - 删除用户
- `GET /api/v1/users` - 获取用户列表

### 消息队列
- `POST /api/v1/hello/publish` - 发布消息到队列

### 系统
- `GET /ping` - 服务健康检查

## 💡 使用示例

### 创建用户
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 发布消息
```bash
curl -X POST http://localhost:8080/api/v1/hello/publish \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, World!",
    "sender": "user123"
  }'
```

### 获取用户列表
```bash
curl "http://localhost:8080/api/v1/users?page=1&page_size=10"
```

## 🔧 开发工具

### Makefile 命令
```bash
# 构建相关
make build           # 构建所有二进制文件
make build-api       # 构建 API 服务
make build-consumer  # 构建消费者服务
make clean           # 清理构建产物

# 运行相关
make run-api         # 运行 API 服务
make run-consumer    # 运行消费者服务
make mq-api          # 运行消息队列 API 服务
make mq-consumer     # 运行消息队列消费者

# 代码质量
make test            # 运行测试
make test-coverage   # 运行测试并生成覆盖率报告
make fmt             # 格式化代码
make lint            # 代码检查
make vet             # 代码静态分析

# 数据库操作
make db-migrate      # 数据库迁移
make db-seed         # 创建种子数据

# 依赖管理
make wire            # 生成依赖注入代码
make deps            # 更新依赖
make install-tools   # 安装开发工具

# Docker
make up       # 启动 Docker 服务
make down     # 停止 Docker 服务

# 帮助信息
make help            # 显示所有可用命令
```

## 🏗️ 架构设计

### 依赖注入
使用 Wire 自动生成依赖注入代码，保证组件解耦：
```go
// 典型的依赖链
Database -> Repository -> Service -> Handler
```

### 中间件链
```go
r.Use(middleware.RequestID())      // 请求 ID
r.Use(middleware.NewLogger(logger)) // 日志记录
r.Use(middleware.NewRecovery(logger)) // 错误恢复
r.Use(middleware.CORS())           // 跨域处理
```

### 统一响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {...},
  "request_id": "uuid"
}
```

### 错误处理
- 统一的错误处理机制
- 结构化错误信息
- HTTP 状态码映射
- 错误日志记录

## 📚 详细文档

- [消息队列使用指南](docs/MESSAGE_QUEUE.md) - RabbitMQ 完整使用指南
- [Wire 架构文档](docs/WIRE_ARCHITECTURE.md) - 依赖注入架构

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 测试特定功能
make test-mq-api    # 测试消息队列 API
```

## 📦 部署

### 本地开发
```bash
make run-api         # 启动 API 服务
make run-consumer    # 启动消费者服务 (可选)
```

### Docker 部署
```bash
make docker-up       # 使用 Docker Compose 启动所有服务
```

### 生产构建
```bash
make build           # 构建生产版本二进制文件
```

## 🔍 监控和调试

### 健康检查
- API 服务: `GET /ping`
- 数据库连接状态检查
- Redis 连接状态检查
- RabbitMQ 连接状态检查

### 日志监控
- 结构化 JSON 日志
- 请求 ID 追踪
- 错误栈跟踪
- 性能指标记录

### 调试工具
- RabbitMQ 管理界面: http://localhost:15672
- 详细的错误信息和堆栈跟踪
- 开发模式下的详细日志

## 🛠️ 扩展指南

### 添加新的 API 端点
1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问层
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/handler/` 中实现 HTTP 处理器
5. 在 `internal/router/` 中注册路由
6. 在 `internal/wire/` 中配置依赖注入

### 添加新的中间件
1. 在 `internal/middleware/` 中创建中间件文件
2. 在路由中注册中间件

### 添加新的消息处理器
1. 在 `internal/messaging/processors/` 中创建处理器
2. 在配置文件中添加队列配置
3. 注册到消息消费服务

## 📄 许可证

本项目基于 MIT 许可证开源。

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

---

**快速开始**: `make run-api` → 访问 http://localhost:8080/ping 

## 🐳 Docker 部署

### 快速开始

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd skeleton
   ```

2. **配置环境变量**
   ```bash
   cp .env.example .env
   # 根据需要修改 .env 文件中的配置
   ```

3. **启动开发环境**
   ```bash
   # 使用 Make 命令
   make docker-up
   
   # 或使用脚本
   ./scripts/docker-dev.sh up
   ```

4. **访问服务**
   - API 服务: http://localhost:8080
   - API 文档: http://localhost:8080/api/v1/docs
   - 数据库管理: http://localhost:8081
   - RabbitMQ 管理: http://localhost:15672 (admin/admin123)

### Docker 命令

#### 开发环境管理
```bash
# 启动开发环境
make docker-up

# 停止开发环境
make docker-down

# 重启开发环境
make docker-restart

# 查看服务日志
make docker-logs

# 查看容器状态
make docker-ps

# 进入 API 容器
make docker-shell

# 清理环境
make docker-clean
```

#### 数据库操作
```bash
# 运行数据库迁移
make docker-migrate

# 运行数据库种子
make docker-seed

# 重置数据库
./scripts/docker-dev.sh db reset
```

#### 镜像构建
```bash
# 构建所有服务镜像
make docker-build

# 构建特定服务
docker build --build-arg SERVICE=api -t skeleton/api:latest .
docker build --build-arg SERVICE=scheduler -t skeleton/scheduler:latest .
```

### 部署模式

#### 开发环境
```bash
# 启动开发环境（包含热重载）
docker compose -f docker compose.yaml -f docker compose.override.yaml up -d
```

#### 生产环境
```bash
# 构建生产镜像
make docker-build

# 启动生产环境
make docker-prod

# 或手动启动
docker compose -f docker compose.yaml -f docker compose.prod.yaml up -d
```

### 服务配置

#### 端口映射
- API 服务: 8080
- PostgreSQL: 5432
- Redis: 6379
- RabbitMQ: 5672 (AMQP), 15672 (管理界面)
- Adminer: 8081

#### 环境变量
主要环境变量配置（详见 `.env.example`）：
```bash
# 应用配置
APP_ENV=development
API_PORT=8080

# 数据库配置
POSTGRES_PASSWORD=123456

# Redis 配置
REDIS_PASSWORD=redis123

# RabbitMQ 配置
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin123
```

### 健康检查

所有服务都配置了健康检查：
- API: `curl http://localhost:8080/health`
- PostgreSQL: `pg_isready`
- Redis: `redis-cli ping`
- RabbitMQ: `rabbitmq-diagnostics ping`

### 监控和日志

#### 日志管理
```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f api
docker compose logs -f scheduler
```

#### 生产环境监控
生产环境包含 Prometheus 和 Grafana：
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### 故障排查

#### 常见问题
1. **端口占用**
   ```bash
   # 检查端口占用
   lsof -i :8080
   
   # 修改端口映射
   # 编辑 .env 文件中的端口配置
   ```

2. **容器启动失败**
   ```bash
   # 查看容器日志
   docker compose logs <service_name>
   
   # 重新构建镜像
   docker compose build --no-cache <service_name>
   ```

3. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker compose ps postgres
   
   # 重启数据库
   docker compose restart postgres
   ```

#### 调试模式
```bash
# 以调试模式启动 API 服务（包含 Delve 调试器）
docker compose -f docker compose.yaml -f docker compose.override.yaml up api

# 连接调试器
dlv connect localhost:2345
```

### 安全考虑

1. **生产环境配置**
   - 修改默认密码
   - 配置防火墙规则
   - 限制网络访问

2. **敏感信息管理**
   - 使用环境变量注入敏感配置
   - 不要将 `.env` 文件提交到版本控制

3. **容器安全**
   - 使用非 root 用户运行应用
   - 定期更新基础镜像
   - 扫描镜像漏洞