import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import Home from './Home.vue'

describe('Home.vue', () => {
  it('should render the component', () => {
    const wrapper = mount(Home)
    expect(wrapper.exists()).toBe(true)
  })

  it('should contain expected text content', () => {
    const wrapper = mount(Home)
    const text = wrapper.text()
    // Check for some key content from the component
    expect(text).toContain('AgentFramework')
    expect(text).toContain('AI 代理框架')
    expect(text).toContain('支持工作流编排')
    expect(text).toContain('技能管理')
    expect(text).toContain('文件系统操作')
  })

  it('should render the home container', () => {
    const wrapper = mount(Home)
    expect(wrapper.find('.home-container').exists()).toBe(true)
  })

  it('should render the feature grid', () => {
    const wrapper = mount(Home)
    expect(wrapper.find('.feature-grid').exists()).toBe(true)
  })

  it('should render feature cards', () => {
    const wrapper = mount(Home)
    const featureCards = wrapper.findAll('.feature-card')
    expect(featureCards.length).toBeGreaterThan(0)
  })

  it('should render feature cards with expected content', () => {
    const wrapper = mount(Home)
    const text = wrapper.text()
    // Check for feature card titles in the rendered text
    expect(text).toContain('可视化工作流编辑器')
    expect(text).toContain('技能管理')
    expect(text).toContain('配置管理')
    expect(text).toContain('文件系统浏览器')
  })

  it('should contain navigation buttons', () => {
    const wrapper = mount(Home)
    const text = wrapper.text()
    // Check for button text instead of trying to select Element Plus buttons directly
    expect(text).toContain('进入工作流编辑器')
    expect(text).toContain('进入技能管理')
    expect(text).toContain('进入配置管理')
    expect(text).toContain('进入文件浏览器')
  })
})
