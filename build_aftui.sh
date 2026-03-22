#!/bin/bash
# AgentFramework TUI 独立编译脚本 (Linux/Mac)

set -e

BINARY_NAME="aftui"
OUTPUT_DIR="build"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"

echo "========================================"
echo " AgentFramework TUI - 独立编译"
echo "========================================"
echo ""

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo "[1/3] 编译 AFTUI 独立程序..."
echo ""

# 编译 AFTUI 独立程序
go build -v -o "$BINARY_PATH" ./cmd/aftui/

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
echo "   $BINARY_PATH                 # 启动 TUI"
echo "   ./$BINARY_NAME                # 从当前目录启动"
echo ""
echo " 快捷键:"
echo "   Tab       - 切换视图"
echo "   Ctrl+R    - 刷新数据"
echo "   Enter     - 执行命令"
echo "   Q         - 退出"
echo ""

# 创建便捷的启动脚本
cat > "$OUTPUT_DIR/tui.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
./aftui "$@"
EOF
chmod +x "$OUTPUT_DIR/tui.sh"

echo "✅ 已创建便捷启动脚本: $OUTPUT_DIR/tui.sh"
echo ""
