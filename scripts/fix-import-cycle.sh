#!/bin/bash
# AgentFramework 快速修复脚本
# 解决导入循环问题

set -e

echo "🔧 AgentFramework 快速修复脚本"
echo "================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    echo -e "${RED}错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

echo -e "${YELLOW}步骤 1: 检查导入循环${NC}"
echo "----------------------------------------"
go list -f '{{.ImportPath}}: {{join .Deps "\n  "}}' ./pkg/iot 2>&1 | grep -i "cycle" && echo -e "${RED}发现导入循环${NC}" || echo -e "${GREEN}未发现导入循环${NC}"

echo ""
echo -e "${YELLOW}步骤 2: 备份原始文件${NC}"
echo "----------------------------------------"
if [ -d "pkg/iot/adapters" ]; then
    mkdir -p .backup
    cp -r pkg/iot/adapters .backup/iot_adapters
    echo -e "${GREEN}已备份到 .backup/iot_adapters${NC}"
else
    echo -e "${YELLOW}pkg/iot/adapters 目录不存在，跳过备份${NC}"
fi

echo ""
echo -e "${YELLOW}步骤 3: 修复导入循环${NC}"
echo "----------------------------------------"

# 方案 C: 合并包
if [ -d "pkg/iot/adapters" ]; then
    echo "正在合并 pkg/iot/adapters 到 pkg/iot..."

    # 移动所有 .go 文件
    find pkg/iot/adapters -name "*.go" -type f -exec mv {} pkg/iot/ \;

    # 删除空目录
    rm -rf pkg/iot/adapters

    # 更新所有 Go 文件中的导入路径
    find . -name "*.go" -type f -exec sed -i 's|AgentFramework/pkg/iot/adapters|AgentFramework/pkg/iot|g' {} \;

    echo -e "${GREEN}导入循环已修复${NC}"
else
    echo -e "${YELLOW}pkg/iot/adapters 不存在，无需修复${NC}"
fi

echo ""
echo -e "${YELLOW}步骤 4: 验证编译${NC}"
echo "----------------------------------------"

# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct

# 更新依赖
echo "正在运行 go mod tidy..."
go mod tidy

# 尝试编译
echo "正在编译项目..."
if go build -v ./... 2>&1 | tee /tmp/build.log; then
    echo -e "${GREEN}✅ 编译成功！${NC}"
else
    echo -e "${RED}❌ 编译失败，查看日志${NC}"
    tail -20 /tmp/build.log
    echo ""
    echo -e "${YELLOW}提示: 可能需要手动修复其他问题${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}步骤 5: 清理${NC}"
echo "----------------------------------------"
echo "清理临时文件..."
rm -f /tmp/build.log

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✅ 修复完成！${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "下一步:"
echo "  1. 运行测试: go test ./..."
echo "  2. 构建二进制: go build -o bin/agent-framework ./cmd/agent-cli"
echo "  3. 运行应用: ./bin/agent-framework serve --config host.yaml"
echo ""
echo "如果需要回滚:"
echo "  mv .backup/iot_adapters/* pkg/iot/adapters/"
echo ""
