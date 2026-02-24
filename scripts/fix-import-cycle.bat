@echo off
REM AgentFramework 快速修复脚本 (Windows)
REM 解决导入循环问题

setlocal enabledelayedexpansion

echo ========================================
echo AgentFramework 快速修复脚本 (Windows)
echo ========================================
echo.

REM 检查是否在项目根目录
if not exist "go.mod" (
    echo [错误] 请在项目根目录运行此脚本
    exit /b 1
)

echo [步骤 1] 检查导入循环
echo ----------------------------------------
go list -f "{{.ImportPath}}: {{join .Deps \"\n  \"}}" ./pkg/iot 2>&1 | findstr /i "cycle" >nul
if %errorlevel% equ 0 (
    echo [发现] 导入循环存在
) else (
    echo [OK] 未发现导入循环
)

echo.
echo [步骤 2] 备份原始文件
echo ----------------------------------------
if exist "pkg\iot\adapters" (
    if not exist ".backup" mkdir .backup
    xcopy /E /I /Y pkg\iot\adapters .backup\iot_adapters >nul
    echo [OK] 已备份到 .backup\iot_adapters
) else (
    echo [跳过] pkg\iot\adapters 目录不存在
)

echo.
echo [步骤 3] 修复导入循环
echo ----------------------------------------

if exist "pkg\iot\adapters" (
    echo [正在] 合并 pkg\iot\adapters 到 pkg\iot...

    REM 移动所有 .go 文件
    for %%f in (pkg\iot\adapters\*.go) do (
        move "%%f" pkg\iot\ >nul
    )

    REM 删除空目录
    rmdir /Q pkg\iot\adapters

    REM 更新所有 Go 文件中的导入路径
    REM 使用 PowerShell 进行替换
    powershell -Command "(Get-ChildItem -Recurse -Filter '*.go') | ForEach-Object { (Get-Content $_.FullName) -replace 'AgentFramework/pkg/iot/adapters', 'AgentFramework/pkg/iot' | Set-Content $_.FullName }"

    echo [OK] 导入循环已修复
) else (
    echo [跳过] pkg\iot\adapters 不存在
)

echo.
echo [步骤 4] 验证编译
echo ----------------------------------------

REM 设置 Go 代理
set GOPROXY=https://goproxy.cn,direct

REM 更新依赖
echo [正在] 运行 go mod tidy...
go mod tidy

REM 尝试编译
echo [正在] 编译项目...
go build -v ./... >%TEMP%\build.log 2>&1
if %errorlevel% equ 0 (
    echo [成功] 编译成功！
) else (
    echo [错误] 编译失败，查看日志
    type %TEMP%\build.log | more
    echo.
    echo [提示] 可能需要手动修复其他问题
    pause
    exit /b 1
)

echo.
echo [步骤 5] 清理
echo ----------------------------------------
echo [正在] 清理临时文件...
del %TEMP%\build.log

echo.
echo ========================================
echo [完成] 修复完成！
echo ========================================
echo.
echo 下一步:
echo   1. 运行测试: go test ./...
echo   2. 构建二进制: go build -o bin\agent-framework.exe .\cmd\agent-cli
echo   3. 运行应用: bin\agent-framework.exe serve --config host.yaml
echo.
echo 如果需要回滚:
echo   move .backup\iot_adapters\* pkg\iot\adapters\
echo.
pause
