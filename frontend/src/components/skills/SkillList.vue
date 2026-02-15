<template>
  <div class="skill-list-container">
    <!-- 搜索和筛选栏 -->
    <div class="filter-bar">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索技能名称或描述..."
        clearable
        style="width: 300px"
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <el-select
        v-model="selectedCategory"
        placeholder="选择分类"
        clearable
        style="width: 180px; margin-left: 10px"
        @change="handleFilter"
      >
        <el-option label="全部分类" value=""></el-option>
        <el-option
          v-for="cat in categories"
          :key="cat.value"
          :label="cat.label"
          :value="cat.value"
        />
      </el-select>

      <div class="stats-info">
        <el-tag type="info">总计: {{ filteredSkills.length }} 个技能</el-tag>
        <el-tag type="success" style="margin-left: 8px">
          启用: {{ enabledCount }} 个
        </el-tag>
      </div>
    </div>

    <!-- 技能表格 -->
    <el-table
      v-loading="loading"
      :data="paginatedSkills"
      style="width: 100%; margin-top: 16px"
      :row-class-name="getRowClassName"
      @row-click="handleRowClick"
      stripe
      highlight-current-row
    >
      <el-table-column prop="name" label="技能名称" width="200" fixed>
        <template #default="{ row }">
          <div class="skill-name-cell">
            <el-icon
              :color="row.enabled ? '#67c23a' : '#909399'"
              :size="18"
            >
              <component :is="row.enabled ? 'CircleCheck' : 'CircleClose'" />
            </el-icon>
            <div class="name-content">
              <div class="name-text">{{ row.name }}</div>
              <div class="name-version">v{{ row.version }}</div>
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />

      <el-table-column prop="category" label="分类" width="120">
        <template #default="{ row }">
          <el-tag :type="getCategoryTagType(row.category)" size="small">
            {{ getCategoryLabel(row.category) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="tags" label="标签" width="180">
        <template #default="{ row }">
          <el-tag
            v-for="tag in row.tags.slice(0, 2)"
            :key="tag"
            size="small"
            style="margin-right: 4px"
          >
            {{ tag }}
          </el-tag>
          <el-tag v-if="row.tags.length > 2" size="small" type="info">
            +{{ row.tags.length - 2 }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="useCount" label="使用统计" width="140" align="center">
        <template #default="{ row }">
          <div class="usage-stats">
            <div class="count">
              <el-icon><DataAnalysis /></el-icon>
              <span>{{ row.useCount }} 次</span>
            </div>
            <div v-if="row.lastUsed" class="last-used">
              {{ formatLastUsed(row.lastUsed) }}
            </div>
            <div v-else class="never-used">未使用</div>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="220" fixed="right" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-tooltip content="查看详情" placement="top">
              <el-button
                type="primary"
                size="small"
                :icon="InfoFilled"
                @click.stop="handleViewDetail(row)"
              />
            </el-tooltip>
            <el-tooltip content="执行技能" placement="top">
              <el-button
                type="success"
                size="small"
                :icon="VideoPlay"
                @click.stop="handleExecute(row)"
              />
            </el-tooltip>
            <el-tooltip :content="row.enabled ? '禁用技能' : '启用技能'" placement="top">
              <el-button
                :type="row.enabled ? 'warning' : 'success'"
                size="small"
                @click.stop="handleToggle(row)"
              >
                {{ row.enabled ? '禁用' : '启用' }}
              </el-button>
            </el-tooltip>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="filteredSkills.length"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Search,
  CircleCheck,
  CircleClose,
  InfoFilled,
  VideoPlay,
  DataAnalysis,
} from '@element-plus/icons-vue'
import type { SkillListItem } from '@/types/skill'

// Props
interface Props {
  skills?: SkillListItem[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  skills: () => [],
  loading: false,
})

// Emits
interface Emits {
  (e: 'refresh'): void
  (e: 'view-detail', skill: SkillListItem): void
  (e: 'execute', skill: SkillListItem): void
  (e: 'toggle', skill: SkillListItem): void
}

const emit = defineEmits<Emits>()

// 状态
const searchKeyword = ref('')
const selectedCategory = ref('')
const currentPage = ref(1)
const pageSize = ref(20)

// 分类选项
const categories = [
  { label: 'HTTP 请求', value: 'http' },
  { label: 'API 调用', value: 'api' },
  { label: '文件操作', value: 'file' },
  { label: '数据处理', value: 'data' },
  { label: '代码执行', value: 'code' },
  { label: '自定义', value: 'custom' },
]

// 计算属性
const filteredSkills = computed(() => {
  let result = [...props.skills]

  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(
      (skill) =>
        skill.name.toLowerCase().includes(keyword) ||
        skill.description.toLowerCase().includes(keyword) ||
        skill.tags.some((tag) => tag.toLowerCase().includes(keyword))
    )
  }

  // 分类筛选
  if (selectedCategory.value) {
    result = result.filter((skill) => skill.category === selectedCategory.value)
  }

  return result
})

const paginatedSkills = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredSkills.value.slice(start, end)
})

const enabledCount = computed(() => {
  return props.skills.filter((s) => s.enabled).length
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

const handleRowClick = (row: SkillListItem) => {
  emit('view-detail', row)
}

const handleViewDetail = (skill: SkillListItem) => {
  emit('view-detail', skill)
}

const handleExecute = (skill: SkillListItem) => {
  emit('execute', skill)
}

const handleToggle = async (skill: SkillListItem) => {
  emit('toggle', skill)
}

const getRowClassName = ({ row }: { row: SkillListItem }) => {
  return row.enabled ? '' : 'disabled-row'
}

const getCategoryTagType = (category: string) => {
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

const getCategoryLabel = (category: string) => {
  const cat = categories.find((c) => c.value === category)
  return cat?.label || category
}

const formatLastUsed = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return '刚刚'
  if (diffMins < 60) return `${diffMins} 分钟前`
  if (diffHours < 24) return `${diffHours} 小时前`
  if (diffDays < 7) return `${diffDays} 天前`
  return date.toLocaleDateString()
}

// 暴露方法供父组件调用
defineExpose({
  refresh: () => emit('refresh'),
})
</script>

<style scoped>
.skill-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.stats-info {
  margin-left: auto;
  display: flex;
  align-items: center;
}

.skill-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.name-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.name-text {
  font-weight: 500;
  color: #303133;
}

.name-version {
  font-size: 12px;
  color: #909399;
}

.usage-stats {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
}

.usage-stats .count {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: #606266;
}

.usage-stats .last-used {
  color: #909399;
  font-size: 11px;
}

.usage-stats .never-used {
  color: #c0c4cc;
  font-size: 11px;
}

.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.disabled-row {
  opacity: 0.6;
  background-color: #f5f7fa;
}

.el-table {
  flex: 1;
}

.el-table :deep(.el-table__body-wrapper) {
  max-height: calc(100vh - 400px);
  overflow-y: auto;
}
</style>
