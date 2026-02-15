import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import FileExplorer from './FileExplorer.vue'
import * as main from '../../wailsjs/go/main/App'

// 模拟后端API调用
vi.mock('../../wailsjs/go/main/App', () => ({
  ListFiles: vi.fn().mockResolvedValue([
    {
      name: 'file1.txt',
      path: '/file1.txt',
      type: 'file',
      size: 1024,
      modified: new Date().toISOString(),
      created: new Date().toISOString()
    },
    {
      name: 'folder1',
      path: '/folder1',
      type: 'directory',
      size: 0,
      modified: new Date().toISOString(),
      created: new Date().toISOString()
    }
  ]),
  CopyFile: vi.fn().mockResolvedValue(undefined),
  MoveFile: vi.fn().mockResolvedValue(undefined),
  CreateFile: vi.fn().mockResolvedValue(undefined),
  CreateDirectory: vi.fn().mockResolvedValue(undefined),
  DeleteFile: vi.fn().mockResolvedValue(undefined),
  DeleteDirectory: vi.fn().mockResolvedValue(undefined),
  ReadFile: vi.fn().mockResolvedValue('test content'),
  WriteFile: vi.fn().mockResolvedValue(undefined),
  DeleteSkill: vi.fn().mockResolvedValue(undefined),
  GetWorkflowVersions: vi.fn().mockResolvedValue([]),
  GetWorkflowVersion: vi.fn().mockResolvedValue({}),
  RestoreWorkflowVersion: vi.fn().mockResolvedValue(undefined),
  CreateWorkflow: vi.fn().mockResolvedValue('workflow-id')
}))

// 模拟ElMessage
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn().mockResolvedValue(true)
  }
}))

describe('FileExplorer.vue', () => {
  // 在每个测试用例前重置模拟函数
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render the component', () => {
    const wrapper = mount(FileExplorer)
    expect(wrapper.exists()).toBe(true)
  })

  it('should render the file list', () => {
    const wrapper = mount(FileExplorer)
    expect(wrapper.find('.file-list').exists()).toBe(true)
  })

  it('should render the file explorer header', () => {
    const wrapper = mount(FileExplorer)
    expect(wrapper.find('.card-header span').exists()).toBe(true)
    expect(wrapper.find('.header-actions button').exists()).toBe(true)
  })

  it('should have action buttons', () => {
    const wrapper = mount(FileExplorer)
    const buttons = wrapper.findAll('.header-actions button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('should call copyFile method when copy button is clicked', async () => {
    const wrapper = mount(FileExplorer)
    
    // 模拟文件数据
    const files = [
      {
        name: 'file1.txt',
        path: '/file1.txt',
        type: 'file',
        size: 1024,
        modified: new Date().toISOString(),
        created: new Date().toISOString()
      }
    ]
    
    // 更新文件列表
    await wrapper.vm.$nextTick()
    wrapper.vm.files = files
    
    // 模拟选择文件
    await wrapper.vm.selectFile(files[0])
    
    // 调用复制方法
    await wrapper.vm.copyFile(files[0])
    
    // 检查复制状态
    expect(wrapper.vm.copiedFile).toEqual(files[0])
    expect(wrapper.vm.movingFile).toBeNull()
  })

  it('should call moveFile method when move button is clicked', async () => {
    const wrapper = mount(FileExplorer)
    
    // 模拟文件数据
    const files = [
      {
        name: 'file1.txt',
        path: '/file1.txt',
        type: 'file',
        size: 1024,
        modified: new Date().toISOString(),
        created: new Date().toISOString()
      }
    ]
    
    // 更新文件列表
    await wrapper.vm.$nextTick()
    wrapper.vm.files = files
    
    // 模拟选择文件
    await wrapper.vm.selectFile(files[0])
    
    // 调用移动方法
    await wrapper.vm.moveFile(files[0])
    
    // 检查移动状态
    expect(wrapper.vm.movingFile).toEqual(files[0])
    expect(wrapper.vm.copiedFile).toBeNull()
  })

  it('should call CopyFile when pasting a copied file', async () => {
    const wrapper = mount(FileExplorer)
    
    // 模拟文件数据
    const files = [
      {
        name: 'file1.txt',
        path: '/file1.txt',
        type: 'file',
        size: 1024,
        modified: new Date().toISOString(),
        created: new Date().toISOString()
      }
    ]
    
    // 更新文件列表和状态
    await wrapper.vm.$nextTick()
    wrapper.vm.files = files
    wrapper.vm.copiedFile = files[0]
    wrapper.vm.currentPath = '/'
    
    // 调用粘贴方法
    await wrapper.vm.pasteFile()
    
    // 检查是否调用了后端API
    expect(main.CopyFile).toHaveBeenCalledWith('/file1.txt', '/file1.txt')
  })

  it('should call MoveFile when pasting a moved file', async () => {
    const wrapper = mount(FileExplorer)
    
    // 模拟文件数据
    const files = [
      {
        name: 'file1.txt',
        path: '/file1.txt',
        type: 'file',
        size: 1024,
        modified: new Date().toISOString(),
        created: new Date().toISOString()
      }
    ]
    
    // 更新文件列表和状态
    await wrapper.vm.$nextTick()
    wrapper.vm.files = files
    wrapper.vm.movingFile = files[0]
    wrapper.vm.currentPath = '/'
    
    // 调用粘贴方法
    await wrapper.vm.pasteFile()
    
    // 检查是否调用了后端API
    expect(main.MoveFile).toHaveBeenCalledWith('/file1.txt', '/file1.txt')
  })

  it('should call navigateToPath method when breadcrumb is clicked', async () => {
    const wrapper = mount(FileExplorer)
    
    // 模拟当前路径
    wrapper.vm.currentPath = '/folder1/subfolder'
    
    // 调用导航方法
    await wrapper.vm.navigateToPath('/folder1')
    
    // 检查路径是否更新
    expect(wrapper.vm.currentPath).toBe('/folder1')
  })

  it('should refresh files when refreshFiles is called', async () => {
    const wrapper = mount(FileExplorer)
    
    // 调用刷新方法
    await wrapper.vm.refreshFiles()
    
    // 检查是否调用了后端API
    expect(main.ListFiles).toHaveBeenCalled()
  })

  it('should handle paste when no file is copied or moved', async () => {
    const wrapper = mount(FileExplorer)
    
    // 确保没有文件被复制或移动
    wrapper.vm.copiedFile = null
    wrapper.vm.movingFile = null
    
    // 调用粘贴方法
    await wrapper.vm.pasteFile()
    
    // 检查是否没有调用后端API
    expect(main.CopyFile).not.toHaveBeenCalled()
    expect(main.MoveFile).not.toHaveBeenCalled()
  })
})
