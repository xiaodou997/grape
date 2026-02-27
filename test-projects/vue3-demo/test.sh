#!/bin/bash

# Grape Registry Vue3 Demo 快速测试脚本
# 用于自动测试依赖下载功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$SCRIPT_DIR/vue3-demo"
REGISTRY_URL="http://localhost:4874"

echo "🍇 Grape Registry - Vue3 Demo 测试脚本"
echo "======================================"
echo ""

# 检查 Grape 服务是否运行
echo "📋 步骤 1: 检查 Grape Registry 服务..."
if curl -s "$REGISTRY_URL/-/health" | grep -q "ok"; then
    echo "✅ Grape Registry 服务运行正常"
else
    echo "❌ Grape Registry 服务未运行"
    echo "请先启动服务：cd /path/to/grape && ./bin/grape"
    exit 1
fi
echo ""

# 进入项目目录
echo "📋 步骤 2: 进入项目目录..."
cd "$DEMO_DIR"
echo "✅ 当前目录：$(pwd)"
echo ""

# 检查 .npmrc 配置
echo "📋 步骤 3: 检查 .npmrc 配置..."
if [ -f ".npmrc" ]; then
    echo "✅ .npmrc 文件存在"
    echo "配置内容:"
    cat .npmrc
else
    echo "❌ .npmrc 文件不存在"
    exit 1
fi
echo ""

# 清理旧的 node_modules 和 lock 文件
echo "📋 步骤 4: 清理旧的依赖..."
if [ -d "node_modules" ]; then
    echo "删除 node_modules..."
    rm -rf node_modules
fi
if [ -f "package-lock.json" ]; then
    echo "删除 package-lock.json..."
    rm -f package-lock.json
fi
echo "✅ 清理完成"
echo ""

# 安装依赖
echo "📋 步骤 5: 安装依赖（这可能需要几分钟）..."
echo "npm install --registry $REGISTRY_URL"
echo "----------------------------------------"

START_TIME=$(date +%s)

npm install 2>&1 | tee /tmp/grape-test-install.log

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo "----------------------------------------"
echo ""

# 检查安装结果
if [ -d "node_modules" ]; then
    echo "✅ 依赖安装成功！"
    echo ""
    
    # 统计信息
    PKG_COUNT=$(find node_modules -maxdepth 1 -type d | wc -l | tr -d ' ')
    PKG_COUNT=$((PKG_COUNT - 1))  # 减去 node_modules 本身
    
    TOTAL_SIZE=$(du -sh node_modules 2>/dev/null | cut -f1)
    
    echo "📊 安装统计:"
    echo "  - 安装包数量：$PKG_COUNT 个"
    echo "  - 总大小：$TOTAL_SIZE"
    echo "  - 耗时：$DURATION 秒"
    echo ""
    
    # 验证关键依赖
    echo "📋 步骤 6: 验证关键依赖..."
    CRITICAL_DEPS=("vue" "vue-router" "pinia" "axios" "element-plus" "vite" "@vitejs/plugin-vue")
    
    for dep in "${CRITICAL_DEPS[@]}"; do
        if [ -d "node_modules/$dep" ]; then
            echo "  ✅ $dep"
        else
            echo "  ❌ $dep (未找到)"
        fi
    done
    echo ""
    
    # 验证 scoped 包
    echo "📋 步骤 7: 验证 scoped 包..."
    SCOPED_DEPS=("@element-plus/icons-vue" "@types/node")
    
    for dep in "${SCOPED_DEPS[@]}"; do
        if [ -d "node_modules/$dep" ]; then
            echo "  ✅ $dep"
        else
            echo "  ❌ $dep (未找到)"
        fi
    done
    echo ""
    
    # 测试启动项目
    echo "📋 步骤 8: 测试启动开发服务器..."
    echo "提示：运行 'npm run dev' 启动项目"
    echo ""
    
    echo "======================================"
    echo "🎉 测试完成！"
    echo "======================================"
    echo ""
    echo "下一步操作:"
    echo "  1. 运行 'npm run dev' 启动开发服务器"
    echo "  2. 访问 http://localhost:3000 查看应用"
    echo "  3. 验证所有功能正常工作"
    echo ""
    
else
    echo "❌ 依赖安装失败！"
    echo "请查看日志文件：/tmp/grape-test-install.log"
    exit 1
fi
