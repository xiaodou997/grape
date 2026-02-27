#!/bin/bash

# Grape 测试项目安装脚本
# 用于快速测试 Grape npm 仓库功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/vue3-demo"

echo "🍇 Grape 测试项目安装脚本"
echo "================================"
echo ""

# 检查 Grape 服务是否运行
echo "📡 检查 Grape 服务状态..."
if curl -s http://localhost:4873/-/health > /dev/null 2>&1; then
    echo "✅ Grape 服务正在运行"
else
    echo "❌ Grape 服务未运行"
    echo ""
    echo "请先启动 Grape 服务："
    echo "  cd /path/to/grape"
    echo "  ./bin/grape"
    echo ""
    exit 1
fi

echo ""
echo "📦 进入测试项目目录..."
cd "$PROJECT_DIR"

echo ""
echo "🔍 检查 .npmrc 配置..."
if [ -f ".npmrc" ]; then
    echo "✅ .npmrc 文件存在"
    echo "   配置内容:"
    cat .npmrc | sed 's/^/   /'
else
    echo "❌ .npmrc 文件不存在"
    exit 1
fi

echo ""
echo "📥 开始安装依赖..."
echo "   这将从 Grape 下载包（第一次可能较慢）"
echo ""

# 安装依赖
npm install

echo ""
echo "✅ 依赖安装完成！"
echo ""

# 显示安装的依赖
echo "📦 已安装的依赖:"
npm list --depth=0

echo ""
echo "🚀 启动开发服务器..."
echo "   访问：http://localhost:5173"
echo ""

# 启动开发服务器
npm run dev
