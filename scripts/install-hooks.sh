#!/bin/bash
# 安装 Git hooks 脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🔧 正在安装 Git hooks..."

# 创建 .git/hooks 目录（如果不存在）
mkdir -p "$REPO_ROOT/.git/hooks"

# 安装 pre-commit hook
echo "📝 安装 pre-commit hook..."
ln -sf "../../scripts/pre-commit.sh" "$REPO_ROOT/.git/hooks/pre-commit"
chmod +x "$REPO_ROOT/.git/hooks/pre-commit"

# 安装 pre-push hook
echo "📝 安装 pre-push hook..."
ln -sf "../../scripts/pre-push.sh" "$REPO_ROOT/.git/hooks/pre-push"
chmod +x "$REPO_ROOT/.git/hooks/pre-push"

echo ""
echo "✅ Git hooks 安装完成！"
echo ""
echo "已安装的 hooks:"
echo "  📝 pre-commit  - 检查许可证合规性"
echo "  🚀 pre-push    - 运行测试和格式检查"
echo ""
echo "💡 提示: 如需跳过检查，使用 --no-verify 标志:"
echo "  git commit --no-verify -m 'message'"
echo "  git push --no-verify"
echo ""
