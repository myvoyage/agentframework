import {createApp} from 'vue'
import App from './App.vue'
import './style.css'
import router from './router'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// Import stores
import { useAppStore } from './stores/appStore'
import { useSecurityStore } from './stores/securityStore'
import { usePerformanceStore } from './stores/performanceStore'

const app = createApp(App)
const pinia = createPinia()

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(router)
app.use(pinia)
app.use(ElementPlus)

// Initialize stores
const appStore = useAppStore(pinia)
const securityStore = useSecurityStore(pinia)
const performanceStore = usePerformanceStore(pinia)

// Initialize security and performance features
securityStore.initSecurityConfig()
performanceStore.initFromStorage()

// Auto-start performance monitoring in production
if (import.meta.env.PROD) {
  performanceStore.startMonitoring(5000)
}

app.mount('#app')
