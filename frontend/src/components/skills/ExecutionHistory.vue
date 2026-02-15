<template>
  <div class="execution-history-container">
    <!-- 工具栏 -->
    <el-card shadow="hover" class="toolbar-card">
      <div class="toolbar">
        <div class="search-section">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索技能名称或输入..."
            clearable
            style="width: 300px"
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-select
            v-model="statusFilter"
            placeholder="状态筛选"
            clearable
            style="width: 150px; margin-left: 10px"
            @change="handleFilter"
          >
            <el-option label="全部" value="" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="error" />
          </el-select>

          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 260px; margin-left: 10px"
            @change="handleFilter"
          />
        </div>

        <div class="action-section">
          <el-statistic group>
            <el-statistic-item title="总执行次数" :value="totalExecutions" />
            <el-statistic-item title="成功率" :value="successRate" suffix="%" />
          </el-statistic>
          <el-button type="danger" @click="handleClearHistory">
            <el-icon><Delete /></el-icon>
            清空历史
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 历史记录列表 -->
    <el-card shadow="hover" class="history-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Clock /></el-icon>
            执行历史记录
          </span>
          <el-tag type="info">{{ filteredHistory.length }} 条记录</el-tag>
        </div>
      </template>

      <!-- 时间线视图 -->
      <div v-if="displayHistory.length > 0" class="timeline-container">
        <el-timeline>
          <el-timeline-item
            v-for="item in displayHistory"
            :key="item.id"
            :timestamp="formatTimestamp(item.timestamp)"
            :type="item.success ? 'success' : 'danger'"
            placement="top"
            :hollow="!item.success"
          >
            <el-card
              shadow="hover"
              class="history-item-card"
              @click="showDetail(item)"
            >
              <div class="history-item-header">
                <div class="item-title">
                  <el-icon :size="18">
                    <component :is="item.success ? 'CircleCheck' : 'CircleClose'" />
                  </el-icon>
                  <strong>{{ item.skillName }}</strong>
                  <el-tag
                    :type="item.success ? 'success' : 'danger'"
                    size="small"
                  >
                    {{ item.success ? '成功' : '失败' }}
                  </el-tag>
                </div>
                <div class="item-meta">
                  <el-text type="info" size="small">
                    <el-icon><Clock /></el-icon>
                    {{ formatDuration(item.duration) }}
                  </el-text>
                </div>
              </div>

              <div class="history-item-body">
                <div class="input-preview">
                  <el-text type="info" size="small">输入:</el-text>
                  <el-text
                    truncated
                    size="small"
                  >
                    {{ truncateText(item.input, 100) }}
                  </el-text>
                </div>

                <div v-if="item.success && item.output" class="output-preview">
                  <el-text type="success" size="small">输出:</el-text>
                  <el-text truncated size="small">
                    {{ truncateText(formatOutput(item.output), 100) }}
                  </el-text>
                </div>

                <div v-if="!item.success && item.error" class="error-preview">
                  <el-text type="danger" size="small">错误:</el-text>
                  <el-text truncated size="small" type="danger">
                    {{ item.error }}
                  </el-text>
                </div>
              </div>

              <div class="history-item-footer">
                <el-button size="small" type="primary" link>
                  查看详情
                  <el-icon><ArrowRight /></el-icon>
                </el-button>
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>
      </div>

      <!-- 空状态 -->
      <el-empty v-else description="暂无执行历史记录">
        <el-icon :size="64" color="#909399">
          <Clock />
        </el-icon>
      </el-empty>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="filteredHistory.length"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="执行详情"
      width="70%"
      @close="closeDetail"
    >
      <div v-if="selectedItem" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="技能名称">
            {{ selectedItem.skillName }}
          </el-descriptions-item>
          <el-descriptions-item label="执行状态">
            <el-tag :type="selectedItem.success ? 'success' : 'danger'">
              {{ selectedItem.success ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="执行时间">
            {{ formatTimestamp(selectedItem.timestamp) }}
          </el-descriptions-item>
          <el-descriptions-item label="执行耗时">
            {{ formatDuration(selectedItem.duration) }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>输入参数</el-divider>
        <div class="code-block">
          <pre>{{ selectedItem.input }}</pre>
        </div>

        <template v-if="selectedItem.success">
          <el-divider>执行结果</el-divider>
          <div class="code-block">
            <pre>{{ formatOutput(selectedItem.output) }}</pre>
          </div>
        </template>

        <template v-if="!selectedItem.success && selectedItem.error">
          <el-divider>错误信息</el-divider>
          <el-alert
            :title="selectedItem.error"
            type="error"
            :closable="false"
          />
        </template>
      </div>

      <template #footer>
        <el-button @click="closeDetail">关闭</el-button>
        <el-button type="primary" @click="handleReExecute">
          <el-icon><RefreshRight /></el-icon>
          重新执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Delete,
  Clock,
  CircleCheck,
  CircleClose,
  ArrowRight,
  RefreshRight,
} from '@element-plus/icons-vue'
import type { ExecutionHistoryItem } from '@/types/skill'

// Props
interface Props {
  history?: ExecutionHistoryItem[]
}

const props = withDefaults(defineProps<Props>(), {
  history: () => [],
})

// Emits
interface Emits {
  (e: 're-execute', item: ExecutionHistoryItem): void
  (e: 'clear-history'): void
}

const emit = defineEmits<Emits>()

// 状态
const searchKeyword = ref('')
const statusFilter = ref('')
const dateRange = ref<[Date, Date] | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)
const detailDialogVisible = ref(false)
const selectedItem = ref<ExecutionHistoryItem | null>(null)

// 计算属性
const filteredHistory = computed(() => {
  let result = [...props.history]

  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(
      (item) =>
        item.skillName.toLowerCase().includes(keyword) ||
        item.input.toLowerCase().includes(keyword)
    )
  }

  // 状态筛选
  if (statusFilter.value === 'success') {
    result = result.filter((item) => item.success)
  } else if (statusFilter.value === 'error') {
    result = result.filter((item) => !item.success)
  }

  // 日期筛选
  if (dateRange.value && dateRange.value.length === 2) {
    const [start, end] = dateRange.value
    result = result.filter((item) => {
      const itemDate = new Date(item.timestamp)
      return itemDate >= start && itemDate <= end
    })
  }

  return result
})

