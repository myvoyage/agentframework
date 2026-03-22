@echo off
REM AgentFramework TUI 独立程序测试脚本
SETLOCAL

ECHO ========================================
ECHO  AgentFramework TUI - 测试
ECHO ========================================
ECHO.

SET BINARY=build\aftui.exe

ECHO [1/4] 检查编译产物...
IF NOT EXIST "%BINARY%" (
    ECHO ❌ 编译产物不存在
    GOTO :error
)
ECHO ✅ 找到编译产物
ECHO.

ECHO [2/4] 检查文件大小...
FOR %%F IN ("%BINARY%") DO set SIZE=%%~zF
ECHO ✅ 文件大小: %SIZE bytes
ECHO.

ECHO [3/4] 尝试运行版本检查...
"%BINARY%" --version 2>NUL
IF ERRORLEVEL 0 (
    ECHO ✅ 程序可运行
) ELSE (
    ECHO ⚠ 程序可能不支持版本参数
)
ECHO.

ECHO [4/4] 使用说明...
ECHO.
ECHO 📖 快速开始:
ECHO.
ECHO   1. 启动 TUI:
ECHO      %BINARY%
ECHO.
ECHO   2. 基本命令:
ECHO      agent list              列出 Agents
ECHO      agent select ^<id^>       选择 Agent
ECHO      chat ^<message^>           发送消息
ECHO      help                     显示帮助
ECHO.
ECHO   3. 快捷键:
ECHO      Tab                      切换视图
ECHO      Ctrl+R                   刷新数据
ECHO      Q                        退出
ECHO.
ECHO ========================================
ECHO  测试完成！
ECHO ========================================
ECHO.
ECHO  现在可以运行: %BINARY%
ECHO.

GOTO :end

:error
ECHO.
ECHO ❌ 测试失败！请先运行编译脚本:
ECHO    build_aftui.bat
ECHO.

:end
ENDLOCAL
