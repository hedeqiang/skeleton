#!/bin/bash

# Docker环境消费者服务测试脚本
# 用于独立测试Consumer服务

set -e

echo "🐳 清理现有consumer容器..."
docker compose rm -f consumer

echo "🏗️  构建Consumer Docker镜像..."
docker compose build consumer

echo "🚀 确保依赖服务运行中..."
docker compose up -d postgres redis rabbitmq

echo "⏳ 等待依赖服务就绪..."
sleep 5

echo "📊 检查依赖服务状态..."
docker compose ps postgres redis rabbitmq

echo "🚀 启动Consumer服务..."
docker compose up consumer

# 如果Consumer服务启动失败，显示日志
if [ $? -ne 0 ]; then
    echo "❌ Consumer服务启动失败，显示错误日志："
    docker compose logs consumer
fi 