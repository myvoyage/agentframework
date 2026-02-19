#!/bin/bash
# 批量为所有缺失许可声明的 Go 文件添加 AGPL-3.0 许可

LICENSE_HEADER='// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
'

total_count=0
updated_count=0
skipped_count=0

echo "正在扫描所有 Go 文件..."
echo ""

# 查找所有 .go 文件
find . -name "*.go" -type f | while read file; do
    total_count=$((total_count + 1))

    # 检查是否已有许可声明
    if ! grep -q "SPDX-License-Identifier\|GNU Affero General Public License" "$file"; then
        echo "添加许可: $file"
        updated_count=$((updated_count + 1))

        # 创建临时文件
        temp_file=$(mktemp)

        # 写入许可声明
        echo "$LICENSE_HEADER" > "$temp_file"
        echo "" >> "$temp_file"

        # 追加原文件内容
        cat "$file" >> "$temp_file"

        # 替换原文件
        mv "$temp_file" "$file"
    else
        skipped_count=$((skipped_count + 1))
    fi
done

echo ""
echo "=== 处理完成 ==="
echo "总文件数: $total_count"
echo "已更新: $updated_count"
echo "已跳过: $skipped_count"
