# AgentFramework 增强技能

这个目录包含增强技能系统的所有技能定义。

## 目录结构

```
.skills/
├── github_pr/              # GitHub PR 管理技能
│   └── SKILL.md
├── network_diagnostics/    # 网络诊断技能
│   └── SKILL.md
├── file_manager/          # 文件管理技能
│   └── SKILL.md
├── git_helper/            # Git 助手技能
│   └── SKILL.md
└── template/              # 技能开发模板
    └── SKILL.md
```

## 创建新技能

### 快速开始

1. 复制模板：
```bash
cp -r template my_new_skill
cd my_new_skill
```

2. 编辑 `SKILL.md` 文件，修改 YAML frontmatter

3. 测试技能：
```bash
# 列出技能
agentframework enhanced-skill list

# 检查依赖
agentframework enhanced-skill check com.example.skill.my-new-skill

# 执行动作
agentframework enhanced-skill execute com.example.skill.my-new-skill my_action --vars "Param=Value"
```

### 详细指南

请参阅 [技能开发指南](../docs/SKILL_DEVELOPMENT.md) 获取完整的开发教程。

## 技能列表

| 技能 | ID | 分类 | 状态 |
|------|----|----|----|
| GitHub PR 管理 | `com.agentframework.skill.github_pr` | development | ✅ |
| 网络诊断 | `com.agentframework.skill.network_diagnostics` | system | ⚠️ 需要依赖 |
| 文件管理 | `com.agentframework.skill.file_manager` | utility | ✅ |
| Git 助手 | `com.agentframework.skill.git_helper` | development | ✅ |

## 管理技能

### 列出技能

```bash
agentframework enhanced-skill list
```

### 搜索技能

```bash
# 根据触发器搜索
agentframework enhanced-skill search command "/git"

# 根据关键词搜索
agentframework enhanced-skill search keyword "git"
```

### 查看详情

```bash
agentframework enhanced-skill get com.agentframework.skill.git_helper
```

### 检查依赖

```bash
agentframework enhanced-skill check com.agentframework.skill.git_helper
```

### 执行动作

```bash
# 基本执行
agentframework enhanced-skill execute com.agentframework.skill.git_helper status --vars "RepoPath=."

# 带多个变量
agentframework enhanced-skill execute com.agentframework.skill.file_manager find_large_files --vars "Path=.,Size=100M"
```

### 安装新技能

```bash
# 从 GitHub 安装
agentframework enhanced-skill install username/skill-repo
```

## 贡献技能

欢迎分享您的技能！

1. Fork 项目
2. 创建技能目录和 SKILL.md
3. 测试技能功能
4. 提交 Pull Request

## 技能规范

### ID 命名

使用反向域名格式：
```
com.organization.category.skill-name
```

### 文件命名

- 目录名：小写，下划线分隔
- SKILL.md：固定名称

### 最小要求

每个技能必须包含：
- ✅ 唯一的 ID
- ✅ 名称和描述
- ✅ 至少一个动作
- ✅ 依赖声明（如需要）

## 相关文档

- [CLI 技能使用指南](../docs/CLI_SKILLS.md)
- [技能开发指南](../docs/SKILL_DEVELOPMENT.md)
- [增强技能系统架构](../agent/skills/README.md)

## 支持

- 问题反馈：[GitHub Issues](https://github.com/your-repo/issues)
- 功能建议：[GitHub Discussions](https://github.com/your-repo/discussions)

---

**开始创建你的第一个技能吧！** 🚀
