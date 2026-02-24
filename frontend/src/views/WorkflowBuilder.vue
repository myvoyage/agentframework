<template>
  <div class="workflow-builder">
    <!-- 顶部工具栏 -->
    <div class="builder-toolbar">
      <div class="toolbar-left">
        <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
        <el-divider direction="vertical" />
        <el-input
          v-model="workflowName"
          placeholder="工作流名称"
          style="width: 200px"
        />
        <el-input
          v-model="workflowDescription"
          placeholder="描述"
          style="width: 250px"
        />
      </div>
      <div class="toolbar-right">
        <el-button-group>
          <el-tooltip content="导入工作流">
            <el-button :icon="Upload" @click="importWorkflow" />
          </el-tooltip>
          <el-tooltip content="导出工作流">
            <el-button :icon="Download" @click="exportWorkflow" />
          </el-tooltip>
          <el-tooltip content="保存">
            <el-button type="primary" :icon="Check" @click="saveWorkflow">
              保存
            </el-button>
          </el-tooltip>
          <el-tooltip content="运行">
            <el-button type="success" :icon="VideoPlay" @click="runWorkflow">
              运行
            </el-button>
          </el-tooltip>
        </el-button-group>
      </div>
    </div>

    <div class="builder-content">
      <!-- 左侧：节点面板 -->
      <div class="nodes-panel">
        <div class="panel-header">
          <el-icon><Grid /></el-icon>
          <span>节点库</span>
        </div>
        <div class="panel-search">
          <el-input
            v-model="nodeSearchText"
            placeholder="搜索节点..."
            :prefix-icon="Search"
            size="small"
            clearable
          />
        </div>
        <div class="node-categories">
          <el-collapse v-model="activeCategories" accordion>
            <el-collapse-item title="触发器" name="triggers">
              <div class="node-list">
                <div
                  v-for="node in filteredNodes('trigger')"
                  :key="node.type"
                  class="node-item"
                  draggable="true"
                  @dragstart="onNodeDragStart($event, node)"
                >
                  <div class="node-icon" :style="{ background: node.color }">
                    <el-icon>
                      <component :is="node.icon" />
                    </el-icon>
                  </div>
                  <span class="node-label">{{ node.label }}</span>
                </div>
              </div>
            </el-collapse-item>

            <el-collapse-item title="操作" name="actions">
              <div class="node-list">
                <div
                  v-for="node in filteredNodes('action')"
                  :key="node.type"
                  class="node-item"
                  draggable="true"
                  @dragstart="onNodeDragStart($event, node)"
                >
                  <div class="node-icon" :style="{ background: node.color }">
                    <el-icon>
                      <component :is="node.icon" />
                    </el-icon>
                  </div>
                  <span class="node-label">{{ node.label }}</span>
                </div>
              </div>
            </el-collapse-item>

            <el-collapse-item title="控制" name="control">
              <div class="node-list">
                <div
                  v-for="node in filteredNodes('control')"
                  :key="node.type"
                  class="node-item"
                  draggable="true"
                  @dragstart="onNodeDragStart($event, node)"
                >
                  <div class="node-icon" :style="{ background: node.color }">
                    <el-icon>
                      <component :is="node.icon" />
                    </el-icon>
                  </div>
                  <span class="node-label">{{ node.label }}</span>
                </div>
              </div>
            </el-collapse-item>

            <el-collapse-item title="集成" name="integration">
              <div class="node-list">
                <div
                  v-for="node in filteredNodes('integration')"
                  :key="node.type"
                  class="node-item"
                  draggable="true"
                  @dragstart="onNodeDragStart($event, node)"
                >
                  <div class="node-icon" :style="{ background: node.color }">
                    <el-icon>
                      <component :is="node.icon" />
                    </el-icon>
                  </div>
                  <span class="node-label">{{ node.label }}</span>
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>

      <!-- 中间：画布区域 -->
      <div class="canvas-container">
        <div class="canvas-toolbar">
          <el-button-group size="small">
            <el-button :icon="ZoomOut" @click="zoomOut" />
            <el-button disabled>{{ Math.round(zoom * 100) }}%</el-button>
            <el-button :icon="ZoomIn" @click="zoomIn" />
          </el-button-group>
          <el-divider direction="vertical" />
          <el-button-group size="small">
            <el-button :icon="RefreshLeft" @click="undo" :disabled="!canUndo" />
            <el-button :icon="RefreshRight" @click="redo" :disabled="!canRedo" />
          </el-button-group>
          <el-divider direction="vertical" />
          <el-button-group size="small">
            <el-button :icon="Grid" @click="fitToScreen">适应屏幕</el-button>
            <el-button :icon="Compass" @click="autoLayout">自动布局</el-button>
          </el-button-group>
          <div class="canvas-info">
            <span>节点: {{ workflowNodes.length }}</span>
            <span>连接: {{ workflowConnections.length }}</span>
          </div>
        </div>

        <div
          ref="canvasRef"
          class="workflow-canvas"
          :style="{ transform: `scale(${zoom})` }"
          @drop="onCanvasDrop"
          @dragover="onCanvasDragOver"
        >
          <!-- SVG 连接线 -->
          <svg class="connections-layer" :width="canvasSize.width" :height="canvasSize.height">
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="7"
                refX="9"
                refY="3.5"
                orient="auto"
              >
                <polygon points="0 0, 10 3.5, 0 7" fill="#909399" />
              </marker>
            </defs>
            <path
              v-for="conn in workflowConnections"
              :key="conn.id"
              :d="getConnectionPath(conn)"
              stroke="#909399"
              stroke-width="2"
              fill="none"
              marker-end="url(#arrowhead)"
              :class="{ 'connection-selected': conn.id === selectedConnectionId }"
              @click="selectConnection(conn.id)"
            />
          </svg>

          <!-- 节点 -->
          <div
            v-for="node in workflowNodes"
            :key="node.id"
            class="workflow-node"
            :class="{ 'node-selected': node.id === selectedNodeId }"
            :style="{ left: node.x + 'px', top: node.y + 'px' }"
            @mousedown="startNodeDrag($event, node)"
            @click="selectNode(node.id)"
          >
            <div class="node-header" :style="{ background: node.color }">
              <el-icon size="16">
                <component :is="node.icon" />
              </el-icon>
              <span class="node-title">{{ node.label }}</span>
              <el-button
                class="node-delete"
                :icon="Close"
                size="small"
                text
                @click.stop="deleteNode(node.id)"
              />
            </div>
            <div class="node-body">
              <div class="node-config">
                <el-input
                  v-if="node.type === 'agent'"
                  v-model="node.config.agentId"
                  placeholder="Agent ID"
                  size="small"
                />
                <el-input
                  v-if="node.type === 'skill'"
                  v-model="node.config.skillId"
                  placeholder="Skill ID"
                  size="small"
                />
                <el-input
                  v-if="node.type === 'condition'"
                  v-model="node.config.condition"
                  placeholder="条件表达式"
                  size="small"
                />
                <el-input
                  v-if="node.type === 'delay'"
                  v-model.number="node.config.duration"
                  type="number"
                  placeholder="延迟(秒)"
                  size="small"
                />
              </div>
              <!-- 端口 -->
              <div class="node-port input-port" data-port="input"></div>
              <div class="node-port output-port" data-port="output"></div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="workflowNodes.length === 0" class="canvas-empty">
            <el-empty description="拖拽节点到此处开始构建工作流">
              <template #image>
                <el-icon :size="80" color="#c0c4cc">
                  <Operation />
                </el-icon>
              </template>
            </el-empty>
          </div>
        </div>
      </div>

      <!-- 右侧：属性面板 -->
      <div class="properties-panel">
        <div class="panel-header">
          <el-icon><Setting /></el-icon>
          <span>属性</span>
        </div>
        <el-tabs v-model="activeTab" type="border-card">
          <el-tab-pane label="节点" name="node">
            <div v-if="selectedNode" class="node-properties">
              <el-form label-width="80px" size="small">
                <el-form-item label="名称">
                  <el-input v-model="selectedNode.label" />
                </el-form-item>
                <el-form-item label="类型">
                  <el-input v-model="selectedNode.type" disabled />
                </el-form-item>
                <el-form-item label="位置">
                  <el-row :gutter="8">
                    <el-col :span="12">
                      <el-input-number
                        v-model="selectedNode.x"
                        :controls="false"
                        @change="updateNodePosition"
                      />
                    </el-col>
                    <el-col :span="12">
                      <el-input-number
                        v-model="selectedNode.y"
                        :controls="false"
                        @change="updateNodePosition"
                      />
                    </el-col>
                  </el-row>
                </el-form-item>
                <!-- 特定节点配置 -->
                <template v-if="selectedNode.type === 'agent'">
                  <el-form-item label="Agent">
                    <el-select v-model="selectedNode.config.agentId" placeholder="选择 Agent">
                      <el-option
                        v-for="agent in agents"
                        :key="agent.id"
                        :label="agent.name"
                        :value="agent.id"
                      />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="提示词">
                    <el-input
                      v-model="selectedNode.config.prompt"
                      type="textarea"
                      :rows="3"
                    />
                  </el-form-item>
                </template>
                <template v-if="selectedNode.type === 'condition'">
                  <el-form-item label="条件">
                    <el-input
                      v-model="selectedNode.config.condition"
                      placeholder="input.value > 100"
                    />
                  </el-form-item>
                </template>
                <template v-if="selectedNode.type === 'loop'">
                  <el-form-item label="循环次数">
                    <el-input-number v-model="selectedNode.config.count" :min="1" />
                  </el-form-item>
                </template>
              </el-form>
            </div>
            <el-empty v-else description="选择节点查看属性" :image-size="80" />
          </el-tab-pane>

          <el-tab-pane label="工作流" name="workflow">
            <el-form label-width="80px" size="small">
              <el-form-item label="名称">
                <el-input v-model="workflowName" />
              </el-form-item>
              <el-form-item label="描述">
                <el-input
                  v-model="workflowDescription"
                  type="textarea"
                  :rows="2"
                />
              </el-form-item>
              <el-form-item label="类型">
                <el-select v-model="workflowType">
                  <el-option label="顺序执行" value="sequential" />
                  <el-option label="并行执行" value="parallel" />
                  <el-option label="DAG" value="dag" />
                </el-select>
              </el-form-item>
              <el-form-item label="超时">
                <el-input-number v-model="workflowTimeout" :min="0" />
                <span style="margin-left: 8px">秒</span>
              </el-form-item>
              <el-form-item label="重试">
                <el-input-number v-model="workflowRetries" :min="0" :max="5" />
                <span style="margin-left: 8px">次</span>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="变量" name="variables">
            <div class="variables-list">
              <el-button
                type="primary"
                size="small"
                :icon="Plus"
                @click="addVariable"
                style="width: 100%; margin-bottom: 12px"
              >
                添加变量
              </el-button>
              <div
                v-for="(variable, index) in workflowVariables"
                :key="index"
                class="variable-item"
              >
                <el-input
                  v-model="variable.name"
                  placeholder="变量名"
                  size="small"
                />
                <el-input
                  v-model="variable.value"
                  placeholder="值"
                  size="small"
                />
                <el-button
                  :icon="Delete"
                  size="small"
                  text
                  @click="removeVariable(index)"
                />
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft,
  Upload,
  Download,
  Check,
  VideoPlay,
  Grid,
  Search,
  ZoomIn,
  ZoomOut,
  RefreshLeft,
  RefreshRight,
  Compass,
  Operation,
  Setting,
  Plus,
  Delete,
  Close,
  Clock,
  MagicStick,
  ChatDotRound,
  Document,
  Timer,
  Warning,
  Share,
  Connection
} from '@element-plus/icons-vue'
import { apiService } from '@/services/api'

