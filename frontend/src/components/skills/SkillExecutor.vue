<template>
  <div class="skill-executor-container">
    <!-- 技能选择区域 -->
    <el-card shadow="hover" class="selector-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><VideoPlay /></el-icon>
            技能执行器
          </span>
          <el-button
            v-if="!executing"
            type="primary"
            size="small"
            @click="handleRefreshSkills"
          >
            <el-icon><Refresh /></el-icon>
            刷新技能列表
          </el-button>
        </div>
      </template>

      <el-form label-width="100px" label-position="left">
        <el-form-item label="选择技能">
          <el-select
            v-model="selectedSkillId"
            placeholder="请选择要执行的技能"
            filterable
            style="width: 100%"
            :disabled="executing"
            @change="handleSkillChange"
          >
            <el-option
              v-for="skill in availableSkills"
              :key="skill.id"
              :label="skill.name"
              :value="skill.id"
            >
              <div class="skill-option">
                <span class="skill-name">{{ skill.name }}</span>
                <span class="skill-category">{{ skill.category }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item v-if="selectedSkill" label="技能描述">
          <el-text type="info">{{ selectedSkill.description }}</el-text>
        </el-form-item>

        <el-form-item label="工作空间">
          <el-input
            v-model="workspace"
            placeholder="/workspace"
            :disabled="executing"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 输入参数区域 -->
    <el-card
      v-if="selectedSkill"
      shadow="hover"
      class="input-card"
      :body-style="{ padding: '20px' }"
    >
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Edit /></el-icon>
            输入参数
          </span>
          <el-tag v-if="selectedSkill" type="info" size="small">
            {{ selectedSkill.category }}
          </el-tag>
        </div>
      </template>

      <el-form
        ref="inputFormRef"
        :model="inputForm"
        :rules="inputRules"
        label-width="120px"
        label-position="left"
      >
        <el-form-item label="输入数据" prop="inputData">
          <el-input
            v-model="inputForm.inputData"
            type="textarea"
            :rows="6"
            placeholder='请输入 JSON 格式的数据，例如：{"url": "https://api.example.com", "method": "GET"}'
            :disabled="executing"
          />
          <div class="form-tip">
            <el-text type="info" size="small">
              提示：输入将作为 JSON 解析，请确保格式正确
            </el-text>
          </div>
        </el-form-item>

        <!-- 动态参数（基于技能定义） -->
        <template v-if="selectedSkill && hasParameters">
          <el-divider content-position="left">额外参数</el-divider>
          <el-form-item
            v-for="param in dynamicParameters"
            :key="param.name"
            :label="param.label"
            :prop="param.name"
          >
            <el-input
              v-if="param.type === 'string'"
              v-model="inputForm.parameters[param.name]"
              :placeholder="param.placeholder"
              :disabled="executing"
            />
            <el-input-number
              v-else-if="param.type === 'number'"
              v-model="inputForm.parameters[param.name]"
              :placeholder="param.placeholder"
              :disabled="executing"
              style="width: 100%"
            />
            <el-switch
              v-else-if="param.type === 'boolean'"
              v-model="inputForm.parameters[param.name]"
              :disabled="executing"
            />
          </el-form-item>
        </template>

        <el-form-item>
          <el-button
            type="primary"
            :loading="executing"
            :disabled="!canExecute"
            @click="handleExecute"
            size="large"
            style="width: 100%"
          >
            <el-icon v-if="!executing"><VideoPlay /></el-icon>
            {{ executing ? '执行中...' : '执行技能' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 执行结果区域 -->
    <el-card
      v-if="executionResult || executionError"
      shadow="hover"
      class="result-card"
    >
      <template #header>
        <div class="card-header">
          <span>
            <el-icon>
              <component :is="executionResult ? 'CircleCheck' : 'CircleClose'" />
            </el-icon>
            执行结果
          </span>
          <div class="header-actions">
            <el-tag
              :type="executionResult ? 'success' : 'danger'"
              size="small"
            >
              {{ executionResult ? '成功' : '失败' }}
            </el-tag>
            <el-button
              v-if="executionResult"
              type="primary"
              size="small"
              @click="handleCopyResult"
            >
              <el-icon><DocumentCopy /></el-icon>
              复制结果
            </el-button>
          </div>
        </div>
      </template>

      <!-- 错误信息 -->
      <el-alert
        v-if="executionError"
        :title="executionError"
        type="error"
        :closable="false"
        show-icon
      />

      <!-- 成功结果 -->
      <div v-if="executionResult" class="result-content">
        <el-tabs v-model="activeResultTab">
          <el-tab-pane label="格式化视图" name="formatted">
            <div class="formatted-result">
              <pre>{{ formattedResult }}</pre>
            </div>
          </el-tab-pane>
          <el-tab-pane label="JSON 视图" name="json">
            <div class="json-result">
              <pre>{{ jsonResult }}</pre>
            </div>
          </el-tab-pane>
          <el-tab-pane label="原始输出" name="raw">
            <div class="raw-result">
              <pre>{{ rawResult }}</pre>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-card>

    <!-- 执行日志区域 -->
    <el-card
      v-if="executionLogs.length > 0"
      shadow="hover"
      class="logs-card"
    >
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><List /></el-icon>
            执行日志
          </span>
          <el-button size="small" @click="handleClearLogs">
            <el-icon><Delete /></el-icon>
            清空
          </el-button>
        </div>
      </template>

      <el-timeline>
        <el-timeline-item
          v-for="(log, index) in executionLogs"
          :key="index"
          :timestamp="log.timestamp"
          :type="log.type"
          placement="top"
        >
          <div class="log-content">
            <strong>{{ log.title }}</strong>
            <p v-if="log.message">{{ log.message }}</p>
          </div>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  VideoPlay,
  Refresh,
  Edit,
  CircleCheck,
  CircleClose,
  DocumentCopy,
  List,
  Delete,
} from '@element-plus/icons-vue'
import type { SkillListItem, ExecuteSkillInput, ExecuteSkillOutput } from '@/types/skill'

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
  (e: 'execute', input: ExecuteSkillInput): Promise<ExecuteSkillOutput>
  (e: 'refresh-skills'): void
}

const emit = defineEmits<Emits>()

// 状态
const selectedSkillId = ref('')
const workspace = ref('/workspace')
const executing = ref(false)
const executionResult = ref<any>(null)
const executionError = ref('')
const activeResultTab = ref('formatted')
const executionLogs = ref<Array<{ timestamp: string; type: string; title: string; message?: string }>>([])

// 表单
const inputFormRef = ref<FormInstance>()
const inputForm = ref({
  inputData: '',
  parameters: {} as Record<string, any>,
})

const inputRules: FormRules = {
  inputData: [
    { required: true, message: '请输入执行数据', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value) {
          callback()
          return
        }
        try {
          JSON.parse(value)
          callback()
        } catch (error) {
          callback(new Error('输入必须是有效的 JSON 格式'))
        }
      },
      trigger: 'blur',
    },
  ],
}

