<template>
  <div class="file-explorer">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>文件浏览器</span>
          <div class="header-actions">
            <el-button type="primary" size="small" @click="refreshFiles">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
            <el-button size="small" @click="createNewFile">
              <el-icon><Plus /></el-icon> 新建文件
            </el-button>
            <el-button size="small" @click="createNewFolder">
              <el-icon><FolderAdd /></el-icon> 新建文件夹
            </el-button>
            <el-button 
              type="success" 
              size="small" 
              @click="pasteFile"
              :disabled="!copiedFile && !movingFile"
            >
              <el-icon><DocumentCopy /></el-icon> 
              {{ copiedFile ? '粘贴' : movingFile ? '粘贴剪切' : '粘贴' }}
            </el-button>
          </div>
        </div>
      </template>
      
      <div class="file-content">
        <!-- 路径导航 -->
        <div class="path-navigator">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item 
              v-for="(path, index) in currentPathArray" 
              :key="index"
              @click="navigateToPath(path.fullPath)"
            >
              {{ path.name }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <!-- 文件列表 -->
        <div 
          class="file-list"
          @dragover.prevent
          @drop="onDrop"
        >
          <div 
            v-for="file in files" 
            :key="file.path"
            class="file-item"
            @click="selectFile(file)"
            :class="{ 'selected': selectedFile?.path === file.path }"
            draggable="true"
            @dragstart="onFileDragStart($event, file)"
          >
            <el-icon :size="28" :color="file.type === 'directory' ? '#409eff' : '#67c23a'">
              {{ file.type === 'directory' ? 'Folder' : 'Document' }}
            </el-icon>
            <div class="file-info">
              <div class="file-name">{{ file.name }}</div>
              <div class="file-meta">
                <span v-if="file.type === 'file'" class="file-size">{{ formatFileSize(file.size) }}</span>
                <span class="file-time">{{ formatDate(file.modified) }}</span>
              </div>
            </div>
            <div class="file-actions">
              <el-button 
                v-if="file.type === 'directory'"
                type="text" 
                size="small" 
                @click.stop="navigateToPath(file.path)"
              >
                <el-icon><Right /></el-icon>
              </el-button>
              <el-button 
                v-else
                type="text" 
                size="small" 
                @click.stop="openFile(file)"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
            </div>
          </div>
          
          <!-- 空状态 -->
          <div v-if="files.length === 0" class="empty-state">
            <el-icon :size="48" color="#909399"><DocumentRemove /></el-icon>
            <p>该目录为空</p>
          </div>
        </div>
        
        <!-- 选中文件信息 -->
        <div v-if="selectedFile" class="selected-file-info">
          <el-card size="small" shadow="hover">
            <h4>文件信息</h4>
            <el-descriptions :column="2" :size="'small'">
              <el-descriptions-item label="名称">{{ selectedFile.name }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ selectedFile.type }}</el-descriptions-item>
              <el-descriptions-item label="大小">{{ formatFileSize(selectedFile.size) }}</el-descriptions-item>
              <el-descriptions-item label="修改时间">{{ formatDate(selectedFile.modified) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDate(selectedFile.created) }}</el-descriptions-item>
              <el-descriptions-item label="路径">{{ selectedFile.path }}</el-descriptions-item>
            </el-descriptions>
            <div class="file-operations">
              <el-button type="primary" size="small" @click="openFile(selectedFile)">
                <el-icon><Edit /></el-icon> 打开
              </el-button>
              <el-button size="small" @click="copyFile(selectedFile)">
                <el-icon><CopyDocument /></el-icon> 复制
              </el-button>
              <el-button size="small" @click="moveFile(selectedFile)">
                <el-icon><DocumentCopy /></el-icon> 移动
              </el-button>
              <el-button type="danger" size="small" @click="deleteFile(selectedFile)">
                <el-icon><Delete /></el-icon> 删除
              </el-button>
            </div>
          </el-card>
        </div>
        
        <!-- 新建文件对话框 -->
        <el-dialog
          v-model="newFileDialogVisible"
          title="新建文件"
          width="40%"
        >
          <el-form label-position="top" size="small">
            <el-form-item label="文件名">
              <el-input v-model="newFileName" placeholder="请输入文件名"></el-input>
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="newFileDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmCreateFile">确定</el-button>
            </span>
          </template>
        </el-dialog>
        
        <!-- 新建文件夹对话框 -->
        <el-dialog
          v-model="newFolderDialogVisible"
          title="新建文件夹"
          width="40%"
        >
          <el-form label-position="top" size="small">
            <el-form-item label="文件夹名">
              <el-input v-model="newFolderName" placeholder="请输入文件夹名"></el-input>
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="newFolderDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmCreateFolder">确定</el-button>
            </span>
          </template>
        </el-dialog>
        
        <!-- 文件编辑对话框 -->
        <el-dialog
          v-model="fileEditDialogVisible"
          :title="'编辑文件: ' + (selectedFile?.name || '')"
          width="70%"
          fullscreen
        >
          <el-input
            v-if="editingFile"
            v-model="fileContent"
            type="textarea"
            rows="20"
            placeholder="文件内容"
            monospaced
          ></el-input>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="fileEditDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="saveFileContent">保存</el-button>
            </span>
          </template>
        </el-dialog>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as main from '../../wailsjs/go/main/App'