const router = useRouter()

// 节点类型定义
const nodeTypes = [
  // 触发器
  { type: 'trigger', label: '定时触发', icon: Clock, color: '#409eff', category: 'triggers' },
  { type: 'trigger', label: 'Webhook', icon: Connection, color: '#409eff', category: 'triggers' },

  // 操作
  { type: 'agent', label: 'Agent 调用', icon: ChatDotRound, color: '#67c23a', category: 'actions' },
  { type: 'skill', label: '技能执行', icon: MagicStick, color: '#67c23a', category: 'actions' },
  { type: 'http', label: 'HTTP 请求', icon: Share, color: '#67c23a', category: 'actions' },
  { type: 'code', label: '代码执行', icon: Document, color: '#67c23a', category: 'actions' },

  // 控制
  { type: 'condition', label: '条件分支', icon: Warning, color: '#e6a23c', category: 'control' },
  { type: 'loop', label: '循环', icon: RefreshRight, color: '#e6a23c', category: 'control' },
  { type: 'delay', label: '延迟', icon: Timer, color: '#e6a23c', category: 'control' },
  { type: 'parallel', label: '并行', icon: Grid, color: '#e6a23c', category: 'control' },

  // 集成
  { type: 'file', label: '文件操作', icon: Folder, color: '#909399', category: 'integration' },
  { type: 'database', label: '数据库', icon: Coin, color: '#909399', category: 'integration' },
  { type: 'notify', label: '通知', icon: Bell, color: '#909399', category: 'integration' }
]

