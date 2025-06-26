#!/bin/bash

# 生产环境部署脚本
# 用于部署所有生产服务，包括API、Consumer、Scheduler

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查必要的环境变量
check_env_vars() {
    echo -e "${BLUE}🔍 检查环境变量...${NC}"
    
    required_vars=(
        "POSTGRES_PASSWORD"
        "REDIS_PASSWORD"
        "RABBITMQ_USER"
        "RABBITMQ_PASSWORD"
        "JWT_SECRET"
    )
    
    missing_vars=()
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo -e "${RED}❌ 缺少必要的环境变量:${NC}"
        printf '%s\n' "${missing_vars[@]}"
        echo -e "${YELLOW}请设置这些环境变量后重试${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 环境变量检查通过${NC}"
}

# 构建生产镜像
build_images() {
    echo -e "${BLUE}🏗️  构建生产Docker镜像...${NC}"
    
    # 设置构建参数
    export VERSION=${VERSION:-$(git describe --tags --always --dirty)}
    export BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    export GIT_COMMIT=$(git rev-parse HEAD)
    
    echo -e "${YELLOW}版本信息:${NC}"
    echo -e "  Version: ${VERSION}"
    echo -e "  Build Time: ${BUILD_TIME}"
    echo -e "  Git Commit: ${GIT_COMMIT}"
    
    # 构建所有服务镜像
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml build \
        api consumer scheduler
    
    echo -e "${GREEN}✅ 镜像构建完成${NC}"
}

# 创建必要的目录
create_directories() {
    echo -e "${BLUE}📁 创建必要的目录...${NC}"
    
    sudo mkdir -p /var/log/skeleton
    sudo mkdir -p /var/lib/postgresql/data
    sudo mkdir -p /var/lib/redis
    sudo mkdir -p /var/lib/rabbitmq
    
    # 设置适当的权限
    sudo chown -R 999:999 /var/lib/postgresql/data || true
    sudo chown -R 999:999 /var/lib/redis || true
    sudo chown -R 999:999 /var/lib/rabbitmq || true
    
    echo -e "${GREEN}✅ 目录创建完成${NC}"
}

# 停止现有服务
stop_services() {
    echo -e "${BLUE}🛑 停止现有服务...${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml down || true
    echo -e "${GREEN}✅ 服务已停止${NC}"
}

# 启动生产服务
start_services() {
    echo -e "${BLUE}🚀 启动生产服务...${NC}"
    
    # 首先启动基础设施服务
    echo -e "${YELLOW}启动基础设施服务 (PostgreSQL, Redis, RabbitMQ)...${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d \
        postgres redis rabbitmq
    
    # 等待基础设施服务就绪
    echo -e "${YELLOW}等待基础设施服务就绪...${NC}"
    sleep 20
    
    # 运行数据库迁移
    echo -e "${YELLOW}运行数据库迁移...${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml run --rm migrate || true
    
    # 启动应用服务
    echo -e "${YELLOW}启动应用服务 (API, Consumer, Scheduler)...${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d \
        api consumer scheduler
    
    # 启动监控服务（如果需要）
    if [ "$ENABLE_MONITORING" = "true" ]; then
        echo -e "${YELLOW}启动监控服务 (Prometheus, Grafana)...${NC}"
        docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d \
            prometheus grafana
    fi
    
    echo -e "${GREEN}✅ 生产服务已启动${NC}"
}

# 显示服务状态
show_status() {
    echo -e "${BLUE}📊 服务状态:${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml ps
    
    echo -e "\n${BLUE}🔗 服务端点:${NC}"
    echo -e "  API服务: http://localhost:8080"
    echo -e "  RabbitMQ管理界面: http://localhost:15672"
    if [ "$ENABLE_MONITORING" = "true" ]; then
        echo -e "  Prometheus: http://localhost:9090"
        echo -e "  Grafana: http://localhost:3000"
    fi
}

# 健康检查
health_check() {
    echo -e "${BLUE}🏥 执行健康检查...${NC}"
    
    # 等待服务启动
    sleep 10
    
    # 检查API健康状态
    if curl -f -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✅ API服务健康${NC}"
    else
        echo -e "${RED}❌ API服务不健康${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ 健康检查通过${NC}"
}

# 主函数
main() {
    echo -e "${BLUE}🚀 开始生产环境部署...${NC}"
    
    check_env_vars
    create_directories
    stop_services
    build_images
    start_services
    show_status
    health_check
    
    echo -e "${GREEN}🎉 生产环境部署完成！${NC}"
    echo -e "${YELLOW}💡 使用 'make prod-down' 或 './scripts/docker-prod.sh stop' 停止服务${NC}"
}

# 停止函数
stop() {
    echo -e "${BLUE}🛑 停止生产环境...${NC}"
    docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml down
    echo -e "${GREEN}✅ 生产环境已停止${NC}"
}

# 参数处理
case "${1:-}" in
    "stop")
        stop
        ;;
    *)
        main
        ;;
esac 