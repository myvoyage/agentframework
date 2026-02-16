---
id: "com.agentframework.skill.git_helper"
name: "git_helper"
version: "1.0.0"
category: "development"
tags:
  - git
  - version-control
  - development
description: "Git 版本控制助手"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "git"
      version: ">=2.0.0"
      install:
        apt: "sudo apt install git"
        brew: "brew install git"
        apk: "apk add git"
  env: []

triggers:
  - type: "command"
    pattern: "/git"
    priority: 10
  - type: "keyword"
    pattern: "git helper"
    priority: 5

actions:
  - id: "status"
    type: "shell"
    description: "显示仓库状态"
    config:
      command: "cd {{.RepoPath}} && git status -sb"
    timeout: "5s"

  - id: "quick_commit"
    type: "shell"
    description: "快速提交更改"
    config:
      command: "cd {{.RepoPath}} && git add -A && git commit -m \"{{.Message}}\""
    timeout: "10s"

  - id: "create_branch"
    type: "shell"
    description: "创建新分支"
    config:
      command: "cd {{.RepoPath}} && git checkout -b {{.BranchName}}"
    timeout: "5s"

  - id: "list_branches"
    type: "shell"
    description: "列出所有分支"
    config:
      command: "cd {{.RepoPath}} && git branch -a"
    timeout: "5s"

  - id: "sync_branch"
    type: "shell"
    description: "同步分支（拉取并变基）"
    config:
      command: "cd {{.RepoPath}} && git fetch && git rebase origin/{{.BranchName}}"
    timeout: "30s"

  - id: "view_history"
    type: "shell"
    description: "查看提交历史"
    config:
      command: "cd {{.RepoPath}} && git log --oneline --graph --decorate -{{.Count}}"
    timeout: "10s"

  - id: "diff_staged"
    type: "shell"
    description: "查看暂存的更改"
    config:
      command: "cd {{.RepoPath}} && git diff --cached"
    timeout: "10s"

  - id: "discard_changes"
    type: "shell"
    description: "丢弃所有未提交的更改"
    config:
      command: "cd {{.RepoPath}} && git reset --hard && git clean -fd"
    timeout: "10s"

  - id: "cherry_pick"
    type: "shell"
    description: "拣选指定提交"
    config:
      command: "cd {{.RepoPath}} && git cherry-pick {{.CommitHash}}"
    timeout: "30s"

  - id: "show_remote"
    type: "shell"
    description: "显示远程仓库信息"
    config:
      command: "cd {{.RepoPath}} && git remote -v"
    timeout: "5s"

config:
  max_output_size: 5242880
  max_execution_time: "60s"
  enable_cache: false
  cache_ttl: "0s"

always: false

---

# Git 助手技能

简化 Git 工作流程的实用工具集。

## 功能

- **状态查看**: 快速查看仓库状态
- **快速提交**: 一键添加并提交所有更改
- **分支管理**: 创建、列出、同步分支
- **历史查看**: 查看提交历史
- **更改管理**: 查看差异、丢弃更改
- **提交操作**: 拣选提交、查看远程

## 使用示例

### 查看状态

```bash
# 查看当前仓库状态
agentframework enhanced-skill execute com.agentframework.skill.git_helper status --vars "RepoPath=."

# 查看指定目录状态
agentframework enhanced-skill execute com.agentframework.skill.git_helper status --vars "RepoPath=/path/to/repo"
```

### 快速提交

```bash
# 提交所有更改
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=fix: 修复登录问题"

# 功能开发提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=feat: 添加用户设置页面"
```

### 分支操作

```bash
# 创建新分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper create_branch --vars "RepoPath=.,BranchName=feature/new-ui"

# 列出所有分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper list_branches --vars "RepoPath=."

# 同步主分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper sync_branch --vars "RepoPath=.,BranchName=main"
```

### 查看历史

```bash
# 查看最近 10 次提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper view_history --vars "RepoPath=.,Count=10"

# 查看最近 20 次提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper view_history --vars "RepoPath=.,Count=20"
```

### 更改管理

```bash
# 查看暂存的更改
agentframework enhanced-skill execute com.agentframework.skill.git_helper diff_staged --vars "RepoPath=."

# ⚠️ 丢弃所有未提交的更改（危险操作）
agentframework enhanced-skill execute com.agentframework.skill.git_helper discard_changes --vars "RepoPath=."
```

