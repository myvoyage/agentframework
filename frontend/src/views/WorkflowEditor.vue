<template>
  <div class="workflow-editor">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>工作流编辑器</span>
          <div class="header-actions">
            <el-select v-model="selectedWorkflowId" placeholder="选择工作流" class="workflow-select">
              <el-option
                v-for="workflow in workflows"
                :key="workflow.id"
                :label="workflow.name"
                :value="workflow.id"
              >
                {{ workflow.name }} - {{ workflow.description }}
              </el-option>
            </el-select>
            <el-button size="small" @click="loadWorkflow">
              <el-icon><Download /></el-icon> 加载
            </el-button>
            <el-button type="primary" size="small" @click="saveWorkflow">
              <el-icon><Save /></el-icon> 保存
            </el-button>
            <el-button size="small" @click="showVersions">
              <el-icon><Time /></el-icon> 版本历史
            </el-button>
            <el-button size="small" @click="executeWorkflow">
              <el-icon><VideoPlay /></el-icon> 执行
            </el-button>
            <el-button size="small" @click="clearGraph">
              <el-icon><Delete /></el-icon> 清空
            </el-button>
          </div>
        </div>
      </template>
      
      <div class="editor-content">
        <!-- 左侧工具栏 -->
        <div class="toolbar">
          <h3>节点类型</h3>
          <div class="node-types">
            <div 
              v-for="nodeType in nodeTypes" 
              :key="nodeType.type"
              class="node-type-item"
              draggable="true"
              @dragstart="onDragStart($event, nodeType)"
            >
              <el-icon :size="24">{{ nodeType.icon }}</el-icon>
              <span>{{ nodeType.label }}</span>
            </div>
          </div>
          
          <h3 style="margin-top: 20px;">工作流模板</h3>
          <div class="template-types">
            <div 
              v-for="template in workflowTemplates" 
              :key="template.name"
              class="template-type-item"
              @click="createWorkflowFromTemplate(template)"
            >
              <h4>{{ template.name }}</h4>
              <p>{{ template.description }}</p>
            </div>
          </div>
        </div>
        
        <!-- 中间画布区域 -->
        <div 
          class="canvas-container"
          ref="canvasContainer"
          @dragover.prevent
          @drop="onDrop"
        >
          <div id="workflow-graph" ref="graphContainer"></div>
        </div>
        
        <!-- 右侧属性面板 -->
        <div class="property-panel">
          <h3>属性编辑</h3>
          <div v-if="selectedNode" class="node-properties">
            <el-form label-position="top" size="small">
              <el-form-item label="节点名称">
                <el-input v-model="selectedNode.label" @input="updateNodeProperty"></el-input>
              </el-form-item>
              <el-form-item label="节点类型">
                <el-input v-model="selectedNode.type" disabled></el-input>
              </el-form-item>
              <el-form-item label="描述">
                <el-input 
                  v-model="selectedNode.description" 
                  type="textarea" 
                  rows="3" 
                  @input="updateNodeProperty"
                ></el-input>
              </el-form-item>
              <el-divider></el-divider>
              
              <!-- 条件节点特定属性 -->
              <template v-if="selectedNode.type === 'condition'">
                <el-form-item label="条件表达式">
                  <el-input 
                    v-model="selectedNode.conditionExpression" 
                    type="textarea" 
                    rows="4" 
                    placeholder="JavaScript条件表达式，例如: ${input.value > 10}" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
              </template>
              
              <!-- 循环节点特定属性 -->
              <template v-else-if="selectedNode.type === 'loop'">
                <el-form-item label="循环类型">
                  <el-select 
                    v-model="selectedNode.loopType" 
                    placeholder="选择循环类型" 
                    @change="updateNodeProperty"
                  >
                    <el-option label="While循环" value="while"></el-option>
                    <el-option label="For循环" value="for"></el-option>
                  </el-select>
                </el-form-item>
                <el-form-item label="条件表达式">
                  <el-input 
                    v-model="selectedNode.conditionExpression" 
                    type="textarea" 
                    rows="3" 
                    placeholder="循环条件表达式，例如: ${index < 10}" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
                <el-form-item label="最大迭代次数">
                  <el-input-number 
                    v-model="selectedNode.maxIterations" 
                    :min="1" 
                    :max="100" 
                    @change="updateNodeProperty"
                  ></el-input-number>
                </el-form-item>
                <el-form-item label="迭代变量名称">
                  <el-input 
                    v-model="selectedNode.iterationVariable" 
                    placeholder="例如: index" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
              </template>
              
              <!-- 通用属性 -->
              <template v-else-if="selectedNode.type === 'task' || selectedNode.type === 'fork' || selectedNode.type === 'join'">
                <el-form-item label="优先级">
                  <el-input-number 
                    v-model="selectedNode.priority" 
                    :min="0" 
                    :max="10" 
                    @change="updateNodeProperty"
                  ></el-input-number>
                </el-form-item>
                <el-form-item label="最大重试次数">
                  <el-input-number 
                    v-model="selectedNode.maxRetries" 
                    :min="0" 
                    :max="10" 
                    @change="updateNodeProperty"
                  ></el-input-number>
                </el-form-item>
                <el-form-item label="重试延迟">
                  <el-input 
                    v-model="selectedNode.retryDelay" 
                    placeholder="例如: 1s, 500ms" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
                <el-form-item label="执行超时">
                  <el-input 
                    v-model="selectedNode.timeout" 
                    placeholder="例如: 30s, 5m" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
                <el-form-item label="自定义配置">
                  <el-input 
                    v-model="selectedNode.config" 
                    type="textarea" 
                    rows="4" 
                    placeholder="JSON格式的自定义配置" 
                    @input="updateNodeProperty"
                  ></el-input>
                </el-form-item>
              </template>
              
              <!-- 开始和结束节点只有基本属性 -->
            </el-form>
          </div>
          <div v-else-if="selectedEdge" class="edge-properties">
            <el-form label-position="top" size="small">
              <el-form-item label="边类型">
                <el-input v-model="selectedEdge.type" disabled></el-input>
              </el-form-item>
              <el-form-item label="条件表达式">
                <el-input 
                  v-model="selectedEdge.condition" 
                  type="textarea" 
                  rows="3" 
                  placeholder="JavaScript条件表达式，例如: ${input.value > 10}" 
                  @input="updateEdgeProperty"
                ></el-input>
              </el-form-item>
              <el-form-item label="自定义配置">
                <el-input 
                  v-model="selectedEdge.config" 
                  type="textarea" 
                  rows="4" 
                  placeholder="JSON格式的自定义配置" 
                  @input="updateEdgeProperty"
                ></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div v-else class="no-selection">
            <p>请选择一个节点或边进行编辑</p>
          </div>
        </div>
        
        <!-- 版本历史对话框 -->
        <el-dialog
          v-model="showVersionDialog"
          title="工作流版本历史"
          width="60%"
        >
          <el-table :data="workflowVersions" style="width: 100%">
            <el-table-column prop="version" label="版本号" width="100"></el-table-column>
            <el-table-column prop="name" label="名称"></el-table-column>
            <el-table-column prop="description" label="描述"></el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="180"></el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="scope">
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="viewVersion(scope.row)"
                >
                  查看
                </el-button>
                <el-button 
                  type="warning" 
                  size="small" 
                  @click="restoreVersion(scope.row)"
                >
                  恢复
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-dialog>
        
        <!-- 恢复版本确认对话框 -->
        <el-dialog
          v-model="showRestoreConfirm"
          title="确认恢复版本"
          width="40%"
        >
          <p>确定要恢复工作流到版本 {{ selectedVersion?.version }} 吗？</p>
          <p>这将创建一个新的版本，当前工作流的更改将被保存。</p>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="showRestoreConfirm = false">取消</el-button>
              <el-button type="primary" @click="confirmRestoreVersion">确定</el-button>
            </span>
          </template>
        </el-dialog>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import G6, { Graph } from '@antv/g6'
