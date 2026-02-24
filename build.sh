#!/bin/bash
# Agent Framework - 统一构建脚本
# 构建所有组件：主程序、CLI、TUI、示例

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Agent Framework - 统一构建脚本${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 解析参数
BUILD_ALL=true
BUILD_MAIN=false
BUILD_CLI=false
BUILD_TUI=false
BUILD_SERVER=false
BUILD_SIMPLEBOT=false
BUILD_EXAMPLES=false
CLEAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN=true
            shift
            ;;
        --main)
            BUILD_ALL=false
            BUILD_MAIN=true
            shift
            ;;
        --cli)
            BUILD_ALL=false
            BUILD_CLI=true
            shift
            ;;
        --tui)
            BUILD_ALL=false
            BUILD_TUI=true
            shift
            ;;
        --server)
            BUILD_ALL=false
            BUILD_SERVER=true
            shift
            ;;
        --simplebot)
            BUILD_ALL=false
            BUILD_SIMPLEBOT=true
            shift
            ;;
        --examples)
            BUILD_ALL=false
            BUILD_EXAMPLES=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --clean        清理构建产物"
            echo "  --main         只构建主程序"
            echo "  --cli          只构建 CLI 工具"
            echo "  --tui          只构建 TUI"
            echo "  --server       只构建服务器演示"
            echo "  --simplebot    只构建简单机器人"
            echo "  --examples     只构建示例"
            echo "  --help, -h     显示此帮助"
            echo ""
            echo "默认: 构建所有组件"
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            echo "运行 '$0 --help' 查看帮助"
            exit 1
            ;;
    esac
done

# 清理构建产物
if [ "$CLEAN" = true ]; then
    echo -e "${YELLOW}🧹 清理构建产物...${NC}"
    rm -f *.exe
    echo -e "${GREEN}✓ 清理完成${NC}"
    echo ""
fi

# 构建函数
build_component() {
    local name=$1
    local target=$2
    local output=$3

    echo -e "${BLUE}🔨 构建 $name...${NC}"
    if go build -o "$output" $target 2>&1; then
        if [ -f "$output" ]; then
            local size=$(ls -lh "$output" | awk '{print $5}')
            echo -e "${GREEN}✓ $name 构建成功 ($size)${NC}"
        else
            echo -e "${GREEN}✓ $name 构建成功${NC}"
        fi
        return 0
    else
        echo -e "${RED}✗ $name 构建失败${NC}"
        return 1
    fi
}

# 构建主程序
if [ "$BUILD_ALL" = true ] || [ "$BUILD_MAIN" = true ]; then
    build_component "主程序 (AgentFramework.exe)" "." "AgentFramework.exe"
fi

# 构建 CLI 工具
if [ "$BUILD_ALL" = true ] || [ "$BUILD_CLI" = true ]; then
    build_component "CLI 工具 (agent-cli.exe)" "./cmd/cli" "agent-cli.exe"
fi

# 构建 TUI
if [ "$BUILD_ALL" = true ] || [ "$BUILD_TUI" = true ]; then
    build_component "TUI (tui.exe)" "./cmd/tui" "tui.exe"
fi

# 构建服务器演示
if [ "$BUILD_ALL" = true ] || [ "$BUILD_SERVER" = true ]; then
    build_component "服务器演示 (server_demo.exe)" "./cmd/server_demo" "server_demo.exe"
fi

# 构建简单机器人
if [ "$BUILD_ALL" = true ] || [ "$BUILD_SIMPLEBOT" = true ]; then
    build_component "简单机器人 (simplebot.exe)" "./cmd/simplebot" "simplebot.exe"
fi

# 构建示例
if [ "$BUILD_ALL" = true ] || [ "$BUILD_EXAMPLES" = true ]; then
    echo -e "${BLUE}🔨 构建示例...${NC}"

    # 尝试构建各个示例
    for example_dir in examples/*/; do
        if [ -d "$example_dir" ]; then
            example_name=$(basename "$example_dir")
            main_file="$example_dir/main.go"

            if [ -f "$main_file" ]; then
                if go build -o "examples/${example_name}/${example_name}.exe" "$main_file" 2>/dev/null; then
                    echo -e "${GREEN}✓ ${example_name} 构建成功${NC}"
                else
                    echo -e "${YELLOW}⚠ ${example_name} 需要额外依赖或修复${NC}"
                fi
            fi
        fi
    done
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}🎉 构建完成！${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "可执行文件:"
ls -lh *.exe 2>/dev/null | awk '{printf "  • %s (%s)\n", $9, $5}' | grep -E "\.exe"
echo ""
echo "运行方式:"
echo "  • 桌面应用: ./run.sh 或 ./AgentFramework.exe"
echo "  • TUI:       ./run.sh --tui 或 ./tui.exe"
echo "  • CLI:       ./run.sh cli --help 或 ./agent-cli.exe"
echo ""