### 高级操作

```bash
# 拣选指定提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper cherry_pick --vars "RepoPath=.,CommitHash=abc123"

# 查看远程仓库
agentframework enhanced-skill execute com.agentframework.skill.git_helper show_remote --vars "RepoPath=."
```

## 参数说明

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| RepoPath | 仓库路径 | 当前目录 | `.`, `/path/to/repo` |
| BranchName | 分支名称 | - | `feature/new-ui`, `hotfix/bug-123` |
| CommitHash | 提交哈希 | - | `abc123`, `def456` |
| Message | 提交消息 | - | `fix: 修复问题` |
| Count | 数量 | 10 | `20`, `50` |

## 提交消息规范

使用约定式提交格式：

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat:` | 新功能 | `feat: 添加用户设置` |
| `fix:` | 修复 Bug | `fix: 修复登录问题` |
| `docs:` | 文档更新 | `docs: 更新 README` |
| `style:` | 代码格式 | `style: 格式化代码` |
| `refactor:` | 重构 | `refactor: 重构 API 层` |
| `test:` | 测试 | `test: 添加单元测试` |
| `chore:` | 构建/工具 | `chore: 更新依赖` |

## 典型工作流

### 功能开发

```bash
# 1. 创建功能分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper create_branch --vars "RepoPath=.,BranchName=feature/new-feature"

# 2. 开发并提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=feat: 添加新功能"

# 3. 同步主分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper sync_branch --vars "RepoPath=.,BranchName=main"
```

### Bug 修复

```bash
# 1. 创建修复分支
agentframework enhanced-skill execute com.agentframework.skill.git_helper create_branch --vars "RepoPath=.,BranchName=hotfix/critical-bug"

# 2. 修复并提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=fix: 修复关键 Bug"

# 3. 查看状态
agentframework enhanced-skill execute com.agentframework.skill.git_helper status --vars "RepoPath=."
```

## ⚠️ 危险操作

### discard_changes

此命令会**永久删除**所有未提交的更改：

```bash
# 这将丢弃所有工作！
agentframework enhanced-skill execute com.agentframework.skill.git_helper discard_changes --vars "RepoPath=."
```

**执行前请确认：**
1. 是否有重要的未保存更改？
2. 是否已经提交了所有必要的工作？
3. 是否可以接受数据丢失？

## 最佳实践

### 1. 定期提交

```bash
# 完成一个功能后立即提交
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=feat: 完成用户认证"
```

### 2. 清晰的提交消息

```bash
# 好的提交消息
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=feat(auth): 添加 OAuth2 登录支持"

# 不好的提交消息
agentframework enhanced-skill execute com.agentframework.skill.git_helper quick_commit --vars "RepoPath=.,Message=update"
```

### 3. 分支策略

```bash
# 功能分支
feature/功能名称

# 修复分支
hotfix/问题描述

# 发布分支
release/版本号
```

## 故障排除

### 权限错误
```bash
# 检查仓库权限
ls -la .git/

# 修复权限
chmod -R g+rw .git/
```

### 合并冲突
```bash
# 查看冲突
git status

# 手动解决后继续
git add .
git rebase --continue
```

### 远步同步问题
```bash
# 查看远程
agentframework enhanced-skill execute com.agentframework.skill.git_helper show_remote --vars "RepoPath=."

# 添加远程
git remote add origin https://github.com/user/repo.git
```

## 集成建议

### 与 IDE 集成

将这些命令配置到 IDE 的快捷键：
- `Ctrl+Shift+C`: 快速提交
- `Ctrl+Shift+S`: 查看状态
- `Ctrl+Shift+B`: 创建分支

### 与 CI/CD 集成

在 CI/CD 流程中使用：
```yaml
# 示例 CI 配置
script:
  - git_helper status
  - git_helper sync_branch main
```

## 相关资源

- [Git 官方文档](https://git-scm.com/doc)
- [约定式提交](https://www.conventionalcommits.org/)
- [Git 分支管理最佳实践](https://www.atlassian.com/git/tutorials/comparing-workflows)
