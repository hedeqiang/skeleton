#!/bin/bash

# 开发环境启动脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
ENV_FILE=".env"
COMPOSE_FILES="-f docker compose.yaml -f docker compose.override.yaml"
PROJECT_NAME="skeleton"

# 显示帮助信息
show_help() {
    echo -e "${GREEN}🚀 区块链交换项目 - 开发环境管理脚本${NC}"
    echo ""
    echo -e "${YELLOW}用法:${NC}"
    echo "  $0 [命令] [选项]"
    echo ""
    echo -e "${YELLOW}命令:${NC}"
    echo "  up         启动开发环境"
    echo "  down       停止开发环境"
    echo "  restart    重启开发环境"
    echo "  logs       查看日志"
    echo "  ps         查看运行状态"
    echo "  build      重新构建镜像"
    echo "  clean      清理环境"
    echo "  shell      进入容器shell"
    echo "  db         数据库操作"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo "  -h, --help     显示帮助信息"
    echo "  -v, --verbose  详细输出"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo "  $0 up                    # 启动开发环境"
    echo "  $0 logs api              # 查看API服务日志"
    echo "  $0 shell api             # 进入API容器"
    echo "  $0 db migrate            # 运行数据库迁移"
}

# 检查环境
check_environment() {
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker 未安装${NC}"
        exit 1
    fi

    # 检查 Docker Compose
    if ! command -v docker compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose 未安装${NC}"
        exit 1
    fi

    # 检查环境文件
    if [ ! -f "${ENV_FILE}" ]; then
        echo -e "${YELLOW}⚠️  环境文件不存在，从示例文件创建${NC}"
        cp .env.example ${ENV_FILE}
        echo -e "${GREEN}✅ 已创建 ${ENV_FILE}，请根据需要修改配置${NC}"
    fi
}

# 启动服务
start_services() {
    echo -e "${GREEN}🚀 启动开发环境${NC}"
    
    # 创建必要的目录
    mkdir -p logs

    # 启动服务
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} up -d
    
    echo ""
    echo -e "${GREEN}✅ 开发环境启动完成${NC}"
    echo ""
    echo -e "${YELLOW}服务地址:${NC}"
    echo "  • API 服务:        http://localhost:8080"
    echo "  • API 文档:        http://localhost:8080/api/v1/docs"
    echo "  • 数据库管理:      http://localhost:8081"
    echo "  • RabbitMQ 管理:   http://localhost:15672 (admin/admin123)"
    echo ""
    echo -e "${YELLOW}数据库连接信息:${NC}"
    echo "  • 主机: localhost:5432"
    echo "  • 数据库: skeleton"
    echo "  • 用户名: postgres"
    echo "  • 密码: 123456"
}

# 停止服务
stop_services() {
    echo -e "${YELLOW}🛑 停止开发环境${NC}"
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} down
    echo -e "${GREEN}✅ 开发环境已停止${NC}"
}

# 重启服务
restart_services() {
    echo -e "${YELLOW}🔄 重启开发环境${NC}"
    stop_services
    sleep 2
    start_services
}

# 查看日志
show_logs() {
    local service=$1
    if [ -n "$service" ]; then
        docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} logs -f $service
    else
        docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} logs -f
    fi
}

# 查看状态
show_status() {
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} ps
}

# 重新构建
rebuild_services() {
    echo -e "${YELLOW}🔨 重新构建镜像${NC}"
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} build --no-cache
    echo -e "${GREEN}✅ 镜像重新构建完成${NC}"
}

# 清理环境
clean_environment() {
    echo -e "${YELLOW}🧹 清理开发环境${NC}"
    
    # 停止并删除容器
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} down -v --remove-orphans
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 删除未使用的网络
    docker network prune -f
    
    echo -e "${GREEN}✅ 环境清理完成${NC}"
}

# 进入容器 shell
enter_shell() {
    local service=${1:-api}
    echo -e "${BLUE}🐚 进入 $service 容器${NC}"
    docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} exec $service sh
}

# 数据库操作
database_operations() {
    local operation=$1
    case $operation in
        migrate)
            echo -e "${BLUE}📊 运行数据库迁移${NC}"
            docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} run --rm migrate
            ;;
        seed)
            echo -e "${BLUE}🌱 运行数据库种子${NC}"
            docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} run --rm seed
            ;;
        reset)
            echo -e "${YELLOW}⚠️  重置数据库${NC}"
            read -p "确认要重置数据库吗? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} stop postgres
                docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} rm -f postgres
                docker volume rm ${PROJECT_NAME}_postgres_dev_data 2>/dev/null || true
                docker compose ${COMPOSE_FILES} --project-name ${PROJECT_NAME} up -d postgres
                sleep 5
                database_operations migrate
                database_operations seed
            fi
            ;;
        *)
            echo -e "${RED}❌ 未知的数据库操作: $operation${NC}"
            echo "可用操作: migrate, seed, reset"
            ;;
    esac
}

# 主函数
main() {
    check_environment
    
    case $1 in
        up|start)
            start_services
            ;;
        down|stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        logs)
            show_logs $2
            ;;
        ps|status)
            show_status
            ;;
        build)
            rebuild_services
            ;;
        clean)
            clean_environment
            ;;
        shell)
            enter_shell $2
            ;;
        db)
            database_operations $2
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}❌ 未知命令: $1${NC}"
            echo "使用 '$0 --help' 查看可用命令"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 