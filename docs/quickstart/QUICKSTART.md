# 快速上手

> 5 分钟从零跑起来

---

## 前置要求

| 工具 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 必须 |
| Git | 任意 | 必须 |
| Ollama 或 LM Studio | 任意 | 本地模型（可选，也可用 API Key） |

---

## 第一步：克隆并构建

```bash
git clone <仓库地址>
cd AgentFramework

# 下载依赖
go mod download

# 构建 CLI 工具
go build -o build/afcli.exe ./cmd/afcli/   # Windows
go build -o build/afcli      ./cmd/afcli/   # Linux/macOS
```

---

## 第二步：配置模型

编辑项目根目录的 `host.yaml`：

```yaml
# 使用 LM Studio（推荐本地测试）
models:
  default:
    provider: "lmstudio"
    baseurl: "http://localhost:1234/v1"
    model: "local-model"

# 或者使用 Ollama
models:
  default:
    provider: "ollama"
    baseurl: "http://localhost:11434/api"
    model: "llama3"

# 或者使用云端 API（以 DeepSeek 为例）
models:
  default:
    provider: "deepseek"
    baseurl: "https://api.deepseek.com/v1"
    model: "deepseek-chat"
    apikey: "${DEEPSEEK_API_KEY}"   # 环境变量注入
```

---

## 第三步：运行

### 方式 A：命令行对话

```bash
# 单次问答
./build/afcli agent chat "你好，介绍一下自己"

# 进入多轮对话（Ctrl+C 退出）
./build/afcli agent chat
```

### 方式 B：启动 Gateway 服务

```bash
# 默认端口 18640
./build/afcli gateway

# 指定端口
./build/afcli gateway --port 8080

# 详细日志
./build/afcli gateway --verbose

# 强制释放端口（杀掉占用进程后启动）
./build/afcli gateway --force
```

Gateway 启动后，WebSocket 连接到 `ws://localhost:18640/`，HTTP API 在 `http://localhost:18640/v1/`。

### 方式 C：TUI 终端界面

```bash
# 构建 TUI
go build -o build/aftui.exe ./cmd/tui/

# 启动
./build/aftui.exe
```

---

## 第四步：接入渠道（可选）

### 飞书（推荐，WebSocket 模式无需公网）

```bash
# 安装飞书插件
./build/afcli lark install

# 配置 App ID 和 Secret
./build/afcli lark config --app-id cli_xxx --app-secret xxx
```

然后在 `host.yaml` 里启用：

```yaml
channels:
  lark:
    enabled: true
    connectionMode: "websocket"   # 无需公网 URL
    appId: "${LARK_APP_ID}"
    appSecret: "${LARK_APP_SECRET}"
```

### 企业微信

```bash
./build/afcli wechat install
./build/afcli wechat config --corp-id xxx --agent-id xxx --corp-secret xxx
```

---

## 第五步：添加 Skill（可选）

在项目目录创建 `.skills/hello/SKILL.md`：

```markdown
---
name: hello
description: 向用户打招呼
parameters:
  name:
    type: string
    description: 用户名字
    required: true
---

# 使用说明

调用此工具时，用热情的方式向用户打招呼，并夸奖他们的名字好听。
```

重启服务后，Agent 会自动发现并使用这个 Skill。

---

## 常见问题

**Q：模型连接失败？**
- 确认本地模型服务已启动（Ollama: `ollama serve`，LM Studio: 打开应用）
- 检查 `baseurl` 是否正确
- 检查端口是否被防火墙拦截

**Q：Gateway 端口被占用？**
- 加 `--force` 参数：`./build/afcli gateway --force`

**Q：飞书消息收不到？**
- 用 `connectionMode: "websocket"` 模式，不需要配置 Webhook URL
- 确认 AppID 和 AppSecret 正确

---

## 下一步

- [架构概览](../architecture/ARCHITECTURE_OVERVIEW.md) — 理解框架设计
- [配置指南](../configuration/CONFIGURATION.md) — 完整配置参考
- [Skill 开发](../SKILL_DEVELOPMENT.md) — 开发自定义 Skill
- [渠道集成](../CHANNEL_INTEGRATION.md) — 接入更多渠道
