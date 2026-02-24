/**
 * AgentFramework - Skill System Pinia Store
 * 使用Pinia实现全局状态管理，减少prop drilling
 * 重构以使用统一的 API 服务层
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { apiService, handleApiError } from '@/services/api'
import type {
  SkillListItem,
  SkillDefinitionInfo,
  SkillSystemStats,
  CacheStats,
  PoolStats,
  ExecuteSkillInput,
  ExecuteSkillOutput,
  ExecutionHistoryItem,
} from '@/types/skill'

export const useSkillStore = defineStore('skill', () => {
  // ===== State =====
  const loading = ref(false)
  const refreshing = ref(false)
  const skills = ref<SkillListItem[]>([])
  const definitions = ref<SkillDefinitionInfo[]>([])
  const executionHistory = ref<ExecutionHistoryItem[]>([])
  const selectedSkill = ref<SkillListItem | undefined>(undefined)
  const selectedDefinition = ref<SkillDefinitionInfo | undefined>(undefined)
  const executing = ref(false)
  const executionResult = ref<ExecuteSkillOutput | undefined>(undefined)
  const error = ref<string | undefined>(undefined)

  // Statistics
  const systemStats = ref<SkillSystemStats | undefined>(undefined)
  const cacheStats = ref<CacheStats | undefined>(undefined)
  const poolStats = ref<PoolStats | undefined>(undefined)

  // ===== Getters =====
  const enabledSkills = computed(() => {
    return skills.value.filter((s) => s.enabled)
  })

  const skillCategories = computed(() => {
    const categories = new Set<string>()
    skills.value.forEach((s) => categories.add(s.category))
    return Array.from(categories).sort()
  })

  const totalExecutions = computed(() => {
    return skills.value.reduce((sum, s) => sum + s.useCount, 0)
  })

  const mostUsedSkills = computed(() => {
    return [...skills.value]
      .sort((a, b) => b.useCount - a.useCount)
      .slice(0, 5)
  })

  const isReady = computed(() => {
    return !loading.value && skills.value.length > 0
  })

  // ===== Actions =====

  /**
   * Load all registered skills
   */
  async function loadSkills() {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.listSkills()
      if (!result.success) {
        throw new Error(result.error || 'Failed to load skills')
      }

      // Transform API response to local format
      skills.value = (result.data || []).map((skill: any) => ({
        id: skill.id,
        name: skill.name,
        description: skill.description || '',
        category: skill.metadata?.category || 'general',
        version: skill.version || '1.0.0',
        enabled: skill.enabled !== false,
        useCount: skill.metadata?.useCount || 0,
        createdAt: new Date(skill.metadata?.createdAt || Date.now()),
      }))
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load skills'
      handleApiError(err, '加载技能')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Get a single skill by ID
   */
  async function getSkill(id: string) {
    try {
      const result = await apiService.getSkill(id)
      if (!result.success) {
        throw new Error(result.error || 'Failed to get skill')
      }

      const skill = result.data
      return {
        id: skill.id,
        name: skill.name,
        description: skill.description || '',
        category: skill.metadata?.category || 'general',
        version: skill.version || '1.0.0',
        enabled: skill.enabled !== false,
        useCount: skill.metadata?.useCount || 0,
        createdAt: new Date(skill.metadata?.createdAt || Date.now()),
      } as SkillListItem
    } catch (err) {
      handleApiError(err, '获取技能')
      throw err
    }
  }

  /**
   * Load skill definitions
   * Note: This may need to be implemented in the API if not already available
   */
  async function loadDefinitions() {
    try {
      // For now, we'll skip this as the API might not have this endpoint yet
      // const result = await apiService.getSkillDefinitions()
      definitions.value = []
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to load definitions'
      error.value = errMsg
      handleApiError(err, '加载技能定义')
      throw err
    }
  }

  /**
   * Load system statistics
   * Note: This may need additional API endpoints
   */
  async function loadStats() {
    try {
      // Stats endpoints may need to be added to the API
      // For now, we'll compute basic stats from loaded skills
      systemStats.value = {
        totalSkills: skills.value.length,
        enabledSkills: enabledSkills.value.length,
        totalExecutions: totalExecutions.value,
      }
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to load stats'
      error.value = errMsg
      handleApiError(err, '加载统计信息')
      throw err
    }
  }

  /**
   * Refresh all data
   */
  async function refresh() {
    refreshing.value = true
    try {
      await Promise.all([
        loadSkills(),
        loadDefinitions(),
        loadStats(),
      ])
    } finally {
      refreshing.value = false
    }
  }

  /**
   * Select a skill
   */
  function selectSkill(skill: SkillListItem | undefined) {
    selectedSkill.value = skill
  }

  /**
   * Execute a skill
   * Note: This uses the Wails binding directly for now
   * TODO: Implement skill execution in the API
   */
  async function executeSkill(input: ExecuteSkillInput): Promise<ExecuteSkillOutput> {
    executing.value = true
    error.value = undefined
    executionResult.value = undefined

    try {
      // Import Wails binding dynamically
      const main = await import('../../wailsjs/go/main/App')
      const result = await main.ExecuteSkill(input)

      if (result.error) {
        throw new Error(result.error)
      }

      executionResult.value = result.output

      // Add to execution history
      const historyItem: ExecutionHistoryItem = {
        id: Date.now().toString(),
        skillId: input.skillId,
        skillName: skills.value.find((s) => s.id === input.skillId)?.name || input.skillId,
        input: input.input,
        output: result.output,
        timestamp: new Date(),
        duration: result.output?.duration || 0,
        success: !result.error,
      }
      executionHistory.value.unshift(historyItem)

      // Update skill use count
      const skill = skills.value.find((s) => s.id === input.skillId)
      if (skill) {
        skill.useCount++
      }

      return result.output!
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to execute skill'
      handleApiError(err, '执行技能')
      throw err
    } finally {
      executing.value = false
    }
  }

  /**
   * Toggle skill enabled state
   */
  async function toggleSkill(skillId: string, enabled: boolean) {
    try {
      const result = enabled
        ? await apiService.enableSkill(skillId)
        : await apiService.disableSkill(skillId)

      if (!result.success) {
        throw new Error(result.error || 'Failed to toggle skill')
      }

      const skill = skills.value.find((s) => s.id === skillId)
      if (skill) {
        skill.enabled = enabled
      }

      ElMessage.success(enabled ? '技能已启用' : '技能已禁用')
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to toggle skill'
      error.value = errMsg
      handleApiError(err, '切换技能状态')
      throw err
    }
  }

  /**
   * Register a new skill
   */
  async function registerSkill(skillData: {
    id: string
    name: string
    description: string
    metadata?: Record<string, any>
  }) {
    try {
      const result = await apiService.registerSkill(skillData)
      if (!result.success) {
        throw new Error(result.error || 'Failed to register skill')
      }

      ElMessage.success('技能注册成功')
      await loadSkills() // Refresh list
      return result.data
    } catch (err) {
      handleApiError(err, '注册技能')
      throw err
    }
  }

  /**
   * Delete a skill
   */
  async function deleteSkill(skillId: string) {
    try {
      const result = await apiService.deleteSkill(skillId)
      if (!result.success) {
        throw new Error(result.error || 'Failed to delete skill')
      }

      ElMessage.success('技能删除成功')
      await loadSkills() // Refresh list
    } catch (err) {
      handleApiError(err, '删除技能')
      throw err
    }
  }

  /**
   * Clear execution history
   */
  function clearHistory() {
    executionHistory.value = []
  }

  /**
   * Reset store state
   */
  function reset() {
    loading.value = false
    refreshing.value = false
    skills.value = []
    definitions.value = []
    executionHistory.value = []
    selectedSkill.value = undefined
    selectedDefinition.value = undefined
    executing.value = false
    executionResult.value = undefined
    error.value = undefined
    systemStats.value = undefined
    cacheStats.value = undefined
    poolStats.value = undefined
  }

  // ===== WebSocket Event Handling =====

  /**
   * Set up WebSocket event listeners for real-time updates
   */
  function setupWebSocketListeners() {
    // Listen for skill execution events
    apiService.onWsEvent('skill_executed', (data: any) => {
      const { skillId, result } = data
      // Update skill statistics
      const skill = skills.value.find((s) => s.id === skillId)
      if (skill) {
        skill.useCount++
      }
    })

    // Listen for skill registration events
    apiService.onWsEvent('skill_registered', (data: any) => {
      // Refresh skill list
      loadSkills()
    })

    // Listen for skill toggle events
    apiService.onWsEvent('skill_toggled', (data: any) => {
      const { skillId, enabled } = data
      const skill = skills.value.find((s) => s.id === skillId)
      if (skill) {
        skill.enabled = enabled
      }
    })
  }

  /**
   * Clean up WebSocket listeners
   */
  function cleanupWebSocketListeners() {
    apiService.offWsEvent('skill_executed', () => {})
    apiService.offWsEvent('skill_registered', () => {})
    apiService.offWsEvent('skill_toggled', () => {})
  }

  return {
    // State
    loading,
    refreshing,
    skills,
    definitions,
    executionHistory,
    selectedSkill,
    selectedDefinition,
    executing,
    executionResult,
    error,
    systemStats,
    cacheStats,
    poolStats,

    // Getters
    enabledSkills,
    skillCategories,
    totalExecutions,
    mostUsedSkills,
    isReady,

    // Actions
    loadSkills,
    getSkill,
    loadDefinitions,
    loadStats,
    refresh,
    selectSkill,
    executeSkill,
    toggleSkill,
    registerSkill,
    deleteSkill,
    clearHistory,
    reset,
    setupWebSocketListeners,
    cleanupWebSocketListeners,
  }
})