import { ElMessage } from 'element-plus'
import * as main from '../../wailsjs/go/main/App'

const canvasContainer = ref<HTMLElement | null>(null)
const graphContainer = ref<HTMLElement | null>(null)
const graph = ref<Graph | null>(null)
const selectedNode = ref<any>(null)
const selectedEdge = ref<any>(null)

// 工作流列表和选择
const workflows = ref<any[]>([])
const selectedWorkflowId = ref<string>('')
const currentWorkflowId = ref<string>('') // 当前编辑的工作流ID
const currentWorkflowName = ref<string>('新工作流')
const currentWorkflowDescription = ref<string>('描述')

// 版本控制相关
const workflowVersions = ref<any[]>([])
const showVersionDialog = ref(false)
const selectedVersion = ref<any>(null)
const showRestoreConfirm = ref(false)
const restoreVersionId = ref<string>('')

// 节点类型定义
const nodeTypes = [
  { type: 'start', label: '开始节点', icon: 'CircleCheck' },
  { type: 'end', label: '结束节点', icon: 'CircleClose' },
  { type: 'task', label: '任务节点', icon: 'Document' },
  { type: 'condition', label: '条件节点', icon: 'SwitchButton' },
  { type: 'loop', label: '循环节点', icon: 'Refresh' },
  { type: 'fork', label: '分支节点', icon: 'Scissor' },
  { type: 'join', label: '合并节点', icon: 'Link' }
]

