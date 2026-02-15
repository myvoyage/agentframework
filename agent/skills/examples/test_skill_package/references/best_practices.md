# 文本处理最佳实践

## 概述

本文档介绍了文本处理的最佳实践和常见模式。

## 性能优化

### 1. 批量处理

对于大文本，建议分块处理：

```python
CHUNK_SIZE = 4096

def process_large_text(text):
    chunks = [text[i:i+CHUNK_SIZE] for i in range(0, len(text), CHUNK_SIZE)]
    results = []
    for chunk in chunks:
        result = process_chunk(chunk)
        results.append(result)
    return merge_results(results)
```

### 2. 内存管理

及时释放不再需要的大文本对象：

```python
# 处理完成后清除大文本
del large_text
import gc
gc.collect()
```

### 3. 编码处理

始终使用 UTF-8 编码：

```python
with open(file, 'r', encoding='utf-8') as f:
    text = f.read()
```

## 错误处理

### 1. 输入验证

始终验证输入数据：

```python
def validate_input(text):
    if not isinstance(text, str):
        raise TypeError("Input must be a string")
    if len(text) == 0:
        raise ValueError("Input cannot be empty")
    if len(text) > MAX_SIZE:
        raise ValueError("Input too large")
    return True
```

### 2. 异常捕获

使用具体的异常类型：

```python
try:
    result = process_text(text)
except UnicodeDecodeError as e:
    handle_encoding_error(e)
except MemoryError:
    handle_memory_error()
except Exception as e:
    log_error(e)
    raise
```

## 安全考虑

### 1. 输入清理

移除潜在的恶意内容：

```python
import re

def sanitize_input(text):
    # 移除控制字符
    text = re.sub(r'[\x00-\x08\x0b-\x0c\x0e-\x1f\x7f-\x9f]', '', text)
    # 限制长度
    text = text[:MAX_LENGTH]
    return text
```

### 2. 路径遍历防护

处理文件路径时要小心：

```python
import os

def safe_path(base_dir, filename):
    full_path = os.path.join(base_dir, filename)
    full_path = os.path.abspath(full_path)
    if not full_path.startswith(os.path.abspath(base_dir)):
        raise ValueError("Invalid path")
    return full_path
```

## 测试建议

### 1. 单元测试

为每个功能编写单元测试：

```python
def test_word_count():
    text = "Hello World"
    assert count_words(text) == 2

def test_empty_input():
    with pytest.raises(ValueError):
        process_text("")
```

### 2. 边界测试

测试边界情况：

```python
def test_very_long_text():
    text = "a" * 1000000
    assert process_text(text) is not None

def test_unicode():
    text = "你好世界 🌍"
    assert process_text(text) is not None
```

## 文档规范

### 1. 代码注释

为复杂逻辑添加注释：

```python
# 使用滑动窗口算法进行模式匹配
# 时间复杂度: O(n)
# 空间复杂度: O(1)
def sliding_window_match(text, pattern):
    # ...
```

### 2. API文档

使用清晰的文档字符串：

```python
def process_text(text: str, options: dict) -> dict:
    """
    处理文本并返回统计信息。

    参数:
        text: 要处理的文本
        options: 处理选项字典

    返回:
        包含统计信息的字典

    异常:
        ValueError: 如果输入无效
        TypeError: 如果类型错误
    """
    pass
```

## 更多资源

- [Python文本处理库](https://docs.python.org/3/library/text.html)
- [正则表达式指南](https://docs.python.org/3/library/re.html)
- [Unicode标准](https://unicode.org/)
