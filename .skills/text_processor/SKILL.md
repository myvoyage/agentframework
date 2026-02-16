---
id: "com.agentframework.skill.text_processor"
name: "text_processor"
version: "1.0.0"
category: "utility"
tags:
  - text
  - utilities
  - conversion
description: "文本处理和转换工具"
author: "AgentFramework Team"
license: "MIT"
enabled: true

prerequisites:
  bins:
    - name: "sed"
      version: "any"
      install:
        apt: "sed ships with GNU sed"
        brew: "sed ships with macOS"
        apk: "apk add sed"
    - name: "awk"
      version: "any"
      install:
        apt: "awk ships with GNU awk"
        brew: "awk ships with macOS"
        apk: "apk add awk"
    - name: "tr"
      version: "any"
      install:
        apt: "tr ships with coreutils"
        brew: "tr ships with macOS"
        apk: "apk add coreutils"
  env: []

triggers:
  - type: "command"
    pattern: "/text"
    priority: 10
  - type: "keyword"
    pattern: "text processor"
    priority: 5

actions:
  - id: "to_upper"
    type: "shell"
    description: "转换为大写"
    config:
      command: "echo '{{.Text}}' | tr '[:lower:]' '[:upper:]'"
    timeout: "5s"

  - id: "to_lower"
    type: "shell"
    description: "转换为小写"
    config:
      command: "echo '{{.Text}}' | tr '[:upper:]' '[:lower:]'"
    timeout: "5s"

  - id: "to_title"
    type: "shell"
    description: "转换为首字母大写"
    config:
      command: "echo '{{.Text}}' | sed 's/\\b\\(\\w\\)/\\u\\1/g'"
    timeout: "5s"

  - id: "trim_spaces"
    type: "shell"
    description: "删除前后空格"
    config:
      command: "echo '{{.Text}}' | xargs"
    timeout: "5s"

  - id: "remove_lines"
    type: "shell"
    description: "删除匹配的行"
    config:
      command: "echo '{{.Text}}' | sed '/{{.Pattern}}/d'"
    timeout: "5s"

  - id: "extract_lines"
    type: "shell"
    description: "提取匹配的行"
    config:
      command: "echo '{{.Text}}' | grep '{{.Pattern}}' || echo '{{.Text}}' | grep -E '{{.Pattern}}'"
    timeout: "5s"

  - id: "replace_text"
    type: "shell"
    description: "替换文本"
    config:
      command: "echo '{{.Text}}' | sed 's/{{.Old}}/{{.New}}/g'"
    timeout: "5s"

  - id: "count_lines"
    type: "shell"
    description: "统计行数"
    config:
      command: "echo '{{.Text}}' | wc -l"
    timeout: "5s"

  - id: "count_words"
    type: "shell"
    description: "统计字数"
    config:
      command: "echo '{{.Text}}' | wc -w"
    timeout: "5s"

  - id: "count_chars"
    type: "shell"
    description: "统计字符数"
    config:
      command: "echo '{{.Text}}' | wc -c"
    timeout: "5s"

  - id: "reverse_lines"
    type: "shell"
    description: "反转行顺序"
    config:
      command: "echo '{{.Text}}' | tac"
    timeout: "5s"

  - id: "sort_lines"
    type: "shell"
    description: "排序行"
    config:
      command: "echo '{{.Text}}' | sort"
    timeout: "5s"

  - id: "unique_lines"
    type: "shell"
    description: "去重行"
    config:
      command: "echo '{{.Text}}' | sort -u"
    timeout: "10s"

  - id: "base64_encode"
    type: "shell"
    description: "Base64 编码"
    config:
      command: "echo '{{.Text}}' | base64"
    timeout: "5s"

  - id: "base64_decode"
    type: "shell"
    description: "Base64 解码"
    config:
      command: "echo '{{.Text}}' | base64 -d"
    timeout: "5s"

  - id: "url_encode"
    type: "shell"
    description: "URL 编码"
    config:
      command: "echo '{{.Text}}' | jq -sRr @uri"
    timeout: "5s"

  - id: "url_decode"
    type: "shell"
    description: "URL 解码"
    config:
      command: "echo '{{.Text}}' | sed 's/+/ /g' | sed 's/%/\\x/g' | xargs -0 printf '%b'"
    timeout: "5s"

  - id: "remove_duplicates"
    type: "shell"
    description: "删除重复的单词"
    config:
      command: "echo '{{.Text}}' | tr ' ' '\\n' | sort -u | tr '\\n' ' '"
    timeout: "5s"

