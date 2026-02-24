@echo off
REM Agent Framework - Windows 统一构建脚本
REM 构建所有组件：主程序、CLI、TUI、示例

setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

echo ========================================
echo   Agent Framework - 统一构建脚本
echo ========================================
echo.

REM 解析参数
set "BUILD_ALL=1"
set "BUILD_MAIN=0"
set "BUILD_CLI=0"
set "BUILD_TUI=0"
set "BUILD_SERVER=0"
set "BUILD_SIMPLEBOT=0"
set "CLEAN=0"

:parse_args
if "%~1"=="" goto :start_build
if "%~1"=="--clean" (
    set "CLEAN=1"
    shift
    goto :parse_args
)
if "%~1"=="--main" (
    set "BUILD_ALL=0"
    set "BUILD_MAIN=1"
    shift
    goto :parse_args
)
if "%~1"=="--cli" (
    set "BUILD_ALL=0"
    set "BUILD_CLI=1"
    shift
    goto :parse_args
)
if "%~1"=="--tui" (
    set "BUILD_ALL=0"
    set "BUILD_TUI=1"
    shift
    goto :parse_args
)
if "%~1"=="--server" (
    set "BUILD_ALL=0"
    set "BUILD_SERVER=1"
    shift
    goto :parse_args
)
if "%~1"=="--simplebot" (
    set "BUILD_ALL=0"
    set "BUILD_SIMPLEBOT=1"
    shift
    goto :parse_args
)
if "%~1"=="--help" goto :show_help
if "%~1"=="-h" goto :show_help
echo 未知选项: %~1
echo 运行 '%0 --help' 查看帮助
exit /b 1

:show_help
echo 用法: %0 [选项]
echo.
echo 选项:
echo   --clean        清理构建产物
echo   --main         只构建主程序
echo   --cli          只构建 CLI 工具
echo   --tui          只构建 TUI
echo   --server       只构建服务器演示
echo   --simplebot    只构建简单机器人
echo   --help, -h     显示此帮助
echo.
echo 默认: 构建所有组件
exit /b 0

:start_build
REM 清理构建产物
if %CLEAN%==1 (
    echo 🧹 清理构建产物...
    del /Q *.exe 2>nul
    echo ✓ 清理完成
    echo.
)

REM 构建主程序
if %BUILD_ALL%==1 (
    echo 🔨 构建主程序...
    go build -o AgentFramework.exe .
    if %ERRORLEVEL%==0 (
        echo ✓ 主程序构建成功
    ) else (
        echo ✗ 主程序构建失败
    )
)

if %BUILD_MAIN%==1 (
    echo 🔨 构建主程序...
    go build -o AgentFramework.exe .
    if %ERRORLEVEL%==0 (
        echo ✓ 主程序构建成功
    ) else (
        echo ✗ 主程序构建失败
    )
)

REM 构建 CLI 工具
if %BUILD_ALL%==1 (
    echo 🔨 构建 CLI 工具...
    go build -o agent-cli.exe ./cmd/cli
    if %ERRORLEVEL%==0 (
        echo ✓ CLI 工具构建成功
    ) else (
        echo ✗ CLI 工具构建失败
    )
)

if %BUILD_CLI%==1 (
    echo 🔨 构建 CLI 工具...
    go build -o agent-cli.exe ./cmd/cli
    if %ERRORLEVEL%==0 (
        echo ✓ CLI 工具构建成功
    ) else (
        echo ✗ CLI 工具构建失败
    )
)

REM 构建 TUI
if %BUILD_ALL%==1 (
    echo 🔨 构建 TUI...
    go build -o tui.exe ./cmd/tui
    if %ERRORLEVEL%==0 (
        echo ✓ TUI 构建成功
    ) else (
        echo ✗ TUI 构建失败
    )
)

if %BUILD_TUI%==1 (
    echo 🔨 构建 TUI...
    go build -o tui.exe ./cmd/tui
    if %ERRORLEVEL%==0 (
        echo ✓ TUI 构建成功
    ) else (
        echo ✗ TUI 构建失败
    )
)

REM 构建服务器演示
if %BUILD_ALL%==1 (
    echo 🔨 构建服务器演示...
    go build -o server_demo.exe ./cmd/server_demo
    if %ERRORLEVEL%==0 (
        echo ✓ 服务器演示构建成功
    ) else (
        echo ✗ 服务器演示构建失败
    )
)

if %BUILD_SERVER%==1 (
    echo 🔨 构建服务器演示...
    go build -o server_demo.exe ./cmd/server_demo
    if %ERRORLEVEL%==0 (
        echo ✓ 服务器演示构建成功
    ) else (
        echo ✗ 服务器演示构建失败
    )
)

REM 构建简单机器人
if %BUILD_ALL%==1 (
    echo 🔨 构建简单机器人...
    go build -o simplebot.exe ./cmd/simplebot
    if %ERRORLEVEL%==0 (
        echo ✓ 简单机器人构建成功
    ) else (
        echo ✗ 简单机器人构建失败
    )
)

if %BUILD_SIMPLEBOT%==1 (
    echo 🔨 构建简单机器人...
    go build -o simplebot.exe ./cmd/simplebot
    if %ERRORLEVEL%==0 (
        echo ✓ 简单机器人构建成功
    ) else (
        echo ✗ 简单机器人构建失败
    )
)

echo.
echo ========================================
echo 🎉 构建完成！
echo ========================================
echo.

REM 列出可执行文件
echo 可执行文件:
for %%f in (*.exe) do (
    echo   • %%f
)
echo.

REM 运行方式说明
echo 运行方式:
echo   • 桌面应用: run.bat 或 AgentFramework.exe
echo   • TUI:       run.bat --tui 或 tui.exe
echo   • CLI:       run.bat cli --help 或 agent-cli.exe
echo.

endlocal