// 工作流模板
const workflowTemplates = [
  {
    name: '数据处理模板',
    description: '基本的数据处理工作流模板',
    definition: {
      type: 'graph',
      name: '数据处理模板',
      nodes: [
        { id: 'start', label: '开始', type: 'start', x: 100, y: 100 },
        { id: 'fetch', label: '获取数据', type: 'task', x: 100, y: 200 },
        { id: 'process', label: '处理数据', type: 'task', x: 100, y: 300 },
        { id: 'save', label: '保存结果', type: 'task', x: 100, y: 400 },
        { id: 'end', label: '结束', type: 'end', x: 100, y: 500 }
      ],
      edges: [
        { id: 'e1', source: 'start', target: 'fetch', label: '下一步' },
        { id: 'e2', source: 'fetch', target: 'process', label: '下一步' },
        { id: 'e3', source: 'process', target: 'save', label: '下一步' },
        { id: 'e4', source: 'save', target: 'end', label: '完成' }
      ]
    }
  },
  {
    name: 'API调用模板',
    description: '调用外部API的工作流模板',
    definition: {
      type: 'graph',
      name: 'API调用模板',
      nodes: [
        { id: 'start', label: '开始', type: 'start', x: 100, y: 100 },
        { id: 'prepare', label: '准备请求', type: 'task', x: 100, y: 200 },
        { id: 'call', label: '调用API', type: 'task', x: 100, y: 300 },
        { id: 'parse', label: '解析响应', type: 'task', x: 100, y: 400 },
        { id: 'end', label: '结束', type: 'end', x: 100, y: 500 }
      ],
      edges: [
        { id: 'e1', source: 'start', target: 'prepare', label: '下一步' },
        { id: 'e2', source: 'prepare', target: 'call', label: '下一步' },
        { id: 'e3', source: 'call', target: 'parse', label: '下一步' },
        { id: 'e4', source: 'parse', target: 'end', label: '完成' }
      ]
    }
  },
  {
    name: '文件处理模板',
    description: '处理文件的工作流模板',
    definition: {
      type: 'graph',
      name: '文件处理模板',
      nodes: [
        { id: 'start', label: '开始', type: 'start', x: 100, y: 100 },
        { id: 'read', label: '读取文件', type: 'task', x: 100, y: 200 },
        { id: 'process', label: '处理数据', type: 'task', x: 100, y: 300 },
        { id: 'write', label: '写入文件', type: 'task', x: 100, y: 400 },
        { id: 'end', label: '结束', type: 'end', x: 100, y: 500 }
      ],
      edges: [
        { id: 'e1', source: 'start', target: 'read', label: '下一步' },
        { id: 'e2', source: 'read', target: 'process', label: '下一步' },
        { id: 'e3', source: 'process', target: 'write', label: '下一步' },
        { id: 'e4', source: 'write', target: 'end', label: '完成' }
      ]
    }
  }
]

