<template>
  <div class="logs-view">
    <div class="logs-header">
      <el-page-header title="系统日志">
        <template #extra>
          <el-button-group>
            <el-button :icon="Refresh" @click="refreshLogs">刷新</el-button>
            <el-button :icon="Download" @click="downloadLogs">下载</el-button>
            <el-button :icon="Delete" @click="clearLogs">清空</el-button>
          </el-button-group>
        </template>
      </el-page-header>
    </div>

    <!-- 过滤器 -->
    <el-card shadow="never" class="filters-card">
      <el-row :gutter="16">
        <el-col :span="4">
          <el-select v-model="filters.level" placeholder="日志级别" @change="applyFilters">
            <el-option label="全部" value="" />
            <el-option label="调试" value="debug" />
            <el-option label="信息" value="info" />
            <el-option label="警告" value="warn" />
            <el-option label="错误" value="error" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.module" placeholder="模块" @change="applyFilters">
            <el-option label="全部" value="" />
            <el-option label="Agent" value="agent" />
            <el-option label="Workflow" value="workflow" />
            <el-option label="Skill" value="skill" />
            <el-option label="API" value="api" />
          </el-select>
        </el-col>
        <el-col :span="8">
          <el-input
            v-model="filters.search"
            placeholder="搜索日志..."
            :prefix-icon="Search"
            clearable
            @input="applyFilters"
          />
        </el-col>
        <el-col :span="4">
          <el-switch
            v-model="filters.autoRefresh"
            active-text="自动刷新"
            @change="toggleAutoRefresh"
          />
        </el-col>
        <el-col :span="4">
          <el-checkbox v-model="filters.followTail" @change="scrollToBottom">
            跟随最新
          </el-checkbox>
        </el-col>
      </el-row>
    </el-card>

    <!-- 日志内容 -->
    <el-card shadow="never" class="logs-card" ref="logsCardRef">
      <div class="logs-content" ref="logsContentRef">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-entry"
          :class="`log-${log.level}`"
        >
          <span class="log-time">{{ formatTime(log.timestamp) }}</span>
          <span class="log-level">{{ log.level.toUpperCase() }}</span>
          <span class="log-module">{{ log.module }}</span>
          <span class="log-message">{{ log.message }}</span>
          <el-button
            v-if="log.details"
            text
            size="small"
            @click="toggleDetails(index)"
          >
            {{ log.showDetails ? '隐藏' : '详情' }}
          </el-button>
          <pre v-if="log.showDetails && log.details" class="log-details">{{ log.details }}</pre>
        </div>
        <el-empty
          v-if="filteredLogs.length === 0"
          description="暂无日志"
          :image-size="80"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Download, Delete, Search } from '@element-plus/icons-vue'

interface LogEntry {
  timestamp: Date
  level: string
  module: string
  message: string
  details?: string
  showDetails?: boolean
}

const filters = ref({
  level: '',
  module: '',
  search: '',
  autoRefresh: false,
  followTail: true
})

const logs = ref<LogEntry[]>([])
const logsContentRef = ref<HTMLElement>()
const logsCardRef = ref<HTMLElement>()

let refreshInterval: any = null

const filteredLogs = computed(() => {
  let result = logs.value

  if (filters.value.level) {
    result = result.filter(log => log.level === filters.value.level)
  }

  if (filters.value.module) {
    result = result.filter(log => log.module === filters.value.module)
  }

  if (filters.value.search) {
    const search = filters.value.search.toLowerCase()
    result = result.filter(log =>
      log.message.toLowerCase().includes(search) ||
      log.module.toLowerCase().includes(search)
    )
  }

  return result
})

const formatTime = (date: Date) => {
  return new Date(date).toLocaleTimeString('zh-CN', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const applyFilters = () => {
  // 触发计算属性更新
}

const refreshLogs = () => {
  // 模拟日志加载
  const sampleLogs: LogEntry[] = [
    {
      timestamp: new Date(Date.now() - 5000),
      level: 'info',
      module: 'agent',
      message: 'Agent "助手" 已加载'
    },
    {
      timestamp: new Date(Date.now() - 4000),
      level: 'info',
      module: 'workflow',
      message: '工作流 "数据处理" 执行完成'
    },
    {
      timestamp: new Date(Date.now() - 3000),
      level: 'warn',
      module: 'skill',
      message: '技能 "数据分析" 响应时间较长',
      details: '执行耗时: 3.5s\n预期耗时: < 1s'
    },
    {
      timestamp: new Date(Date.now() - 2000),
      level: 'error',
      module: 'api',
      message: 'HTTP 请求失败',
      details: 'Error: Connection refused\nURL: http://localhost:8080/api/test'
    },
    {
      timestamp: new Date(Date.now() - 1000),
      level: 'info',
      module: 'system',
      message: 'API 服务器已启动在端口 8080'
    }
  ]

  logs.value = [...sampleLogs, ...logs.value].slice(0, 100)

  if (filters.value.followTail) {
    nextTick(() => {
      scrollToBottom()
    })
  }
}

const downloadLogs = () => {
  const content = filteredLogs.value.map(log =>
    `[${formatTime(log.timestamp)}] [${log.level.toUpperCase()}] [${log.module}] ${log.message}${log.details ? '\n' + log.details : ''}`
  ).join('\n')

  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `logs-${Date.now()}.txt`
  a.click()
  URL.revokeObjectURL(url)

  ElMessage.success('日志已下载')
}

const clearLogs = async () => {
  logs.value = []
  ElMessage.success('日志已清空')
}

const toggleDetails = (index: number) => {
  if (filteredLogs.value[index].showDetails) {
    delete filteredLogs.value[index].showDetails
  } else {
    filteredLogs.value[index].showDetails = true
  }
}

const toggleAutoRefresh = () => {
  if (filters.value.autoRefresh) {
    refreshInterval = setInterval(refreshLogs, 5000)
  } else {
    clearInterval(refreshInterval)
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (logsContentRef.value && logsCardRef.value) {
      logsContentRef.value.scrollTop = logsContentRef.value.scrollHeight
    }
  })
}

onMounted(() => {
  refreshLogs()
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.logs-view {
  padding: 20px;
}

.filters-card {
  margin-bottom: 20px;
}

.logs-card {
  height: calc(100vh - 300px);
}

.logs-content {
  height: 100%;
  overflow-y: auto;
  background-color: #1e1e1e;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.log-entry {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #333;
  line-height: 1.5;
}

.log-time {
  color: #909399;
}

.log-level {
  font-weight: 500;
}

.log-debug .log-level {
  color: #909399;
}

.log-info .log-level {
  color: #409eff;
}

.log-warn .log-level {
  color: #e6a23c;
}

.log-error .log-level {
  color: #f56c6c;
}

.log-module {
  color: #67c23a;
}

.log-message {
  flex: 1;
  color: #c0c4cc;
}

.log-details {
  width: 100%;
  margin-top: 8px;
  padding: 8px;
  background-color: #2d2d2d;
  border-radius: 4px;
  color: #909399;
  white-space: pre-wrap;
}
</style>