// 状态
const workflowName = ref('新建工作流')
const workflowDescription = ref('')
const workflowType = ref('sequential')
const workflowTimeout = ref(300)
const workflowRetries = ref(0)

const nodeSearchText = ref('')
const activeCategories = ref(['triggers', 'actions'])
const activeTab = ref('node')

const zoom = ref(1)
const selectedNodeId = ref<string>('')
const selectedConnectionId = ref<string>('')

const workflowNodes = ref<Array<{
  id: string
  type: string
  label: string
  icon: any
  color: string
  x: number
  y: number
  config: Record<string, any>
}>>([])

const workflowConnections = ref<Array<{
  id: string
  from: string
  to: string
  fromPort: string
  toPort: string
}>>([])

const workflowVariables = ref<Array<{ name: string; value: string }>>([])

const agents = ref<Array<{ id: string; name: string }>>([])

const canvasRef = ref<HTMLElement>()
const canvasSize = ref({ width: 2000, height: 1500 })

// 历史记录
const history = ref<Array<any>>([])
const historyIndex = ref(-1)

const canUndo = computed(() => historyIndex.value > 0)
const canRedo = computed(() => historyIndex.value < history.value.length - 1)

// 方法
const goBack = () => {
  router.push('/workflow')
}

const filteredNodes = (category: string) => {
  return nodeTypes.filter(node => {
    const matchCategory = node.category === category
    const matchSearch = !nodeSearchText.value ||
      node.label.toLowerCase().includes(nodeSearchText.value.toLowerCase())
    return matchCategory && matchSearch
  })
}