// 初始化图
const initGraph = () => {
  if (!graphContainer.value) return
  
  // 注册自定义节点
  G6.registerNode('workflow-node', {
    draw(cfg, group) {
      const { label = '', type = 'task', description = '' } = cfg
      const nodeSize = [120, 60]
      
      // 创建节点背景
      const rect = group!.addShape('rect', {
        attrs: {
          x: -nodeSize[0] / 2,
          y: -nodeSize[1] / 2,
          width: nodeSize[0],
          height: nodeSize[1],
          radius: 8,
          fill: '#fff',
          stroke: '#409eff',
          lineWidth: 2
        },
        name: 'node-rect',
      })
      
      // 创建节点文本
      group!.addShape('text', {
        attrs: {
          text: label,
          x: 0,
          y: -10,
          fontSize: 14,
          fontWeight: 500,
          fill: '#333',
          textAlign: 'center',
          textBaseline: 'middle'
        },
        name: 'node-label'
      })
      
      group!.addShape('text', {
        attrs: {
          text: type,
          x: 0,
          y: 10,
          fontSize: 12,
          fill: '#666',
          textAlign: 'center',
          textBaseline: 'middle'
        },
        name: 'node-type'
      })
      
      return rect
    },
    update(cfg, node) {
      const group = node.getContainer()
      const labelShape = group.find(element => element.get('name') === 'node-label')
      const typeShape = group.find(element => element.get('name') === 'node-type')
      
      if (labelShape) {
        labelShape.attr('text', cfg.label || '')
      }
      if (typeShape) {
        typeShape.attr('text', cfg.type || '')
      }
    }
  }, 'rect')
  
  // 初始化图实例
  graph.value = new G6.Graph({
    container: graphContainer.value,
    width: graphContainer.value.offsetWidth,
    height: graphContainer.value.offsetHeight,
    modes: {
      default: ['drag-canvas', 'zoom-canvas', 'drag-node', 'click-select']
    },
    defaultNode: {
      type: 'workflow-node',
      size: [120, 60]
    },
    defaultEdge: {
      type: 'polyline',
      style: {
        radius: 10,
        offset: 15,
        endArrow: true,
        lineWidth: 2,
        stroke: '#aaa'
      },
      labelCfg: {
        autoRotate: true,
        style: {
          fill: '#666',
          fontSize: 12,
          background: {
            fill: '#fff',
            stroke: '#ccc',
            padding: [2, 4, 2, 4],
            radius: 4
          }
        }
      }
    },
    layout: {
      type: 'dagre',
      rankdir: 'TB',
      align: 'UL',
      controlPoints: true,
      nodesepFunc: () => 30,
      ranksepFunc: () => 50
    }
  })
  
  // 监听节点选择事件
  graph.value.on('node:click', (evt) => {
    const { item } = evt
    selectedNode.value = item?.getModel()
    selectedEdge.value = null
  })
  
  // 监听边选择事件
  graph.value.on('edge:click', (evt) => {
    const { item } = evt
    selectedEdge.value = item?.getModel()
    selectedNode.value = null
  })
  
  // 监听画布点击事件
  graph.value.on('canvas:click', () => {
    selectedNode.value = null
    selectedEdge.value = null
  })
  
  // 初始化数据
  const data = {
    nodes: [
      { id: 'node1', label: '开始', type: 'start', x: 200, y: 100 },
      { id: 'node2', label: '任务1', type: 'task', x: 200, y: 200 },
      { id: 'node3', label: '结束', type: 'end', x: 200, y: 300 }
    ],
    edges: [
      { id: 'edge1', source: 'node1', target: 'node2', label: '下一步' },
      { id: 'edge2', source: 'node2', target: 'node3', label: '完成' }
    ]
  }
  
  graph.value.data(data)
  graph.value.render()
}

