<template>
  <div class="chat-container">
    <!-- 顶部工具栏 -->
    <div class="chat-header">
      <div class="agent-selector">
        <el-select
          v-model="selectedAgentId"
          placeholder="选择 Agent"
          @change="onAgentChange"
          style="width: 200px"
        >
          <el-option
            v-for="agent in agents"
            :key="agent.id"
            :label="agent.name"
            :value="agent.id"
          >
            <div class="agent-option">
              <span>{{ agent.name }}</span>
              <span class="agent-type">{{ agent.type }}</span>
            </div>
          </el-option>
        </el-select>
      </div>
      <div class="header-actions">
        <el-button-group>
          <el-tooltip content="新建对话">
            <el-button :icon="Plus" @click="newConversation" />
          </el-tooltip>
          <el-tooltip content="导出对话">
            <el-button :icon="Download" @click="exportConversation" />
          </el-tooltip>
          <el-tooltip content="清空对话">
            <el-button :icon="Delete" @click="clearConversation" />
          </el-tooltip>
        </el-button-group>
      </div>
    </div>

    <!-- 对话历史侧边栏 -->
    <div v-if="showHistory" class="history-sidebar">
      <div class="history-header">
        <span>对话历史</span>
        <el-button text :icon="Close" @click="showHistory = false" />
      </div>
      <div class="history-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="history-item"
          :class="{ active: conv.id === currentConversationId }"
          @click="loadConversation(conv.id)"
        >
          <div class="history-title">{{ conv.title }}</div>
          <div class="history-time">{{ formatTime(conv.updatedAt) }}</div>
        </div>
      </div>
    </div>

    <!-- 消息区域 -->
    <div class="chat-messages" ref="messagesContainer">
      <div v-if="messages.length === 0" class="empty-state">
        <el-empty description="开始对话吧">
          <template #image>
            <el-icon :size="80" color="#909399">
              <ChatDotRound />
            </el-icon>
          </template>
          <el-button
            v-if="!selectedAgentId"
            type="primary"
            @click="showAgentSelector"
          >
            选择 Agent
          </el-button>
        </el-empty>
      </div>

      <div
        v-for="(message, index) in messages"
        :key="index"
        class="message"
        :class="`message-${message.role}`"
      >
        <div class="message-avatar">
          <el-avatar v-if="message.role === 'user'" :icon="User" />
          <el-avatar v-else :icon="Robot" />
        </div>
        <div class="message-content">
          <div class="message-header">
            <span class="message-role">
              {{ message.role === 'user' ? '你' : selectedAgent?.name || 'Assistant' }}
            </span>
            <span class="message-time">{{ formatTime(message.timestamp) }}</span>
          </div>
          <div class="message-text">
            <template v-if="message.role === 'assistant' && message.streaming">
              <span v-html="renderStreamingText(message.content)"></span>
              <span class="cursor">|</span>
            </template>
            <template v-else>
              <span v-html="renderMarkdown(message.content)"></span>
            </template>
          </div>
          <div v-if="message.metadata" class="message-metadata">
            <el-tag
              v-for="(value, key) in message.metadata"
              :key="key"
              size="small"
              type="info"
            >
              {{ key }}: {{ value }}
            </el-tag>
          </div>
        </div>
        <div class="message-actions">
          <el-button-group size="small">
            <el-tooltip content="复制">
              <el-button
                :icon="CopyDocument"
                @click="copyMessage(message.content)"
              />
            </el-tooltip>
            <el-tooltip content="重新生成">
              <el-button
                v-if="message.role === 'assistant'"
                :icon="Refresh"
                @click="regenerateMessage(index)"
              />
            </el-tooltip>
          </el-button-group>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="isLoading" class="message message-assistant">
        <div class="message-avatar">
          <el-avatar :icon="Robot" />
        </div>
        <div class="message-content">
          <div class="typing-indicator">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="chat-input-area">
      <div class="input-toolbar">
        <el-button-group size="small">
          <el-tooltip content="对话历史">
            <el-button
              :type="showHistory ? 'primary' : ''"
              @click="showHistory = !showHistory"
            >
              <el-icon><Clock /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="上传文件">
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              accept=".txt,.md,.json"
              @change="handleFileUpload"
            >
              <el-button>
                <el-icon><Folder /></el-icon>
              </el-button>
            </el-upload>
          </el-tooltip>
          <el-tooltip content="代码块">
            <el-button @click="insertCodeBlock">
              <el-icon><Document /></el-icon>
            </el-button>
          </el-tooltip>
        </el-button-group>
        <div class="input-stats">
          <span>{{ inputText.length }} / 4000</span>
        </div>
      </div>
      <div class="input-wrapper">
        <el-input
          v-model="inputText"
          type="textarea"
          :rows="3"
          :maxlength="4000"
          show-word-limit
          placeholder="输入消息... (Ctrl+Enter 发送)"
          @keydown.ctrl.enter="sendMessage"
          :disabled="!selectedAgentId || isLoading"
        />
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="isLoading"
          :disabled="!canSend"
          @click="sendMessage"
          class="send-btn"
        >
          发送
        </el-button>
      </div>
      <!-- 附件预览 -->
      <div v-if="attachments.length > 0" class="attachments">
        <el-tag
          v-for="(file, index) in attachments"
          :key="index"
          closable
          @close="removeAttachment(index)"
        >
          {{ file.name }}
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  User,
  Robot,
  Plus,
  Download,
  Delete,
  Close,
  ChatDotRound,
  CopyDocument,
  Refresh,
  Clock,
  Folder,
  Document,
  Promotion
} from '@element-plus/icons-vue'
import { apiService, handleApiError } from '@/services/api'
import { useAppStore } from '@/stores/appStore'
import { marked } from 'marked'

