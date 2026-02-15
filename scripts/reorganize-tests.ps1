# AgentFramework 测试文件重新组织脚本
# 修复包结构问题

$ProjectRoot = "e:/myVibeCoding/AgentFramework"
$TestRoot = Join-Path $ProjectRoot "tests/unit"

Write-Host "AgentFramework 测试文件重新组织脚本"
Write-Host "========================================="
Write-Host ""

# 函数：移动测试文件到正确的位置
function Move-TestFiles {
    param(
        [string]$SourceDir,
        [string]$Pattern,
        [string]$TargetDir
    )

    $SourcePath = Join-Path $ProjectRoot $SourceDir
    if (-not (Test-Path $SourcePath)) {
        Write-Host "⚠️  源目录不存在: $SourceDir"
        return
    }

    $TestFiles = Get-ChildItem -Path $SourcePath -Filter $Pattern -Recurse
    if ($TestFiles.Count -eq 0) {
        Write-Host "ℹ️  没有找到测试文件: $Pattern"
        return
    }

    $TargetPath = Join-Path $TestRoot $TargetDir
    if (-not (Test-Path $TargetPath)) {
        New-Item -ItemType Directory -Path $TargetPath -Force | Out-Null
    }

    $CopiedCount = 0
    foreach ($File in $TestFiles) {
        $FileName = Split-Path $File -Leaf
        $DestPath = Join-Path $TargetPath $FileName

        try {
            Copy-Item -Path $File.FullName -Destination $DestPath -Force -ErrorAction Stop
            $CopiedCount++
            Write-Host "  ✓ 复制 $FileName"
        } catch {
            Write-Host "  ✗ 复制失败: $FileName ($($_)"
        }
    }

    Write-Host "  → 共复制 $CopiedCount 个文件到 $TargetDir"
}

# 步骤 1：处理 agent/ 包测试
Write-Host "步骤 1：处理 agent/ 包测试..."
Move-TestFiles "agent" "*_test.go" "agent"
Write-Host ""

# 步骤 2：处理 internal/ 包测试
Write-Host "步骤 2：处理 internal/ 包测试..."
Move-TestFiles "internal" "*_test.go" "internal"
Write-Host ""

# 步骤 3：处理 pkg/ 包测试
Write-Host "步骤 3：处理 pkg/ 包测试..."
Move-TestFiles "pkg" "*_test.go" "pkg"
Write-Host ""

Write-Host "========================================="
Write-Host "测试文件重新组织完成！"
Write-Host ""
Write-Host "下一步："
Write-Host "1. 验证包结构：go list ./tests/unit/..."
Write-Host "2. 运行测试：go test ./tests/unit/... -v"
Write-Host "3. 查看测试覆盖率：go test ./tests/unit/... -cover -coverprofile=coverage.out"