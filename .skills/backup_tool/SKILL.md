---
id: "com.agentframework.skill.backup_tool"
name: "backup_tool"
version: "1.0.0"
category: "utility"
tags:
  - backup
  - automation
  - productivity
description: "简单备份工具"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "tar"
      version: "any"
      install:
        apt: "tar ships with GNU tar"
        brew: "tar ships with macOS"
        apk: "apk add tar"
    - name: "rsync"
      version: "any"
      install:
        apt: "sudo apt install rsync"
        brew: "brew install rsync"
        apk: "apk add rsync"
  env: []

triggers:
  - type: "command"
    pattern: "/backup"
    priority: 10
  - type: "keyword"
    pattern: "backup tool"
    priority: 5

actions:
  - id: "create_archive"
    type: "shell"
    description: "创建压缩归档"
    config:
      command: "tar -czf {{.Output}} {{.Input}}"
    timeout: "300s"

  - id: "extract_archive"
    type: "shell"
    description: "提取压缩归档"
    config:
      command: "tar -xzf {{.Input}} -C {{.Output}}"
    timeout: "300s"

  - id: "sync_dirs"
    type: "shell"
    description: "同步目录"
    config:
      command: "rsync -av --delete {{.Source}}/ {{.Dest}}/"
    timeout: "600s"

  - id: "backup_with_timestamp"
    type: "shell"
    description: "带时间戳的备份"
    config:
      command: "tar -czf {{.Dest}}/backup_$(date +%Y%m%d_%H%M%S).tar.gz {{.Source}}"
    timeout: "300s"

  - id: "incremental_backup"
    type: "shell"
    description: "增量备份"
    config:
      command: "rsync -av --backup --backup-dir={{.BackupDir}} {{.Source}}/ {{.Dest}}/"
    timeout: "600s"

  - id: "list_archives"
    type: "shell"
    description: "列出归档文件"
    config:
      command: "ls -lh {{.Dir}}/*.tar.gz 2>/dev/null || ls -lh {{.Dir}}/*.tgz 2>/dev/null || echo 'No archives found'"
    timeout: "10s"

  - id: "verify_backup"
    type: "shell"
    description: "验证备份完整性"
    config:
      command: "tar -tzf {{.Archive}} | head -20"
    timeout: "30s"

  - id: "get_backup_size"
    type: "shell"
    description: "获取备份大小"
    config:
      command: "du -sh {{.Path}}"
    timeout: "10s"

  - id: "rotate_backups"
    type: "shell"
    description: "轮转备份（保留最近的N个）"
    config:
      command: "ls -t {{.Dir}}/backup_*.tar.gz | tail -n +{{.Keep+1}} | xargs -r rm"
    timeout: "30s"

  - id: "compress_file"
    type: "shell"
    description: "压缩单个文件"
    config:
      command: "gzip -c {{.Input}} > {{.Output}}.gz"
    timeout: "300s"

  - id: "decompress_file"
    type: "shell"
    description: "解压缩文件"
    config:
      command: "gunzip -c {{.Input}} > {{.Output}}"
    timeout: "300s"

config:
  max_output_size: 10485760
  max_execution_time: "600s"
  enable_cache: false
  cache_ttl: "0s"

always: false

---

# 备份工具技能

简单实用的数据备份和归档工具。

## 功能

- **创建归档**: 创建 .tar.gz 压缩归档
- **提取归档**: 解压缩归档文件
- **目录同步**: 使用 rsync 同步目录
- **时间戳备份**: 自动添加时间戳
- **增量备份**: 只备份更改的文件
- **备份验证**: 验证备份完整性
- **备份轮转**: 自动清理旧备份
- **文件压缩**: 压缩单个文件

## 使用示例

### 创建备份

```bash
# 创建压缩归档
agentframework enhanced-skill execute com.agentframework.skill.backup_tool create_archive --vars "Input=/path/to/source,Output=backup.tar.gz"

# 带时间戳的备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/path/to/source,Dest=/backup/dir"

# 压缩单个文件
agentframework enhanced-skill execute com.agentframework.skill.backup_tool compress_file --vars "Input=file.txt,Output=file.txt"
```

### 提取备份

```bash
# 提取归档
agentframework enhanced-skill execute com.agentframework.skill.backup_tool extract_archive --vars "Input=backup.tar.gz,Output=/path/to/extract"

# 解压文件
agentframework enhanced-skill execute com.agentframework.skill.backup_tool decompress_file --vars "Input=file.txt.gz,Output=file.txt"
```

### 同步目录

```bash
# 同步目录（镜像）
agentframework enhanced-skill execute com.agentframework.skill.backup_tool sync_dirs --vars "Source=/path/to/source,Dest=/path/to/dest"

# 增量备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool incremental_backup --vars "Source=/path/to/source,Dest=/path/to/dest,BackupDir=/path/to/old"
```

### 管理备份

```bash
# 列出所有备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool list_archives --vars "Dir=/backup/dir"

# 验证备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool verify_backup --vars "Archive=backup.tar.gz"

# 获取备份大小
agentframework enhanced-skill execute com.agentframework.skill.backup_tool get_backup_size --vars "Path=/backup/dir"

# 轮转备份（保留最近的 7 个）
agentframework enhanced-skill execute com.agentframework.skill.backup_tool rotate_backups --vars "Dir=/backup/dir,Keep=7"
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| Input | 源文件/目录 | `/path/to/source` |
| Output | 输出文件/目录 | `backup.tar.gz` |
| Source | 源目录 | `/path/to/source` |
| Dest | 目标目录 | `/path/to/dest` |
| Dir | 目录路径 | `/backup/dir` |
| Archive | 归档文件 | `backup.tar.gz` |
| BackupDir | 备份目录 | `/path/to/old` |
| Keep | 保留数量 | `7` |
| Path | 路径 | `/backup/dir` |

## 备份策略

### 1. 完整备份

```bash
# 每天完整备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/data,Dest=/backups"
```

### 2. 增量备份

```bash
# 使用 rsync 进行增量备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool incremental_backup --vars "Source=/data,Dest=/backups/current,BackupDir=/backups/incremental"
```

### 3. 镜像备份

```bash
# 完全镜像源目录
agentframework enhanced-skill execute com.agentframework.skill.backup_tool sync_dirs --vars "Source=/data,Dest=/backups/mirror"
```

## 自动化备份

### 定期备份脚本

```bash
#!/bin/bash
# daily_backup.sh