config:
  max_output_size: 1048576
  max_execution_time: "30s"
  enable_cache: false
  cache_ttl: "0s"

always: false

---

# 文本处理器技能

强大的文本处理和转换工具集。

## 功能

- **大小写转换**: 大写、小写、首字母大写
- **空格处理**: 删除空格、格式化
- **行操作**: 删除、提取、排序、反转
- **文本替换**: 模式匹配替换
- **统计功能**: 行数、字数、字符数
- **编码转换**: Base64、URL 编码
- **去重**: 行去重、单词去重

## 使用示例

### 大小写转换

```bash
# 转换为大写
agentframework enhanced-skill execute com.agentframework.skill.text_processor to_upper --vars "Text=hello world"

# 转换为小写
agentframework enhanced-skill execute com.agentframework.skill.text_processor to_lower --vars "Text=HELLO WORLD"

# 首字母大写
agentframework enhanced-skill execute com.agentframework.skill.text_processor to_title --vars "Text=hello world"
```

### 空格处理

```bash
# 删除前后空格
agentframework enhanced-skill execute com.agentframework.skill.text_processor trim_spaces --vars "Text=  hello world  "
```

### 行操作

```bash
# 删除匹配的行
agentframework enhanced-skill execute com.agentframework.skill.text_processor remove_lines --vars "Text=hello\\nworld\\ntest,Pattern=test"

# 提取匹配的行
agentframework enhanced-skill execute com.agentframework.skill.text_processor extract_lines --vars "Text=hello\\nworld\\ntest,Pattern=hello"

# 反转行顺序
agentframework enhanced-skill execute com.agentframework.skill.text_processor reverse_lines --vars "Text=line1\\nline2\\nline3"

# 排序行
agentframework enhanced-skill execute com.agentframework.skill.text_processor sort_lines --vars "Text=zebra\\napple\\nmango"

# 去重行
agentframework enhanced-skill execute com.agentframework.skill.text_processor unique_lines --vars "Text=line1\\nline2\\nline1"
```

### 文本替换

```bash
# 替换文本
agentframework enhanced-skill execute com.agentframework.skill.text_processor replace_text --vars "Text=hello world,Old=world,New=there"
```

### 统计功能

```bash
# 统计行数
agentframework enhanced-skill execute com.agentframework.skill.text_processor count_lines --vars "Text=line1\\nline2\\nline3"

# 统计字数
agentframework enhanced-skill execute com.agentframework.skill.text_processor count_words --vars "Text=hello world test"

# 统计字符数
agentframework enhanced-skill execute com.agentframework.skill.text_processor count_chars --vars "Text=hello"
```

### 编码转换

```bash
# Base64 编码
agentframework enhanced-skill execute com.agentframework.skill.text_processor base64_encode --vars "Text=hello world"

# Base64 解码
agentframework enhanced-skill execute com.agentframework.skill.text_processor base64_decode --vars "Text=aGVsbG8gd29ybGQ="

# URL 编码
agentframework enhanced-skill execute com.agentframework.skill.text_processor url_encode --vars "Text=hello world"

# URL 解码
agentframework enhanced-skill execute com.agentframework.skill.text_processor url_decode --vars "Text=hello%20world"
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| Text | 输入文本（使用\\n表示换行） | `hello world` |
| Pattern | 匹配模式 | `test`, `^hello` |
| Old | 要替换的文本 | `old` |
| New | 替换后的文本 | `new` |

## 实用场景

### 1. 格式化输出

```bash
# 统一大小写
agentframework enhanced-skill execute com.agentframework.skill.text_processor to_upper --vars "Text=mixed CASE Text"

# 去除空格
agentframework enhanced-skill execute com.agentframework.skill.text_processor trim_spaces --vars "Text=  messy  text  "
```

### 2. 日志处理

```bash
# 提取错误日志
agentframework enhanced-skill execute com.agentframework.skill.text_processor extract_lines --vars "Text=log1\\nERROR: something\\nlog3,Pattern=ERROR"