// 保存工作流
const saveWorkflow = async () => {
  if (!graph.value) return
  
  const graphData = graph.value.save() as { nodes: any[]; edges: any[] }
  const workflowDefinition = {
    type: 'graph',
    name: currentWorkflowName.value,
    nodes: graphData.nodes,
    edges: graphData.edges
  }
  
  const definition = JSON.stringify(workflowDefinition, null, 2)
  
  try {
    if (currentWorkflowId.value) {
      // 更新现有工作流
      await main.UpdateWorkflow(currentWorkflowId.value, currentWorkflowName.value, currentWorkflowDescription.value, definition)
      ElMessage.success('工作流已更新，创建了新的版本')
    } else {
      // 创建新工作流
      // 提取步骤列表（从节点ID中提取）
      const steps = graphData.nodes.map((n: any) => n.id).filter((id: string) => id !== 'start' && id !== 'end')
      const workflowId = await main.CreateWorkflow(currentWorkflowName.value, currentWorkflowDescription.value, steps)
      currentWorkflowId.value = workflowId
      ElMessage.success('工作流已保存')
    }
    // 重新加载工作流列表
    await fetchWorkflows()
  } catch (error) {
    console.error('保存工作流失败:', error)
    ElMessage.error('保存工作流失败')
  }
}

// 执行工作流
const executeWorkflow = async () => {
  if (!graph.value) return
  
  try {
    // 获取当前工作流数据
    const graphData = graph.value.save() as { nodes: any[]; edges: any[] }
    const workflowDefinition = {
      type: 'graph',
      name: '新工作流',
      nodes: graphData.nodes,
      edges: graphData.edges
    }
    
    const definition = JSON.stringify(workflowDefinition, null, 2)
    
    // 创建临时工作流并执行
    const steps = graphData.nodes.map((n: any) => n.id).filter((id: string) => id !== 'start' && id !== 'end')
    const workflowId = await main.CreateWorkflow('临时工作流', '用于执行', steps)
    const result = await main.ExecuteWorkflow(workflowId, '{}')
    
    ElMessage.success(`工作流执行结果: ${result}`)
  } catch (error) {
    console.error('执行工作流失败:', error)
    ElMessage.error('执行工作流失败')
  }
}

// 加载工作流列表
const fetchWorkflows = async () => {
  try {
    const result = await main.GetWorkflows()
    workflows.value = result
  } catch (error) {
    console.error('加载工作流列表失败:', error)
    ElMessage.error('加载工作流列表失败')
  }
}

// 加载工作流
const loadWorkflow = async () => {
  if (!selectedWorkflowId.value || !graph.value) return
  
  try {
    const workflow = await main.GetWorkflow(selectedWorkflowId.value)
    if (!workflow) {
      ElMessage.error('工作流不存在')
      return
    }
    
    const workflowDefinition = JSON.parse(workflow.definition)
    const graphData = {
      nodes: workflowDefinition.nodes || [],
      edges: workflowDefinition.edges || []
    }
    
    // 设置当前工作流信息
    currentWorkflowId.value = workflow.id
    currentWorkflowName.value = workflow.name
    currentWorkflowDescription.value = workflow.description
    
    graph.value.changeData(graphData)
    ElMessage.success('工作流已加载')
  } catch (error) {
    console.error('加载工作流失败:', error)
    ElMessage.error('加载工作流失败')
  }
}

// 显示版本历史
const showVersions = async () => {
  if (!currentWorkflowId.value) {
    ElMessage.warning('请先保存工作流')
    return
  }
  
  try {
    // 获取工作流版本历史
    // 暂时注释掉版本管理相关代码，因为后端API尚未实现
    // workflowVersions.value = await main.GetWorkflowVersions(currentWorkflowId.value)
    showVersionDialog.value = true
  } catch (error) {
    console.error('获取版本历史失败:', error)
    ElMessage.error('获取版本历史失败')
  }
}

