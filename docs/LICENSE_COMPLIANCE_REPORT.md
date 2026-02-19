# AgentFramework 许可证合规性报告

**检查日期**: 2026-02-19
**许可证**: AGPL-3.0-or-later
**检查范围**: 所有 Go 源代码文件

---

## 📋 执行摘要

AgentFramework 项目已采用 **GNU Affero General Public License v3.0 (AGPL-3.0-or-later)** 许可证。

**关键发现**:
- ✅ 项目根目录包含完整的 AGPL-3.0 许可证文件
- ✅ 69% 的 Go 文件包含许可声明（380/551）
- ⚠️ 31% 的 Go 文件缺少许可声明（171/551）
- ✅ 新增的 channels 模块包含许可声明
- ❌ 新增的 IoT 模块缺少许可声明

---

## 📊 许可证覆盖统计

### 整体统计

| 类别 | 数量 | 百分比 |
|------|------|--------|
| **总 Go 文件数** | 551 | 100% |
| **有许可声明的文件** | 380 | 69% |
| **缺少许可声明的文件** | 171 | 31% |
| **完整 AGPL 声明** | 238 | 43% |
| **仅 Copyright 声明** | 142 | 26% |

### 按模块统计

| 模块 | 状态 | 说明 |
|------|------|------|
| `agent/` | ✅ 良好 | 大部分文件有完整声明 |
| `pkg/channels/` | ✅ 完整 | 包含 SPDX 标识符 |
| `pkg/beads/` | ✅ 良好 | 大部分有声明 |
| `pkg/iot/` | ❌ 缺失 | **需要添加许可声明** |
| `pkg/framework/` | ✅ 良好 | 大部分有声明 |
| `pkg/tools/` | ⚠️ 部分 | 部分文件缺少声明 |
| `pkg/skills/` | ✅ 良好 | 大部分有声明 |
| `examples/` | ⚠️ 部分 | 示例代码通常需要声明 |

---

## ✅ 正确的许可声明格式

### 标准格式（推荐）

```go
// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package packagename
```

### 简化格式（可接受）

```go
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package packagename
```

### 优秀示例（pkg/channels/types.go）

```go
// Package channels provides a unified multi-channel messaging system
// supporting various platforms like Telegram, Discord, Slack, Feishu, and WeWork.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels
```

---

## ❌ 缺少许可声明的文件

### 高优先级（核心模块）

#### pkg/iot/ 模块 - 全部缺失

**需要添加的文件**:
- `pkg/iot/types.go`
- `pkg/iot/device.go`
- `pkg/iot/adapter.go`
- `pkg/iot/adapters/*.go` (所有适配器)
- `pkg/iot/workflow_engine.go`
- `examples/iot/*.go` (所有示例)

#### 部分工具文件

- `pkg/tools/sandbox/` 下的部分文件
- 测试文件可能需要简化声明

### 中优先级（辅助模块）

- 部分示例代码
- 部分测试代码
- 配置文件

---

## 🔧 修复建议

### 1. 自动添加许可声明（推荐）

创建脚本自动添加许可声明：

```bash
#!/bin/bash
# add_license.sh

LICENSE_HEADER='// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
'

# 查找没有许可声明的 Go 文件
find . -name "*.go" -type f ! -exec grep -q "SPDX-License-Identifier" {} \; | while read file; do
    echo "Adding license to: $file"

    # 创建临时文件
    temp_file=$(mktemp)

    # 写入许可声明
    echo "$LICENSE_HEADER" > "$temp_file"
    echo "" >> "$temp_file"

    # 追加原文件内容（跳过现有的 shebang 或空行）
    awk 'NF{p=1} p' "$file" >> "$temp_file"

    # 替换原文件
    mv "$temp_file" "$file"
done
```

### 2. 手动添加

对于关键文件，建议手动添加以确保准确性：

**pkg/iot/types.go**:
```go
// Package iot provides unified IoT device management and protocol abstraction
// supporting Zigbee, Thread, Z-Wave, and NearLink protocols.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package iot
```

### 3. 使用工具

使用 **addlicense** 工具：

```bash
# 安装工具
go install github.com/google/addlicense@latest

# 添加许可声明
addlicense -c "Agent Framework Contributors" \
  -l "AGPL-3.0-or-later" \
  -skip yaml \
  -skip json \
  ./pkg/iot/...

# 验证
addlicense -check ./...
```

---

## 📝 AGPL-3.0 要求检查清单

### 文件级要求