// 状态
const selectedAgentId = ref<string>('')
const agents = ref<Array<{ id: string; name: string; type: string }>>([])
const selectedAgent = computed(() => agents.value.find(a => a.id === selectedAgentId.value))

const messages = ref<Array<{
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  metadata?: Record<string, any>
  streaming?: boolean
}>>([])

const inputText = ref('')
const isLoading = ref(false)
const showHistory = ref(false)
const currentConversationId = ref<string>('')

const conversations = ref<Array<{
  id: string
  agentId: string
  title: string
  updatedAt: Date
}>>([])

const attachments = ref<Array<File>>([])
const messagesContainer = ref<HTMLElement>()

// 计算属性
const canSend = computed(() => {
  return selectedAgentId.value && inputText.value.trim().length > 0 && !isLoading.value
})

// 方法
const loadAgents = async () => {
  try {
    const result = await apiService.listAgents()
    if (result.success && result.data) {
      agents.value = (result.data as string[]).map(id => ({
        id,
        name: id,
        type: 'agent'
      }))
    }
  } catch (error) {
    handleApiError(error, '加载 Agents')
  }
}

const onAgentChange = () => {
  newConversation()
}

const newConversation = () => {
  messages.value = []
  currentConversationId.value = Date.now().toString()

  // 添加欢迎消息
  if (selectedAgent.value) {
    messages.value.push({
      role: 'system',
      content: `你正在与 ${selectedAgent.value.name} 对话。输入你的问题开始对话。`,
      timestamp: new Date()
    })
  }

  scrollToBottom()
}

const loadConversation = async (convId: string) => {
  // TODO: 从本地存储加载对话历史
  currentConversationId.value = convId
  showHistory.value = false
}

const sendMessage = async () => {
  if (!canSend.value) return

  const userMessage = inputText.value.trim()
  const userMsgObj = {
    role: 'user' as const,
    content: userMessage,
    timestamp: new Date()
  }

  messages.value.push(userMsgObj)
  inputText.value = ''
  scrollToBottom()

  isLoading.value = true

  try {
    // 创建流式响应消息
    const assistantMsg = {
      role: 'assistant' as const,
      content: '',
      timestamp: new Date(),
      streaming: true
    }
    messages.value.push(assistantMsg)

    // 调用 API
    const result = await apiService.chatWithAgent(selectedAgentId.value, userMessage)

    if (result.success && result.data) {
      assistantMsg.content = result.data as string
    } else {
      assistantMsg.content = `错误: ${result.error || '发送失败'}`
    }
  } catch (error) {
    messages.value.push({
      role: 'assistant',
      content: `抱歉，发生了错误: ${error}`,
      timestamp: new Date()
    })
  } finally {
    isLoading.value = false
    // 移除流式状态
    const lastMsg = messages.value[messages.value.length - 1]
    if (lastMsg && lastMsg.role === 'assistant') {
      lastMsg.streaming = false
    }
    scrollToBottom()
  }
}

