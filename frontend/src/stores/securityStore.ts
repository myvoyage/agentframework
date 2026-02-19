/**
 * AgentFramework - Enhanced Security Store
 * 安全功能状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useSecurityStore = defineStore('security', () => {
  // ===== State =====
  const isAuthenticated = ref(false)
  const currentUser = ref<string | null>(null)
  const userRoles = ref<string[]>([])
  const jwtToken = ref<string | null>(null)
  const permissions = ref<Record<string, boolean>>({})

  // Security configuration
  const securityConfig = ref({
    jwtEnabled: true,
    rbacEnabled: true,
    inputValidationEnabled: true,
    maxInputLength: 10000,
  })

  // Security statistics
  const stats = ref({
    jwtValidations: 0,
    inputValidations: 0,
    permissionChecks: 0,
    deniedRequests: 0,
  })

  // ===== Getters =====
  const isAdmin = computed(() => {
    return currentUser.value === 'admin' || userRoles.value.includes('admin')
  })

  const canExecuteAgents = computed(() => {
    return permissions.value['agent:execute'] || permissions.value['*:*'] || isAdmin.value
  })

  const canExecuteWorkflows = computed(() => {
    return permissions.value['workflow:execute'] || permissions.value['*:*'] || isAdmin.value
  })

  const canManageSkills = computed(() => {
    return permissions.value['skill:manage'] || permissions.value['*:*'] || isAdmin.value
  })

  const canManageConfig = computed(() => {
    return permissions.value['config:manage'] || permissions.value['*:*'] || isAdmin.value
  })

  // ===== Actions =====

  /**
   * 登录用户
   */
  async login(token: string, userInfo: { id: string; name: string; roles?: string[] }) {
    try {
      // 在实际应用中，这里会调用后端API验证JWT
      // const response = await api.post('/auth/login', { token })

      // 模拟JWT验证成功
      isAuthenticated.value = true
      currentUser.value = userInfo.id
      userRoles.value = userInfo.roles || ['user']
      jwtToken.value = token

      // 加载用户权限
      await loadUserPermissions()

      addNotification('success', '登录成功')
      return true
    } catch (error) {
      addNotification('error', `登录失败: ${error}`)
      return false
    }
  }

  /**
   * 登出用户
   */
  async logout() {
    isAuthenticated.value = false
    currentUser.value = null
    userRoles.value = []
    jwtToken.value = null
    permissions.value = {}

    addNotification('info', '已登出')
  }

  /**
   * 检查权限
   */
  function hasPermission(resource: string, action: string): boolean {
    const key = `${resource}:${action}`
    return permissions.value[key] || permissions.value['*:*'] || isAdmin.value
  }

  /**
   * 需要权限
   */
  function requirePermission(resource: string, action: string): boolean {
    if (!hasPermission(resource, action)) {
      addNotification('warning', `缺少权限: ${resource}:${action}`)
      return false
    }
    return true
  }

  /**
   * 验证输入
   */
  function validateInput(input: string, maxLength?: number): { valid: boolean; error?: string } {
    stats.value.inputValidations++

    if (!input) {
      return { valid: false, error: '输入不能为空' }
    }

    const maxLen = maxLength || securityConfig.value.maxInputLength
    if (input.length > maxLen) {
      return { valid: false, error: `输入长度不能超过 ${maxLen} 个字符` }
    }

    // 检查危险模式
    const dangerousPatterns = [
      /<script/i,
      /javascript:/i,
      /\.\.\\/i,
      /DROP TABLE/i,
      /UNION SELECT/i,
    ]

    for (const pattern of dangerousPatterns) {
      if (pattern.test(input)) {
        return { valid: false, error: `输入包含危险内容: ${pattern}` }
      }
    }

    return { valid: true }
  }

  /**
   * 刷新权限
   */
  async function refreshPermissions() {
    if (!currentUser.value) {
      return
    }

    try {
      // 在实际应用中，这里会调用后端API获取权限
      // const response = await api.get(`/api/users/${currentUser.value}/permissions`)
      // permissions.value = response.data

      addNotification('success', '权限已刷新')
    } catch (error) {
      addNotification('error', `权限刷新失败: ${error}`)
    }
  }

  /**
   * 加载用户权限
   */
  async function loadUserPermissions() {
    if (!currentUser.value) {
      return
    }

    // 根据角色设置默认权限
    const rolePermissions: Record<string, string[]> = {
      admin: ['*:*'],
      user: ['agent:read', 'agent:execute:own', 'workflow:read', 'workflow:execute:own', 'workflow:create'],
      viewer: ['agent:read', 'workflow:read', 'device:read'],
      operator: ['*:read', 'system:restart', 'system:configure', 'logs:read'],
    }

    const userPerms: string[] = []
    for (const role of userRoles.value) {
      if (rolePermissions[role]) {
        userPerms.push(...rolePermissions[role])
      }
    }

    // 将权限转换为权限对象
    const permsObj: Record<string, boolean> = {}
    for (const perm of userPerms) {
      permsObj[perm] = true
    }

    permissions.value = permsObj
  }

  /**
   * 添加通知
   */
  function addNotification(type: 'success' | 'error' | 'warning' | 'info', message: string) {
    // 将通知传递给应用store
    const { useAppStore } = require('./appStore')
    const appStore = useAppStore()
    appStore.addNotification(type, message)
  }

  // ===== Configuration =====

  /**
   * 更新安全配置
   */
  function updateSecurityConfig(config: Partial<typeof securityConfig.value>) {
    Object.assign(securityConfig.value, config)
  }

  /**
   * 初始化安全配置
   */
  function initSecurityConfig() {
    // 从localStorage加载配置
    const saved = localStorage.getItem('securityConfig')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        updateSecurityConfig(config)
      } catch (error) {
        console.error('Failed to parse security config:', error)
      }
    }
  }

  /**
   * 保存安全配置
   */
  function saveSecurityConfig() {
    localStorage.setItem('securityConfig', JSON.stringify(securityConfig.value))
  }

  // Return state and actions
  return {
    // State
    isAuthenticated,
    currentUser,
    userRoles,
    jwtToken,
    permissions,
    securityConfig,
    stats,

    // Getters
    isAdmin,
    canExecuteAgents,
    canExecuteWorkflows,
    canManageSkills,
    canManageConfig,

    // Actions
    login,
    logout,
    hasPermission,
    requirePermission,
    validateInput,
    refreshPermissions,
    updateSecurityConfig,
    initSecurityConfig,
    saveSecurityConfig,
  }
})
