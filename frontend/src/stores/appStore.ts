/**
 * AgentFramework - Application Global Store
 * 全局应用状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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

  // ===== Getters =====
  const isDarkMode = computed(() => theme.value === 'dark')
  const unreadNotifications = computed(() => {
    return notifications.value.length
  })

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

  return {
    // State
    theme,
    sidebarCollapsed,
    currentRoute,
    notifications,
    
    // Getters
    isDarkMode,
    unreadNotifications,
    
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
  }
})
