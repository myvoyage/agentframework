@echo off
REM AgentFramework TUI 独立编译脚本 (Windows)

SETLOCAL ENABLEDELAYEDEXPANSION

ECHO ========================================
ECHO  AgentFramework TUI - 独立编译
ECHO ========================================
ECHO.

REM 设置编译参数
set BINARY_NAME=aftui
set OUTPUT_DIR=build
set BINARY_PATH=%OUTPUT_DIR%\%BINARY_NAME%.exe

REM 创建输出目录
IF NOT EXIST "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

ECHO [1/3] 编译 TUI 独立程序...
ECHO.

REM 编译 TUI 独立程序
go build -v -o "%BINARY_PATH%" ./cmd/tui/
IF ERRORLEVEL 1 (
    ECHO ❌ 编译失败！
    EXIT /B 1
)

ECHO ✅ 编译成功
ECHO.

ECHO [2/3] 检查编译结果...
IF NOT EXIST "%BINARY_PATH%" (
    ECHO ❌ 编译产物未找到: %BINARY_PATH%
    EXIT /B 1
)

ECHO ✅ 编译产物已生成
ECHO.

ECHO [3/3] 显示文件信息...
FOR %%F IN ("%BINARY_PATH%") DO (
    set SIZE=%%~zF bytes
    set TIME=%%~tF
)

ECHO ========================================
ECHO  编译完成！
ECHO ========================================
ECHO.
ECHO  输出文件: %BINARY_PATH%
ECHO  文件大小: %SIZE%
ECHO  编译时间: %TIME%
ECHO.
ECHO  使用方法:
ECHO    %BINARY_PATH%                    # 启动 TUI
ECHO.
ECHO  或在当前目录:
ECHO    %OUTPUT_DIR%\%BINARY_NAME%        # 启动 TUI
ECHO.

ENDLOCAL