const onNodeDragStart = (event: DragEvent, node: any) => {
  event.dataTransfer!.setData('application/json', JSON.stringify(node))
}

const onCanvasDragOver = (event: DragEvent) => {
  event.preventDefault()
}

const onCanvasDrop = (event: DragEvent) => {
  event.preventDefault()
  const data = event.dataTransfer!.getData('application/json')
  if (!data) return

  const node = JSON.parse(data)
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  addNode(node, x, y)
}

const addNode = (nodeTemplate: any, x: number, y: number) => {
  const newNode = {
    id: `node-${Date.now()}`,
    type: nodeTemplate.type,
    label: nodeTemplate.label,
    icon: nodeTemplate.icon,
    color: nodeTemplate.color,
    x: x - 75,
    y: y - 40,
    config: getDefaultConfig(nodeTemplate.type)
  }

  workflowNodes.value.push(newNode)
  saveToHistory()
}

const getDefaultConfig = (type: string) => {
  const configs: Record<string, any> = {
    agent: { agentId: '', prompt: '' },
    skill: { skillId: '', input: '{}' },
    condition: { condition: '' },
    loop: { count: 3 },
    delay: { duration: 1 },
    http: { url: '', method: 'GET', headers: {} },
    code: { language: 'javascript', code: '' },
    file: { operation: 'read', path: '' },
    database: { operation: 'query', query: '' },
    notify: { type: 'email', to: '', subject: '' }
  }
  return configs[type] || {}
}