const currentPath = ref('/')
const files = ref<any[]>([])
const selectedFile = ref<any>(null)
const newFileDialogVisible = ref(false)
const newFolderDialogVisible = ref(false)
const fileEditDialogVisible = ref(false)
const newFileName = ref('')
const newFolderName = ref('')
const editingFile = ref<any>(null)
const fileContent = ref('')
const copiedFile = ref<any>(null) // 存储复制的源文件
const movingFile = ref<any>(null) // 存储移动的源文件

// 计算当前路径数组
const currentPathArray = computed(() => {
  const pathParts = currentPath.value.split('/').filter(Boolean)
  const pathArray = [{ name: '根目录', fullPath: '/' }]
  
  let fullPath = ''
  for (const part of pathParts) {
    fullPath += '/' + part
    pathArray.push({ name: part, fullPath })
  }
  
  return pathArray
})

// 获取文件列表
const getFiles = async (path: string) => {
  try {
    // 调用后端API获取文件列表
    const result = await main.ListFiles(path)
    files.value = result
  } catch (error) {
    console.error('获取文件列表失败:', error)
    ElMessage.error('获取文件列表失败')
  }
}

// 刷新文件列表
const refreshFiles = () => {
  getFiles(currentPath.value)
}

// 导航到路径
const navigateToPath = (path: string) => {
  currentPath.value = path
  getFiles(path)
  selectedFile.value = null
}

// 选择文件
const selectFile = (file: any) => {
  selectedFile.value = file
}

// 创建新文件
const createNewFile = () => {
  newFileName.value = ''
  newFileDialogVisible.value = true
}

// 确认创建文件
const confirmCreateFile = async () => {
  if (!newFileName.value) return
  
  try {
    const filePath = currentPath.value + '/' + newFileName.value
    // 调用后端API创建文件
    await main.CreateFile(filePath, '')
    
    getFiles(currentPath.value)
    newFileDialogVisible.value = false
    ElMessage.success(`文件 ${newFileName.value} 已创建`)
  } catch (error) {
    console.error('创建文件失败:', error)
    ElMessage.error('创建文件失败')
  }
}

// 创建新文件夹
const createNewFolder = () => {
  newFolderName.value = ''
  newFolderDialogVisible.value = true
}

// 确认创建文件夹
const confirmCreateFolder = async () => {
  if (!newFolderName.value) return
  
  try {
    const folderPath = currentPath.value + '/' + newFolderName.value
    // 调用后端API创建文件夹
    await main.CreateDirectory(folderPath)
    
    getFiles(currentPath.value)
    newFolderDialogVisible.value = false
    ElMessage.success(`文件夹 ${newFolderName.value} 已创建`)
  } catch (error) {
    console.error('创建文件夹失败:', error)
    ElMessage.error('创建文件夹失败')
  }
}

// 打开文件
const openFile = async (file: any) => {
  try {
    // 调用后端API读取文件内容
    const content = await main.ReadFile(file.path)
    fileContent.value = content
    
    editingFile.value = file
    fileEditDialogVisible.value = true
  } catch (error) {
    console.error('打开文件失败:', error)
    ElMessage.error('打开文件失败')
  }
}

