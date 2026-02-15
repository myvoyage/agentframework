# AgentFramework 测试文件组织迁移脚本
# 使用 PowerShell 实现

$ProjectRoot = "e:/myVibeCoding/AgentFramework"

Write-Host "AgentFramework 测试文件组织迁移脚本"
Write-Host "========================================="
Write-Host ""

# 创建测试目录结构
Write-Host "步骤 1：创建测试目录结构..."
$TestDirectories = @(
    "tests/unit/agent",
    "tests/unit/internal",
    "tests/unit/pkg",
    "tests/integration",
    "tests/e2e",
    "tests/benchmarks"
)

foreach ($dir in $TestDirectories) {
    $FullPath = Join-Path $ProjectRoot $dir
    if (-not (Test-Path $FullPath)) {
        New-Item -ItemType Directory -Path $FullPath -Force | Out-Null
        Write-Host "  ✓ 创建目录 $dir"
    }
}

Write-Host "✓ 测试目录结构创建完成"
Write-Host ""

# 迁移 agent/ 包的测试文件
Write-Host "步骤 2：迁移 agent/ 包测试文件..."
$AgentTestFiles = @(
    "agent/collaboration/collaboration_test.go",
    "agent/config_manager_test.go",
    "agent/edge_agent_test.go",
    "agent/event_bus_test.go",
    "agent/human_node_test.go",
    "agent/memory_monitor_test.go",
    "agent/model_cache_test.go",
    "agent/model_factory_test.go",
    "agent/model_manager_test.go",
    "agent/monitor_test.go",
    "agent/realtime_agent_test.go",
    "agent/sandbox_manager_test.go",
    "agent/skills/skills_benchmark_test.go",
    "agent/skills/skills_test.go",
    "agent/skill_templates_test.go",
    "agent/thread_store_test.go",
    "agent/token/compression_test.go",
    "agent/workflow_graph_test.go",
    "agent/workflow_react_test.go",
    "agent/workflow_test.go",
    "agent/scheduler/scheduler_test.go",
    "agent/skills/markdown/markdown_skill_test.go"
)

foreach ($file in $AgentTestFiles) {
    $SourcePath = Join-Path $ProjectRoot $file
    if (Test-Path $SourcePath) {
        $TargetPath = Join-Path $ProjectRoot "tests/unit" $file
        $TargetDir = Split-Path $TargetPath -Parent
        if (-not (Test-Path $TargetDir)) {
            New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
        }
        Copy-Item -Path $SourcePath -Destination $TargetPath -Force
        Write-Host "  ✓ 迁移 $file"
    } else {
        Write-Host "  ℹ️  文件 $file 不存在"
    }
}

# 迁移 internal/ 包的测试文件
Write-Host ""
Write-Host "步骤 3：迁移 internal/ 包测试文件..."
$InternalTestFiles = @(
    "internal/eino_bridge/eino_rpc_client_http_test.go",
    "internal/pipelineengine/branch_test.go",
    "internal/pipelineengine/engine_test.go",
    "internal/pipelineengine/loop_test.go",
    "internal/pipelineengine/validation_integration_test.go",
    "internal/pipelineengine/validation_test.go",
    "internal/registry/tool_registry_test.go"
)

foreach ($file in $InternalTestFiles) {
    $SourcePath = Join-Path $ProjectRoot $file
    if (Test-Path $SourcePath) {
        $TargetPath = Join-Path $ProjectRoot "tests/unit" $file
        $TargetDir = Split-Path $TargetPath -Parent
        if (-not (Test-Path $TargetDir)) {
            New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
        }
        Copy-Item -Path $SourcePath -Destination $TargetPath -Force
        Write-Host "  ✓ 迁移 $file"
    } else {
        Write-Host "  ℹ️  文件 $file 不存在"
    }
}

