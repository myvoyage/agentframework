@echo off
REM AgentFramework TUI (AFTUI) 独立编译脚本
SETLOCAL

ECHO ========================================
ECHO  AgentFramework TUI - 独立编译
ECHO ========================================
ECHO.

SET BINARY_NAME=aftui
SET OUTPUT_DIR=build
SET BINARY_PATH=%OUTPUT_DIR%\%BINARY_NAME%.exe

REM 创建输出目录
IF NOT EXIST "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

ECHO [1/3] 编译 AFTUI 独立程序...
ECHO.

go build -v -o "%BINARY_PATH%" ./cmd/aftui/
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
ECHO ║    %BINARY_PATH%                    # 启动 TUI  ║
ECHO ║    .\%BINARY_NAME%                       # 从当前目录启动 ║
ECHO ╠════════════════════════════════════════════════════════════╣
ECHO ║  快捷键:                                                    ║
ECHO ║    Tab       - 切换视图                                      ║
ECHO ║    Ctrl+R    - 刷新数据                                       ║
ECHO ║    Enter     - 执行命令                                       ║
ECHO ║    Q         - 退出                                          ║
ECHO ╚════════════════════════════════════════════════════════════╝

REM 创建便捷的启动脚本
echo @echo off > "%OUTPUT_DIR%\tui.bat"
echo start "" "%BINARY_PATH%" >> "%OUTPUT_DIR%\tui.bat"

ECHO.
ECHO ✅ 已创建便捷启动脚本: %OUTPUT_DIR%\tui.bat
ECHO.

ENDLOCAL