// 保存文件内容
const saveFileContent = async () => {
  if (!editingFile.value) return
  
  try {
    // 调用后端API保存文件内容
    await main.WriteFile(editingFile.value.path, fileContent.value)
    
    fileEditDialogVisible.value = false
    ElMessage.success(`文件 ${editingFile.value.name} 已保存`)
  } catch (error) {
    console.error('保存文件失败:', error)
    ElMessage.error('保存文件失败')
  }
}

// 复制文件
const copyFile = (file: any) => {
  copiedFile.value = file
  movingFile.value = null
  ElMessage.success(`已复制: ${file.name}`)
}

// 移动文件
const moveFile = (file: any) => {
  movingFile.value = file
  copiedFile.value = null
  ElMessage.success(`已剪切: ${file.name}`)
}

// 粘贴文件
const pasteFile = async () => {
  if (!copiedFile.value && !movingFile.value) return

  const sourceFile = copiedFile.value || movingFile.value
  const targetPath = currentPath.value

  try {
    // 构建目标文件路径，确保没有双斜杠
    const cleanTargetPath = targetPath.replace(/\/$/, '')
    const targetFilePath = cleanTargetPath + '/' + sourceFile.name
    
    if (copiedFile.value) {
      // 复制文件
      await main.CopyFile(sourceFile.path, targetFilePath)
      ElMessage.success(`已将 ${sourceFile.name} 复制到当前目录`)
    } else if (movingFile.value) {
      // 移动文件
      await main.MoveFile(sourceFile.path, targetFilePath)
      ElMessage.success(`已将 ${sourceFile.name} 移动到当前目录`)
    }
    
    // 刷新文件列表
    getFiles(currentPath.value)
    
    // 清除复制/移动状态
    copiedFile.value = null
    movingFile.value = null
  } catch (error) {
    console.error('粘贴文件失败:', error)
    ElMessage.error(`粘贴文件失败: ${(error as Error).message}`)
  }
}

// 删除文件
const deleteFile = async (file: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${file.name} 吗？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用后端API删除文件
    if (file.type === 'directory') {
      await main.DeleteDirectory(file.path)
    } else {
      await main.DeleteFile(file.path)
    }
    
    getFiles(currentPath.value)
    selectedFile.value = null
    ElMessage.success(`文件 ${file.name} 已删除`)
  } catch (error: any) {
    if (error !== 'cancel' && error.name !== 'CanceledError') {
      console.error('删除文件失败:', error)
      ElMessage.error('删除文件失败')
    }
  }
}

// 文件拖拽开始事件
const onFileDragStart = (event: DragEvent, file: any) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/json', JSON.stringify(file))
  }
}

// 拖拽结束事件
const onDrop = (event: DragEvent) => {
  event.preventDefault()
  // TODO: 实现拖拽文件到工作区的功能
  ElMessage.info('文件已拖拽到工作区')
}

// 格式化文件大小
const formatFileSize = (size: number) => {
  if (size < 1024) {
    return size + ' B'
  } else if (size < 1024 * 1024) {
    return (size / 1024).toFixed(2) + ' KB'
  } else if (size < 1024 * 1024 * 1024) {
    return (size / (1024 * 1024)).toFixed(2) + ' MB'
  } else {
    return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  }
}

// 格式化日期
const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString()
}

// 生命周期钩子
onMounted(() => {
  getFiles(currentPath.value)
})

// 暴露变量和方法给测试
defineExpose({
  currentPath,
  files,
  selectedFile,
  copiedFile,
  movingFile,
  selectFile,
  copyFile,
  moveFile,
  pasteFile,
  navigateToPath,
  refreshFiles
})
</script>

<style scoped>
.file-explorer {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.file-content {
  margin-top: 20px;
}

.path-navigator {
  margin-bottom: 20px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.file-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 15px;
  padding: 20px;
  background-color: #fafafa;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  min-height: 300px;
}

.file-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 15px;
  background-color: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.file-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.file-item.selected {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.file-info {
  margin-top: 10px;
  text-align: center;
  width: 100%;
  overflow: hidden;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-actions {
  position: absolute;
  top: 5px;
  right: 5px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.file-item:hover .file-actions {
  opacity: 1;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #909399;
}

.selected-file-info {
  margin-top: 20px;
}

.file-operations {
  margin-top: 15px;
  display: flex;
  gap: 10px;
}
</style>