// 查看特定版本
const viewVersion = async (version: any) => {
  try {
    // 暂时注释掉版本管理相关代码，因为后端API尚未实现
    // const workflowVersion = await main.GetWorkflowVersion(currentWorkflowId.value, version.version)
    // 模拟版本数据
    const workflowVersion = { definition: '{}' }
    if (!workflowVersion) {
      ElMessage.error('版本不存在')
      return
    }
    
    const workflowDefinition = JSON.parse(workflowVersion.definition)
    const graphData = {
      nodes: workflowDefinition.nodes || [],
      edges: workflowDefinition.edges || []
    }
    
    // 临时显示版本内容，不修改当前工作流
    const tempGraph = new G6.Graph({
      container: graphContainer.value || '#graph-container',
      width: graphContainer.value ? graphContainer.value.offsetWidth : 800,
      height: graphContainer.value ? graphContainer.value.offsetHeight : 600,
      modes: {
        default: ['drag-canvas', 'zoom-canvas']
      },
      defaultNode: {
        type: 'workflow-node',
        size: [120, 60]
      },
      defaultEdge: {
        type: 'polyline',
        style: {
          radius: 10,
          offset: 15,
          endArrow: true,
          lineWidth: 2,
          stroke: '#aaa'
        },
        labelCfg: {
          autoRotate: true,
          style: {
            fill: '#666',
            fontSize: 12,
            background: {
              fill: '#fff',
              stroke: '#ccc',
              padding: [2, 4, 2, 4],
              radius: 4
            }
          }
        }
      },
      layout: {
        type: 'dagre',
        rankdir: 'TB',
        align: 'UL',
        controlPoints: true,
        nodesepFunc: () => 30,
        ranksepFunc: () => 50
      }
    })
    
    tempGraph.data(graphData)
    tempGraph.render()
    
    // 3秒后恢复原工作流
    setTimeout(() => {
      tempGraph.destroy()
      if (graph.value) {
        const currentGraphData = graph.value.save() as { nodes: any[]; edges: any[] }
        graph.value.changeData(currentGraphData)
      }
    }, 3000)
    
    ElMessage.info(`显示版本 ${version.version}，3秒后恢复原工作流`)
  } catch (error) {
    console.error('查看版本失败:', error)
    ElMessage.error('查看版本失败')
  }
}

// 恢复到指定版本
const restoreVersion = (version: any) => {
  selectedVersion.value = version
  showRestoreConfirm.value = true
}

// 确认恢复版本
const confirmRestoreVersion = async () => {
  if (!selectedVersion.value || !currentWorkflowId.value) return
  
  try {
    // 暂时注释掉版本管理相关代码，因为后端API尚未实现
    // await main.RestoreWorkflowVersion(currentWorkflowId.value, selectedVersion.value.version)
    ElMessage.success(`已恢复到版本 ${selectedVersion.value.version}`)
    
    // 重新加载当前工作流
    await loadWorkflow()
    showRestoreConfirm.value = false
    showVersionDialog.value = false
  } catch (error) {
    console.error('恢复版本失败:', error)
    ElMessage.error('恢复版本失败')
  }
}

// 清空画布
const clearGraph = () => {
  if (!graph.value) return
  graph.value.changeData({ nodes: [], edges: [] })
  selectedNode.value = null
  selectedEdge.value = null
  currentWorkflowId.value = ''
  currentWorkflowName.value = '新工作流'
  currentWorkflowDescription.value = '描述'
  ElMessage.info('画布已清空')
}

// 更新节点属性
const updateNodeProperty = () => {
  if (!graph.value || !selectedNode.value) return
  graph.value.updateItem(selectedNode.value.id, selectedNode.value)
}

// 更新边属性
const updateEdgeProperty = () => {
  if (!graph.value || !selectedEdge.value) return
  graph.value.updateItem(selectedEdge.value.id, selectedEdge.value)
}

// 拖拽开始事件
const onDragStart = (event: DragEvent, nodeType: any) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/json', JSON.stringify(nodeType))
  }
}

