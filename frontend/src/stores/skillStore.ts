/**
 * AgentFramework - Skill System Pinia Store
 * 使用Pinia实现全局状态管理，减少prop drilling
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import * as main from '../../wailsjs/go/main/App'
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
      const result = await main.ListSkills()
      if (result.error) {
        throw new Error(result.error)
      }
      skills.value = result.skills || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load skills'
      ElMessage.error(`加载技能失败: ${error.value}`)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Load skill definitions
   */
  async function loadDefinitions() {
    try {
      const result = await main.ListSkillDefinitions()
      if (result.error) {
        throw new Error(result.error)
      }
      definitions.value = result.definitions || []
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to load definitions'
      ElMessage.error(`加载技能定义失败: ${errMsg}`)
      throw err
    }
  }

  /**
   * Load system statistics
   */
  async function loadStats() {
    try {
      const statsResult = await main.GetSkillSystemStats()
      if (statsResult.error) {
        throw new Error(statsResult.error)
      }
      systemStats.value = statsResult.stats

      const cacheResult = await main.GetCacheStats()
      if (!cacheResult.error && cacheResult.stats) {
        cacheStats.value = cacheResult.stats
      }

      const poolResult = await main.GetPoolStats()
      if (!poolResult.error && poolResult.stats) {
        poolStats.value = poolResult.stats
      }
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to load stats'
      ElMessage.error(`加载统计信息失败: ${errMsg}`)
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
   */
  async function executeSkill(input: ExecuteSkillInput): Promise<ExecuteSkillOutput> {
    executing.value = true
    error.value = undefined
    executionResult.value = undefined

    try {
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
      ElMessage.error(`执行技能失败: ${error.value}`)
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
      const result = await main.ToggleSkill(skillId, enabled)
      if (result.error) {
        throw new Error(result.error)
      }

      const skill = skills.value.find((s) => s.id === skillId)
      if (skill) {
        skill.enabled = enabled
      }
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to toggle skill'
      ElMessage.error(`切换技能状态失败: ${errMsg}`)
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
    loadDefinitions,
    loadStats,
    refresh,
    selectSkill,
    executeSkill,
    toggleSkill,
    clearHistory,
    reset,
  }
})