const selectNode = (nodeId: string) => {
  selectedNodeId.value = nodeId
  selectedConnectionId.value = ''
  activeTab.value = 'node'
}

const deleteNode = async (nodeId: string) => {
  try {
    await ElMessageBox.confirm('确定要删除此节点吗？', '确认', {
      type: 'warning'
    })

    // 删除节点和相关连接
    workflowNodes.value = workflowNodes.value.filter(n => n.id !== nodeId)
    workflowConnections.value = workflowConnections.value.filter(
      c => c.from !== nodeId && c.to !== nodeId
    )

    if (selectedNodeId.value === nodeId) {
      selectedNodeId.value = ''
    }

    saveToHistory()
  } catch {
    // 用户取消
  }
}

const selectConnection = (connId: string) => {
  selectedConnectionId.value = connId
  selectedNodeId.value = ''
  activeTab.value = 'node'
}

const getConnectionPath = (conn: any) => {
  const fromNode = workflowNodes.value.find(n => n.id === conn.from)
  const toNode = workflowNodes.value.find(n => n.id === conn.to)

  if (!fromNode || !toNode) return ''

  const fromX = fromNode.x + 150
  const fromY = fromNode.y + 40
  const toX = toNode.x
  const toY = toNode.y + 40

  return `M ${fromX} ${fromY} L ${toX} ${toY}`
}

const startNodeDrag = (event: MouseEvent, node: any) => {
  const startX = event.clientX - node.x
  const startY = event.clientY - node.y

  const onMouseMove = (e: MouseEvent) => {
    node.x = e.clientX - startX
    node.y = e.clientY - startY
  }

  const onMouseUp = () => {
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
    saveToHistory()
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

const updateNodePosition = () => {
  // 触发重新渲染
}

const zoomIn = () => {
  zoom.value = Math.min(zoom.value + 0.1, 2)
}

const zoomOut = () => {
  zoom.value = Math.max(zoom.value - 0.1, 0.5)
}

const fitToScreen = () => {
  zoom.value = 1
}

const autoLayout = () => {
  // 简单的自动布局实现
  let currentY = 100
  let currentX = 100

  workflowNodes.value.forEach((node, index) => {
    if (index > 0 && index % 4 === 0) {
      currentX = 100
      currentY += 150
    }
    node.x = currentX
    node.y = currentY
    currentX += 200
  })

  saveToHistory()
}

const saveToHistory = () => {
  const state = {
    nodes: JSON.parse(JSON.stringify(workflowNodes.value)),
    connections: JSON.parse(JSON.stringify(workflowConnections.value))
  }

  // 移除当前指针后的历史
  history.value = history.value.slice(0, historyIndex.value + 1)
  history.value.push(state)
  historyIndex.value = history.value.length - 1
}

const undo = () => {
  if (canUndo.value) {
    historyIndex.value--
    const state = history.value[historyIndex.value]
    workflowNodes.value = JSON.parse(JSON.stringify(state.nodes))
    workflowConnections.value = JSON.parse(JSON.stringify(state.connections))
  }
}

const redo = () => {
  if (canRedo.value) {
    historyIndex.value++
    const state = history.value[historyIndex.value]
    workflowNodes.value = JSON.parse(JSON.stringify(state.nodes))
    workflowConnections.value = JSON.parse(JSON.stringify(state.connections))
  }
}

const addVariable = () => {
  workflowVariables.value.push({ name: '', value: '' })
}

const removeVariable = (index: number) => {
  workflowVariables.value.splice(index, 1)
}

const saveWorkflow = async () => {
  try {
    const definition = JSON.stringify({
      nodes: workflowNodes.value,
      connections: workflowConnections.value,
      variables: workflowVariables.value,
      config: {
        type: workflowType.value,
        timeout: workflowTimeout.value,
        retries: workflowRetries.value
      }
    })

    const result = await apiService.createWorkflow({
      name: workflowName.value,
      description: workflowDescription.value,
      definition
    })

    if (result.success) {
      ElMessage.success('工作流保存成功')
      router.push('/workflow')
    } else {
      throw new Error(result.error)
    }
  } catch (error) {
    ElMessage.error(`保存失败: ${error}`)
  }
}

const runWorkflow = async () => {
  if (workflowNodes.value.length === 0) {
    ElMessage.warning('请先添加工作流节点')
    return
  }

  try {
    // 保存并运行
    await saveWorkflow()
    // TODO: 获取工作流 ID 并执行
    ElMessage.success('工作流执行已启动')
  } catch (error) {
    ElMessage.error(`执行失败: ${error}`)
  }
}

const importWorkflow = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e: any) => {
    const file = e.target.files[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target?.result as string)
        // 加载工作流
        workflowName.value = data.name || '导入的工作流'
        workflowDescription.value = data.description || ''
        workflowNodes.value = data.nodes || []
        workflowConnections.value = data.connections || []
        workflowVariables.value = data.variables || []
        saveToHistory()
        ElMessage.success('工作流导入成功')
      } catch {
        ElMessage.error('无效的工作流文件')
      }
    }
    reader.readAsText(file)
  }
  input.click()
}