// 计算属性
const availableSkills = computed(() => {
  return props.skills.filter((s) => s.enabled)
})

const selectedSkill = computed(() => {
  return props.skills.find((s) => s.id === selectedSkillId.value)
})

const canExecute = computed(() => {
  return selectedSkill.value && inputForm.value.inputData && !executing.value
})

const dynamicParameters = computed(() => {
  // 这里可以根据技能定义动态生成参数
  // 目前返回示例参数
  return [
    {
      name: 'timeout',
      label: '超时时间(秒)',
      type: 'number',
      placeholder: '默认30秒',
      default: 30,
    },
    {
      name: 'retry',
      label: '失败重试',
      type: 'boolean',
      default: true,
    },
  ]
})

const hasParameters = computed(() => {
  return dynamicParameters.value.length > 0
})

const formattedResult = computed(() => {
  if (!executionResult.value) return ''
  if (typeof executionResult.value === 'string') {
    return executionResult.value
  }
  return JSON.stringify(executionResult.value, null, 2)
})

const jsonResult = computed(() => {
  if (!executionResult.value) return ''
  return JSON.stringify(executionResult.value, null, 2)
})

const rawResult = computed(() => {
  if (!executionResult.value) return ''
  return String(executionResult.value)
})

// 方法
const handleRefreshSkills = () => {
  emit('refresh-skills')
  ElMessage.info('正在刷新技能列表...')
}

