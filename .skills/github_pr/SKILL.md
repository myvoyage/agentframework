---
id: "com.agentframework.skill.github_pr"
name: "github_pr"
version: "1.0.0"
category: "development"
tags:
  - github
  - pull-requests
  - git
description: "管理 GitHub Pull Requests"
author: "AgentFramework Team"
enabled: true

prerequisites:
  bins:
    - name: "gh"
      version: ">=2.0.0"
      install:
        brew: "brew install gh"
        apt: "sudo apt install gh"
        apk: "apk add gh"
  env:
    - name: "GITHUB_TOKEN"
      optional: true
      description: "GitHub Personal Access Token with repo scope"

triggers:
  - type: "command"
    pattern: "/pr"
    priority: 10
  - type: "keyword"
    pattern: "pull request"
    priority: 5

actions:
  - id: "list_prs"
    type: "shell"
    description: "列出 Pull Requests"
    config:
      command: "gh pr list --repo {{.Repo}} --state {{.State}} --limit {{.Limit}}"
    timeout: "30s"

  - id: "view_pr"
    type: "shell"
    description: "查看 PR 详情"
    config:
      command: "gh pr view {{.Number}} --repo {{.Repo}} --json title,state,author,url"
    timeout: "30s"

  - id: "create_pr"
    type: "shell"
    description: "创建 Pull Request"
    config:
      command: "gh pr create --repo {{.Repo}} --title \"{{.Title}}\" --body \"{{.Body}}\" --base {{.Base}}"
    timeout: "60s"

config:
  max_output_size: 10485760
  max_execution_time: "60s"
  enable_cache: true
  cache_ttl: "5m"

always: false

---

# GitHub Pull Requests 技能

管理 GitHub Pull Requests 的强大工具。

## 功能

- 列出 Pull Requests
- 查看 PR 详情
- 创建新的 Pull Request

## 使用示例

### 列出 PR

\`\`\`bash
# 列出所有打开的 PR
/pr list --repo owner/repo --state open --limit 10

# 列出已合并的 PR
/pr list --repo owner/repo --state merged
\`\`\`

### 创建 PR

\`\`\`bash
/pr create --repo owner/repo --title "Fix bug" --body "Description" --base main
\`\`\`

### 查看 PR

\`\`\`bash
/pr view --repo owner/repo --number 123
\`\`\`

## 前置条件

1. **GitHub CLI (gh)**: 版本 >= 2.0.0
   - macOS: `brew install gh`
   - Ubuntu: `sudo apt install gh`
   - Alpine: `apk add gh`

2. **GitHub Token**: 环境变量 `GITHUB_TOKEN`
   - 需要包含 `repo` 权限
   - 可在 https://github.com/settings/tokens 创建

## 故障排除

### 问题：gh 命令未找到
**解决**：安装 GitHub CLI
- macOS: `brew install gh`
- Linux: 根据发行版选择包管理器

### 问题：认证失败
**解决**：设置 GitHub Token
\`\`\`bash
export GITHUB_TOKEN=your_token_here
gh auth login
\`\`\`

### 问题：权限不足
**解决**：确保 Token 包含所需权限
- 需要 `repo` 权限来访问仓库
- 需要 `pull_request` 权限来管理 PR
