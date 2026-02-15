#!/bin/bash
#
# TextProcessor Converter Script
#
# 文本格式转换脚本 - 支持编码转换、换行符转换等
#
# 作者: AgentFramework Team
# 版本: 1.2.0
# 许可证: MIT

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认值
INPUT_ENCODING="utf-8"
OUTPUT_ENCODING="utf-8"
NORMALIZE="false"
NORMALIZE_TO="lf"
BASE64_ENCODE="false"
BASE64_DECODE="false"
INPUT_FILE=""
OUTPUT_FILE=""

# 显示帮助信息
show_help() {
    cat << EOF
TextProcessor Converter - 文本格式转换工具

用法: converter.sh [选项]

选项:
    -i, --input FILE         输入文件路径
    -o, --output FILE        输出文件路径（可选，默认为标准输出）
    -e, --encode ENC         输入编码（默认: utf-8）
    -d, --decode ENC         输出编码（默认: utf-8）
    -n, --normalize TYPE     标准化换行符: lf|crlf|cr（默认: 不转换）
    -b, --base64             Base64 编码
    -B, --base64-decode      Base64 解码
    -h, --help               显示此帮助信息

示例:
    # 转换编码
    converter.sh -i input.txt -e gbk -d utf-8

    # 标准化换行符为 LF
    converter.sh -i input.txt -n lf -o output.txt

    # Base64 编码
    converter.sh -i input.txt -b

    # 组合使用
    converter.sh -i input.txt -e gbk -d utf-8 -n lf -o output.txt

支持的编码:
    utf-8, utf-16, utf-16le, utf-16be, utf-32, utf-32le, utf-32be,
    gbk, gb2312, gb18030, big5, shift_jis, euc-jp, euc-kr, iso-8859-1,
    ascii, latin1, windows-1252

EOF
}

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# 检查依赖
check_dependencies() {
    local deps=("iconv" "base64")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "缺少依赖: $dep"
            exit 1
        fi
    done
}

# 转换编码
convert_encoding() {
    local input_file="$1"
    local input_enc="$2"
    local output_enc="$3"

    iconv -f "$input_enc" -t "$output_enc" "$input_file" 2>/dev/null || {
        log_error "编码转换失败: $input_enc -> $output_enc"
        return 1
    }
}

# 标准化换行符
normalize_newlines() {
    local input="$1"
    local target="$2"

    case "$target" in
        lf)
            # 转换为 LF (\n)
            tr -d '\r' <<< "$input"
            ;;
        crlf)
            # 转换为 CRLF (\r\n)
            awk '{printf "%s\r\n", $0}' <<< "$input"
            ;;
        cr)
            # 转换为 CR (\r)
            tr '\n' '\r' <<< "$input" | tr -d '\n'
            ;;
        *)
            log_error "无效的换行符类型: $target"
            return 1
            ;;
    esac
}

# Base64 编码
base64_encode() {
    base64 -w 0
}

# Base64 解码
base64_decode() {
    base64 -d
}

# 主函数
main() {
    local input_content=""
    local result=""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--input)
                INPUT_FILE="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -e|--encode)
                INPUT_ENCODING="$2"
                shift 2
                ;;
            -d|--decode)
                OUTPUT_ENCODING="$2"
                shift 2
                ;;
            -n|--normalize)
                NORMALIZE="true"
                NORMALIZE_TO="$2"
                shift 2
                ;;
            -b|--base64)
                BASE64_ENCODE="true"
                shift
                ;;
            -B|--base64-decode)
                BASE64_DECODE="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 检查输入文件
    if [[ -z "$INPUT_FILE" ]]; then
        log_error "必须指定输入文件 (-i|--input)"
        show_help
        exit 1
    fi

    if [[ ! -f "$INPUT_FILE" ]]; then
        log_error "文件不存在: $INPUT_FILE"
        exit 1
    fi

    # 检查依赖
    check_dependencies

    log_info "处理文件: $INPUT_FILE"

    # 读取输入文件
    input_content=$(cat "$INPUT_FILE")

    # Base64 解码（如果需要）
    if [[ "$BASE64_DECODE" == "true" ]]; then
        log_info "执行 Base64 解码..."
        result=$(echo "$input_content" | base64_decode)
        input_content="$result"
    fi

    # 转换编码
    if [[ "$INPUT_ENCODING" != "$OUTPUT_ENCODING" ]]; then
        log_info "转换编码: $INPUT_ENCODING -> $OUTPUT_ENCODING"
        result=$(echo "$input_content" | convert_encoding \
                 /dev/stdin "$INPUT_ENCODING" "$OUTPUT_ENCODING")
        input_content="$result"
    fi

    # 标准化换行符
    if [[ "$NORMALIZE" == "true" ]]; then
        log_info "标准化换行符为: $NORMALIZE_TO"
        result=$(normalize_newlines "$input_content" "$NORMALIZE_TO")
        input_content="$result"
    fi

    # Base64 编码（如果需要）
    if [[ "$BASE64_ENCODE" == "true" ]]; then
        log_info "执行 Base64 编码..."
        result=$(echo "$input_content" | base64_encode)
        input_content="$result"
    fi

    # 输出结果
    if [[ -n "$OUTPUT_FILE" ]]; then
        log_info "保存到文件: $OUTPUT_FILE"
        echo "$input_content" > "$OUTPUT_FILE"
        log_info "转换完成！"
    else
        echo "$input_content"
    fi
}

# 执行主函数
main "$@"
