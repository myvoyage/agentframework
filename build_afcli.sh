#!/bin/bash
# AgentFramework CLI (AFCLI) 独立编译脚本 (Linux/Mac)

set -e

BINARY_NAME="afcli"
OUTPUT_DIR="build"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"

echo "========================================"
echo " AgentFramework CLI - 独立编译"
echo "========================================"
echo ""

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo "[1/3] 编译 AFCLI 独立程序..."
echo ""

# 编译 AFCLI 独立程序
go build -v -o "$BINARY_PATH" ./cmd/afcli/

echo "✅ 编译成功"
echo ""

echo "[2/3] 验证编译产物..."
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ 编译产物未找到"
    exit 1
fi

SIZE=$(stat -f%z "$BINARY_PATH" 2>/dev/null || stat -c%s "$BINARY_PATH")
echo "✅ 编译产物验证成功 (大小: $SIZE bytes)"
echo ""

echo "[3/3] 设置执行权限..."
chmod +x "$BINARY_PATH"
echo "✅ 执行权限已设置"
echo ""

echo "========================================"
echo "           编译完成！"
echo "========================================"
echo ""
echo " 输出文件: $BINARY_PATH"
echo " 文件大小: $SIZE bytes"
echo ""
echo " 使用方法:"
echo "   $BINARY_PATH --help            # 显示帮助"
echo "   $BINARY_PATH agent list        # 列出agents"
echo "   ./$BINARY_NAME --help          # 从当前目录"
echo ""
echo " 常用命令:"
echo "   agent list              列出所有 agents"
echo "   agent select <id>        选择 agent"
echo "   workflow list           列出工作流"
echo "   skill list              列出技能"
echo "   config get              查看配置"
echo "   gateway --port 18789    启动 OpenClaw 网关"
echo ""

# 创建便捷的启动脚本
cat > "$OUTPUT_DIR/afcli.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
./afcli "$@"
EOF
chmod +x "$OUTPUT_DIR/afcli.sh"

echo "✅ 已创建便捷启动脚本: $OUTPUT_DIR/afcli.sh"
echo ""
echo "💡 提示: 将 $OUTPUT_DIR 添加到 PATH 环境变量中，即可在任何目录使用 'afcli' 命令"
echo ""
