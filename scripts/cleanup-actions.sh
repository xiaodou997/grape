#!/bin/bash

# GitHub Actions 运行记录清理脚本
# 用法：./cleanup-actions.sh [选项]

set -e

REPO="xiaodou997/grape"
DRY_RUN=true

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_usage() {
    echo "用法：$0 [选项]"
    echo ""
    echo "选项:"
    echo "  -a, --all           删除所有运行记录"
    echo "  -f, --failed        删除所有失败的运行记录"
    echo "  -s, --success       删除所有成功的运行记录"
    echo "  -l, --limit NUM     限制删除数量 (默认：全部)"
    echo "  -e, --execute       实际执行删除 (默认：仅预览)"
    echo "  -h, --help          显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 -a -e            删除所有运行记录"
    echo "  $0 -f -l 10 -e      删除最近 10 个失败的运行记录"
    echo "  $0 -s --execute     删除所有成功的运行记录"
}

delete_runs() {
    local run_ids=$1
    local count=0
    
    if [ -z "$run_ids" ]; then
        echo -e "${YELLOW}没有找到符合条件的运行记录${NC}"
        return
    fi
    
    echo "$run_ids" | while read -r run_id; do
        if [ -n "$run_id" ]; then
            count=$((count + 1))
            if [ "$DRY_RUN" = true ]; then
                echo -e "${YELLOW}[预览] 将删除运行记录：$run_id${NC}"
            else
                echo -e "${GREEN}[删除] 正在删除运行记录：$run_id${NC}"
                gh run delete "$run_id"
            fi
        fi
    done
    
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}预览模式：共找到 $count 个运行记录${NC}"
        echo -e "${YELLOW}使用 -e 参数实际执行删除${NC}"
    else
        echo -e "${GREEN}完成！共删除 $count 个运行记录${NC}"
    fi
}

# 解析参数
LIMIT=""
FILTER=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--all)
            FILTER="all"
            shift
            ;;
        -f|--failed)
            FILTER="failed"
            shift
            ;;
        -s|--success)
            FILTER="success"
            shift
            ;;
        -l|--limit)
            LIMIT="$2"
            shift 2
            ;;
        -e|--execute)
            DRY_RUN=false
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo "未知选项：$1"
            print_usage
            exit 1
            ;;
    esac
done

# 检查 gh 是否安装
if ! command -v gh &> /dev/null; then
    echo -e "${RED}错误：GitHub CLI (gh) 未安装${NC}"
    echo "安装方法：brew install gh (macOS) 或查看 https://cli.github.com/"
    exit 1
fi

# 检查是否已认证
if ! gh auth status &> /dev/null; then
    echo -e "${RED}错误：未认证 GitHub${NC}"
    echo "运行 gh auth login 进行认证"
    exit 1
fi

echo "========================================="
echo "GitHub Actions 运行记录清理工具"
echo "========================================="
echo "仓库：$REPO"
echo "模式：$([ "$DRY_RUN" = true ] && echo "预览" || echo "执行")"
echo "筛选：${FILTER:-all}"
echo "限制：${LIMIT:-无}"
echo "========================================="
echo ""

# 获取运行记录 ID
case $FILTER in
    all)
        if [ -n "$LIMIT" ]; then
            RUN_IDS=$(gh run list --repo $REPO --limit $LIMIT --json databaseId --jq '.[].databaseId')
        else
            RUN_IDS=$(gh run list --repo $REPO --json databaseId --jq '.[].databaseId')
        fi
        ;;
    failed)
        if [ -n "$LIMIT" ]; then
            RUN_IDS=$(gh run list --repo $REPO --status failure --limit $LIMIT --json databaseId --jq '.[].databaseId')
        else
            RUN_IDS=$(gh run list --repo $REPO --status failure --json databaseId --jq '.[].databaseId')
        fi
        ;;
    success)
        if [ -n "$LIMIT" ]; then
            RUN_IDS=$(gh run list --repo $REPO --status success --limit $LIMIT --json databaseId --jq '.[].databaseId')
        else
            RUN_IDS=$(gh run list --repo $REPO --status success --json databaseId --jq '.[].databaseId')
        fi
        ;;
    *)
        RUN_IDS=$(gh run list --repo $REPO --json databaseId --jq '.[].databaseId')
        ;;
esac

delete_runs "$RUN_IDS"

echo ""
echo "========================================="
echo "完成！"
echo "========================================="