# 迁移 pkg/ 包的测试文件
Write-Host ""
Write-Host "步骤 4：迁移 pkg/ 包测试文件..."
$PkgTestFiles = @(
    "pkg/heads/context/memory/memory_tier_test.go",
    "pkg/heads/edge_test.go",
    "pkg/heads/hardware_test.go",
    "pkg/heads/mcp/task_tracker_mcp_test.go",
    "pkg/heads/models_test.go",
    "pkg/heads/security_test.go",
    "pkg/heads/store/json_property_test.go",
    "pkg/heads/store/jsonl_store_test.go",
    "pkg/heads/store/sqlite_property_test.go",
    "pkg/heads/store/sqlite_store_test.go",
    "pkg/heads/stream_test.go",
    "pkg/heads/tracker/dependency_resolver_test.go",
    "pkg/heads/tracker/event_processor_test.go",
    "pkg/heads/tracker/sync_daemon_test.go",
    "pkg/heads/tracker/task_tracker_test.go",
    "pkg/framework/event/event_bus_test.go",
    "pkg/framework/workflow/workflow_graph_test.go",
    "pkg/framework/workflow/workflow_react_test.go",
    "pkg/framework/workflow/workflow_test.go",
    "pkg/framework/memory/memory_monitor_test.go",
    "pkg/framework/memory/thread_store_test.go",
    "pkg/tools/sandbox/sandbox_integration_test.go",
    "pkg/tools/sandbox/sandbox_test.go",
    "pkg/tools/sandbox/code/analyzer_benchmark_test.go",
    "pkg/tools/sandbox/code/cache_test.go",
    "pkg/tools/sandbox/code/code_analyzer_test.go",
    "pkg/tools/sandbox/code/code_exec_test.go",
    "pkg/tools/sandbox/code/config_test.go",
    "pkg/tools/sandbox/code/container_executor_test.go",
    "pkg/tools/sandbox/code/container_pool_test.go",
    "pkg/tools/sandbox/code/integration_test.go",
    "pkg/tools/sandbox/code/mcp_tools_test.go",
    "pkg/tools/sandbox/code/security_test.go"
)

foreach ($file in $PkgTestFiles) {
    $SourcePath = Join-Path $ProjectRoot $file
    if (Test-Path $SourcePath) {
        $TargetPath = Join-Path $ProjectRoot "tests/unit" $file
        $TargetDir = Split-Path $TargetPath -Parent
        if (-not (Test-Path $TargetDir)) {
            New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
        }
        Copy-Item -Path $SourcePath -Destination $TargetPath -Force
        Write-Host "  ✓ 迁移 $file"
    } else {
        Write-Host "  ℹ️  文件 $file 不存在"
    }
}

# 验证迁移结果
Write-Host ""
Write-Host "步骤 5：验证迁移结果..."
$UnitTestCount = (Get-ChildItem -Path (Join-Path $ProjectRoot "tests/unit") -Filter "*_test.go" -Recurse).Count
$AllTestCount = (Get-ChildItem -Path $ProjectRoot -Filter "*_test.go" -Recurse).Count

Write-Host "✓ 共找到 $AllTestCount 个测试文件"
Write-Host "✓ 成功迁移 $UnitTestCount 个测试文件到新目录"
Write-Host ""
Write-Host "========================================="
Write-Host "测试文件迁移完成！"
Write-Host ""
Write-Host "目录结构："
Write-Host "tests/"
Write-Host "├── unit/            # 单元测试"
Write-Host "│   ├── agent/     # agent 包测试"
Write-Host "│   ├── internal/   # internal 包测试"
Write-Host "│   └── pkg/       # pkg 包测试"
Write-Host "├── integration/     # 集成测试"
Write-Host "├── e2e/             # 端到端测试"
Write-Host "└── benchmarks/      # 性能测试"
Write-Host ""
Write-Host "下一步："
Write-Host "1. 验证测试编译：cd $ProjectRoot; go test ./tests/unit/... -v"
Write-Host "2. 验证测试覆盖率：cd $ProjectRoot; go test ./tests/unit/... -cover -coverprofile=coverage.out"
Write-Host "3. 更新 go.mod 中的测试路径"
Write-Host "4. 删除原位置的测试文件（可选）"