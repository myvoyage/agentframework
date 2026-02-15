<template>
  <div class="definition-viewer-container">
    <!-- 技能选择器 -->
    <el-card shadow="hover" class="selector-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Document /></el-icon>
            技能定义查看器
          </span>
        </div>
      </template>

      <el-select
        v-model="selectedDefinitionId"
        placeholder="选择要查看的技能定义"
        filterable
        style="width: 100%"
        @change="handleDefinitionChange"
      >
        <el-option
          v-for="def in definitions"
          :key="def.id"
          :label="def.name"
          :value="def.id"
        >
          <div class="definition-option">
            <span class="def-name">{{ def.name }}</span>
            <el-tag size="small" type="info">{{ def.category }}</el-tag>
          </div>
        </el-option>
      </el-select>
    </el-card>

    <!-- 定义详情 -->
    <el-card v-if="currentDefinition" shadow="hover" class="detail-card">
      <template #header>
        <div class="card-header">
          <span>{{ currentDefinition.name }}</span>
          <div class="header-actions">
            <el-tag type="info" size="small">v{{ currentDefinition.version }}</el-tag>
            <el-button size="small" @click="handleCopyYAML">
              <el-icon><DocumentCopy /></el-icon>
              复制 YAML
            </el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- 概览 -->
        <el-tab-pane label="概览" name="overview">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="技能ID">
              {{ currentDefinition.id }}
            </el-descriptions-item>
            <el-descriptions-item label="名称">
              {{ currentDefinition.name }}
            </el-descriptions-item>
            <el-descriptions-item label="分类" :span="2">
              <el-tag :type="getCategoryTagType(currentDefinition.category)">
                {{ getCategoryLabel(currentDefinition.category) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">
              {{ currentDefinition.description }}
            </el-descriptions-item>
            <el-descriptions-item label="作者">
              {{ currentDefinition.author || '未知' }}
            </el-descriptions-item>
            <el-descriptions-item label="许可证">
              <el-tag size="small">{{ currentDefinition.license }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>

          <!-- 配置信息 -->
          <div class="config-section">
            <h4>配置参数</h4>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="缓存启用">
                <el-tag :type="currentDefinition.config.cache_enabled ? 'success' : 'info'">
                  {{ currentDefinition.config.cache_enabled ? '是' : '否' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="缓存TTL">
                {{ currentDefinition.config.cache_ttl }}
              </el-descriptions-item>
              <el-descriptions-item label="默认超时">
                {{ currentDefinition.config.default_timeout }}
              </el-descriptions-item>
              <el-descriptions-item label="最大重试次数">
                {{ currentDefinition.config.max_retries }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </el-tab-pane>

        <!-- 工作流可视化 -->
        <el-tab-pane label="工作流可视化" name="workflow">
          <div class="workflow-container">
            <div class="workflow-graph" ref="graphRef">
              <div class="workflow-nodes">
                <div
                  v-for="(step, stepId) in currentDefinition.workflow"
                  :key="stepId"
                  class="workflow-node"
                  :class="getStepNodeClass(step)"
                >
                  <div class="node-header">
                    <el-icon>
                      <component :is="getStepIcon(step.action)" />
                    </el-icon>
                    <span class="node-name">{{ step.name }}</span>
                  </div>
                  <div class="node-body">
                    <div class="node-action">
                      <el-tag :type="getActionTagType(step.action)" size="small">
                        {{ step.action }}
                      </el-tag>
                    </div>
                    <div v-if="step.timeout" class="node-timeout">
                      <el-text type="info" size="small">
                        <el-icon><Clock /></el-icon>
                        {{ step.timeout }}
                      </el-text>
                    </div>
                    <div v-if="step.skip_if" class="node-skip">
                      <el-text type="warning" size="small">
                        <el-icon><Warning /></el-icon>
                        条件跳过
                      </el-text>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 工作流步骤列表 -->
            <div class="workflow-steps-list">
              <h4>工作流步骤</h4>
              <el-timeline>
                <el-timeline-item
                  v-for="(step, stepId, index) in currentDefinition.workflow"
                  :key="stepId"
                  :timestamp="`步骤 ${index + 1}`"
                  placement="top"
                >
                  <el-card>
                    <div class="step-detail">
                      <div class="step-header">
                        <strong>{{ step.name }}</strong>
                        <el-tag :type="getActionTagType(step.action)" size="small">
                          {{ step.action }}
                        </el-tag>
                      </div>
                      <div class="step-info">
                        <div v-if="step.timeout" class="info-item">
                          <el-icon><Clock /></el-icon>
                          <span>超时: {{ step.timeout }}</span>
                        </div>
                        <div v-if="step.skip_if" class="info-item">
                          <el-icon><Warning /></el-icon>
                          <span>跳过条件: {{ step.skip_if }}</span>
                        </div>
                      </div>
                    </div>
                  </el-card>
                </el-timeline-item>
              </el-timeline>
            </div>
          </div>
        </el-tab-pane>

        <!-- YAML 代码 -->
        <el-tab-pane label="YAML 代码" name="yaml">
          <div class="yaml-container">
            <pre class="yaml-code">{{ yamlCode }}</pre>
          </div>
        </el-tab-pane>

        <!-- 示例 -->
        <el-tab-pane label="使用示例" name="example">
          <div class="example-container">
            <el-alert
              title="执行示例"
              type="info"
              :closable="false"
              style="margin-bottom: 16px"
            >
              <template #default>
                以下是如何调用此技能的示例代码和输入参数
              </template>
            </el-alert>

            <div class="example-section">
              <h4>Go 代码示例</h4>
              <pre class="code-block">{{ goExampleCode }}</pre>
            </div>

            <div class="example-section">
              <h4>输入参数示例 (JSON)</h4>
              <pre class="code-block">{{ jsonExample }}</pre>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 空状态 -->
    <el-card v-else shadow="hover" class="empty-card">
      <el-empty description="请选择一个技能定义查看详情">
        <el-icon :size="64" color="#909399">
          <Document />
        </el-icon>
      </el-empty>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document,
  DocumentCopy,
  Clock,
  Warning,
  VideoPlay,
  Setting,
  Check,
  DocumentChecked,
  MagicStick,
  Delete,
  Tools,
} from '@element-plus/icons-vue'
import type { SkillDefinitionInfo } from '@/types/skill'

// Props
interface Props {
  definitions?: SkillDefinitionInfo[]
}

const props = withDefaults(defineProps<Props>(), {
  definitions: () => [],
})

// 状态
const selectedDefinitionId = ref('')
const activeTab = ref('overview')
const graphRef = ref<HTMLElement>()

// 计算属性
const currentDefinition = computed(() => {
  return props.definitions.find((d) => d.id === selectedDefinitionId.value)
})

const yamlCode = computed(() => {
  if (!currentDefinition.value) return ''

  const def = currentDefinition.value
  return `id: "${def.id}"
name: "${def.name}"
description: "${def.description}"
version: "${def.version}"
category: "${def.category}"
author: "${def.author || 'Unknown'}"
license: "${def.license}"

config:
  cache_enabled: ${def.config.cache_enabled}
  cache_ttl: ${def.config.cache_ttl}
  default_timeout: ${def.config.default_timeout}
  max_retries: ${def.config.max_retries}

workflow:
${Object.entries(def.workflow).map(([id, step]) => `  - id: "${id}"
    name: "${step.name}"
    action: "${step.action}"${step.timeout ? `
    timeout: ${step.timeout}` : ''}${step.skip_if ? `
    skip_if: "${step.skip_if}"` : ''}`).join('\n')}
`
})

const goExampleCode = computed(() => {
  if (!currentDefinition.value) return ''

  const def = currentDefinition.value
  return `package main

import (
    "context"
    "fmt"
    "agentframework/agent/skills"
)

func main() {
    ctx := context.Background()

    // 创建执行上下文
    execCtx := skills.NewExecutionContext()
    execCtx.Workspace = "/workspace"

    // 准备输入参数
    input := \`{
        "param1": "value1",
        "param2": "value2"
    }\`

    // 执行技能
    result, err := skillSystem.ExecuteSkill(
        ctx,
        "${def.id}",
        input,
        execCtx,
    )

    if err != nil {
        fmt.Printf("技能执行失败: %v\\n", err)
        return
    }

    fmt.Printf("执行结果: %v\\n", result)
}`
})

const jsonExample = computed(() => {
  return `{
  "param1": "value1",
  "param2": "value2",
  "options": {
    "timeout": 30000,
    "retry": true
  }
}`
})

// 方法
const handleDefinitionChange = () => {
  activeTab.value = 'overview'
}

const handleCopyYAML = () => {
  navigator.clipboard.writeText(yamlCode.value)
  ElMessage.success('YAML 代码已复制到剪贴板')
}

const getStepNodeClass = (step: any) => {
  return `step-${step.action}`
}

const getStepIcon = (action: string) => {
  const iconMap: Record<string, string> = {
    validate: 'DocumentChecked',
    prepare: 'Setting',
    execute: 'VideoPlay',
    check_exists: 'Check',
    generate_code: 'MagicStick',
    cleanup: 'Delete',
  }
  return iconMap[action] || 'Tools'
}

const getActionTagType = (action: string) => {
  const typeMap: Record<string, string> = {
    validate: 'warning',
    prepare: 'info',
    execute: 'success',
    check_exists: 'primary',
    generate_code: '',
    cleanup: 'danger',
  }
  return typeMap[action] || ''
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
.definition-viewer-container {
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

.definition-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.definition-option .def-name {
  font-weight: 500;
}

/* 配置区域 */
.config-section {
  margin-top: 24px;
}

.config-section h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

/* 工作流容器 */
.workflow-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.workflow-graph {
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 8px;
  min-height: 200px;
}

.workflow-nodes {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  justify-content: center;
}

.workflow-node {
  background-color: white;
  border: 2px solid #dcdfe6;
  border-radius: 8px;
  padding: 12px;
  min-width: 180px;
  transition: all 0.3s;
}

.workflow-node:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.2);
  transform: translateY(-2px);
}

.workflow-node.step-validate {
  border-color: #e6a23c;
}

.workflow-node.step-execute {
  border-color: #67c23a;
}

.workflow-node.step-check_exists {
  border-color: #409eff;
}

.workflow-node.step-cleanup {
  border-color: #f56c6c;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 600;
  color: #303133;
}

.node-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.node-timeout,
.node-skip {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 工作流步骤列表 */
.workflow-steps-list h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.step-detail {
  padding: 8px;
}

.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.step-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.step-info .info-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #606266;
}

/* YAML 代码 */
.yaml-container {
  background-color: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.yaml-code {
  margin: 0;
  padding: 20px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #d4d4d4;
  white-space: pre-wrap;
  word-break: break-word;
}

/* 示例容器 */
.example-container {
  padding: 16px 0;
}

.example-section {
  margin-bottom: 24px;
}

.example-section h4 {
  margin: 0 0 12px 0;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.code-block {
  margin: 0;
  padding: 16px;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
}

/* 空状态 */
.empty-card {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
