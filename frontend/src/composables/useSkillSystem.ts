/**
 * AgentFramework - Skill System Composable
 * 技能系统的状态管理
 */

import { ref, computed, reactive } from 'vue'
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
  SkillSystemUIState,
  SkillSystemActions,
} from '@/types/skill'

export function useSkillSystem() {
  // ===== 状态 =====
  const state = reactive<SkillSystemUIState>({
    // 加载状态
    loading: false,
    refreshing: false,

    // 数据
    skills: [],
    definitions: [],
    executionHistory: [],

    // 选中的技能
    selectedSkill: undefined,
    selectedDefinition: undefined,

    // 执行状态
    executing: false,
    executionResult: undefined,

    // 错误
    error: undefined,
  })

  // 统计数据（独立响应式对象）
  const systemStats = ref<SkillSystemStats | undefined>(undefined)
  const cacheStats = ref<CacheStats | undefined>(undefined)
  const poolStats = ref<PoolStats | undefined>(undefined)

  // ===== 计算属性 =====
  const enabledSkills = computed(() => {
    return state.skills.filter((s) => s.enabled)
  })

  const skillCategories = computed(() => {
    const categories = new Set<string>()
    state.skills.forEach((s) => categories.add(s.category))
    return Array.from(categories).sort()
  })

  const totalExecutions = computed(() => {
    return state.skills.reduce((sum, s) => sum + s.useCount, 0)
  })

  const mostUsedSkills = computed(() => {
    return [...state.skills]
      .sort((a, b) => b.useCount - a.useCount)
      .slice(0, 5)
  })

  const isReady = computed(() => {
    return !state.loading && state.skills.length > 0
  })

  // ===== 数据加载方法 =====

  /**
   * 加载所有已注册的技能
   */
  const loadSkills = async () => {
    state.loading = true
    state.error = undefined

    try {
      const skills = await main.ListRegisteredSkills()
      state.skills = skills || []
      console.log(`[useSkillSystem] 加载了 ${state.skills.length} 个技能`)
    } catch (error: any) {
      console.error('[useSkillSystem] 加载技能列表失败:', error)
      state.error = error.message || '加载技能列表失败'
      ElMessage.error('加载技能列表失败')
    } finally {
      state.loading = false
    }
  }

  /**
   * 加载所有技能定义
   */
  const loadDefinitions = async () => {
    try {
      const definitions = await main.ListSkillDefinitions()
      state.definitions = definitions || []
      console.log(`[useSkillSystem] 加载了 ${state.definitions.length} 个技能定义`)
    } catch (error: any) {
      console.error('[useSkillSystem] 加载技能定义失败:', error)
      ElMessage.error('加载技能定义失败')
    }
  }

  /**
   * 加载系统统计数据
   */
  const loadStats = async () => {
    try {
      // 加载技能系统统计
      const sysStats = await main.GetSkillSystemStats()
      systemStats.value = sysStats

      // 加载缓存统计
      const cache = await main.GetCacheStats()
      cacheStats.value = cache

      // 加载连接池统计
      const pool = await main.GetPoolStats()
      poolStats.value = pool

      console.log('[useSkillSystem] 统计数据已更新')
    } catch (error: any) {
      console.error('[useSkillSystem] 加载统计数据失败:', error)
    }
  }

  /**
   * 刷新所有数据
   */
  const refresh = async () => {
    state.refreshing = true
    try {
      await Promise.all([loadSkills(), loadDefinitions(), loadStats()])
      ElMessage.success('数据已刷新')
    } catch (error) {
      ElMessage.error('刷新失败')
    } finally {
      state.refreshing = false
    }
  }

  // ===== 技能操作方法 =====

  /**
   * 选择技能
   */
  const selectSkill = (skill: SkillListItem) => {
    state.selectedSkill = skill
    console.log(`[useSkillSystem] 选择了技能: ${skill.name}`)
  }

  /**
   * 选择技能定义
   */
  const selectDefinition = async (definitionId: string) => {
    try {
      const definition = await main.GetSkillDefinition(definitionId)
      state.selectedDefinition = definition
    } catch (error: any) {
      console.error('[useSkillSystem] 加载技能定义失败:', error)
      ElMessage.error('加载技能定义失败')
    }
  }

  /**
   * 执行技能
   */
  const executeSkill = async (input: ExecuteSkillInput): Promise<ExecuteSkillOutput> => {
    state.executing = true
    state.executionResult = undefined
    state.error = undefined

    try {
      console.log(`[useSkillSystem] 执行技能: ${input.skillName}`)
      const result = await main.ExecuteSkillByName(input)

      state.executionResult = result

      // 记录执行历史
      state.executionHistory.unshift({
        id: Date.now().toString(),
        skillName: input.skillName,
        input: input.input,
        output: result.result,
        error: result.error,
        timestamp: new Date().toISOString(),
        duration: 0, // TODO: 从后端获取
        success: result.success,
      })

      // 只保留最近 100 条记录
      if (state.executionHistory.length > 100) {
        state.executionHistory = state.executionHistory.slice(0, 100)
      }

      if (!result.success) {
        throw new Error(result.error)
      }

      return result
    } catch (error: any) {
      state.error = error.message || '执行技能失败'
      console.error('[useSkillSystem] 执行技能失败:', error)
      ElMessage.error(`执行技能失败: ${error.message}`)
      throw error
    } finally {
      state.executing = false
    }
  }

  /**
   * 切换技能状态
   */
  const toggleSkill = async (skill: SkillListItem) => {
    try {
      // 调用后端 API 切换技能状态
      const newState = await main.ToggleSkill(skill.id)

      // 更新前端状态
      const index = state.skills.findIndex((s) => s.id === skill.id)
      if (index !== -1) {
        state.skills[index].enabled = newState
        ElMessage.success(`技能 ${skill.name} 已${newState ? '启用' : '禁用'}`)
      }
    } catch (error: any) {
      console.error('[useSkillSystem] 切换技能状态失败:', error)
      ElMessage.error(`操作失败: ${error.message}`)
      throw error
    }
  }

  // ===== 系统操作方法 =====

  /**
   * 清空缓存
   */
  const clearCache = async () => {
    try {
      await main.ClearCache()
      await loadStats() // 重新加载统计数据
      ElMessage.success('缓存已清空')
    } catch (error: any) {
      console.error('[useSkillSystem] 清空缓存失败:', error)
      ElMessage.error('清空缓存失败')
    }
  }

  /**
   * 重新加载技能定义
   */
  const reloadDefinitions = async () => {
    try {
      await main.ReloadSkillDefinitions()
      await loadDefinitions()
      ElMessage.success('技能定义已重新加载')
    } catch (error: any) {
      console.error('[useSkillSystem] 重新加载技能定义失败:', error)
      ElMessage.error('重新加载技能定义失败')
    }
  }

  /**
   * 清空执行历史
   */
  const clearHistory = () => {
    state.executionHistory = []
    ElMessage.info('执行历史已清空')
  }

  // ===== 工具方法 =====

  /**
   * 获取技能详情
   */
  const getSkillById = (id: string): SkillListItem | undefined => {
    return state.skills.find((s) => s.id === id)
  }

  /**
   * 根据分类筛选技能
   */
  const getSkillsByCategory = (category: string): SkillListItem[] => {
    return state.skills.filter((s) => s.category === category)
  }

  /**
   * 搜索技能
   */
  const searchSkills = (keyword: string): SkillListItem[] => {
    const lowerKeyword = keyword.toLowerCase()
    return state.skills.filter(
      (s) =>
        s.name.toLowerCase().includes(lowerKeyword) ||
        s.description.toLowerCase().includes(lowerKeyword) ||
        s.tags.some((tag) => tag.toLowerCase().includes(lowerKeyword))
    )
  }

  /**
   * 重置状态
   */
  const reset = () => {
    state.loading = false
    state.refreshing = false
    state.skills = []
    state.definitions = []
    state.executionHistory = []
    state.selectedSkill = undefined
    state.selectedDefinition = undefined
    state.executing = false
    state.executionResult = undefined
    state.error = undefined

    systemStats.value = undefined
    cacheStats.value = undefined
    poolStats.value = undefined
  }

  // 返回状态和方法
  return {
    // 状态
    state,
    systemStats,
    cacheStats,
    poolStats,

    // 计算属性
    enabledSkills,
    skillCategories,
    totalExecutions,
    mostUsedSkills,
    isReady,

    // 方法
    loadSkills,
    loadDefinitions,
    loadStats,
    refresh,

    selectSkill,
    selectDefinition,
    executeSkill,
    toggleSkill,

    clearCache,
    reloadDefinitions,
    clearHistory,

    getSkillById,
    getSkillsByCategory,
    searchSkills,
    reset,
  }
}

// 导出类型
export type UseSkillSystemReturn = ReturnType<typeof useSkillSystem>
