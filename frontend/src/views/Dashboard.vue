<template>
  <div class="dashboard-container">
    <!-- 顶部状态栏 -->
    <el-row :gutter="20" class="status-bar">
      <el-col :span="18">
        <el-alert
          :type="apiStore.apiConnected ? 'success' : 'warning'"
          :closable="false"
          show-icon
        >
          <template #title>
            <span class="status-title">
              API 状态: {{ apiStore.apiConnected ? '已连接' : '未连接' }}
              ({{ apiStore.usingHttpApi ? 'HTTP 模式' : 'Wails 模式' }})
            </span>
          </template>
        </el-alert>
      </el-col>
      <el-col :span="6">
        <div class="theme-toggle">
          <el-switch
            v-model="isDark"
            active-text="暗色"
            inactive-text="亮色"
            @change="toggleTheme"
          />
        </div>
      </el-col>
    </el-row>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <stat-card
          title="总 Agents"
          :value="stats.agents"
          icon="User"
          color="#409eff"
          :loading="loading"
        />
      </el-col>
      <el-col :span="6">
        <stat-card
          title="工作流"
          :value="stats.workflows"
          icon="Operation"
          color="#67c23a"
          :loading="loading"
        />
      </el-col>
      <el-col :span="6">
        <stat-card
          title="技能"
          :value="stats.skills"
          :value2="stats.enabledSkills"
          value2-label="已启用"
          icon="MagicStick"
          color="#e6a23c"
          :loading="loading"
        />
      </el-col>
      <el-col :span="6">
        <stat-card
          title="对话消息"
          :value="stats.messages"
          icon="ChatDotRound"
          color="#f56c6c"
          :loading="loading"
        />
      </el-col>
    </el-row>

    <!-- 主要内容区 -->
    <el-row :gutter="20" class="content-row">
      <!-- 左侧：快速操作 -->
      <el-col :span="8">
        <el-card shadow="hover" class="action-card">
          <template #header>
            <div class="card-header">
              <el-icon><Grid /></el-icon>
              <span>快速操作</span>
            </div>
          </template>
          <div class="action-list">
            <el-button
              type="primary"
              plain
              size="large"
              :icon="ChatDotRound"
              @click="goToChat"
              class="action-btn"
            >
              开始对话
            </el-button>
            <el-button
              type="success"
              plain
              size="large"
              :icon="Operation"
              @click="goToWorkflow"
              class="action-btn"
            >
              创建工作流
            </el-button>
            <el-button
              type="warning"
              plain
              size="large"
              :icon="MagicStick"
              @click="goToSkills"
              class="action-btn"
            >
              管理技能
            </el-button>
            <el-button
              type="info"
              plain
              size="large"
              :icon="Setting"
              @click="goToConfig"
              class="action-btn"
            >
              配置设置
            </el-button>
          </div>
        </el-card>

        <!-- 最近活动 -->
        <el-card shadow="hover" class="activity-card">
          <template #header>
            <div class="card-header">
              <el-icon><Clock /></el-icon>
              <span>最近活动</span>
            </div>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="activity in recentActivities"
              :key="activity.id"
              :timestamp="formatTime(activity.timestamp)"
              placement="top"
            >
              <div class="activity-item">
                <el-tag :type="activity.type" size="small">{{ activity.category }}</el-tag>
                <span>{{ activity.message }}</span>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>

      <!-- 中间：工作流执行状态 -->
      <el-col :span="10">
        <el-card shadow="hover" class="workflow-card">
          <template #header>
            <div class="card-header">
              <el-icon><DataAnalysis /></el-icon>
              <span>工作流状态</span>
              <el-button
                text
                :icon="Refresh"
                @click="refreshWorkflows"
                :loading="refreshing"
              >
                刷新
              </el-button>
            </div>
          </template>
          <div class="workflow-list">
            <el-empty
              v-if="workflowStore.workflows.length === 0 && !workflowStore.loading"
              description="暂无工作流"
            >
              <el-button type="primary" @click="goToWorkflow">创建工作流</el-button>
            </el-empty>
            <div
              v-for="workflow in recentWorkflows"
              :key="workflow.id"
              class="workflow-item"
              @click="viewWorkflow(workflow.id)"
            >
              <div class="workflow-info">
                <div class="workflow-name">{{ workflow.name }}</div>
                <div class="workflow-desc">{{ workflow.description }}</div>
              </div>
              <el-tag
                :type="getWorkflowStatusType(workflow.status)"
                size="small"
              >
                {{ workflow.status }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- 执行历史 -->
        <el-card shadow="hover" class="execution-card">
          <template #header>
            <div class="card-header">
              <el-icon><List /></el-icon>
              <span>执行历史</span>
            </div>
          </template>
          <el-table
            :data="recentExecutions"
            stripe
            size="small"
            :show-header="false"
          >
            <el-table-column prop="workflowId" label="工作流" width="120" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag
                  :type="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'danger' : 'info'"
                  size="small"
                >
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时间">
              <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 右侧：系统状态和通知 -->
      <el-col :span="6">
        <!-- 系统状态 -->
        <el-card shadow="hover" class="system-card">
          <template #header>
            <div class="card-header">
              <el-icon><Monitor /></el-icon>
              <span>系统状态</span>
            </div>
          </template>
          <div class="system-status">
            <div class="status-item">
              <span class="status-label">CPU</span>
              <el-progress
                :percentage="systemStats.cpu"
                :color="getProgressColor(systemStats.cpu)"
                :show-text="false"
              />
              <span class="status-value">{{ systemStats.cpu }}%</span>
            </div>
            <div class="status-item">
              <span class="status-label">内存</span>
              <el-progress
                :percentage="systemStats.memory"
                :color="getProgressColor(systemStats.memory)"
                :show-text="false"
              />
              <span class="status-value">{{ systemStats.memory }}%</span>
            </div>
            <div class="status-item">
              <span class="status-label">磁盘</span>
              <el-progress
                :percentage="systemStats.disk"
                :color="getProgressColor(systemStats.disk)"
                :show-text="false"
              />
              <span class="status-value">{{ systemStats.disk }}%</span>
            </div>
          </div>
        </el-card>

        <!-- 通知中心 -->
        <el-card shadow="hover" class="notification-card">
          <template #header>
            <div class="card-header">
              <el-icon><Bell /></el-icon>
              <span>通知 ({{ appStore.unreadNotifications }})</span>
              <el-button
                v-if="appStore.notifications.length > 0"
                text
                size="small"
                @click="clearNotifications"
              >
                清空
              </el-button>
            </div>
          </template>
          <div class="notification-list">
            <div
              v-for="notification in appStore.notifications.slice(0, 5)"
              :key="notification.id"
              class="notification-item"
              :class="`notification-${notification.type}`"
            >
              <div class="notification-message">{{ notification.message }}</div>
              <div class="notification-time">{{ formatTime(notification.timestamp) }}</div>
            </div>
            <el-empty
              v-if="appStore.notifications.length === 0"
              description="暂无通知"
              :image-size="60"
            />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/appStore'
import { useWorkflowStore } from '@/stores/workflowStore'
import { useSkillStore } from '@/stores/skillStore'
import { apiService } from '@/services/api'
import {
  Grid,
  ChatDotRound,
  Operation,
  MagicStick,
  Setting,
  Clock,
  DataAnalysis,
  List,
  Monitor,
  Bell,
  Refresh,
  User
} from '@element-plus/icons-vue'

const router = useRouter()
const appStore = useAppStore()
const workflowStore = useWorkflowStore()
const skillStore = useSkillStore()

// 状态
const loading = ref(false)
const refreshing = ref(false)
const isDark = computed(() => appStore.isDarkMode)

// 统计数据
const stats = ref({
  agents: 0,
  workflows: 0,
  skills: 0,
  enabledSkills: 0,
  messages: 0
})

// 系统状态
const systemStats = ref({
  cpu: 25,
  memory: 45,
  disk: 60
})

// 最近活动
const recentActivities = ref([
  {
    id: '1',
    category: '工作流',
    type: 'success',
    message: '工作流 "数据处理" 执行完成',
    timestamp: new Date(Date.now() - 300000)
  },
  {
    id: '2',
    category: '技能',
    type: 'info',
    message: '技能 "数据分析" 已启用',
    timestamp: new Date(Date.now() - 600000)
  },
  {
    id: '3',
    category: 'Agent',
    type: 'warning',
    message: 'Agent "助手" 响应时间较长',
    timestamp: new Date(Date.now() - 900000)
  }
])

// 计算属性
const recentWorkflows = computed(() => {
  return workflowStore.workflows.slice(0, 5)
})

const recentExecutions = computed(() => {
  return workflowStore.recentExecutions.slice(0, 5)
})

// 方法
const toggleTheme = () => {
  appStore.toggleTheme()
}

const goToChat = () => {
  router.push('/chat')
}

const goToWorkflow = () => {
  router.push('/workflow')
}

const goToSkills = () => {
  router.push('/skills')
}

const goToConfig = () => {
  router.push('/config')
}

const viewWorkflow = (id: string) => {
  router.push(`/workflow/${id}`)
}

const refreshWorkflows = async () => {
  refreshing.value = true
  try {
    await workflowStore.loadWorkflows()
    await loadStats()
  } finally {
    refreshing.value = false
  }
}

const clearNotifications = () => {
  appStore.clearNotifications()
}

const formatTime = (date: Date) => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (days > 0) return `${days} 天前`
  if (hours > 0) return `${hours} 小时前`
  if (minutes > 0) return `${minutes} 分钟前`
  return '刚刚'
}

