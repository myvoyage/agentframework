<template>
  <div class="agent-studio">
    <div class="studio-header">
      <el-page-header title="Agent Studio">
        <template #extra>
          <el-button type="primary" :icon="Plus" @click="createAgent">
            新建 Agent
          </el-button>
        </template>
      </el-page-header>
    </div>

    <div class="studio-content">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card shadow="hover">
            <template #header>
              <span>Agent 列表</span>
            </template>
            <div class="agent-list">
              <div
                v-for="agent in agents"
                :key="agent.id"
                class="agent-item"
                :class="{ active: selectedAgentId === agent.id }"
                @click="selectAgent(agent)"
              >
                <div class="agent-icon">
                  <el-avatar :icon="User" />
                </div>
                <div class="agent-info">
                  <div class="agent-name">{{ agent.name }}</div>
                  <div class="agent-type">{{ agent.type }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :span="16">
          <el-card v-if="selectedAgent" shadow="hover">
            <template #header>
              <div class="card-header">
                <span>Agent 配置</span>
                <el-button-group>
                  <el-button :icon="ChatDotRound" @click="chatWithAgent">
                    对话
                  </el-button>
                  <el-button :icon="Delete" type="danger" @click="deleteAgent">
                    删除
                  </el-button>
                </el-button-group>
              </div>
            </template>
            <el-form label-width="100px">
              <el-form-item label="名称">
                <el-input v-model="selectedAgent.name" />
              </el-form-item>
              <el-form-item label="类型">
                <el-input v-model="selectedAgent.type" disabled />
              </el-form-item>
              <el-form-item label="系统提示词">
                <el-input
                  v-model="selectedAgent.systemPrompt"
                  type="textarea"
                  :rows="4"
                />
              </el-form-item>
              <el-form-item label="模型">
                <el-select v-model="selectedAgent.model">
                  <el-option label="ollama/llama3" value="ollama/llama3" />
                  <el-option label="gpt-4" value="gpt-4" />
                  <el-option label="claude-3" value="claude-3" />
                </el-select>
              </el-form-item>
              <el-form-item label="温度">
                <el-slider v-model="selectedAgent.temperature" :min="0" :max="1" :step="0.1" />
              </el-form-item>
            </el-form>
          </el-card>

          <el-card v-else shadow="hover">
            <el-empty description="选择一个 Agent 查看详情">
              <el-button type="primary" @click="createAgent">新建 Agent</el-button>
            </el-empty>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, User, ChatDotRound, Delete } from '@element-plus/icons-vue'
import { apiService } from '@/services/api'

const router = useRouter()

const agents = ref<Array<{
  id: string
  name: string
  type: string
  systemPrompt?: string
  model?: string
  temperature?: number
}>>([])

const selectedAgentId = ref<string>('')
const selectedAgent = ref<any>(null)

const selectAgent = (agent: any) => {
  selectedAgentId.value = agent.id
  selectedAgent.value = agent
}

const createAgent = () => {
  // TODO: 实现创建 Agent
  ElMessage.info('新建 Agent 功能开发中')
}

const chatWithAgent = () => {
  router.push(`/chat?agent=${selectedAgentId.value}`)
}

const deleteAgent = async () => {
  // TODO: 实现删除 Agent
  ElMessage.info('删除 Agent 功能开发中')
}

// 加载 agents
const loadAgents = async () => {
  const result = await apiService.listAgents()
  if (result.success && result.data) {
    agents.value = (result.data as string[]).map(id => ({
      id,
      name: id,
      type: 'agent'
    }))
  }
}

loadAgents()
</script>

<style scoped>
.agent-studio {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.agent-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.agent-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.agent-item:hover,
.agent-item.active {
  background-color: var(--el-fill-color-light);
}

.agent-icon {
  flex-shrink: 0;
}

.agent-info {
  flex: 1;
}

.agent-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.agent-type {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
