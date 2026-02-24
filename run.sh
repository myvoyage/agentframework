#!/bin/bash
# Agent Framework - 启动脚本
# 支持三种模式：桌面UI (默认)、TUI、CLI

MODE="${1:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 显示帮助信息
show_help() {
    cat << EOF
Agent Framework - 多界面 AI Agent 框架

用法:
    ./run.sh [选项] [参数]

选项:
    (无参数)     启动桌面 GUI 应用 (Wails)
    --tui, -t    启动终端用户界面 (TUI)
    --cli, -c    进入命令行模式
    --help, -h   显示此帮助信息

命令行模式子命令:
    ./run.sh cli agent list              列出所有 agents
    ./run.sh cli workflow list           列出所有工作流
    ./run.sh cli skill list              列出所有技能
    ./run.sh cli chat <agent_id> <msg>   与 agent 对话

示例:
    ./run.sh                           # 启动桌面应用
    ./run.sh --tui                     # 启动 TUI
    ./run.sh cli agent list            # CLI: 列出 agents
    ./run.sh cli chat default "你好"   # CLI: 对话

更多信息请访问: https://github.com/your-repo/agent-framework
EOF
}

# 根据参数选择模式
case "$MODE" in
    --help|-h|help)
        show_help
        exit 0
        ;;
    --tui|-t)
        echo "🖥️  启动 TUI 模式..."
        if [ -f "tui.exe" ]; then
            exec ./tui.exe
        else
            echo "❌ 错误: tui.exe 不存在，请先运行: go build ./cmd/tui"
            exit 1
        fi
        ;;
    --cli|-c|cli)
        shift
        echo "📟 CLI 模式"
        exec ./agent-cli.exe "$@"
        ;;
    "")
        echo "🖥️  启动桌面应用模式..."
        if [ -f "AgentFramework.exe" ]; then
            exec ./AgentFramework.exe
        else
            echo "❌ 错误: AgentFramework.exe 不存在，请先运行: go build ."
            exit 1
        fi
        ;;
    *)
        echo "❌ 未知选项: $MODE"
        echo "运行 './run.sh --help' 查看帮助"
        exit 1
        ;;
esac
