<template>
  <div class="skill-manager">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>
          <el-icon :size="24"><Grid /></el-icon>
          技能管理系统
        </h2>
        <el-text type="info">一体化技能管理、执行和监控平台</el-text>
      </div>
      <div class="header-right">
        <el-button-group>
          <el-button @click="handleRefreshAll" :loading="refreshing">
            <el-icon><Refresh /></el-icon>
            刷新全部
          </el-button>
          <el-button @click="handleShowHelp">
            <el-icon><QuestionFilled /></el-icon>
            帮助
          </el-button>
        </el-button-group>
      </div>
    </div>

    <!-- 系统状态概览 -->
    <el-card shadow="hover" class="overview-card" v-if="!loading && isReady">
      <div class="overview-stats">
        <div class="stat-item">
          <div class="stat-value">{{ skills.length }}</div>
          <div class="stat-label">已注册技能</div>
        </div>
        <el-divider direction="vertical" />
        <div class="stat-item">
          <div class="stat-value">{{ enabledSkills.length }}</div>
          <div class="stat-label">启用技能</div>
        </div>
        <el-divider direction="vertical" />
        <div class="stat-item">
          <div class="stat-value">{{ totalExecutions }}</div>
          <div class="stat-label">总执行次数</div>
        </div>
        <el-divider direction="vertical" />
        <div class="stat-item">
          <div class="stat-value">{{ executionHistory.length }}</div>
          <div class="stat-label">历史记录</div>
        </div>
      </div>
    </el-card>

    <!-- 主内容区域 - 标签页 -->
    <el-card shadow="hover" class="main-card" v-loading="loading">
      <el-tabs v-model="activeTab" type="card" @tab-change="handleTabChange">
        <!-- Tab 1: 技能列表 -->
        <el-tab-pane name="list">
          <template #label>
            <span class="tab-label">
              <el-icon><List /></el-icon>
              技能列表
              <el-badge v-if="skills.length > 0" :value="skills.length" />
            </span>
          </template>
          <SkillList
            :skills="skills"
            :loading="loading"
            @refresh="handleRefreshSkills"
            @view-detail="handleViewSkillDetail"
            @execute="handleExecuteSkill"
            @toggle="handleToggleSkill"
          />
        </el-tab-pane>

        <!-- Tab 2: 技能执行 -->
        <el-tab-pane name="execute">
          <template #label>
            <span class="tab-label">
              <el-icon><VideoPlay /></el-icon>
              技能执行
            </span>
          </template>
          <SkillExecutor
            :skills="enabledSkills"
            :loading="executing"
            @refresh-skills="handleRefreshSkills"
            @execute="handleExecuteSkillRequest"
          />
        </el-tab-pane>

        <!-- Tab 3: 技能定义 -->
        <el-tab-pane name="definition">
          <template #label>
            <span class="tab-label">
              <el-icon><Document /></el-icon>
              技能定义
            </span>
          </template>
          <SkillDefinitionViewer :definitions="definitions" />
        </el-tab-pane>

        <!-- Tab 4: 性能监控 -->
        <el-tab-pane name="monitor">
          <template #label>
            <span class="tab-label">
              <el-icon><TrendCharts /></el-icon>
              性能监控
            </span>
          </template>
          <PerformanceMonitor
            :cache-stats="cacheStats"
            :pool-stats="poolStats"
            :system-stats="systemStats"
            :loading="loading"
            @refresh="handleRefreshStats"
          />
        </el-tab-pane>

        <!-- Tab 5: 执行历史 -->
        <el-tab-pane name="history">
          <template #label>
            <span class="tab-label">
              <el-icon><Clock /></el-icon>
              执行历史
              <el-badge v-if="executionHistory.length > 0" :value="executionHistory.length" />
            </span>
          </template>
          <ExecutionHistory
            :history="executionHistory"
            @re-execute="handleReExecute"
            @clear-history="handleClearHistory"
          />
        </el-tab-pane>

        <!-- Tab 6: 导入技能 -->
        <el-tab-pane name="import">
          <template #label>
            <span class="tab-label">
              <el-icon><Upload /></el-icon>
              导入技能
            </span>
          </template>
          <SkillImport
            :workspace="workspace"
            @imported="handleSkillImported"
            @execute="handleExecuteImportedSkill"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 技能详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="`技能详情 - ${selectedSkill?.name}`"
      width="60%"
    >
      <div v-if="selectedSkill">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="技能ID">
            {{ selectedSkill.id }}
          </el-descriptions-item>
          <el-descriptions-item label="名称">
            {{ selectedSkill.name }}
          </el-descriptions-item>
          <el-descriptions-item label="分类">
            <el-tag :type="getCategoryTagType(selectedSkill.category)">
              {{ getCategoryLabel(selectedSkill.category) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="版本">
            {{ selectedSkill.version }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ selectedSkill.description }}
          </el-descriptions-item>
          <el-descriptions-item label="标签" :span="2">
            <el-tag
              v-for="tag in selectedSkill.tags"
              :key="tag"
              style="margin-right: 8px"
            >
              {{ tag }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="使用次数">
            {{ selectedSkill.useCount }}
          </el-descriptions-item>
          <el-descriptions-item label="最后使用">
            {{ selectedSkill.lastUsed || '从未使用' }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleExecuteFromDetail">
          <el-icon><VideoPlay /></el-icon>
          执行此技能
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Grid,
  Refresh,
  QuestionFilled,
  List,
  VideoPlay,
  Document,
  TrendCharts,
  Clock,
  Upload,
} from '@element-plus/icons-vue'
import { useSkillSystem } from '@/composables/useSkillSystem'
import SkillList from '@/components/skills/SkillList.vue'
import SkillExecutor from '@/components/skills/SkillExecutor.vue'
import SkillDefinitionViewer from '@/components/skills/SkillDefinitionViewer.vue'
import PerformanceMonitor from '@/components/skills/PerformanceMonitor.vue'
import ExecutionHistory from '@/components/skills/ExecutionHistory.vue'
import SkillImport from '@/components/skills/SkillImport.vue'
import type { SkillListItem, ExecuteSkillInput, ExecuteSkillOutput, ExecutionHistoryItem } from '@/types/skill'

// 使用技能系统 composable
const {
  state,
  systemStats,
  cacheStats,
  poolStats,
  enabledSkills,
  totalExecutions,
  isReady,
  loadSkills,
  loadDefinitions,
  loadStats,
  refresh,
  selectSkill,
  executeSkill,
  toggleSkill,
  clearHistory,
  reset,
} = useSkillSystem()

// 状态
const activeTab = ref('list')
const detailDialogVisible = ref(false)
const selectedSkill = ref<SkillListItem | undefined>(undefined)
const refreshing = ref(false)
const workspace = ref('/workspace')

// 计算属性
const loading = computed(() => state.loading)
const executing = computed(() => state.executing)
const skills = computed(() => state.skills)
const definitions = computed(() => state.definitions)
const executionHistory = computed(() => state.executionHistory)

// 生命周期
onMounted(async () => {
  console.log('[SkillManager] 组件已挂载，开始加载数据...')
  await refresh()
})

onUnmounted(() => {
  console.log('[SkillManager] 组件已卸载')
})

// 方法
const handleRefreshAll = async () => {
  refreshing.value = true
  try {
    await refresh()
    ElMessage.success('所有数据已刷新')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    refreshing.value = false
  }
}

const handleRefreshSkills = async () => {
  await loadSkills()
}

const handleRefreshStats = async () => {
  await loadStats()
}

const handleTabChange = (tabName: string) => {
  console.log('[SkillManager] 切换到标签页:', tabName)

  // 根据标签页加载相应数据
  switch (tabName) {
    case 'definition':
      if (definitions.value.length === 0) {
        loadDefinitions()
      }
      break
    case 'monitor':
      loadStats()
      break
  }
}

const handleViewSkillDetail = (skill: SkillListItem) => {
  selectedSkill.value = skill
  detailDialogVisible.value = true
}

const handleExecuteSkill = (skill: SkillListItem) => {
  selectedSkill.value = skill
  activeTab.value = 'execute'
}

const handleExecuteFromDetail = () => {
  if (selectedSkill.value) {
    detailDialogVisible.value = false
    activeTab.value = 'execute'
  }
}

const handleExecuteSkillRequest = async (input: ExecuteSkillInput): Promise<ExecuteSkillOutput> => {
  return await executeSkill(input)
}

const handleToggleSkill = async (skill: SkillListItem) => {
  await toggleSkill(skill)
}

const handleReExecute = async (item: ExecutionHistoryItem) => {
  const input: ExecuteSkillInput = {
    skillName: item.skillName,
    input: item.input,
    workspace: '/workspace',
  }
  activeTab.value = 'execute'
  // SkillExecutor 组件会自动填充输入
}

const handleClearHistory = () => {
  clearHistory()
}

const handleShowHelp = () => {
  ElMessage.info('帮助文档正在开发中...')
}

const handleSkillImported = async (skill: any) => {
  console.log('[SkillManager] 技能导入成功:', skill)
  ElMessage.success(`技能 "${skill.skillName || skill.name}" 导入成功！`)

  // 刷新技能列表
  await loadSkills()

  // 切换到技能列表标签页
  activeTab.value = 'list'
}

const handleExecuteImportedSkill = (skill: any) => {
  console.log('[SkillManager] 执行导入的技能:', skill)

  // 切换到执行标签页
  activeTab.value = 'execute'

  // TODO: 将技能信息传递给SkillExecutor组件
  selectedSkill.value = skill
}

const getCategoryTagType = (category: string): string => {
  const typeMap: Record<string, string> = {
    http: 'primary',
    api: 'success',
    file: 'warning',
    data: 'info',
    code: 'danger',
    custom: '',
  }
  return typeMap[category] || ''
}

const getCategoryLabel = (category: string): string => {
  const labelMap: Record<string, string> = {
    http: 'HTTP 请求',
    api: 'API 调用',
    file: '文件操作',
    data: '数据处理',
    code: '代码执行',
    custom: '自定义',
  }
  return labelMap[category] || category
}
</script>

<style scoped>
.skill-manager {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: white;
}

.header-left h2 {
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 600;
}

.header-right :deep(.el-button-group) {
  display: flex;
}

.header-right :deep(.el-button) {
  background-color: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
  color: white;
}

.header-right :deep(.el-button:hover) {
  background-color: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.4);
}

/* 概览卡片 */
.overview-card {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.overview-stats {
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 16px 0;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #409eff;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

/* 主卡片 */
.main-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.main-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.main-card :deep(.el-tabs) {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.main-card :deep(.el-tabs__content) {
  flex: 1;
  overflow-y: auto;
}

.main-card :deep(.el-tab-pane) {
  height: 100%;
}

/* 标签页标签 */
.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 对话框 */
:deep(.el-dialog__body) {
  padding: 20px;
}

/* 响应式 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .overview-stats {
    flex-direction: column;
    gap: 16px;
  }

  .stat-value {
    font-size: 24px;
  }
}
</style>
