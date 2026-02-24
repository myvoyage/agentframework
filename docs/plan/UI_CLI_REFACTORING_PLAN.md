# AgentFramework UI + CLI 优化重构计划

## 🔍 当前 AgentFramework 项目分析

### 现有优势

1. **架构完善**：已具备 core → agent → skills 的分层结构
2. **功能丰富**：工作流引擎、多渠道通信、IoT 支持
3. **CLI 基础**：基于 Cobra 的命令行框架已搭建
4. **服务化**：Application 层抽象良好，支持多界面

### 待改进点

1. **CLI 体验**：交互性不足，缺乏 TUI 界面
2. **UI 缺失**：无桌面/Web 界面，仅靠命令行
3. **本地优先**：数据存储和配置分散，未实现完全本地化
4. **沙箱机制**：虽有 sandbox_manager，但未深度集成到执行流程
5. **技能发现**：技能注册和使用流程不够直观

---

## 💡 UI + CLI 优化重构建议

### 一、CLI 层优化

#### 1.1 引入 TUI (Terminal User Interface)

建议引入 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 框架，这是 Go 语言最流行的 TUI 框架：

```go
// 建议新增 cmd/tui/main.go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Model 定义 TUI 状态
type Model struct {
    agentType    string
    messages     []Message
    input        textinput.Model
    viewport     viewport.Model
    spinner      spinner.Model
    // ...
}

// 提供交互式 Agent 选择、实时对话、工作流可视化
```

**预期效果**：
- 交互式 Agent 选择界面
- 实时对话流显示（类似 ChatGPT CLI）
- 工作流执行进度可视化
- 文件浏览器集成

#### 1.2 CLI 命令结构优化

当前命令结构较扁平，建议重构为：

```bash
# 当前结构
agentframework workflow list
agentframework agent chat "message"

# 建议结构
af agent list                    # 列出可用 agents
af agent chat <name>             # 启动对话
af agent run <name> --task "xxx" # 执行任务

af workflow list                 # 工作流管理
af workflow create --from-file workflow.yaml
af workflow run <id> --watch     # 实时查看执行

af skill list                    # 技能管理
af skill install <name>
af skill run <name> --input "{}"

af config init                   # 交互式配置初始化
af config set model.default gpt-4

af tui                           # 启动交互式 TUI
```

#### 1.3 配置文件管理优化

本地优先理念：

```go
// config/local_config.go
type LocalConfig struct {
    // 统一配置存储
    Settings    AppSettings    `json:"settings"`
    Agents      []AgentConfig  `json:"agents"`
    Workflows   []WorkflowDef  `json:"workflows"`
    Skills      []SkillConfig  `json:"skills"`
    Channels    []ChannelConfig `json:"channels"`
}

// 使用 SQLite 存储配置，支持版本管理
func (c *LocalConfig) Save() error
func (c *LocalConfig) Load() error
func (c *LocalConfig) Migrate() error
```

---

### 二、UI 层设计

#### 2.1 桌面应用架构选择

**方案 A：Electron + React **

优点：
- 成熟的跨平台方案
- 丰富的 UI 组件生态
- 与现有 Go 后端通过 HTTP/gRPC 通信

架构：
```
┌─────────────────────────────────────┐
│         Electron Frontend           │
│  (React + Redux + TailwindCSS)      │
├─────────────────────────────────────┤
│      Go HTTP/gRPC Server            │
│  (复用现有 core.Application)        │
├─────────────────────────────────────┤
│      AgentFramework Core            │
└─────────────────────────────────────┘
```

**方案 B：Wails（纯 Go + Web 前端）**

优点：
- 更小的包体积
- 纯 Go 开发体验
- 原生性能

#### 2.2 UI 功能模块设计

```typescript
// 建议的页面结构
/src
  /pages
    /Dashboard          # 主控制台
    /AgentStudio        # Agent 可视化编排
    /WorkflowBuilder    # 工作流构建器
    /SkillMarket        # 技能市场
    /Settings           # 设置中心
    /Logs               # 日志与监控
  /components
    /Chat               # 对话组件
    /WorkflowGraph      # 工作流图可视化
    /SkillCard          # 技能卡片
```

#### 2.3 核心界面功能

1. **Agent Studio**：
   - 可视化 Agent 配置
   - 提示词编辑器
   - 工具链组装界面

2. **Workflow Builder**：
   - 拖拽式工作流编排
   - 实时执行监控
   - DAG 可视化

3. **Chat Interface**：
   - 多 Agent 对话切换
   - 上下文管理
   - 代码高亮与复制

---

### 三、架构重构建议

#### 3.1 目录结构优化

```
AgentFramework/
├── cmd/
│   ├── agent-cli/          # 现有 CLI（精简版）
│   ├── agent-tui/          # 新增 TUI 版本
│   └── agent-desktop/      # 桌面应用入口
├── internal/               # 私有代码
│   ├── core/               # 核心业务逻辑（从根目录移入）
│   ├── cli/                # CLI 共享代码
│   ├── tui/                # TUI 组件
│   └── server/             # HTTP/gRPC 服务
├── pkg/                    # 可导出包
│   ├── api/                # API 定义
│   ├── config/             # 配置管理
│   └── ui/                 # UI 共享组件
├── web/                    # 前端代码（Electron/React）
│   ├── src/
│   └── package.json
└── desktop/                # 桌面应用包装
    ├── main.go             # Electron 主进程
    └── embed.go            # 嵌入前端资源
```

#### 3.2 核心接口抽象