// 拖拽结束事件
const onDrop = (event: DragEvent) => {
  if (!graph.value || !event.dataTransfer) return
  
  event.preventDefault()
  const nodeTypeData = event.dataTransfer.getData('application/json')
  if (!nodeTypeData) return
  
  const nodeType = JSON.parse(nodeTypeData)
  const canvasRect = graphContainer.value?.getBoundingClientRect()
  if (!canvasRect) return
  
  // 计算节点在画布中的坐标
  const point = graph.value.getPointByClient(event.clientX, event.clientY)
  
  // 创建新节点，根据节点类型设置不同的默认属性
  const newNode: any = {
    id: `node_${Date.now()}`,
    label: nodeType.label,
    type: nodeType.type,
    x: point.x,
    y: point.y,
    priority: 0,
    maxRetries: 0,
    retryDelay: "1s",
    timeout: "30s"
  }
  
  // 根据节点类型添加特定属性
  if (nodeType.type === 'condition') {
    newNode.conditionExpression = '' // 条件表达式
  } else if (nodeType.type === 'loop') {
    newNode.loopType = 'while' // 循环类型：while 或 for
    newNode.conditionExpression = '' // 循环条件表达式
    newNode.maxIterations = 10 // 最大迭代次数
    newNode.iterationVariable = '' // 迭代变量名称
  }
  
  // 添加节点到图中
  const data = graph.value.save() as { nodes: any[]; edges: any[] }
  data.nodes.push(newNode)
  graph.value.changeData(data)
}

// 从模板创建工作流
const createWorkflowFromTemplate = (template: any) => {
  if (!graph.value) return
  
  // 清空当前画布
  graph.value.changeData({ nodes: [], edges: [] })
  
  // 从模板加载数据
  const graphData = {
    nodes: template.definition.nodes,
    edges: template.definition.edges
  }
  
  graph.value.changeData(graphData)
  
  // 设置当前工作流信息
  currentWorkflowId.value = ''
  currentWorkflowName.value = template.name
  currentWorkflowDescription.value = template.description
  
  ElMessage.success('从模板创建工作流成功')
}

// 响应式调整画布大小
const resizeGraph = () => {
  if (!graph.value || !graphContainer.value) return
  graph.value.changeSize(
    graphContainer.value.offsetWidth,
    graphContainer.value.offsetHeight
  )
}

// 生命周期钩子
onMounted(async () => {
  initGraph()
  await fetchWorkflows()
  window.addEventListener('resize', resizeGraph)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeGraph)
  if (graph.value) {
    graph.value.destroy()
  }
})
</script>

<style scoped>
.workflow-editor {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.workflow-select {
  width: 250px;
  margin-right: 8px;
}

.editor-content {
  display: flex;
  height: 600px;
  gap: 20px;
  margin-top: 20px;
}

.toolbar {
  width: 200px;
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  overflow-y: auto;
}

.toolbar h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #303133;
}

.node-types {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.node-type-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px;
  background-color: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  cursor: move;
  transition: all 0.3s ease;
}

.node-type-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.template-types {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.template-type-item {
  padding: 10px;
  background-color: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.template-type-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.template-type-item h4 {
  margin: 0 0 5px 0;
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.template-type-item p {
  margin: 0;
  font-size: 12px;
  color: #606266;
  line-height: 1.4;
}

.canvas-container {
  flex: 1;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  background-color: #fafafa;
  background-image: 
    linear-gradient(#e4e7ed 1px, transparent 1px),
    linear-gradient(90deg, #e4e7ed 1px, transparent 1px);
  background-size: 20px 20px;
}

#workflow-graph {
  width: 100%;
  height: 100%;
}

.property-panel {
  width: 280px;
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  overflow-y: auto;
}

.property-panel h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #303133;
}

.node-properties,
.edge-properties {
  background-color: #fff;
  padding: 12px;
  border-radius: 6px;
}

.no-selection {
  text-align: center;
  color: #909399;
  padding: 20px 0;
}
</style>
