<template>
  <div class="performance-monitor-container">
    <!-- 统计概览卡片 -->
    <div class="stats-cards">
      <el-card
        v-for="card in statsCards"
        :key="card.key"
        shadow="hover"
        class="stat-card"
        :class="card.colorClass"
      >
        <div class="stat-content">
          <div class="stat-icon" :style="{ backgroundColor: card.iconBg }">
            <el-icon :size="24" :color="card.iconColor">
              <component :is="card.icon" />
            </el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">
              <span class="value">{{ card.value }}</span>
              <span v-if="card.unit" class="unit">{{ card.unit }}</span>
            </div>
            <div class="stat-label">{{ card.label }}</div>
            <div v-if="card.trend" class="stat-trend" :class="card.trendClass">
              <el-icon :size="14">
                <component :is="card.trendIcon" />
              </el-icon>
              <span>{{ card.trendText }}</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 详细数据表格 -->
    <el-card shadow="hover" class="detail-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><DataAnalysis /></el-icon>
            详细统计数据
          </span>
          <div class="header-actions">
            <el-switch
              v-model="autoRefresh"
              active-text="自动刷新"
              inactive-text="手动刷新"
              @change="handleAutoRefreshChange"
            />
            <el-button
              type="primary"
              size="small"
              @click="handleRefresh"
              :loading="refreshing"
            >
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- 缓存统计 -->
        <el-tab-pane label="缓存统计" name="cache">
          <div v-if="cacheStats" class="stats-table">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="缓存命中率">
                <div class="stat-highlight">
                  <el-progress
                    :percentage="cacheStats.hitRate * 100"
                    :color="getHitRateColor(cacheStats.hitRate)"
                    :show-text="true"
                    :format="(percentage: number) => `${percentage.toFixed(1)}%`"
                  />
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="缓存命中次数">
                <el-tag type="success" size="large">{{ cacheStats.hits }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="缓存未命中次数">
                <el-tag type="warning" size="large">{{ cacheStats.misses }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="总条目数">
                <el-tag type="info" size="large">{{ cacheStats.totalEntries }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="缓存大小">
                <el-tag>{{ formatBytes(cacheStats.size) }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- 缓存性能图表 -->
            <div class="chart-container">
              <div class="chart-title">缓存性能趋势</div>
              <div class="chart-placeholder">
                <el-empty
                  description="图表功能待实现（可集成 ECharts 或其他图表库）"
                  :image-size="120"
                />
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无缓存统计数据" />
        </el-tab-pane>

        <!-- 连接池统计 -->
        <el-tab-pane label="连接池统计" name="pool">
          <div v-if="poolStats" class="stats-table">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="活跃连接数">
                <el-tag type="success" size="large">
                  {{ poolStats.activeConnections }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="空闲连接数">
                <el-tag type="info" size="large">
                  {{ poolStats.idleConnections }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="最大连接数">
                <el-tag type="warning" size="large">
                  {{ poolStats.maxConnections }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="最小连接数">
                <el-tag size="large">{{ poolStats.minConnections }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="利用率">
                <div class="stat-highlight">
                  <el-progress
                    :percentage="poolStats.utilizationRate * 100"
                    :color="getUtilizationColor(poolStats.utilizationRate)"
                    :show-text="true"
                    :format="(percentage: number) => `${percentage.toFixed(1)}%`"
                  />
                </div>
              </el-descriptions-item>
            </el-descriptions>

            <!-- 连接池可视化 -->
            <div class="pool-visualization">
              <div class="chart-title">连接池状态</div>
              <div class="pool-bars">
                <div class="pool-bar">
                  <div class="bar-label">活跃连接</div>
                  <el-progress
                    :percentage="(poolStats.activeConnections / poolStats.maxConnections) * 100"
                    :show-text="false"
                    color="#67c23a"
                  />
                  <div class="bar-value">
                    {{ poolStats.activeConnections }} / {{ poolStats.maxConnections }}
                  </div>
                </div>
                <div class="pool-bar">
                  <div class="bar-label">空闲连接</div>
                  <el-progress
                    :percentage="(poolStats.idleConnections / poolStats.maxConnections) * 100"
                    :show-text="false"
                    color="#409eff"
                  />
                  <div class="bar-value">
                    {{ poolStats.idleConnections }} / {{ poolStats.maxConnections }}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无连接池统计数据" />
        </el-tab-pane>

        <!-- 技能系统统计 -->
        <el-tab-pane label="技能系统统计" name="system">
          <div v-if="systemStats" class="stats-table">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="总技能数">
                <el-tag type="primary" size="large">
                  {{ systemStats.totalSkills }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="总执行次数">
                <el-tag type="success" size="large">
                  {{ systemStats.totalUses }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- 分类统计 -->
            <div v-if="Object.keys(systemStats.categories).length > 0" class="categories-section">
              <div class="chart-title">分类统计</div>
              <div class="categories-grid">
                <div
                  v-for="(count, category) in systemStats.categories"
                  :key="category"
                  class="category-item"
                >
                  <el-tag :type="getCategoryTagType(category)" size="large">
                    {{ getCategoryLabel(category) }}: {{ count }}
                  </el-tag>
                </div>
              </div>
            </div>

            <!-- 最常用技能 -->
            <div v-if="systemStats.mostUsedSkills.length > 0" class="top-skills-section">
              <div class="chart-title">最常用技能 TOP 5</div>
              <el-table :data="systemStats.mostUsedSkills" style="width: 100%">
                <el-table-column type="index" label="排名" width="80" />
                <el-table-column prop="name" label="技能名称" />
                <el-table-column prop="count" label="使用次数" width="120" align="center">
                  <template #default="{ row }">
                    <el-tag type="primary">{{ row.count }}</el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
          <el-empty v-else description="暂无技能系统统计数据" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 性能建议 -->
    <el-card v-if="hasRecommendations" shadow="hover" class="recommendations-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Notification /></el-icon>
            性能优化建议
          </span>
        </div>
      </template>
      <div class="recommendations">
        <el-alert
          v-for="(rec, index) in recommendations"
          :key="index"
          :title="rec.title"
          :type="rec.type"
          :description="rec.description"
          show-icon
          :closable="false"
          style="margin-bottom: 12px"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  DataAnalysis,
  Refresh,
  TrendCharts,
  Notification,
  CircleCheck,
  Warning,
  ArrowUp,
  ArrowDown,
  Minus,
} from '@element-plus/icons-vue'
import type { CacheStats, PoolStats, SkillSystemStats } from '@/types/skill'

// StatCard类型定义
interface StatCard {
  key: string
  label: string
  value: string
  unit: string
  icon: string
  iconBg: string
  iconColor: string
  colorClass: string
  trend: boolean
  trendValue?: number
  trendClass?: string
  trendIcon?: string
  trendText?: string
}

// Props
interface Props {
  cacheStats?: CacheStats
  poolStats?: PoolStats
  systemStats?: SkillSystemStats
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
})

// Emits
interface Emits {
  (e: 'refresh'): void
}

const emit = defineEmits<Emits>()

// 状态
const activeTab = ref('cache')
const autoRefresh = ref(true)
const refreshing = ref(false)
let refreshTimer: number | null = null

// 计算属性 - 统计卡片
const statsCards = computed((): StatCard[] => {
  const cards: StatCard[] = [
    {
      key: 'cache',
      label: '缓存命中率',
      value: props.cacheStats ? (props.cacheStats.hitRate * 100).toFixed(1) : '0.0',
      unit: '%',
      icon: 'TrendCharts',
      iconBg: '#ecf5ff',
      iconColor: '#409eff',
      colorClass: 'blue-card',
      trend: true,
      trendValue: props.cacheStats?.hitRate || 0,
      trendClass: '',
      trendIcon: '',
      trendText: '',
    },
    {
      key: 'pool',
      label: '连接池利用率',
      value: props.poolStats ? (props.poolStats.utilizationRate * 100).toFixed(1) : '0.0',
      unit: '%',
      icon: 'Connection',
      iconBg: '#f0f9ff',
      iconColor: '#67c23a',
      colorClass: 'green-card',
      trend: true,
      trendValue: props.poolStats?.utilizationRate || 0,
      trendClass: '',
      trendIcon: '',
      trendText: '',
    },
    {
      key: 'executions',
      label: '总执行次数',
      value: props.systemStats?.totalUses?.toString() || '0',
      unit: '',
      icon: 'DataAnalysis',
      iconBg: '#fdf6ec',
      iconColor: '#e6a23c',
      colorClass: 'orange-card',
      trend: false,
      trendClass: '',
      trendIcon: '',
      trendText: '',
    },
    {
      key: 'skills',
      label: '已注册技能',
      value: props.systemStats?.totalSkills?.toString() || '0',
      unit: '',
      icon: 'CircleCheck',
      iconBg: '#f0f9ff',
      iconColor: '#67c23a',
      colorClass: 'green-card',
      trend: false,
      trendClass: '',
      trendIcon: '',
      trendText: '',
    },
  ]

  // 添加趋势信息
  cards.forEach((card) => {
    if (card.trend && card.trendValue !== undefined) {
      if (card.trendValue > 0.7) {
        card.trendClass = 'trend-up'
        card.trendIcon = 'ArrowUp'
        card.trendText = '优秀'
      } else if (card.trendValue > 0.4) {
        card.trendClass = 'trend-stable'
        card.trendIcon = 'Minus'
        card.trendText = '正常'
      } else {
        card.trendClass = 'trend-down'
        card.trendIcon = 'ArrowDown'
        card.trendText = '需优化'
      }
    }
  })

  return cards
})

// 计算属性 - 性能建议
const recommendations = computed(() => {
  const recs: Array<{ type: string; title: string; description: string }> = []

  // 缓存建议
  if (props.cacheStats) {
    if (props.cacheStats.hitRate < 0.5) {
      recs.push({
        type: 'warning',
        title: '缓存命中率较低',
        description: '当前缓存命中率为 ${(props.cacheStats.hitRate * 100).toFixed(1)}%，建议检查缓存配置或增加缓存容量。',
      })
    }
  }

  // 连接池建议
  if (props.poolStats) {
    if (props.poolStats.utilizationRate > 0.9) {
      recs.push({
        type: 'warning',
        title: '连接池利用率过高',
        description: `当前利用率为 ${(props.poolStats.utilizationRate * 100).toFixed(1)}%，建议增加最大连接数以避免性能瓶颈。`,
      })
    }
    if (props.poolStats.activeConnections === props.poolStats.maxConnections) {
      recs.push({
        type: 'error',
        title: '连接池已满',
        description: '所有连接都在使用中，可能存在连接泄漏或连接未正确释放的问题。',
      })
    }
  }

  return recs
})

const hasRecommendations = computed(() => {
  return recommendations.value.length > 0
})

// 方法
const handleRefresh = () => {
  refreshing.value = true
  emit('refresh')
  setTimeout(() => {
    refreshing.value = false
  }, 1000)
}

const handleAutoRefreshChange = (enabled: boolean) => {
  if (enabled) {
    startAutoRefresh()
    ElMessage.success('已开启自动刷新（每5秒）')
  } else {
    stopAutoRefresh()
    ElMessage.info('已关闭自动刷新')
  }
}

const startAutoRefresh = () => {
  refreshTimer = window.setInterval(() => {
    emit('refresh')
  }, 5000)
}

const stopAutoRefresh = () => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

const getHitRateColor = (rate: number): string => {
  if (rate >= 0.8) return '#67c23a'
  if (rate >= 0.5) return '#e6a23c'
  return '#f56c6c'
}

const getUtilizationColor = (rate: number): string => {
  if (rate >= 0.9) return '#f56c6c'
  if (rate >= 0.7) return '#e6a23c'
  return '#67c23a'
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

// 生命周期
onMounted(() => {
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.performance-monitor-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 统计卡片 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.stat-card {
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-info {
  flex: 1;
}

.stat-value {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin-bottom: 4px;
}

.stat-value .value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-value .unit {
  font-size: 14px;
  color: #909399;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 4px;
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
}

.stat-trend.trend-up {
  color: #67c23a;
}

.stat-trend.trend-stable {
  color: #909399;
}

.stat-trend.trend-down {
  color: #f56c6c;
}

/* 详细卡片 */
.detail-card {
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stats-table {
  padding: 16px 0;
}

.stat-highlight {
  width: 100%;
}

/* 图表容器 */
.chart-container {
  margin-top: 24px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 8px;
}

.chart-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.chart-placeholder {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 连接池可视化 */
.pool-visualization {
  margin-top: 24px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 8px;
}

.pool-bars {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pool-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pool-bar .bar-label {
  width: 100px;
  font-size: 14px;
  color: #606266;
}

.pool-bar .bar-value {
  width: 100px;
  font-size: 14px;
  color: #909399;
  text-align: right;
}

/* 分类统计 */
.categories-section {
  margin-top: 24px;
}

.categories-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
}

.category-item {
  display: flex;
}

/* 最常用技能 */
.top-skills-section {
  margin-top: 24px;
}

/* 性能建议 */
.recommendations-card {
  border-left: 4px solid #e6a23c;
}

.recommendations {
  display: flex;
  flex-direction: column;
}

/* 卡片颜色主题 */
.blue-card :deep(.el-card__body) {
  border-left: 4px solid #409eff;
}

.green-card :deep(.el-card__body) {
  border-left: 4px solid #67c23a;
}

.orange-card :deep(.el-card__body) {
  border-left: 4px solid #e6a23c;
}

/* 响应式 */
@media (max-width: 768px) {
  .stats-cards {
    grid-template-columns: 1fr;
  }
}
</style>
