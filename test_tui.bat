@echo off
REM AgentFramework TUI 快速测试脚本
SETLOCAL

ECHO ========================================
ECHO  AgentFramework TUI - 测试启动
ECHO ========================================
ECHO.

REM 检查编译
IF NOT EXIST "build\af_tui_test.exe" (
    ECHO [1/3] 正在编译...
    go build -v -o build\af_tui_test.exe
    IF ERRORLEVEL 1 (
        ECHO 编译失败！
        EXIT /B 1
    )
    ECHO 编译成功
) ELSE (
    ECHO [1/3] 使用已编译的版本
)

ECHO.
ECHO [2/3] 启动 TUI 模式...
ECHO.
ECHO 提示:
ECHO   - 按 Tab 切换视图
ECHO   - 按 Ctrl+R 刷新数据
ECHO   - 输入 'help' 查看命令
ECHO   - 按 Q 或 Ctrl+C 退出
ECHO.
ECHO ========================================

REM 启动 TUI
build\af_tui_test.exe -tui

ECHO.
ECHO ========================================
ECHO  TUI 已退出
ECHO ========================================

ENDLOCAL
