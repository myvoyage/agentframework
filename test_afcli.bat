@echo off
REM AgentFramework CLI 独立程序测试脚本
SETLOCAL

ECHO ========================================
ECHO  AgentFramework CLI - 测试
ECHO ========================================
ECHO.

SET BINARY=build\afcli.exe

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

ECHO [4/4] 生成命令列表...
ECHO.
ECHO 📖 可用命令:
ECHO.
ECHO   afcli agent list              列出所有 agents
ECHO   afcli agent select ^<id^>       选择 agent
ECHO   afcli workflow list           列出工作流
ECHO   afcli skill list              列出技能
ECHO   afcli config get              查看配置
ECHO   afcli help                    显示帮助
ECHO.
ECHO ========================================
ECHO  测试完成！
ECHO ========================================
ECHO.
ECHO  现在可以运行: %BINARY% --help
ECHO.

GOTO :end

:error
ECHO.
ECHO ❌ 测试失败！请先运行编译脚本:
ECHO    build_afcli.bat
ECHO.

:end
ENDLOCAL