const getWorkflowStatusType = (status: string) => {
  const types: Record<string, any> = {
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    pending: 'info'
  }
  return types[status] || 'info'
}

const getProgressColor = (value: number) => {
  if (value < 50) return '#67c23a'
  if (value < 80) return '#e6a23c'
  return '#f56c6c'
}

const loadStats = async () => {
  try {
    // 加载各模块数据
    await Promise.all([
      workflowStore.loadWorkflows(),
      skillStore.loadSkills()
    ])

    stats.value = {
      agents: 5, // TODO: 从 API 获取
      workflows: workflowStore.workflows.length,
      skills: skillStore.skills.length,
      enabledSkills: skillStore.enabledSkills.length,
      messages: recentActivities.value.length
    }
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

// 生命周期
onMounted(async () => {
  loading.value = true
  try {
    await loadStats()

    // 设置 WebSocket 监听
    workflowStore.setupWebSocketListeners()
    skillStore.setupWebSocketListeners()

    // 定期刷新统计数据
    const interval = setInterval(loadStats, 30000)
    onUnmounted(() => clearInterval(interval))
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  workflowStore.cleanupWebSocketListeners()
  skillStore.cleanupWebSocketListeners()
})

// StatCard 组件
const StatCard = {
  name: 'StatCard',
  props: {
    title: String,
    value: [Number, String],
    value2: Number,
    value2Label: String,
    icon: String,
    color: String,
    loading: Boolean
  },
  template: `
    <el-card shadow="hover" class="stat-card" v-loading="loading">
      <div class="stat-content">
        <div class="stat-icon" :style="{ color: color }">
          <el-icon :size="32"><component :is="icon" /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-title">{{ title }}</div>
          <div class="stat-value">{{ value }}</div>
          <div v-if="value2 !== undefined" class="stat-sub">
            <span class="stat-label">{{ value2Label }}:</span>
            <span class="stat-sub-value">{{ value2 }}</span>
          </div>
        </div>
      </div>
    </el-card>
  `
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

.status-bar {
  margin-bottom: 20px;
}

.status-title {
  font-weight: 500;
}

.theme-toggle {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.stats-row {
  margin-bottom: 20px;
}

.content-row {
  margin-top: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

/* Action Card */
.action-card {
  margin-bottom: 20px;
}

.action-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-btn {
  width: 100%;
}

/* Activity Card */
.activity-card {
  height: calc(100% - 20px);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Workflow Card */
.workflow-card {
  margin-bottom: 20px;
}

.workflow-list {
  max-height: 300px;
  overflow-y: auto;
}

.workflow-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.workflow-item:hover {
  background-color: var(--el-fill-color-light);
}

.workflow-info {
  flex: 1;
}

.workflow-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.workflow-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Execution Card */
.execution-card {
  margin-bottom: 20px;
}

/* System Card */
.system-card {
  margin-bottom: 20px;
}

.system-status {
  padding: 10px 0;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.status-label {
  width: 50px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.status-item .el-progress {
  flex: 1;
}

.status-value {
  width: 40px;
  text-align: right;
  font-size: 12px;
  font-weight: 500;
}

/* Notification Card */
.notification-card {
  height: calc(100% - 20px);
}

.notification-list {
  max-height: 300px;
  overflow-y: auto;
}

.notification-item {
  padding: 10px;
  border-radius: 6px;
  margin-bottom: 8px;
  border-left: 3px solid;
}

.notification-success {
  border-left-color: #67c23a;
  background-color: rgba(103, 194, 58, 0.1);
}

.notification-error {
  border-left-color: #f56c6c;
  background-color: rgba(245, 108, 108, 0.1);
}

.notification-warning {
  border-left-color: #e6a23c;
  background-color: rgba(230, 162, 60, 0.1);
}

.notification-info {
  border-left-color: #409eff;
  background-color: rgba(64, 158, 255, 0.1);
}

.notification-message {
  font-size: 13px;
  margin-bottom: 4px;
}

.notification-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

/* Stat Card */
.stat-card {
  height: 100px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background-color: var(--el-fill-color-light);
}

.stat-info {
  flex: 1;
}

.stat-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
}

.stat-sub {
  font-size: 12px;
  margin-top: 4px;
}

.stat-label {
  color: var(--el-text-color-secondary);
}

.stat-sub-value {
  font-weight: 500;
  margin-left: 4px;
}
</style>
