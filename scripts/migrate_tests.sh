#!/bin/bash

# AgentFramework 测试文件组织迁移脚本
# 将测试文件从原位置迁移到新的测试目录结构

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}") && pwd)"
PROJECT_ROOT="e:/myVibeCoding/AgentFramework"

echo "AgentFramework 测试文件组织迁移脚本"
echo "========================================="
echo ""

# 创建测试目录结构
echo "步骤 1：创建测试目录结构..."
mkdir -p "$PROJECT_ROOT/tests/unit/agent"
mkdir -p "$PROJECT_ROOT/tests/unit/internal"
mkdir -p "$PROJECT_ROOT/tests/unit/pkg"
mkdir -p "$PROJECT_ROOT/tests/integration"
mkdir -p "$PROJECT_ROOT/tests/e2e"
mkdir -p "$PROJECT_ROOT/tests/benchmarks"

echo "✓ 测试目录结构创建完成"
echo ""

# 迁移 agent/ 包的测试文件
echo "步骤 2：迁移 agent/ 包测试文件..."
AGENT_TEST_FILES=(
    "agent/collaboration_test.go"
    "agent/config_manager_test.go"
    "agent/edge_agent_test.go"
    "agent/event_bus_test.go"
    "agent/human_node_test.go"
    "agent/memory_monitor_test.go"
    "agent/model_cache_test.go"
    "agent/model_factory_test.go"
    "agent/model_manager_test.go"
    "agent/monitor_test.go"
    "agent/realtime_agent_test.go"
    "agent/sandbox_manager_test.go"
    "agent/skills_benchmark_test.go"
    "agent/skills_test.go"
    "agent/skill_templates_test.go"
    "agent/thread_store_test.go"
    "agent/token/compression_test.go"
    "agent/workflow_graph_test.go"
    "agent/workflow_react_test.go"
    "agent/workflow_test.go"
)

for file in "${AGENT_TEST_FILES[@]}"; do
    if [ -f "$PROJECT_ROOT/$file" ]; then
        mkdir -p "$PROJECT_ROOT/tests/unit/agent/$(dirname "$file")"
        cp "$PROJECT_ROOT/$file" "$PROJECT_ROOT/tests/unit/agent/$file"
        echo "  ✓ 迁移 $file"
    fi
done

# 迁移 internal/ 包的测试文件
echo ""
echo "步骤 3：迁移 internal/ 包测试文件..."
INTERNAL_TEST_FILES=(
    "internal/eino_bridge/eino_rpc_client_http_test.go"
    "internal/pipelineengine/branch_test.go"
    "internal/pipelineengine/engine_test.go"
    "internal/pipelineengine/loop_test.go"
    "internal/pipelineengine/validation_integration_test.go"
    "internal/pipelineengine/validation_test.go"
    "internal/registry/tool_registry_test.go"
)

for file in "${INTERNAL_TEST_FILES[@]}"; do
    if [ -f "$PROJECT_ROOT/$file" ]; then
        mkdir -p "$PROJECT_ROOT/tests/unit/internal/$(dirname "$file")"
        cp "$PROJECT_ROOT/$file" "$PROJECT_ROOT/tests/unit/internal/$file"
        echo "  ✓ 迁移 $file"
    fi
done

# 迁移 pkg/ 包的测试文件
echo ""
echo "步骤 4：迁移 pkg/ 包测试文件..."
PKG_TEST_FILES=(
    "pkg/heads/context/memory/memory_tier_test.go"
    "pkg/heads/edge_test.go"
    "pkg/heads/hardware_test.go"
    "pkg/heads/mcp/task_tracker_mcp_test.go"
    "pkg/heads/models_test.go"
    "pkg/heads/security_test.go"
    "pkg/heads/store/json_property_test.go"
    "pkg/heads/store/jsonl_store_test.go"
    "pkg/heads/store/sqlite_property_test.go"
    "pkg/heads/store/sqlite_store_test.go"
    "pkg/heads/stream_test.go"
    "pkg/heads/tracker/dependency_resolver_test.go"
    "pkg/heads/tracker/event_processor_test.go"
    "pkg/heads/tracker/sync_daemon_test.go"
    "pkg/heads/tracker/task_tracker_test.go"
    "pkg/framework/event/event_bus_test.go"
    "pkg/framework/workflow/workflow_graph_test.go"
    "pkg/framework/workflow/workflow_react_test.go"
    "pkg/framework/workflow/workflow_test.go"
    "pkg/framework/memory/memory_monitor_test.go"
    "pkg/framework/memory/thread_store_test.go"
    "pkg/tools/sandbox/sandbox_integration_test.go"
    "pkg/tools/sandbox/sandbox_test.go"
    "pkg/tools/sandbox/code/analyzer_benchmark_test.go"
    "pkg/tools/sandbox/code/cache_test.go"
    "pkg/tools/sandbox/code/code_analyzer_test.go"
    "pkg/tools/sandbox/code/code_exec_test.go"
    "pkg/tools/sandbox/code/config_test.go"
    "pkg/tools/sandbox/code/container_executor_test.go"
    "pkg/tools/sandbox/code/container_pool_test.go"
    "pkg/tools/sandbox/code/integration_test.go"
    "pkg/tools/sandbox/code/mcp_tools_test.go"
    "pkg/tools/sandbox/code/security_test.go"
    "agent/scheduler/scheduler_test.go"
    "agent/skills/markdown/markdown_skill_test.go"
)

for file in "${PKG_TEST_FILES[@]}"; do
    if [ -f "$PROJECT_ROOT/$file" ]; then
        mkdir -p "$PROJECT_ROOT/tests/unit/pkg/$(dirname "$file")"
        cp "$PROJECT_ROOT/$file" "$PROJECT_ROOT/tests/unit/pkg/$file"
        echo "  ✓ 迁移 $file"
    fi
done

# 迁移 tests/ 目录的测试文件
echo ""
echo "步骤 5：迁移 tests/ 目录测试文件..."
ROOT_TEST_FILES=(
    "tests/integration/config_integration_test.go"
    "tests/concurrency/current_test.go"
    "tests/stage6_integration_test.go"
)

for file in "${ROOT_TEST_FILES[@]}"; do
    if [ -f "$PROJECT_ROOT/$file" ]; then
        target_dir="tests/$(echo "$file" | cut -d'/' -f2)"
        mkdir -p "$PROJECT_ROOT/$target_dir"
        cp "$PROJECT_ROOT/$file" "$PROJECT_ROOT/tests/unit/$file"
        echo "  ✓ 迁移 $file"
    fi
done

echo ""
echo "========================================="
echo "测试文件迁移完成！"
echo ""
echo "目录结构："
echo "tests/"
echo "├── unit/            # 单元测试"
echo "│   ├── agent/     # agent 包测试"
echo "│   ├── internal/   # internal 包测试"
echo "│   └── pkg/       # pkg 包测试"
echo "├── integration/     # 集成测试"
echo "├── e2e/             # 端到端测试"
echo "└── benchmarks/      # 性能测试"
echo ""
echo "下一步："
echo "1. 验证测试编译：go test ./tests/unit/... -v"
echo "2. 验证测试覆盖率：go test ./tests/unit/... -cover -coverprofile=coverage.out"
echo "3. 更新 go.mod 中的测试路径"
echo "4. 删除原位置的测试文件（可选）"