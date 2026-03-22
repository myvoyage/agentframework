@echo off
REM AgentFramework 多模式快速启动脚本 (Windows)

SETLOCAL ENABLEDELAYEDEXPANSION

SET VERSION=1.2.0
SET BINARY=build\agentframework_multimode.exe

REM 颜色设置（Windows 10+）
SET "INFO=[INFO]"
SET "SUCCESS=[SUCCESS]"
SET "WARNING=[WARNING]"
SET "ERROR=[ERROR]"

REM 显示帮助信息
IF "%1"=="-h" GOTO :show_help
IF "%1"=="--help" GOTO :show_help
IF "%1"=="help" GOTO :show_help
IF "%1"=="version" GOTO :show_version
IF "%1"=="-v" GOTO :show_version
IF "%1"=="--version" GOTO :show_version

REM 检查二进制文件
IF NOT EXIST "%BINARY%" (
    ECHO %WARNING% 二进制文件不存在，正在编译...
    IF NOT EXIST build mkdir build
    go build -v -o "%BINARY%"
    IF ERRORLEVEL 1 (
        ECHO %ERROR% 编译失败
        EXIT /B 1
    )
    ECHO %SUCCESS% 编译成功
)

REM 解析模式参数
IF "%1"=="ui" GOTO :run_ui
IF "%1"=="--ui" GOTO :run_ui
IF "%1"=="-ui" GOTO :run_ui
IF "%1"=="tui" GOTO :run_tui
IF "%1"=="--tui" GOTO :run_tui
IF "%1"=="-tui" GOTO :run_tui
IF "%1"=="cli" GOTO :run_cli
IF "%1"=="--cli" GOTO :run_cli
IF "%1"=="-cli" GOTO :run_cli

REM 没有参数或未知命令，使用默认模式
ECHO %INFO% 未指定模式，将使用默认模式 (UI)
ECHO %INFO% 使用 '%0 -h' 查看帮助信息
ECHO.
GOTO :run_ui

:run_ui
ECHO %INFO% 启动 UI 模式...
"%BINARY%"
GOTO :end

:run_tui
ECHO %INFO% 启动 TUI 模式...
"%BINARY%" -tui
GOTO :end

:run_cli
SHIFT
ECHO %INFO% 启动 CLI 模式...
"%BINARY%" -cli %*
GOTO :end

:show_help
ECHO AgentFramework v%VERSION% - 多模式 AI 代理框架
ECHO.
ECHO 使用方法:
ECHO     %0% [模式] [选项]
ECHO.
ECHO 模式:
ECHO     ui, --ui, -ui         启动 Wails 桌面 GUI (默认)
ECHO     tui, --tui, -tui      启动终端用户界面 (TUI)
ECHO     cli, --cli, -cli      进入命令行模式 (CLI)
ECHO.
ECHO CLI 常用命令:
ECHO     agent list            列出所有 agents
ECHO     agent chat "msg"      与 agent 对话
ECHO     workflow list         列出工作流
ECHO     skill list            列出技能
ECHO     config get            查看配置
ECHO     init                  初始化配置
ECHO     version               显示版本
ECHO.
ECHO 选项:
ECHO     -h, --help            显示此帮助信息
ECHO     -v, --verbose         详细输出
ECHO     -c, --config ^<file^>  指定配置文件
ECHO.
ECHO 示例:
ECHO     %0%                   启动桌面 GUI
ECHO     %0% -tui              启动 TUI 界面
ECHO     %0% cli agent list    列出 agents (CLI 模式)
ECHO.
ECHO 更多信息请查看: docs\MULTIMODE_USAGE.md
ECHO.
GOTO :end

:show_version
ECHO AgentFramework v%VERSION%
GOTO :end

:end
ENDLOCAL
