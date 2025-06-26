#!/bin/bash

# Docker 构建脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
REGISTRY=${REGISTRY:-"skeleton"}
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "latest")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# 服务列表
SERVICES=("api" "scheduler")

# 显示信息
echo -e "${GREEN}🐳 构建 Docker 镜像${NC}"
echo -e "${YELLOW}Registry: ${REGISTRY}${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Git Commit: ${GIT_COMMIT}${NC}"
echo ""

# 检查 Docker 是否运行
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker 未运行或无权限访问${NC}"
    exit 1
fi

# 构建函数
build_service() {
    local service=$1
    local image_name="${REGISTRY}/${service}:${VERSION}"
    local latest_name="${REGISTRY}/${service}:latest"
    
    echo -e "${GREEN}📦 构建服务: ${service}${NC}"
    
    docker build \
        --build-arg SERVICE=${service} \
        --build-arg VERSION=${VERSION} \
        --build-arg BUILD_TIME=${BUILD_TIME} \
        --build-arg GIT_COMMIT=${GIT_COMMIT} \
        --tag ${image_name} \
        --tag ${latest_name} \
        .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ ${service} 构建成功${NC}"
        
        # 显示镜像信息
        echo -e "${YELLOW}镜像大小:${NC}"
        docker images ${image_name} --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
        echo ""
    else
        echo -e "${RED}❌ ${service} 构建失败${NC}"
        return 1
    fi
}

# 构建所有服务
for service in "${SERVICES[@]}"; do
    build_service ${service}
done

# 构建完成
echo -e "${GREEN}🎉 所有服务构建完成!${NC}"
echo ""
echo -e "${YELLOW}可用镜像:${NC}"
for service in "${SERVICES[@]}"; do
    echo "  - ${REGISTRY}/${service}:${VERSION}"
    echo "  - ${REGISTRY}/${service}:latest"
done

echo ""
echo -e "${YELLOW}启动命令:${NC}"
echo "  开发环境: docker compose up -d"
echo "  生产环境: docker compose -f docker compose.yaml -f docker compose.prod.yaml up -d"

# 推送镜像 (可选)
if [ "$PUSH" = "true" ]; then
    echo ""
    echo -e "${GREEN}📤 推送镜像到仓库${NC}"
    for service in "${SERVICES[@]}"; do
        docker push ${REGISTRY}/${service}:${VERSION}
        docker push ${REGISTRY}/${service}:latest
    done
fi 