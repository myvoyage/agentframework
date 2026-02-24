<template>
  <div class="workflow-detail">
    <div class="detail-header">
      <el-page-header @back="goBack" :title="workflow?.name || '工作流详情'">
        <template #content>
          <div class="header-content">
            <span>{{ workflow?.name }}</span>
            <el-tag :type="getStatusType(workflow?.status)">
              {{ workflow?.status }}
            </el-tag>
          </div>
        </template>
        <template #extra>
          <el-button-group>
            <el-button :icon="Edit" @click="edit">编辑</el-button>
            <el-button :icon="VideoPlay" type="success" @click="execute">执行</el-button>
          </el-button-group>
        </template>
      </el-page-header>
    </div>

    <div class="detail-content">
      <el-card v-loading="loading" shadow="never">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ workflow?.id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ workflow?.type }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(workflow?.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatTime(workflow?.updatedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ workflow?.description }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card class="definition-card" shadow="never">
        <template #header>
          <span>工作流定义</span>
        </template>
        <pre class="json-preview">{{ formattedDefinition }}</pre>
      </el-card>

      <el-card class="versions-card" shadow="never">
        <template #header>
          <span>版本历史</span>
        </template>
        <el-timeline>
          <el-timeline-item
            v-for="version in versions"
            :key="version.id"
            :timestamp="formatTime(version.createdAt)"
          >
            版本 {{ version.version }}
          </el-timeline-item>
        </el-timeline>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Edit, VideoPlay } from '@element-plus/icons-vue'
import { useWorkflowStore } from '@/stores/workflowStore'

const router = useRouter()
const route = useRoute()
const workflowStore = useWorkflowStore()

const workflow = ref<any>(null)
const versions = ref<any[]>([])
const loading = ref(false)

const formattedDefinition = computed(() => {
  if (!workflow.value?.definition) return ''
  try {
    return JSON.stringify(JSON.parse(workflow.value.definition), null, 2)
  } catch {
    return workflow.value.definition
  }
})

const getStatusType = (status?: string) => {
  const types: Record<string, any> = {
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    pending: 'info'
  }
  return types[status || ''] || 'info'
}

const formatTime = (date?: Date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const goBack = () => {
  router.push('/workflow')
}

const edit = () => {
  router.push(`/workflow/builder?id=${route.params.id}`)
}

const execute = async () => {
  // TODO: 实现工作流执行
  ElMessage.info('工作流执行功能开发中')
}

onMounted(async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    workflow.value = await workflowStore.getWorkflow(id)
    // TODO: 加载版本历史
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.workflow-detail {
  padding: 20px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-content {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.definition-card,
.versions-card {
  margin-top: 20px;
}

.json-preview {
  background-color: var(--el-fill-color-light);
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
