# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
WIRE=wire

# 二进制文件名
API_BINARY=skeleton_api
CONSUMER_BINARY=skeleton_consumer
SCHEDULER_BINARY=skeleton_scheduler

# 构建目录
BUILD_DIR=build

# 默认目标
.PHONY: all
all: clean wire build

# === 代码生成 ===
.PHONY: wire
wire:
	@echo "🔄 生成 Wire 依赖注入代码..."
	cd internal/wire && $(WIRE)

# === 构建命令 ===
.PHONY: build
build: wire
	@echo "🔨 构建所有服务..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(API_BINARY) -v ./cmd/api
	$(GOBUILD) -o $(BUILD_DIR)/$(CONSUMER_BINARY) -v ./cmd/consumer
	$(GOBUILD) -o $(BUILD_DIR)/$(SCHEDULER_BINARY) -v ./cmd/scheduler

.PHONY: api
api: wire
	@echo "🔨 构建 API 服务..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(API_BINARY) -v ./cmd/api

.PHONY: consumer
consumer: wire
	@echo "🔨 构建消费者服务..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(CONSUMER_BINARY) -v ./cmd/consumer

.PHONY: scheduler
scheduler: wire
	@echo "🔨 构建调度器服务..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SCHEDULER_BINARY) -v ./cmd/scheduler

# === 运行命令 ===
.PHONY: run
run: wire
	@echo "🚀 启动 API 服务..."
	$(GOCMD) run ./cmd/api

.PHONY: run-consumer
run-consumer: wire
	@echo "🚀 启动消费者服务..."
	$(GOCMD) run ./cmd/consumer

.PHONY: run-scheduler
run-scheduler: wire
	@echo "🚀 启动调度器服务..."
	$(GOCMD) run ./cmd/scheduler

# === Docker 命令 ===
.PHONY: up
up:
	@echo "🚀 启动 Docker 环境..."
	docker compose up -d

.PHONY: down
down:
	@echo "🛑 停止 Docker 环境..."
	docker compose down

.PHONY: restart
restart:
	@echo "🔄 重启 Docker 环境..."
	docker compose restart

.PHONY: logs
logs:
	@echo "📋 查看服务日志..."
	docker compose logs -f

.PHONY: ps
ps:
	@echo "📊 查看容器状态..."
	docker compose ps

.PHONY: shell
shell:
	@echo "🐚 进入 API 容器..."
	docker compose exec api sh

.PHONY: db-shell
db-shell:
	@echo "🐚 进入数据库容器..."
	docker compose exec postgres psql -U postgres -d skeleton

# === Docker 构建 ===
.PHONY: docker-build
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	@./scripts/docker-build.sh

# === 数据库操作 ===
.PHONY: migrate
migrate:
	@echo "📊 运行数据库迁移..."
	docker compose run --rm migrate

.PHONY: seed
seed:
	@echo "🌱 运行数据库种子..."
	docker compose run --rm seed

.PHONY: db-reset
db-reset:
	@echo "🗑️ 重置数据库..."
	docker compose down postgres
	docker volume rm skeleton_postgres_data || true
	docker compose up -d postgres
	@sleep 5
	@make migrate
	@make seed

# === 测试命令 ===
.PHONY: test
test: wire
	@echo "🧪 运行测试..."
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: wire
	@echo "🧪 运行测试（覆盖率）..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "📊 覆盖率报告: coverage.html"

# === 代码质量 ===
.PHONY: fmt
fmt:
	@echo "🎨 格式化代码..."
	gofmt -s -w .

.PHONY: lint
lint: wire
	@echo "🔍 代码检查..."
	golangci-lint run

.PHONY: vet
vet: wire
	@echo "🔍 代码静态分析..."
	$(GOCMD) vet ./...

# === 依赖管理 ===
.PHONY: deps
deps:
	@echo "📦 更新依赖..."
	$(GOMOD) tidy
	$(GOMOD) download

.PHONY: tools
tools:
	@echo "🛠️ 安装开发工具..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/google/wire/cmd/wire

# === 清理命令 ===
.PHONY: clean
clean:
	@echo "🧹 清理构建产物..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f internal/wire/wire_gen.go

.PHONY: clean-docker
clean-docker:
	@echo "🧹 清理 Docker 环境..."
	docker compose down -v
	docker system prune -f
	docker images | grep skeleton | awk '{print $$3}' | xargs docker rmi -f 2>/dev/null || true

# === 生产环境 ===
.PHONY: prod
prod:
	@echo "🚀 启动生产环境..."
	docker compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d

.PHONY: prod-down
prod-down:
	@echo "🛑 停止生产环境..."
	docker compose -f docker-compose.yaml -f docker-compose.prod.yaml down

# === API 测试 ===
.PHONY: test-api
test-api:
	@echo "🧪 测试 API 端点..."
	@echo "健康检查:"
	@curl -s http://localhost:8080/health | jq . || curl -s http://localhost:8080/health
	@echo "\n用户API测试:"
	@curl -s -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-d '{"username": "testuser", "email": "test@example.com"}' | jq . || true

.PHONY: test-mq
test-mq:
	@echo "🧪 测试消息队列..."
	@curl -s -X POST http://localhost:8080/api/v1/hello/publish \
		-H "Content-Type: application/json" \
		-d '{"content": "Hello from test!", "sender": "test-user"}' | jq . || true

# === 帮助信息 ===
.PHONY: help
help:
	@echo "🎯 Skeleton 项目命令帮助"
	@echo ""
	@echo "📋 基础命令:"
	@echo "  help          显示此帮助信息"
	@echo "  build         构建所有服务"
	@echo "  api           构建 API 服务"
	@echo "  consumer      构建消费者服务"
	@echo "  scheduler     构建调度器服务"
	@echo ""
	@echo "🚀 运行命令:"
	@echo "  run           运行 API 服务"
	@echo "  run-consumer  运行消费者服务"
	@echo "  run-scheduler 运行调度器服务"
	@echo ""
	@echo "🐳 Docker 命令:"
	@echo "  up            启动 Docker 环境"
	@echo "  down          停止 Docker 环境"
	@echo "  restart       重启 Docker 环境"
	@echo "  logs          查看服务日志"
	@echo "  ps            查看容器状态"
	@echo "  shell         进入 API 容器"
	@echo "  db-shell      进入数据库容器"
	@echo "  docker-build  构建 Docker 镜像"
	@echo ""
	@echo "📊 数据库命令:"
	@echo "  migrate       运行数据库迁移"
	@echo "  seed          运行数据库种子"
	@echo "  db-reset      重置数据库"
	@echo ""
	@echo "🧪 测试命令:"
	@echo "  test          运行测试"
	@echo "  test-coverage 运行测试（覆盖率）"
	@echo "  test-api      测试 API 端点"
	@echo "  test-mq       测试消息队列"
	@echo ""
	@echo "🔍 代码质量:"
	@echo "  fmt           格式化代码"
	@echo "  lint          代码检查"
	@echo "  vet           代码静态分析"
	@echo ""
	@echo "🛠️ 工具命令:"
	@echo "  wire          生成依赖注入代码"
	@echo "  deps          更新依赖"
	@echo "  tools         安装开发工具"
	@echo "  clean         清理构建产物"
	@echo "  clean-docker  清理 Docker 环境"
	@echo ""
	@echo "🚀 生产环境:"
	@echo "  prod          启动生产环境"
	@echo "  prod-down     停止生产环境" 