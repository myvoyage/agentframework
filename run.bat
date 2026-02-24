@echo off
REM Agent Framework - Windows 启动脚本
REM 支持三种模式：桌面UI (默认)、TUI、CLI

setlocal enabledelayedexpansion

set "MODE=%~1"
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM 显示帮助信息
if "%MODE%"=="--help" goto :show_help
if "%MODE%"=="-h" goto :show_help
if "%MODE%"=="help" goto :show_help

REM 根据参数选择模式
if "%MODE%"=="--tui" goto :run_tui
if "%MODE%"=="-t" goto :run_tui
if "%MODE%"=="--cli" goto :run_cli
if "%MODE%"=="-c" goto :run_cli
if "%MODE%"=="cli" goto :run_cli
if "%MODE%"=="" goto :run_desktop

echo ❌ 未知选项: %MODE%
echo 运行 'run.bat --help' 查看帮助
exit /b 1

:show_help
echo Agent Framework - 多界面 AI Agent 框架
echo.
echo 用法:
echo     run.bat [选项] [参数]
echo.
echo 选项:
echo     (无参数)     启动桌面 GUI 应用 (Wails)
echo     --tui, -t    启动终端用户界面 (TUI)
echo     --cli, -c    进入命令行模式
echo     --help, -h   显示此帮助信息
echo.
echo 命令行模式子命令:
echo     run.bat cli agent list              列出所有 agents
echo     run.bat cli workflow list           列出所有工作流
echo     run.bat cli skill list              列出所有技能
echo     run.bat cli chat default ^"你好^"   与 agent 对话
echo.
echo 示例:
echo     run.bat                           启动桌面应用
echo     run.bat --tui                     启动 TUI
echo     run.bat cli agent list            CLI: 列出 agents
echo.
exit /b 0

:run_tui
echo 🖥️  启动 TUI 模式...
if exist "tui.exe" (
    start "" tui.exe
    exit /b 0
) else (
    echo ❌ 错误: tui.exe 不存在，请先运行: go build ./cmd/tui
    exit /b 1
)

:run_cli
shift
echo 📟 CLI 模式
if exist "agent-cli.exe" (
    agent-cli.exe %*
    exit /b %ERRORLEVEL%
) else (
    echo ❌ 错误: agent-cli.exe 不存在，请先运行: go build ./cmd/cli
    exit /b 1
)

:run_desktop
echo 🖥️  启动桌面应用模式...
if exist "AgentFramework.exe" (
    start "" AgentFramework.exe
    exit /b 0
) else (
    echo ❌ 错误: AgentFramework.exe 不存在，请先运行: go build .
    exit /b 1
)
