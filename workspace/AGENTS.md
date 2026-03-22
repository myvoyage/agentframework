---
agents:
  - id: default
    name: 默认助手
    description: 默认的AI助手，处理日常对话和通用任务
    skills:
      - file
      - search
      - code
    model: gpt-4
    enabled: true

  - id: coder
    name: 代码助手
    description: 专注于代码开发、调试和重构
    skills:
      - code
      - git
      - terminal
    model: gpt-4
    enabled: true

  - id: analyst
    name: 数据分析师
    description: 专注于数据处理、分析和可视化
    skills:
      - data
      - excel
      - visualization
    model: gpt-4
    enabled: true

  - id: researcher
    name: 研究助手
    description: 专注于信息检索、文献整理和研究支持
    skills:
      - search
      - summarize
      - translate
    model: gpt-4
    enabled: true
