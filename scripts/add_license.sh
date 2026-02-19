#!/bin/bash
# 批量为 Go 文件添加 AGPL-3.0 许可声明

LICENSE_HEADER='// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
'

# 处理 pkg/iot 目录
echo "正在为 pkg/iot/ 目录添加许可声明..."
find pkg/iot -name "*.go" -type f | while read file; do
    # 检查是否已有许可声明
    if ! grep -q "SPDX-License-Identifier" "$file"; then
        echo "处理: $file"

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
        echo "跳过（已有许可）: $file"
    fi
done

# 处理 examples/iot 目录
echo ""
echo "正在为 examples/iot/ 目录添加许可声明..."
find examples/iot -name "*.go" -type f 2>/dev/null | while read file; do
    if ! grep -q "SPDX-License-Identifier" "$file"; then
        echo "处理: $file"

        temp_file=$(mktemp)
        echo "$LICENSE_HEADER" > "$temp_file"
        echo "" >> "$temp_file"
        cat "$file" >> "$temp_file"
        mv "$temp_file" "$file"
    else
        echo "跳过（已有许可）: $file"
    fi
done

echo ""
echo "✅ 完成！"