const handleSkillChange = () => {
  // 重置表单
  inputForm.value.inputData = ''
  inputForm.value.parameters = {}
  executionResult.value = null
  executionError.value = ''
  addLog('info', '技能已选择', `选择了技能: ${selectedSkill.value?.name}`)
}

const handleExecute = async () => {
  if (!inputFormRef.value) return

  try {
    await inputFormRef.value.validate()
  } catch (error) {
    ElMessage.error('请检查输入数据')
    return
  }

  executing.value = true
  executionResult.value = null
  executionError.value = ''

  addLog('primary', '开始执行', `正在执行技能: ${selectedSkill.value?.name}`)

  try {
    const input: ExecuteSkillInput = {
      skillName: selectedSkill.value!.id,
      input: inputForm.value.inputData,
      workspace: workspace.value,
      parameters: inputForm.value.parameters,
    }

    const result = await emit('execute', input)

    if (result.success) {
      executionResult.value = result.result
      addLog('success', '执行成功', '技能执行完成')
      ElMessage.success('技能执行成功')
    } else {
      executionError.value = result.error || '执行失败'
      addLog('danger', '执行失败', result.error)
      ElMessage.error('技能执行失败')
    }
  } catch (error: any) {
    executionError.value = error.message || '未知错误'
    addLog('danger', '执行错误', error.message)
    ElMessage.error('技能执行出错')
  } finally {
    executing.value = false
  }
}

const handleCopyResult = () => {
  navigator.clipboard.writeText(formattedResult.value)
  ElMessage.success('结果已复制到剪贴板')
}

const handleClearLogs = () => {
  executionLogs.value = []
  ElMessage.info('日志已清空')
}

const addLog = (type: string, title: string, message?: string) => {
  const now = new Date()
  const timestamp = now.toLocaleTimeString()
  executionLogs.value.unshift({
    timestamp,
    type,
    title,
    message,
  })
}

// 初始化参数
watch(
  () => dynamicParameters.value,
  (params) => {
    params.forEach((param) => {
      if (!(param.name in inputForm.value.parameters)) {
        inputForm.value.parameters[param.name] = param.default
      }
    })
  },
  { immediate: true }
)

// 暴露方法
defineExpose({
  clearLogs: handleClearLogs,
  reset: () => {
    selectedSkillId.value = ''
    inputForm.value.inputData = ''
    inputForm.value.parameters = {}
    executionResult.value = null
    executionError.value = ''
    executionLogs.value = []
  },
})
</script>

<style scoped>
.skill-executor-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
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

.skill-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.skill-option .skill-name {
  font-weight: 500;
}

.skill-option .skill-category {
  font-size: 12px;
  color: #909399;
}

.form-tip {
  margin-top: 4px;
}

.result-content {
  max-height: 400px;
  overflow-y: auto;
}

.formatted-result pre,
.json-result pre,
.raw-result pre {
  margin: 0;
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.6;
  color: #303133;
  white-space: pre-wrap;
  word-break: break-word;
}

.log-content strong {
  display: block;
  margin-bottom: 4px;
}

.log-content p {
  margin: 0;
  color: #606266;
  font-size: 13px;
}

.el-form :deep(.el-form-item__label) {
  font-weight: 500;
}

.el-divider {
  margin: 20px 0;
}
</style>
