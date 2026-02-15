<template>
  <div class="skill-import-container">
    <!-- 导入方式选择卡片 -->
    <el-card shadow="hover" class="import-method-card">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><Upload /></el-icon>
            导入技能
          </span>
          <el-text type="info" size="small">支持多种导入方式</el-text>
        </div>
      </template>

      <!-- 导入方式选择 -->
      <div class="import-methods">
        <el-radio-group v-model="importMethod" @change="handleMethodChange">
          <el-radio-button label="file">
            <el-icon><Folder /></el-icon>
            本地文件
          </el-radio-button>
          <el-radio-button label="url">
            <el-icon><Link /></el-icon>
            URL导入
          </el-radio-button>
          <el-radio-button label="paste">
            <el-icon><DocumentCopy /></el-icon>
            粘贴内容
          </el-radio-button>
          <el-radio-button label="git">
            <el-icon><Grid /></el-icon>
            Git仓库
          </el-radio-button>
        </el-radio-group>
      </div>

      <!-- 文件上传 -->
      <div v-if="importMethod === 'file'" class="import-section">
        <el-upload
          ref="uploadRef"
          class="upload-area"
          drag
          :auto-upload="false"
          :on-change="handleFileChange"
          accept=".zip,.tar.gz,.tgz"
          :limit="1"
          :on-exceed="handleExceed"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">
            拖拽技能包到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              支持 .zip、.tar.gz 格式的技能包，最大 50MB
              <el-link type="primary" @click.stop="showTemplateFormat = true">
                查看格式说明
              </el-link>
            </div>
          </template>
        </el-upload>

        <div v-if="selectedFile" class="file-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="文件名">
              {{ selectedFile.name }}
            </el-descriptions-item>
            <el-descriptions-item label="文件大小">
              {{ formatFileSize(selectedFile.size) }}
            </el-descriptions-item>
            <el-descriptions-item label="文件类型">
              {{ selectedFile.type || '未知' }}
            </el-descriptions-item>
            <el-descriptions-item label="最后修改">
              {{ formatDate(selectedFile.lastModified) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <!-- URL导入 -->
      <div v-if="importMethod === 'url'" class="import-section">
        <el-form :model="urlForm" :rules="urlRules" ref="urlFormRef" label-width="100px">
          <el-form-item label="技能URL" prop="url">
            <el-input
              v-model="urlForm.url"
              placeholder="请输入技能包的下载URL"
              clearable
            >
              <template #prepend>
                <el-icon><Link /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="认证方式">
            <el-radio-group v-model="urlForm.authType">
              <el-radio label="none">无需认证</el-radio>
              <el-radio label="token">Token认证</el-radio>
              <el-radio label="basic">基础认证</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item
            v-if="urlForm.authType === 'token'"
            label="Token"
            prop="token"
          >
            <el-input
              v-model="urlForm.token"
              type="password"
              placeholder="请输入访问令牌"
              show-password
              clearable
            />
          </el-form-item>

          <el-form-item
            v-if="urlForm.authType === 'basic'"
            label="用户名"
            prop="username"
          >
            <el-input
              v-model="urlForm.username"
              placeholder="请输入用户名"
              clearable
            />
          </el-form-item>

          <el-form-item
            v-if="urlForm.authType === 'basic'"
            label="密码"
            prop="password"
          >
            <el-input
              v-model="urlForm.password"
              type="password"
              placeholder="请输入密码"
              show-password
              clearable
            />
          </el-form-item>
        </el-form>
      </div>

      <!-- 粘贴内容 -->
      <div v-if="importMethod === 'paste'" class="import-section">
        <el-alert
          title="粘贴SKILL.md内容"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px"
        >
          <template #default>
            请粘贴技能定义文件的完整内容，支持 YAML 格式的元数据和 Markdown 格式的文档
          </template>
        </el-alert>

        <el-input
          v-model="pasteContent"
          type="textarea"
          :rows="15"
          placeholder="---&#10;name: MySkill&#10;description: 技能描述&#10;version: 1.0.0&#10;category: custom&#10;---&#10;&#10;# 技能文档&#10;&#10;在此处编写技能的使用文档..."
          class="code-editor"
        />

        <div class="paste-actions">
          <el-button @click="pasteContent = ''">
            <el-icon><Delete /></el-icon>
            清空
          </el-button>
          <el-button @click="loadExample">
            <el-icon><Document /></el-icon>
            加载示例
          </el-button>
        </div>
      </div>

      <!-- Git仓库导入 -->
      <div v-if="importMethod === 'git'" class="import-section">
        <el-form :model="gitForm" :rules="gitRules" ref="gitFormRef" label-width="120px">
          <el-form-item label="仓库地址" prop="url">
            <el-input
              v-model="gitForm.url"
              placeholder="https://github.com/username/skill-repo"
              clearable
            >
              <template #prepend>
                <el-icon><Grid /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="分支">
            <el-input
              v-model="gitForm.branch"
              placeholder="main (默认)"
              clearable
            />
          </el-form-item>

          <el-form-item label="技能路径">
            <el-input
              v-model="gitForm.path"
              placeholder="/skills/my-skill (可选)"
              clearable
            >
              <template #prepend>
                <el-icon><FolderOpened /></el-icon>
              </template>
            </el-input>
            <div class="form-tip">
              指定仓库中技能的相对路径，留空则使用根目录
            </div>
          </el-form-item>

          <el-form-item label="递归克隆">
            <el-switch v-model="gitForm.recursive" />
            <div class="form-tip">
              如果仓库包含子模块，启用此选项
            </div>
          </el-form-item>

          <el-form-item label="深度">
            <el-input-number
              v-model="gitForm.depth"
              :min="1"
              :max="100"
              placeholder="1 (浅克隆)"
            />
            <div class="form-tip">
              浅克隆可以减少下载的数据量
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- 导入选项 -->
      <el-divider>导入选项</el-divider>

      <el-form :model="importOptions" label-width="120px">
        <el-form-item label="技能ID">
          <el-input
            v-model="importOptions.skillId"
            placeholder="自动生成 (可覆盖)"
            clearable
          >
            <template #append>
              <el-button @click="generateSkillId" :icon="RefreshRight" />
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="覆盖已存在">
          <el-switch v-model="importOptions.overwrite" />
          <div class="form-tip">
            如果技能ID已存在，是否覆盖现有技能
          </div>
        </el-form-item>

        <el-form-item label="自动启用">
          <el-switch v-model="importOptions.autoEnable" />
          <div class="form-tip">
            导入后自动启用该技能
          </div>
        </el-form-item>

        <el-form-item label="验证格式">
          <el-switch v-model="importOptions.validate" />
          <div class="form-tip">
            在导入前验证技能定义的格式是否正确
          </div>
        </el-form-item>
      </el-form>

      <!-- 操作按钮 -->
      <div class="action-buttons">
        <el-button @click="handleReset">
          <el-icon><RefreshLeft /></el-icon>
          重置
        </el-button>
        <el-button type="primary" :loading="importing" @click="handleImport">
          <el-icon><Check /></el-icon>
          {{ importing ? '导入中...' : '开始导入' }}
        </el-button>
      </div>
    </el-card>

    <!-- 导入结果对话框 -->
    <el-dialog
      v-model="resultDialogVisible"
      :title="importSuccess ? '导入成功' : '导入失败'"
      width="60%"
    >
      <div v-if="importSuccess && importedSkill" class="import-result">
        <el-result icon="success" title="技能导入成功" sub-title="技能已成功添加到技能库">
          <template #extra>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="技能ID">
                {{ importedSkill.id }}
              </el-descriptions-item>
              <el-descriptions-item label="名称">
                {{ importedSkill.name }}
              </el-descriptions-item>
              <el-descriptions-item label="分类">
                <el-tag :type="getCategoryTagType(importedSkill.category)">
                  {{ getCategoryLabel(importedSkill.category) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="版本">
                {{ importedSkill.version }}
              </el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">
                {{ importedSkill.description }}
              </el-descriptions-item>
            </el-descriptions>

            <el-divider>操作</el-divider>

            <el-button type="primary" @click="viewImportedSkill">
              <el-icon><View /></el-icon>
              查看技能
            </el-button>
            <el-button @click="executeImportedSkill">
              <el-icon><VideoPlay /></el-icon>
              执行技能
            </el-button>
          </template>
        </el-result>
      </div>

      <div v-else class="import-error">
        <el-result icon="error" title="导入失败" :sub-title="errorMessage">
          <template #extra>
            <el-button type="primary" @click="resultDialogVisible = false">
              关闭
            </el-button>
            <el-button @click="handleRetry">
              <el-icon><RefreshRight /></el-icon>
              重试
            </el-button>
          </template>
        </el-result>
      </div>
    </el-dialog>

    <!-- 格式说明对话框 -->
    <el-dialog
      v-model="showTemplateFormat"
      title="技能包格式说明"
      width="70%"
    >
      <div class="format-doc">
        <el-alert
          title="Claude Skills 格式"
          type="success"
          :closable="false"
          show-icon
          style="margin-bottom: 20px"
        >
          AgentFramework兼容 Claude Skills 格式，支持渐进式披露和脚本工具
        </el-alert>

        <h3>📁 基本结构</h3>
        <pre class="code-block"><code>my-skill/
├── SKILL.md          # 必需：技能主文档
├── docs.md           # 可选：详细文档
├── references/       # 可选：参考文件目录
│   ├── doc1.md
│   └── doc2.md
└── scripts/          # 可选：脚本工具目录
    ├── tool.py
    └── script.sh</code></pre>

        <h3>📝 SKILL.md 格式</h3>
        <pre class="code-block"><code>---
name: MySkill
description: 技能描述
version: 1.0.0
category: custom
author: Your Name
license: MIT
tags:
  - tag1
  - tag2
---

# 技能名称

## 概述
简要描述技能的功能和用途。

## 使用方法
### 步骤1
...

### 步骤2
...

## 工具
技能包含以下工具：
- `scripts/tool.py`: Python工具脚本
- `scripts/script.sh`: Bash脚本

## 参考
- `references/doc1.md`: 参考文档1
- `references/doc2.md`: 参考文档2</code></pre>

        <el-divider />

        <h3>🎯 最佳实践</h3>
        <ul>
          <li><strong>渐进式披露</strong>：元数据精简（50 tokens），完整文档详细（500 tokens）</li>
          <li><strong>脚本工具</strong>：将可复用的脚本放在 scripts/ 目录</li>
          <li><strong>参考文档</strong>：大型参考文件放在 references/ 目录，按需加载</li>
          <li><strong>版本管理</strong>：使用 Git 管理技能的版本和变更历史</li>
          <li><strong>文档完善</strong>：提供清晰的使用说明和示例</li>
        </ul>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Upload,
  Folder,
  Link,
  DocumentCopy,
  Grid,
  UploadFilled,
  Delete,
  Document,
  RefreshRight,
  RefreshLeft,
  Check,
  View,
  VideoPlay,
  FolderOpened,
} from '@element-plus/icons-vue'
import type { UploadInstance, UploadProps, UploadRawFile } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

// Props
interface Props {
  workspace?: string
}

const props = withDefaults(defineProps<Props>(), {
  workspace: '/workspace',
})

// Emits
interface Emits {
  (e: 'imported', skill: any): void
  (e: 'execute', skill: any): void
}

const emit = defineEmits<Emits>()

// 状态
const importMethod = ref('file')
const importing = ref(false)
const selectedFile = ref<File | null>(null)
const pasteContent = ref('')
const resultDialogVisible = ref(false)
const importSuccess = ref(false)
const importedSkill = ref<any>(null)
const errorMessage = ref('')
const showTemplateFormat = ref(false)

// URL导入表单
const urlFormRef = ref<FormInstance>()
const urlForm = reactive({
  url: '',
  authType: 'none',
  token: '',
  username: '',
  password: '',
})

const urlRules: FormRules = {
  url: [{ required: true, message: '请输入技能URL', trigger: 'blur' }],
}

// Git导入表单
const gitFormRef = ref<FormInstance>()
const gitForm = reactive({
  url: '',
  branch: 'main',
  path: '',
  recursive: false,
  depth: 1,
})

const gitRules: FormRules = {
  url: [{ required: true, message: '请输入仓库地址', trigger: 'blur' }],
}

// 导入选项
const importOptions = reactive({
  skillId: '',
  overwrite: false,
  autoEnable: true,
  validate: true,
})

// 引用
const uploadRef = ref<UploadInstance>()

// 方法
const handleMethodChange = (method: string) => {
  console.log('[SkillImport] 导入方式切换为:', method)
}

const handleFileChange: UploadProps['onChange'] = (uploadFile) => {
  selectedFile.value = uploadFile.raw as File
  console.log('[SkillImport] 选择了文件:', selectedFile.value?.name)
}

const handleExceed: UploadProps['onExceed'] = (files) => {
  uploadRef.value!.clearFiles()
  const file = files[0] as UploadRawFile
  uploadRef.value!.handleStart(file)
}

const generateSkillId = () => {
  const timestamp = Date.now().toString(36)
  const random = Math.random().toString(36).substring(2, 8)
  importOptions.skillId = `skill_${timestamp}_${random}`
}

const handleReset = () => {
  selectedFile.value = null
  pasteContent.value = ''
  urlForm.url = ''
  urlForm.authType = 'none'
  urlForm.token = ''
  urlForm.username = ''
  urlForm.password = ''
  gitForm.url = ''
  gitForm.branch = 'main'
  gitForm.path = ''
  gitForm.recursive = false
  gitForm.depth = 1
  importOptions.skillId = ''
  importOptions.overwrite = false
  importOptions.autoEnable = true
  importOptions.validate = true

  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }

  ElMessage.info('已重置所有选项')
}

const handleImport = async () => {
  importing.value = true

  try {
    let result: any

    switch (importMethod.value) {
      case 'file':
        if (!selectedFile.value) {
          ElMessage.error('请先选择文件')
          return
        }
        result = await importFromFile()
        break

      case 'url':
        await urlFormRef.value?.validate()
        result = await importFromUrl()
        break

      case 'paste':
        if (!pasteContent.value.trim()) {
          ElMessage.error('请粘贴技能内容')
          return
        }
        result = await importFromPaste()
        break

      case 'git':
        await gitFormRef.value?.validate()
        result = await importFromGit()
        break
    }

    importedSkill.value = result
    importSuccess.value = true
    resultDialogVisible.value = true

    ElMessage.success('技能导入成功！')
    emit('imported', result)
  } catch (error: any) {
    console.error('[SkillImport] 导入失败:', error)
    errorMessage.value = error.message || '未知错误'
    importSuccess.value = false
    resultDialogVisible.value = true
  } finally {
    importing.value = false
  }
}

const importFromFile = async () => {
  // TODO: 实现文件上传和解析逻辑
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        id: importOptions.skillId || 'imported_skill_001',
        name: '导入的技能',
        description: '从文件导入的技能',
        version: '1.0.0',
        category: 'custom',
      })
    }, 1000)
  })
}

