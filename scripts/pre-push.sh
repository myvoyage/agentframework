#!/bin/bash
# Pre-push hook for additional checks
# 安装: cp .git/hooks/pre-push.sample .git/hooks/pre-push
# 或: ln -s ../../scripts/pre-push.sh .git/hooks/pre-push

set -e

echo "🚀 运行 pre-push 检查..."

# 运行测试
echo "📦 运行测试..."
go test ./... -short -timeout 30s

# 检查构建
echo "🔨 检查构建..."
go build ./...

# 检查许可覆盖
echo "📜 检查许可覆盖..."
total=$(find . -name "*.go" -type f | wc -l)
licensed=$(grep -rl "SPDX-License-Identifier\|GNU Affero" --include="*.go" . 2>/dev/null | wc -l)
coverage=$(awk "BEGIN {printf \"%.1f\", $licensed/$total*100}")

echo "  许可覆盖: $coverage% ($licensed/$total)"

if [ "$licensed" -lt "$total" ]; then
    echo "⚠️  警告: 并非所有文件都有许可声明"
    echo "  建议: 运行 bash scripts/add_all_licenses.sh"
fi

# 运行 gofmt
echo "🎨 检查代码格式..."
unformatted=$(gofmt -l . 2>/dev/null || true)
if [ -n "$unformatted" ]; then
    echo "⚠️  以下文件需要格式化:"
    echo "$unformatted" | head -20
    echo ""
    echo "💡 运行: go fmt ./..."
fi

# 运行 go vet
echo "🔍 运行 go vet..."
go vet ./... || true

echo "✅ Pre-push 检查完成"
exit 0