const regenerateMessage = async (index: number) => {
  const msg = messages.value[index]
  if (msg.role !== 'assistant') return

  // 找到这条回复之前的用户消息
  let userMessage = ''
  for (let i = index - 1; i >= 0; i--) {
    if (messages.value[i].role === 'user') {
      userMessage = messages.value[i].content
      break
    }
  }

  if (userMessage) {
    // 删除当前及之后的消息
    messages.value = messages.value.slice(0, index)
    // 重新发送
    inputText.value = userMessage
    await sendMessage()
  }
}

const copyMessage = (content: string) => {
  navigator.clipboard.writeText(content)
  ElMessage.success('已复制到剪贴板')
}

const clearConversation = async () => {
  try {
    await ElMessageBox.confirm('确定要清空当前对话吗？', '确认', {
      type: 'warning'
    })
    newConversation()
    ElMessage.success('对话已清空')
  } catch {
    // 用户取消
  }
}

const exportConversation = () => {
  const content = messages.value.map(m => {
    const role = m.role === 'user' ? '你' : selectedAgent.value?.name || 'Assistant'
    const time = formatTime(m.timestamp)
    return `[${time}] ${role}: ${m.content}`
  }).join('\n\n')

  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `conversation-${currentConversationId.value}.txt`
  a.click()
  URL.revokeObjectURL(url)

  ElMessage.success('对话已导出')
}

const handleFileUpload = (file: any) => {
  if (file.raw) {
    attachments.value.push(file.raw)
  }
}

const removeAttachment = (index: number) => {
  attachments.value.splice(index, 1)
}

const insertCodeBlock = () => {
  inputText.value += '```\n// 在这里输入代码\n```'
}

const showAgentSelector = () => {
  ElMessage.info('请先选择一个 Agent')
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const formatTime = (date: Date) => {
  return new Date(date).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const renderMarkdown = (content: string) => {
  try {
    return marked(content)
  } catch {
    return content
  }
}

const renderStreamingText = (content: string) => {
  return renderMarkdown(content)
}

// 生命周期
onMounted(async () => {
  await loadAgents()

  // 尝试选择最后一个使用的 agent
  const lastAgent = localStorage.getItem('last-selected-agent')
  if (lastAgent && agents.value.find(a => a.id === lastAgent)) {
    selectedAgentId.value = lastAgent
    newConversation()
  }
})

// 监听 agent 选择变化，保存到本地存储
const unwatchAgent = computed(() => selectedAgentId.value)
// 简化处理，实际应使用 watch
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background-color: var(--el-bg-color);
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color);
}

.agent-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.agent-type {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.history-sidebar {
  position: absolute;
  left: 0;
  top: 60px;
  bottom: 200px;
  width: 280px;
  background-color: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color);
  z-index: 10;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color);
  font-weight: 500;
}

.history-list {
  overflow-y: auto;
  height: calc(100% - 57px);
}

.history-item {
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-left: 3px solid transparent;
}

.history-item:hover {
  background-color: var(--el-fill-color-light);
}

.history-item.active {
  background-color: var(--el-color-primary-light-9);
  border-left-color: var(--el-color-primary);
}

.history-title {
  font-size: 14px;
  margin-bottom: 4px;
}

.history-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.message {
  display: flex;
  gap: 12px;
  max-width: 80%;
}

.message-user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message-assistant {
  align-self: flex-start;
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  flex: 1;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.message-role {
  font-weight: 500;
  font-size: 13px;
}

.message-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.message-text {
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}

.message-text :deep(pre) {
  background-color: var(--el-fill-color-light);
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

.message-text :deep(code) {
  background-color: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.message-text :deep(p) {
  margin: 8px 0;
}

.message-text :deep ul), .message-text :deep(ol {
  margin: 8px 0;
  padding-left: 24px;
}

.cursor {
  display: inline-block;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.message-metadata {
  margin-top: 8px;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.message-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.message:hover .message-actions {
  opacity: 1;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--el-text-color-secondary);
  animation: typing 1.4s infinite;
}

.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-indicator span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-10px); }
}

.chat-input-area {
  border-top: 1px solid var(--el-border-color);
  padding: 16px;
  background-color: var(--el-bg-color);
}

.input-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.input-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-wrapper .el-textarea {
  flex: 1;
}

.send-btn {
  height: 74px;
}

.attachments {
  margin-top: 8px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