const importFromUrl = async () => {
  // TODO: 实现URL下载和解析逻辑
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        id: importOptions.skillId || 'imported_skill_002',
        name: '从URL导入的技能',
        description: '从URL导入的技能',
        version: '1.0.0',
        category: 'custom',
      })
    }, 1000)
  })
}

const importFromPaste = async () => {
  // TODO: 实现内容解析逻辑
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        id: importOptions.skillId || 'imported_skill_003',
        name: '粘贴的技能',
        description: '从粘贴内容导入的技能',
        version: '1.0.0',
        category: 'custom',
      })
    }, 500)
  })
}

const importFromGit = async () => {
  // TODO: 实现Git克隆和解析逻辑
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        id: importOptions.skillId || 'imported_skill_004',
        name: '从Git导入的技能',
        description: '从Git仓库导入的技能',
        version: '1.0.0',
        category: 'custom',
      })
    }, 1500)
  })
}

const handleRetry = () => {
  resultDialogVisible.value = false
  handleImport()
}

const viewImportedSkill = () => {
  emit('imported', importedSkill.value)
  resultDialogVisible.value = false
}

const executeImportedSkill = () => {
  emit('execute', importedSkill.value)
  resultDialogVisible.value = false
}

const loadExample = () => {
  pasteContent.value = `---
name: ExampleSkill
description: 这是一个示例技能
version: 1.0.0
category: custom
author: Your Name
license: MIT
tags:
  - example
  - demo
---

# 示例技能

## 概述
这是一个示例技能，展示了如何创建技能定义文件。

## 使用方法
1. 在输入中提供必要参数
2. 技能将执行相应的操作
3. 返回执行结果

## 功能
- 功能1：描述
- 功能2：描述
- 功能3：描述

## 工具
- \`scripts/tool.py\`: Python工具脚本
`
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (timestamp: number): string => {
  return new Date(timestamp).toLocaleString('zh-CN')
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
.skill-import-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.import-methods {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.import-methods :deep(.el-radio-group) {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.import-methods :deep(.el-radio-button) {
  margin: 0;
}

.import-methods :deep(.el-radio-button__inner) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 20px;
}

.import-section {
  padding: 20px 0;
}

.upload-area {
  margin-bottom: 20px;
}

.upload-area :deep(.el-upload) {
  width: 100%;
}

.upload-area :deep(.el-upload-dragger) {
  width: 100%;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.el-icon--upload {
  font-size: 67px;
  color: #409eff;
  margin-bottom: 16px;
}

.file-info {
  margin-top: 20px;
}

.code-editor {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.paste-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #dcdfe6;
}

.import-result,
.import-error {
  padding: 20px;
}

.format-doc h3 {
  margin-top: 20px;
  margin-bottom: 12px;
  color: #303133;
}

.code-block {
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 16px;
  margin: 12px 0;
  overflow-x: auto;
}

.code-block code {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #303133;
}

.format-doc ul {
  list-style-type: disc;
  padding-left: 20px;
}

.format-doc li {
  margin-bottom: 8px;
}

.format-doc strong {
  color: #409eff;
}

/* 响应式 */
@media (max-width: 768px) {
  .import-methods :deep(.el-radio-group) {
    flex-direction: column;
  }

  .import-methods :deep(.el-radio-button__inner) {
    width: 100%;
  }

  .action-buttons {
    flex-direction: column;
  }

  .action-buttons :deep(.el-button) {
    width: 100%;
  }
}
</style>
