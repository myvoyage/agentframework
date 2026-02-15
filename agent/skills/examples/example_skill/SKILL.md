---
name: ExampleHelloWorld
description: 一个简单的示例技能，用于测试技能导入功能
version: 1.0.0
category: custom
author: AgentFramework Team
license: MIT
tags:
  - example
  - demo
  - test
---

# ExampleHelloWorld 技能

## 概述

这是一个示例技能，用于演示和测试 AgentFramework 的技能导入功能。该技能实现了一个简单的"Hello World"功能。

## 使用方法

### 步骤1：准备输入

在输入框中输入您的名字或任意文本。

### 步骤2：执行技能

选择此技能并执行，系统将返回格式化的问候语。

### 步骤3：查看结果

技能将返回一个包含问候语的响应消息。

## 功能特性

- ✅ 简单的文本处理
- ✅ 格式化输出
- ✅ 错误处理
- ✅ 使用统计

## 输入格式

```
输入：任意文本字符串
```

## 输出格式

```
输出：格式化的问候消息
示例：Hello, {输入}! 欢迎使用 ExampleHelloWorld 技能。
```

## 工作流程

1. **验证输入**: 检查输入是否为空
2. **处理数据**: 格式化输入文本
3. **生成响应**: 创建友好的问候消息
4. **返回结果**: 返回格式化的输出

## 配置选项

本技能使用默认配置：
- 缓存启用：是
- 缓存TTL：60秒
- 超时时间：30秒
- 最大重试次数：3次

## 示例

### 示例1：基本使用
```
输入：World
输出：Hello, World! 欢迎使用 ExampleHelloWorld 技能。
```

### 示例2：带空格的输入
```
输入：Agent Framework
输出：Hello, Agent Framework! 欢迎使用 ExampleHelloWorld 技能。
```

## 错误处理

如果输入为空，技能将返回错误消息：
```
错误：输入不能为空
```

## 版本历史

- **v1.0.0** (2025-01-28): 初始版本
  - 实现基本的问候功能
  - 添加输入验证
  - 添加错误处理

## 作者

AgentFramework Team

## 许可证

MIT License

## 参考资源

- [AgentFramework 文档](https://github.com/your-org/AgentFramework)
- [技能系统指南](https://github.com/your-org/AgentFramework/docs)
