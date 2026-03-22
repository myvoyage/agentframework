#!/bin/bash
# AgentFramework 多模式快速启动脚本

VERSION="1.2.0"
BINARY="./build/agentframework_multimode.exe"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
AgentFramework v${VERSION} - 多模式 AI 代理框架

使用方法:
    $0 [模式] [选项]

模式:
    ui, --ui, -ui         启动 Wails 桌面 GUI (默认)
    tui, --tui, -tui      启动终端用户界面 (TUI)
    cli, --cli, -cli      进入命令行模式 (CLI)

CLI 常用命令:
    agent list            列出所有 agents
    agent chat "msg"      与 agent 对话
    workflow list         列出工作流
    skill list            列出技能
    config get            查看配置
    init                  初始化配置
    version               显示版本

选项:
    -h, --help            显示此帮助信息
    -v, --verbose         详细输出
    -c, --config <file>   指定配置文件

示例:
    $0                    启动桌面 GUI
    $0 -tui               启动 TUI 界面
    $0 cli agent list     列出 agents (CLI 模式)
    $0 -tui               启动 TUI

更多信息请查看: docs/MULTIMODE_USAGE.md

EOF
}

# 检查二进制文件是否存在
check_binary() {
    if [ ! -f "$BINARY" ]; then
        print_warning "二进制文件不存在，正在编译..."
        mkdir -p build
        go build -v -o "$BINARY"
        if [ $? -eq 0 ]; then
            print_success "编译成功"
        else
            print_error "编译失败"
            exit 1
        fi
    fi
}

# 主函数
main() {
    # 没有参数时显示帮助
    if [ $# -eq 0 ]; then
        show_help
        echo ""
        print_info "未指定模式，将使用默认模式 (UI)"
        print_info "使用 '$0 -h' 查看帮助信息"
        echo ""
        check_binary
        exec "$BINARY"
    fi

    # 解析参数
    case "$1" in
        -h|--help|help)
            show_help
            exit 0
            ;;
        -v|--version|version)
            echo "AgentFramework v${VERSION}"
            exit 0
            ;;
        ui|--ui|-ui)
            print_info "启动 UI 模式..."
            check_binary
            exec "$BINARY"
            ;;
        tui|--tui|-tui)
            print_info "启动 TUI 模式..."
            check_binary
            exec "$BINARY" -tui
            ;;
        cli|--cli|-cli)
            shift
            print_info "启动 CLI 模式..."
            check_binary
            exec "$BINARY" -cli "$@"
            ;;
        *)
            # 尝试作为 CLI 命令执行
            print_info "执行 CLI 命令: $@"
            check_binary
            exec "$BINARY" "$@"
            ;;
    esac
}

# 运行主函数
main "$@"
