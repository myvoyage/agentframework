---
name: TextProcessor
description: 高级文本处理技能，支持多种文本操作和分析
version: 1.2.0
category: data
author: AgentFramework Team
license: MIT
tags:
  - text
  - processing
  - analysis
  - utility
---

# TextProcessor 技能

## 概述

TextProcessor 是一个强大的文本处理技能，提供多种文本操作和分析功能。该技能可以帮助用户快速处理、转换和分析文本数据。

## 主要功能

### 1. 文本清理
- 移除多余空格
- 删除特殊字符
- 标准化换行符

### 2. 文本分析
- 字数统计
- 行数统计
- 关键词提取
- 情感分析

### 3. 文本转换
- 大小写转换
- 编码转换
- 格式转换

## 使用方法

### 基本用法

在输入中提供需要处理的文本，技能将根据选择的操作类型进行处理。

### 高级用法

指定操作类型和参数：
```json
{
  "operation": "clean|analyze|convert",
  "options": {
    "removeSpaces": true,
    "removeSpecialChars": false,
    "case": "lower|upper|title"
  }
}
```

## 工作流程

1. **输入验证**: 验证输入文本格式
2. **参数解析**: 解析操作选项
3. **文本处理**: 执行指定的文本操作
4. **结果生成**: 生成处理结果和统计信息
5. **返回输出**: 返回格式化的结果

## 脚本工具

本技能包含以下脚本工具：

- `scripts/counter.py`: Python脚本，用于文本统计
- `scripts/converter.sh`: Bash脚本，用于格式转换

详见 [脚本文档](#脚本文档) 章节。

## 参考文档

- [文本处理最佳实践](references/best_practices.md)
- [API参考](references/api_reference.md)

## 配置

```yaml
cache:
  enabled: true
  ttl: 120s

timeout: 60s
retries: 3

performance:
  max_text_size: 10MB
  chunk_size: 4096
```

## 示例

### 示例1：清理文本
**输入**:
```
"  Hello    World!  \n\n  Test  "
```

**输出**:
```
"Hello World!\nTest"
```

### 示例2：统计信息
**输入**:
```
"Hello World! This is a test."
```

**输出**:
```json
{
  "characters": 28,
  "words": 6,
  "lines": 1,
  "sentences": 2
}
```

### 示例3：大小写转换
**输入**:
```
{
  "text": "hello world",
  "case": "title"
}
```

**输出**:
```
"Hello World"
```

## 错误处理

| 错误代码 | 描述 | 解决方案 |
|---------|------|---------|
| EMPTY_INPUT | 输入文本为空 | 提供有效的输入文本 |
| TEXT_TOO_LARGE | 文本超过最大限制 | 分段处理或减小文本大小 |
| INVALID_OPERATION | 无效的操作类型 | 检查操作类型是否正确 |

## 版本历史

- **v1.2.0** (2025-01-28):
  - 添加情感分析功能
  - 优化性能
  - 修复文本清理bug

- **v1.1.0** (2024-12-15):
  - 添加编码转换功能
  - 改进错误处理

- **v1.0.0** (2024-11-01):
  - 初始版本
  - 基本文本处理功能

## 性能指标

- 平均处理时间：<100ms (1KB文本)
- 最大吞吐量：10MB/s
- 内存占用：<50MB

## 依赖项

- Python 3.8+ (用于counter.py)
- Bash 4.0+ (用于converter.sh)

## 作者

AgentFramework Team

## 许可证

MIT License - 详见 LICENSE 文件

## 贡献

欢迎提交 Issue 和 Pull Request！

## 脚本文档

### counter.py

Python脚本，用于执行高级文本统计。

**功能**:
- 字数统计
- 字符统计（含/不含空格）
- 段落统计
- 关键词频率分析

**使用方法**:
```bash
python scripts/counter.py --input "your text" --output json
```

### converter.sh

Bash脚本，用于文本格式转换。

**功能**:
- 编码转换（UTF-8, GBK, ASCII等）
- 换行符转换（LF, CRLF）
- Base64 编码/解码

**使用方法**:
```bash
./scripts/converter.sh --encode utf8 --normalize lf
```
