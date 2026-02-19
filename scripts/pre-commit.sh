#!/bin/bash
# Pre-commit hook for license compliance
# 安装: cp .git/hooks/pre-commit.sample .git/hooks/pre-commit
# 或: ln -s ../../scripts/pre-commit.sh .git/hooks/pre-commit

set -e

echo "🔍 检查许可证合规性..."

missing_files=()

# 检查新添加或修改的 Go 文件
for file in $(git diff --cached --name-only --diff-filter=ACM | grep '\.go$'); do
    if [ -f "$file" ]; then
        # 检查是否包含许可声明
        if ! grep -q "SPDX-License-Identifier\|GNU Affero General Public License" "$file"; then
            missing_files+=("$file")
        fi
    fi
done

# 如果有文件缺少许可声明
if [ ${#missing_files[@]} -gt 0 ]; then
    echo "❌ 错误: 以下文件缺少 AGPL-3.0 许可声明:"
    echo ""
    for file in "${missing_files[@]}"; do
        echo "  - $file"
    done
    echo ""
    echo "💡 解决方法:"
    echo "  1. 手动添加许可声明"
    echo "  2. 或运行: bash scripts/add_all_licenses.sh"
    echo ""
    echo "许可声明格式:"
    cat << 'EOF'
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package packagename
EOF
    echo ""
    exit 1
fi

echo "✅ 许可证检查通过"
exit 0