const displayHistory = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredHistory.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredHistory.value.length / pageSize.value)
})

const totalExecutions = computed(() => {
  return props.history.length
})

const successRate = computed(() => {
  if (totalExecutions.value === 0) return 0
  const successCount = props.history.filter((item) => item.success).length
  return ((successCount / totalExecutions.value) * 100).toFixed(1)
})

// 方法
const handleSearch = () => {
  currentPage.value = 1
}

const handleFilter = () => {
  currentPage.value = 1
}

const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCurrentChange = () => {
  // 页面变化时的处理
}

const showDetail = (item: ExecutionHistoryItem) => {
  selectedItem.value = item
  detailDialogVisible.value = true
}

const closeDetail = () => {
  detailDialogVisible.value = false
  selectedItem.value = null
}

const handleReExecute = () => {
  if (selectedItem.value) {
    emit('re-execute', selectedItem.value)
    closeDetail()
  }
}

const handleClearHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有执行历史记录吗？此操作不可撤销。',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    emit('clear-history')
    ElMessage.success('执行历史已清空')
  } catch {
    // 用户取消
  }
}

const formatTimestamp = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

const formatDuration = (duration: number): string => {
  if (duration < 1000) {
    return `${duration}ms`
  }
  const seconds = (duration / 1000).toFixed(2)
  return `${seconds}s`
}

const formatOutput = (output: any): string => {
  if (typeof output === 'string') {
    return output
  }
  return JSON.stringify(output, null, 2)
}

const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) {
    return text
  }
  return text.substring(0, maxLength) + '...'
}
</script>

<style scoped>
.execution-history-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.search-section {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.action-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 时间线容器 */
.timeline-container {
  max-height: 600px;
  overflow-y: auto;
  padding-right: 16px;
}

.history-item-card {
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 12px;
}

.history-item-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.item-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 4px;
}

.history-item-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 8px;
}

.input-preview,
.output-preview,
.error-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.output-preview {
  background-color: #f0f9ff;
}

.error-preview {
  background-color: #fef0f0;
}

.history-item-footer {
  display: flex;
  justify-content: flex-end;
}

/* 分页 */
.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

/* 详情对话框 */
.detail-content {
  max-height: 60vh;
  overflow-y: auto;
}

.code-block {
  margin: 12px 0;
  padding: 16px;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.code-block pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
}

/* 响应式 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-section {
    flex-direction: column;
    align-items: stretch;
  }

  .action-section {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