# 删除调试日志
agentframework enhanced-skill execute com.agentframework.skill.text_processor remove_lines --vars "Text=log1\\nDEBUG: info\\nlog3,Pattern=DEBUG"
```

### 3. 数据清洗

```bash
# 去重
agentframework enhanced-skill execute com.agentframework.skill.text_processor unique_lines --vars "Text=a\\nb\\na\\nc"

# 排序
agentframework enhanced-skill execute com.agentframework.skill.text_processor sort_lines --vars "Text=3\\n1\\n2"
```

### 4. 编码转换

```bash
# 生成 Base64
agentframework enhanced-skill execute com.agentframework.skill.text_processor base64_encode --vars "Text=username:password"

# URL 编码参数
agentframework enhanced-skill execute com.agentframework.skill.text_processor url_encode --vars "Text=hello world"
```

## 高级用法

### 1. 链式操作

```bash
# 先去重，再排序，再反转
TEXT="line3\\nline1\\nline2\\nline1"

# 去重
UNIQUE=$(agentframework enhanced-skill execute com.agentframework.skill.text_processor unique_lines --vars "Text=$TEXT")

# 排序
SORTED=$(agentframework enhanced-skill execute com.agentframework.skill.text_processor sort_lines --vars "Text=$UNIQUE")

# 反转
agentframework enhanced-skill execute com.agentframework.skill.text_processor reverse_lines --vars "Text=$SORTED"
```

### 2. 批量处理

```bash
# 处理文件内容
CONTENT=$(cat file.txt)

# 统计行数
agentframework enhanced-skill execute com.agentframework.skill.text_processor count_lines --vars "Text=$CONTENT"
```

### 3. 文本分析

```bash
# 分析文本
TEXT="Your text here"

# 字数统计
WORDS=$(agentframework enhanced-skill execute com.agentframework.skill.text_processor count_words --vars "Text=$TEXT")

# 行数统计
LINES=$(agentframework enhanced-skill execute com.agentframework.skill.text_processor count_lines --vars "Text=$TEXT")

# 字符统计
CHARS=$(agentframework enhanced-skill execute com.agentframework.skill.text_processor count_chars --vars "Text=$TEXT")

echo "Words: $WORDS, Lines: $LINES, Chars: $CHARS"
```

## 性能建议

### 1. 大文本处理

对于大文本，建议使用文件操作：

```bash
# 而不是通过 Text 参数
agentframework enhanced-skill execute com.agentframework.skill.file_manager search_content --vars "Path=.,Pattern=error,Include=*.log"
```

### 2. 批量操作

```bash
# 使用管道组合命令
cat input.txt | sort | uniq > output.txt
```

### 3. 内存考虑

- 大文本会占用大量内存
- 考虑使用流式处理
- 分批处理大文件

## 故障排除

### 特殊字符

```bash
# 使用引号包裹
agentframework enhanced-skill execute com.agentframework.skill.text_processor to_upper --vars "Text='hello world'"

# 转义特殊字符
agentframework enhanced-skill execute com.agentframework.skill.text_processor replace_text --vars "Text='hello\\nworld',Old=\\n,New=space"
```

### 模式匹配

```bash
# 使用正确的正则表达式
agentframework enhanced-skill execute com.agentframework.skill.text_processor extract_lines --vars "Text=hello123\\nworld456,Pattern=[0-9]+"
```

### 编码问题

```bash
# 确保正确的编码
export LANG=en_US.UTF-8
```

## 扩展功能

### 1. 自定义脚本

```bash
# 创建自定义文本处理脚本
cat > custom_processor.sh << 'EOF'
#!/bin/bash
# 自定义处理逻辑
echo "$1" | custom_command
EOF

chmod +x custom_processor.sh
```

### 2. 集成其他工具

```bash
# 与 jq 结合处理 JSON
echo '{"key":"value"}' | jq '.key'

# 与 awk 结合处理表格
ps aux | awk '{print $3, $11}'
```

### 3. 多语言支持

```bash
# 设置字符集
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
```

## 相关工具

- `sed`: 流编辑器
- `awk`: 文本处理工具
- `tr`: 字符转换工具
- `jq`: JSON 处理器
- `iconv`: 编码转换
