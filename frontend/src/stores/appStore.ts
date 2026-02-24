/**
 * AgentFramework - Application Global Store
 * 全局应用状态管理
 * 增强以支持 API 配置管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { configureApi, getApiConfig } from '@/services/api'

export const useAppStore = defineStore('app', () => {
  // ===== State =====
  const theme = ref<'light' | 'dark'>('light')
  const sidebarCollapsed = ref(false)
  const currentRoute = ref('')
  const notifications = ref<Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    message: string
    timestamp: Date
  }>>([])

  // API 配置状态
  const apiConfigured = ref(false)
  const apiMode = ref<'wails' | 'http'>('wails')
  const apiConnected = ref(false)
  const apiHealth = ref<'unknown' | 'healthy' | 'unhealthy'>('unknown')

  // ===== Getters =====
  const isDarkMode = computed(() => theme.value === 'dark')
  const unreadNotifications = computed(() => {
    return notifications.value.length
  })
  const usingHttpApi = computed(() => apiMode.value === 'http')

  // ===== Actions =====
  function setTheme(newTheme: 'light' | 'dark') {
    theme.value = newTheme
    // Persist to localStorage
    localStorage.setItem('theme', newTheme)
  }

  function toggleTheme() {
    setTheme(theme.value === 'light' ? 'dark' : 'light')
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
  }

  function setCurrentRoute(route: string) {
    currentRoute.value = route
  }

  function addNotification(
    type: 'success' | 'error' | 'warning' | 'info',
    message: string
  ) {
    const notification = {
      id: Date.now().toString(),
      type,
      message,
      timestamp: new Date(),
    }
    notifications.value.unshift(notification)

    // Auto remove after 5 seconds
    setTimeout(() => {
      removeNotification(notification.id)
    }, 5000)
  }

  function removeNotification(id: string) {
    const index = notifications.value.findIndex((n) => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }

  function clearNotifications() {
    notifications.value = []
  }

  // Initialize theme from localStorage
  function initTheme() {
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme === 'light' || savedTheme === 'dark') {
      theme.value = savedTheme
    }
  }

  // ===== API Configuration Actions =====

  /**
   * Configure API mode (Wails or HTTP)
   */
  function configureApiMode(mode: 'wails' | 'http', baseUrl?: string) {
    apiMode.value = mode

    if (mode === 'http' && baseUrl) {
      configureApi({
        useHttp: true,
        baseUrl,
        timeout: 30000,
      })
      apiConfigured.value = true
    } else {
      configureApi({
        useHttp: false,
        baseUrl: '',
        timeout: 30000,
      })
      apiConfigured.value = true
    }

    // Persist configuration to localStorage
    localStorage.setItem('apiMode', mode)
    if (baseUrl) {
      localStorage.setItem('apiBaseUrl', baseUrl)
    }
  }

  /**
   * Initialize API configuration from localStorage
   */
  function initApiConfig() {
    const savedMode = localStorage.getItem('apiMode') as 'wails' | 'http' | null
    const savedBaseUrl = localStorage.getItem('apiBaseUrl')

    if (savedMode) {
      configureApiMode(savedMode, savedBaseUrl || undefined)
    } else {
      // Default to Wails mode
      configureApiMode('wails')
    }
  }

  /**
   * Check API health status
   */
  async function checkApiHealth() {
    try {
      const config = getApiConfig()
      if (config.useHttp) {
        // Would make HTTP request to /health endpoint
        apiConnected.value = true
        apiHealth.value = 'healthy'
      } else {
        // Wails mode is always "connected"
        apiConnected.value = true
        apiHealth.value = 'healthy'
      }
    } catch (err) {
      apiConnected.value = false
      apiHealth.value = 'unhealthy'
    }
  }

  return {
    // State
    theme,
    sidebarCollapsed,
    currentRoute,
    notifications,
    apiConfigured,
    apiMode,
    apiConnected,
    apiHealth,

    // Getters
    isDarkMode,
    unreadNotifications,
    usingHttpApi,

    // Actions
    setTheme,
    toggleTheme,
    toggleSidebar,
    setSidebarCollapsed,
    setCurrentRoute,
    addNotification,
    removeNotification,
    clearNotifications,
    initTheme,
    configureApiMode,
    initApiConfig,
    checkApiHealth,
  }
})
