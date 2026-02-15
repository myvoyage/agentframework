# TextProcessor API 参考

## 概述

本文档提供了 TextProcessor 技能的完整 API 参考。

## 主要操作

### 1. Clean

清理文本，移除多余的空格和特殊字符。

**参数**:
```json
{
  "operation": "clean",
  "options": {
    "removeSpaces": true,        // 移除多余空格
    "removeSpecialChars": false, // 保留特殊字符
    "normalizeNewlines": true    // 标准化换行符
  }
}
```

**示例**:
```python
# Python
result = textprocessor.clean("  Hello    World!  \n\n")
# 返回: "Hello World!\n"
```

### 2. Analyze

分析文本，返回统计信息。

**参数**:
```json
{
  "operation": "analyze",
  "options": {
    "countWords": true,
    "countChars": true,
    "countLines": true,
    "extractKeywords": true,
    "sentiment": false
  }
}
```

**返回**:
```json
{
  "success": true,
  "result": {
    "words": 42,
    "characters": 256,
    "lines": 5,
    "paragraphs": 2,
    "keywords": ["hello", "world", "test"],
    "readingTime": 12.5
  }
}
```

### 3. Convert

转换文本格式。

**参数**:
```json
{
  "operation": "convert",
  "options": {
    "case": "lower|upper|title|sentence",
    "encoding": "utf-8|gbk|ascii",
    "normalize": "lf|crlf"
  }
}
```

**示例**:
```python
# 转换为大写
result = textprocessor.convert("hello world", case="upper")
# 返回: "HELLO WORLD"
```

## 辅助函数

### extract_keywords(text, limit=10)

从文本中提取关键词。

**参数**:
- `text` (str): 输入文本
- `limit` (int): 返回关键词数量限制

**返回**:
- `List[str]`: 关键词列表

**示例**:
```python
keywords = textprocessor.extract_keywords(
    "Machine learning is a subset of artificial intelligence.",
    limit=5
)
# 返回: ["machine", "learning", "subset", "artificial", "intelligence"]
```

### count_syllables(word)

计算单词的音节数。

**参数**:
- `word` (str): 输入单词

**返回**:
- `int`: 音节数

**示例**:
```python
count = textprocessor.count_syllables("hello")
# 返回: 2
```

### sentiment_score(text)

分析文本的情感倾向。

**参数**:
- `text` (str): 输入文本

**返回**:
- `float`: 情感分数 (-1.0 到 1.0)
  - -1.0: 非常负面
  - 0.0: 中性
  - 1.0: 非常正面

**示例**:
```python
score = textprocessor.sentiment_score("I love this!")
# 返回: 0.8
```

## 配置选项

### 全局配置

```python
config = {
    "cache": {
        "enabled": True,
        "ttl": 120
    },
    "performance": {
        "max_text_size": 10485760,  # 10MB
        "chunk_size": 4096,
        "parallel_processing": True
    },
    "security": {
        "sanitize_input": True,
        "max_retries": 3
    }
}
```

### 操作特定配置

```python
operation_config = {
    "clean": {
        "preserve_numbers": True,
        "preserve_emails": True
    },
    "analyze": {
        "min_word_length": 3,
        "stop_words": ["a", "an", "the"]
    },
    "convert": {
        "fallback_encoding": "utf-8",
        "error_handling": "ignore"
    }
}
```

## 错误处理

### 错误代码

| 代码 | 描述 | HTTP状态码 |
|-----|------|-----------|
| EMPTY_INPUT | 输入为空 | 400 |
| TEXT_TOO_LARGE | 文本超过最大限制 | 413 |
| INVALID_OPERATION | 无效的操作类型 | 400 |
| ENCODING_ERROR | 编码错误 | 422 |
| PROCESSING_ERROR | 处理错误 | 500 |

### 错误响应格式

```json
{
  "success": false,
  "error": {
    "code": "TEXT_TOO_LARGE",
    "message": "Input text exceeds maximum size of 10MB",
    "details": {
      "size": 15728640,
      "max_size": 10485760
    }
  }
}
```

## 批处理

### batch_process(items)

批量处理多个文本。

**参数**:
- `items` (List[dict]): 文本项目列表

**返回**:
- `List[dict]`: 处理结果列表

**示例**:
```python
items = [
    {"text": "Hello", "operation": "clean"},
    {"text": "World", "operation": "analyze"}
]
results = textprocessor.batch_process(items)
```

## 事件

### on_progress

处理进度事件。

```python
def on_progress(progress):
    print(f"Progress: {progress}%")

textprocessor.process("large text", on_progress=on_progress)
```

### on_complete

处理完成事件。

```python
def on_complete(result):
    print(f"Processing complete: {result}")

textprocessor.process("text", on_complete=on_complete)
```

## 版本信息

当前版本: **v1.2.0**

兼容性:
- Python 3.8+
- Node.js 14+
- Go 1.18+
