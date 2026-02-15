import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import SkillManager from './SkillManager.vue'

// Mock Element Plus components and plugins
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

// Mock the main module
vi.mock('../../wailsjs/go/main/App', () => ({
  GetSkills: vi.fn().mockResolvedValue([
    {
      name: 'test-skill',
      description: 'Test skill',
      category: 'test',
      enabled: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      version: '1.0.0',
      author: 'Test Author',
      dependencies: [],
      config: {}
    }
  ]),
  EnableSkill: vi.fn().mockResolvedValue(true),
  DisableSkill: vi.fn().mockResolvedValue(true),
  DeleteSkill: vi.fn().mockResolvedValue(true)
}))

describe('SkillManager.vue', () => {
  it('should render the component', () => {
    const wrapper = mount(SkillManager, {
      global: {
        stubs: {
          // Stub all Element Plus components
          'ElCard': true,
          'ElButton': true,
          'ElIcon': true,
          'ElTable': true,
          'ElTableColumn': true,
          'ElSwitch': true,
          'ElDialog': true,
          'ElForm': true,
          'ElFormItem': true,
          'ElInput': true,
          'ElDescriptions': true,
          'ElDescriptionsItem': true,
          'ElTag': true,
          // Stub icons
          'Refresh': true,
          'Setting': true,
          'InfoFilled': true,
          'Delete': true,
          'Check': true,
          'CircleClose': true
        }
      }
    })
    expect(wrapper.exists()).toBe(true)
  })

  it('should contain expected text content', () => {
    const wrapper = mount(SkillManager, {
      global: {
        stubs: {
          // Stub all Element Plus components
          'ElCard': true,
          'ElButton': true,
          'ElIcon': true,
          'ElTable': true,
          'ElTableColumn': true,
          'ElSwitch': true,
          'ElDialog': true,
          'ElForm': true,
          'ElFormItem': true,
          'ElInput': true,
          'ElDescriptions': true,
          'ElDescriptionsItem': true,
          'ElTag': true,
          // Stub icons
          'Refresh': true,
          'Setting': true,
          'InfoFilled': true,
          'Delete': true,
          'Check': true,
          'CircleClose': true
        }
      }
    })
    const text = wrapper.text()
    expect(text).toContain('技能管理')
    expect(text).toContain('刷新')
  })

  it('should render the skill table', () => {
    const wrapper = mount(SkillManager, {
      global: {
        stubs: {
          // Stub all Element Plus components
          'ElCard': true,
          'ElButton': true,
          'ElIcon': true,
          'ElTable': true,
          'ElTableColumn': true,
          'ElSwitch': true,
          'ElDialog': true,
          'ElForm': true,
          'ElFormItem': true,
          'ElInput': true,
          'ElDescriptions': true,
          'ElDescriptionsItem': true,
          'ElTag': true,
          // Stub icons
          'Refresh': true,
          'Setting': true,
          'InfoFilled': true,
          'Delete': true,
          'Check': true,
          'CircleClose': true
        }
      }
    })
    expect(wrapper.find('.skill-manager').exists()).toBe(true)
  })
})