```go
// internal/interfaces/ui.go
package interfaces

// UIManager 定义 UI 层通用接口
type UIManager interface {
    // 消息展示
    ShowMessage(msg Message) error
    ShowProgress(id string, percent int) error
    
    // 交互
    PromptInput(prompt string) (string, error)
    SelectOption(options []string) (int, error)
    Confirm(message string) (bool, error)
    
    // Agent 管理
    RenderAgentList(agents []AgentInfo) error
    RenderChat(messages []Message) error
    
    // 工作流
    RenderWorkflowGraph(graph WorkflowGraph) error
    UpdateWorkflowStatus(id string, status Status) error
}

// 实现：CLI 版本
type CLIUI struct { /* ... */ }

// 实现：TUI 版本  
type TUI struct { /* ... */ }

// 实现：Web/Desktop 版本
type WebUI struct { /* ... */ }
```

#### 3.3 本地优先架构

借鉴 LobsterAI 的设计理念：

```go
// internal/local/store.go
package local

// LocalStore 本地数据存储接口
type LocalStore interface {
    // Agent 配置
    SaveAgentConfig(cfg AgentConfig) error
    GetAgentConfig(name string) (AgentConfig, error)
    ListAgentConfigs() ([]AgentConfig, error)
    
    // 对话历史
    SaveConversation(conv Conversation) error
    GetConversations(agentID string) ([]Conversation, error)
    
    // 工作流状态
    SaveWorkflowState(state WorkflowState) error
    GetWorkflowState(id string) (WorkflowState, error)
    
    // 技能缓存
    CacheSkill(skill Skill) error
    GetCachedSkill(name string) (Skill, error)
}

// SQLite 实现
type SQLiteStore struct {
    db *sql.DB
}
```

---

### 四、具体实施路线图

#### Phase 1：CLI 增强（1-2 周）
- [ ] 引入 Bubble Tea 实现 TUI 版本
- [ ] 重构命令结构，增强交互性
- [ ] 添加配置初始化向导

#### Phase 2：本地优先（2-3 周）
- [ ] 实现 SQLite 配置存储
- [ ] 统一配置管理接口
- [ ] 本地对话历史管理

#### Phase 3：桌面 UI（3-4 周）
- [ ] 搭建 Electron + React 框架
- [ ] 实现 Dashboard 和 Chat 界面
- [ ] 与 Go 后端 API 对接

#### Phase 4：高级功能（4-6 周）
- [ ] Workflow Builder 可视化
- [ ] Agent Studio 编排界面
- [ ] 技能市场浏览器

---

### 五、关键技术选型

| 组件 | 推荐方案 | 理由 |
|------|----------|------|
| TUI 框架 | Bubble Tea | Go 生态最成熟，组件丰富 |
| 桌面框架 | Electron + React | 跨平台，生态完善 |
| 本地存储 | SQLite | 轻量，零配置，支持复杂查询 |
| 状态管理 | Redux Toolkit |  predictable，调试友好 |
| UI 组件 | TailwindCSS + Headless UI | 原子化 CSS，可定制性强 |
| 通信协议 | gRPC + HTTP/REST | 高效 + 兼容性 |

---

### 六、代码示例

#### TUI 主界面框架

```go
// cmd/tui/main.go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbles/textinput"
)

type mainModel struct {
    width, height int
    currentView   View
    sidebar       list.Model
    viewport      viewport.Model
    input         textinput.Model
    agent         *core.AgentRunnerService
}

func (m mainModel) Init() tea.Cmd {
    return tea.Batch(
        textinput.Blink,
        m.loadAgents(),
    )
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.updateLayout()
        
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "tab":
            m.switchView()
        }
        
    case AgentListMsg:
        m.sidebar.SetItems(msg.Items)
    }
    
    // 更新子组件...
    return m, cmd
}

func (m mainModel) View() string {
    // 布局渲染...
    return lipgloss.JoinHorizontal(
        lipgloss.Left,
        m.sidebarView(),
        m.contentView(),
    )
}
```

#### 桌面应用入口

```go
// desktop/main.go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:web/dist
var assets embed.FS

func main() {
    app := NewApp()
    
    err := wails.Run(&options.App{
        Title:  "AgentFramework Desktop",
        Width:  1280,
        Height: 800,
        Assets: assets,
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}

type App struct {
    ctx context.Context
    core *core.Application
}

func (a *App) Chat(message string) (string, error) {
    // 调用 core 层
    return a.core.AgentChat(a.ctx, message)
}

func (a *App) GetWorkflows() ([]WorkflowInfo, error) {
    // 返回工作流列表
    return a.core.ListWorkflows(a.ctx)
}
```

---

## 📌 总结

1. **TUI 交互体验**：命令行也可以有丰富的交互体验
2. **代码上下文管理**：Agent 与代码的深度集成
3. **增量生成与补丁应用**：高效的内容更新机制
4. **本地优先架构**：数据安全与隐私保护
5. **三层模块化设计**：清晰的职责分离
6. **Electron + React 技术栈**：成熟的桌面应用方案
7. **技能系统可视化**：降低用户使用门槛

### 实施建议优先级

1. **高优先级**：引入 TUI 框架（Bubble Tea）提升 CLI 体验
2. **中优先级**：实现 SQLite 本地配置存储，统一配置管理
3. **长期规划**：Electron + React 桌面应用，提供可视化编排能力

---

**文档版本**: 1.0  
**创建日期**: 2026-02-22  
**作者**: AgentFramework Team