- [x] **根目录 LICENSE 文件** ✅ 存在且完整
- [x] **源文件声明** ⚠️ 69% 覆盖，需要提升到 100%
- [ ] **SPDX 标识符** ⚠️ 部分文件使用
- [x] **Copyright 声明** ✅ 大部分文件包含
- [x] **许可声明** ✅ 使用标准 AGPL-3.0 文本

### AGPL-3.0 特殊要求

- [x] **网络使用源代码可用** ⚠️ 需要在 README 中说明
- [ ] **交互界面法律声明** ⚠️ 如果有 Web UI，需要添加
- [ ] **源代码提供方式** ⚠️ 需要明确说明如何获取源代码

---

## 🎯 立即行动项

### 优先级 P0（本周完成）

1. **为 pkg/iot/ 模块添加许可声明**
   ```bash
   # 使用 addlicense 工具
   addlicense -c "Agent Framework Contributors" \
     -l AGPL-3.0-or-later \
     ./pkg/iot/... ./examples/iot/...
   ```

2. **为 pkg/channels/ 模块添加 SPDX 标识符**
   - 大部分已有声明，需要统一添加 SPDX 标识符

3. **检查并更新 README.md**
   - 添加 AGPL-3.0 说明
   - 说明如何获取源代码
   - 添加网络使用条款说明

### 优先级 P1（本月完成）

4. **为所有 Go 文件添加许可声明**
   - 目标：100% 覆盖
   - 使用自动化工具

5. **为测试文件添加简化声明**
   ```go
   // Copyright (C) 2025 Agent Framework Contributors
   // SPDX-License-Identifier: AGPL-3.0-or-later
   ```

6. **验证许可证合规性**
   ```bash
   # 使用 go-licenses 检查
   go install github.com/google/go-licenses@latest
   go-licenses check ./...
   ```

### 优先级 P2（下月完成）

7. **添加 CI/CD 许可证检查**
   - 在 CI 中集成许可证检查
   - 防止新文件缺少许可声明

8. **创建 CONTRIBUTING.md 许可证部分**
   - 说明贡献者许可协议
   - 指导如何添加许可声明

---

## 📚 参考资料

### 许可证文档

- [AGPL-3.0 完整文本](../LICENSE)
- [Free Software Foundation 许可证指南](https://www.gnu.org/licenses/agpl-3.0.html)
- [SPDX 标识符列表](https://spdx.org/licenses/)

### 工具

- **addlicense**: https://github.com/google/addlicense
- **go-licenses**: https://github.com/google/go-licenses
- **licensecheck**: https://github.com/google/go-licenses

### 最佳实践

- [REUSE 规范](https://reuse.software/)
- [SPDX 使用指南](https://spdx.dev/)

---

## 🔍 持续监控建议

### 自动化检查

在 CI/CD 中添加许可证检查：

```yaml
# .github/workflows/license-check.yml
name: License Check

on: [push, pull_request]

jobs:
  license:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install addlicense
        run: go install github.com/google/addlicense@latest

      - name: Check licenses
        run: |
          addlicense -check ./...

      - name: Check SPDX identifiers
        run: |
          # 查找没有 SPDX 标识符的 .go 文件
          ! grep -rL "SPDX-License-Identifier" --include="*.go" .
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash

# 检查新文件是否有许可声明
for file in $(git diff --cached --name-only --diff-filter=A | grep '\.go$'); do
    if ! grep -q "SPDX-License-Identifier" "$file"; then
        echo "错误: $file 缺少 SPDX-License-Identifier"
        exit 1
    fi
done
```

---

## ✅ 总结

### 当前状态

| 项目 | 状态 |
|------|------|
| 许可证文件 | ✅ 完整 |
| 源代码覆盖 | ⚠️ 69% (目标: 100%) |
| 格式一致性 | ⚠️ 需要改进 |
| SPDX 标识符 | ⚠️ 部分使用 |
| CI/CD 集成 | ❌ 缺失 |

### 行动计划

1. **本周**: 为 pkg/iot/ 模块添加许可声明
2. **本月**: 实现 100% 源代码覆盖
3. **下月**: 添加 CI/CD 自动检查

### 联系方式

如有许可证相关问题，请联系：
- 项目维护者: Agent Framework Team
- 许可证咨询: 见 LICENSE 文件

---

**检查完成时间**: 2026-02-19
**下次检查**: 建议每月检查一次
**负责人**: Agent Framework Team
**状态**: ⚠️ 需要改进
