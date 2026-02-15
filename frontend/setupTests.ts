import { expect, vi } from 'vitest'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

// 创建一个空的Vue应用实例用于测试
const app = createApp({})

// 注册Element Plus和图标
app.use(ElementPlus)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册Pinia
app.use(createPinia())

// 创建一个简单的路由实例
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('./src/views/Home.vue') },
    { path: '/workflow', component: () => import('./src/views/WorkflowEditor.vue') },
    { path: '/skills', component: () => import('./src/views/SkillManager.vue') },
    { path: '/config', component: () => import('./src/views/ConfigManager.vue') },
    { path: '/files', component: () => import('./src/views/FileExplorer.vue') }
  ]
})
app.use(router)

// 全局mocks
vi.mock('./wailsjs/go/main/App', () => ({
  default: {
    // 添加你需要模拟的Wails API
  }
}))

// 添加任何全局测试辅助函数
globalThis.testApp = app
globalThis.testRouter = router