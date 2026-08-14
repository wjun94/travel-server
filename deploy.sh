#!/usr/bin/env bash
set -euo pipefail

# travel-server 后端一键部署脚本
# 用法: ./deploy.sh
# 构建 Docker 镜像并重启线上 travel-server 容器（db/redis/nginx 不受影响）

cd "$(dirname "$0")"

echo "==> [1/3] 构建后端镜像"
docker compose build backend

echo "==> [2/3] 重启后端容器"
docker compose up -d backend

echo "==> [3/3] 等待服务就绪"
sleep 3
if docker ps --format '{{.Names}}' | grep -q '^travel-server$'; then
  echo "==> 部署完成"
else
  echo "!! 容器未正常运行，请检查日志: docker compose logs backend" >&2
  exit 1
fi