# 设置变量
SOURCE="/data"
DEST="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=$SOURCE,Dest=$DEST"

# 轮转备份（保留 7 天）
agentframework enhanced-skill execute com.agentframework.skill.backup_tool rotate_backups --vars "Dir=$DEST,Keep=7"
```

### 添加到 crontab

```bash
# 每天凌晨 2 点备份
0 2 * * * /path/to/daily_backup.sh

# 每小时备份
0 * * * * /path/to/daily_backup.sh

# 每周备份
0 2 * * 0 /path/to/daily_backup.sh
```

## 备份最佳实践

### 1. 3-2-1 备份原则

- **3** 份副本
- **2** 种不同介质
- **1** 份异地备份

### 2. 备份验证

```bash
# 定期验证备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool verify_backup --vars "Archive=/backups/backup.tar.gz"

# 测试恢复
agentframework enhanced-skill execute com.agentframework.skill.backup_tool extract_archive --vars "Input=/backups/backup.tar.gz,Output=/test/restore"
```

### 3. 备份加密

```bash
# 加密备份
tar -czf - /data | gpg -e -r user@example.com > backup.tar.gz.gpg

# 解密恢复
gpg -d backup.tar.gz.gpg | tar -xzf -
```

### 4. 监控备份

```bash
# 检查备份大小
SIZE=$(agentframework enhanced-skill execute com.agentframework.skill.backup_tool get_backup_size --vars "Path=/backups")

# 发送告警（如果大小异常）
if [ "$SIZE" -lt 1000 ]; then
  echo "Backup size too small!" | mail -s "Backup Alert" admin@example.com
fi
```

## 备份方案示例

### 开发环境备份

```bash
# 备份项目代码
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/home/user/projects,Dest=/backups/projects"

# 备份配置文件
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/home/user/.config,Dest=/backups/config"
```

### 数据库备份

```bash
# 备份前先导出数据库
mysqldump -u user -p database > database.sql

# 压缩备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool compress_file --vars "Input=database.sql,Output=database.sql"
```

### 系统备份

```bash
# 备份重要系统文件
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/etc,Dest=/backups/system"

# 备份用户数据
agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/home,Dest=/backups/home"
```

## 恢复流程

### 1. 验证备份

```bash
# 检查备份完整性
agentframework enhanced-skill execute com.agentframework.skill.backup_tool verify_backup --vars "Archive=backup.tar.gz"
```

### 2. 提取备份

```bash
# 提取到临时目录
agentframework enhanced-skill execute com.agentframework.skill.backup_tool extract_archive --vars "Input=backup.tar.gz,Output=/tmp/restore"
```

### 3. 验证恢复

```bash
# 检查恢复的文件
ls -la /tmp/restore

# 验证内容
diff -r /tmp/restore /original
```

### 4. 完成恢复

```bash
# 移动到最终位置
mv /tmp/restore/* /target/location
```

## 性能优化

### 1. 并行压缩

```bash
# 使用 pigz 进行并行压缩
tar -c /data | pigz -p 4 > backup.tar.gz
```

### 2. 排除文件

```bash
# 排除不需要备份的文件
tar -czf backup.tar.gz --exclude='*.tmp' --exclude='*.log' /data
```

### 3. 增量备份

```bash
# 使用 rsync 的 --link-dest 选项进行硬链接增量备份
rsync -av --link-dest=/backups/yesterday /data /backups/today
```

## 故障排除

### 磁盘空间不足

```bash
# 检查磁盘使用
df -h

# 清理旧备份
agentframework enhanced-skill execute com.agentframework.skill.backup_tool rotate_backups --vars "Dir=/backups,Keep=3"
```

### 备份失败

```bash
# 检查文件权限
ls -la /data

# 使用 sudo
sudo agentframework enhanced-skill execute com.agentframework.skill.backup_tool backup_with_timestamp --vars "Source=/data,Dest=/backups"
```

### 恢复失败

```bash
# 验证备份完整性
agentframework enhanced-skill execute com.agentframework.skill.backup_tool verify_backup --vars "Archive=backup.tar.gz"

# 尝试部分恢复
tar -xzf backup.tar.gz --wildcards '*.txt'
```

## 安全建议

### 1. 加密敏感数据

```bash
# 使用 GPG 加密
gpg --symmetric --cipher-algo AES256 backup.tar.gz
```

### 2. 安全存储

```bash
# 设置合适的权限
chmod 600 backup.tar.gz
chown user:group backup.tar.gz
```

### 3. 异地备份

```bash
# 同步到远程服务器
rsync -avz -e ssh /backups/ user@remote:/backups/
```

### 4. 测试恢复

```bash
# 定期测试恢复流程
# 确保备份可用
```

## 相关工具

- `rsync`: 远程同步工具
- `rdiff-backup`: 增量备份工具
- `duplicity`: 加密备份工具
- `borg`: 去重备份工具
- `restic`: 现代备份工具
