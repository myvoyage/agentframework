@echo off
REM AgentFramework CLI (AFCLI) 独立编译脚本 (Windows)

SETLOCAL ENABLEDELAYEDEXPANSION

ECHO ========================================
ECHO  AgentFramework CLI - 独立编译
ECHO ========================================
ECHO.

SET BINARY_NAME=afcli
SET OUTPUT_DIR=build
SET BINARY_PATH=%OUTPUT_DIR%\%BINARY_NAME%.exe

REM 创建输出目录
IF NOT EXIST "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

ECHO [1/3] 编译 AFCLI 独立程序...
ECHO.

go build -v -o "%BINARY_PATH%" ./cmd/afcli/
IF ERRORLEVEL 1 (
    ECHO ❌ 编译失败！
    EXIT /B 1
)

ECHO ✅ 编译成功
ECHO.

ECHO [2/3] 验证编译产物...
IF NOT EXIST "%BINARY_PATH%" (
    ECHO ❌ 编译产物未找到
    EXIT /B 1
)

FOR %%F IN ("%BINARY_PATH%") DO set SIZE=%%~zF

ECHO ✅ 编译产物验证成功 (大小: %SIZE bytes)
ECHO.

ECHO [3/3] 生成使用说明...
ECHO ╔════════════════════════════════════════════════════════════╗
ECHO ║           编译完成！                                        ║
ECHO ╠════════════════════════════════════════════════════════════╣
ECHO ║  输出文件: %BINARY_PATH%                           ║
ECHO ║  文件大小: %SIZE%                                  ║
ECHO ╠════════════════════════════════════════════════════════════╣
ECHO ║  使用方法:                                                 ║
ECHO ║    %BINARY_PATH% help              # 显示帮助  ║
ECHO ║    %BINARY_PATH% agent list        # 列出agents ║
ECHO ║    %BINARY_PATH% --help            # 查看版本  ║
ECHO ╠════════════════════════════════════════════════════════════╣
ECHO ║  常用命令:                                                 ║
ECHO ║    agent list              列出所有 agents             ║
ECHO ║    agent select ^<id^>       选择 agent                 ║
ECHO ║    workflow list           列出工作流                 ║
ECHO ║    skill list              列出技能                   ║
ECHO ║    config get              查看配置                   ║
ECHO ╚════════════════════════════════════════════════════════════╝

REM 创建便捷的启动脚本
echo @echo off > "%OUTPUT_DIR%\afcli.bat"
echo %BINARY_NAME% %%* >> "%OUTPUT_DIR%\afcli.bat"

ECHO.
ECHO ✅ 已创建便捷启动脚本: %OUTPUT_DIR%\afcli.bat
ECHO.
ECHO 💡 提示: 将 %OUTPUT_DIR% 添加到 PATH 环境变量中，即可在任何目录使用 'afcli' 命令
ECHO.

ENDLOCAL