const exportWorkflow = () => {
  const data = {
    name: workflowName.value,
    description: workflowDescription.value,
    nodes: workflowNodes.value,
    connections: workflowConnections.value,
    variables: workflowVariables.value,
    config: {
      type: workflowType.value,
      timeout: workflowTimeout.value,
      retries: workflowRetries.value
    }
  }

  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${workflowName.value}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// 加载 agents
const loadAgents = async () => {
  const result = await apiService.listAgents()
  if (result.success && result.data) {
    agents.value = (result.data as string[]).map(id => ({
      id,
      name: id
    }))
  }
}

// 初始化
loadAgents()
</script>

<style scoped>
.workflow-builder {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 80px);
}

.builder-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color);
  background-color: var(--el-bg-color);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.builder-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 节点面板 */
.nodes-panel {
  width: 250px;
  border-right: 1px solid var(--el-border-color);
  background-color: var(--el-bg-color-page);
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
  font-weight: 500;
  border-bottom: 1px solid var(--el-border-color);
}

.panel-search {
  padding: 12px;
}

.node-categories {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px 12px;
}

.node-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: 6px;
  cursor: move;
  transition: background-color 0.2s;
}

.node-item:hover {
  background-color: var(--el-fill-color-light);
}

.node-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.node-label {
  font-size: 13px;
}

/* 画布区域 */
.canvas-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--el-fill-color-extra-light);
  position: relative;
  overflow: hidden;
}

.canvas-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color);
}

.canvas-info {
  margin-left: auto;
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.workflow-canvas {
  flex: 1;
  position: relative;
  overflow: auto;
  transform-origin: 0 0;
}

.connections-layer {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
}

.connection-selected {
  stroke: var(--el-color-primary) !important;
  stroke-width: 3 !important;
}

.workflow-node {
  position: absolute;
  width: 150px;
  background-color: var(--el-bg-color);
  border: 2px solid var(--el-border-color);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: move;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.workflow-node:hover {
  border-color: var(--el-color-primary);
}

.node-selected {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-8);
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  color: white;
  border-radius: 6px 6px 0 0;
  font-size: 13px;
  font-weight: 500;
}

.node-delete {
  margin-left: auto;
  color: white;
}

.node-body {
  padding: 12px;
}

.node-config {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-port {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: var(--el-color-primary);
  border: 2px solid var(--el-bg-color);
}

.input-port {
  top: 50%;
  left: -6px;
  transform: translateY(-50%);
}

.output-port {
  top: 50%;
  right: -6px;
  transform: translateY(-50%);
}

.canvas-empty {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

/* 属性面板 */
.properties-panel {
  width: 300px;
  border-left: 1px solid var(--el-border-color);
  background-color: var(--el-bg-color-page);
  display: flex;
  flex-direction: column;
}

.properties-panel .el-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.properties-panel .el-tabs__content {
  flex: 1;
  overflow-y: auto;
}

.node-properties {
  padding: 16px;
}

.variables-list {
  padding: 16px;
}

.variable-item {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.variable-item .el-input {
  flex: 1;
}
</style>
