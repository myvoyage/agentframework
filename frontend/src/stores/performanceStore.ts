/**
 * AgentFramework - Performance Store
 * 性能监控状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface PerformanceMetrics {
  timestamp: number
  requestCount: number
  errorCount: number
  averageLatency: number
  minLatency: number
  maxLatency: number
  throughput: number
  cacheHitRate: number
  poolReuseRate: number
}

interface PoolStats {
  messageAllocated: number
  messagePooled: number
  messageReused: number
  eventAllocated: number
  eventPooled: number
  eventReused: number
}

export const usePerformanceStore = defineStore('performance', () => {
  // ===== State =====
  const metrics = ref<PerformanceMetrics>({
    timestamp: Date.now(),
    requestCount: 0,
    errorCount: 0,
    averageLatency: 0,
    minLatency: 0,
    maxLatency: 0,
    throughput: 0,
    cacheHitRate: 0,
    poolReuseRate: 0,
  })

  const poolStats = ref<PoolStats>({
    messageAllocated: 0,
    messagePooled: 0,
    messageReused: 0,
    eventAllocated: 0,
    eventPooled: 0,
    eventReused: 0,
  })

  const isMonitoring = ref(false)
  const updateInterval = ref<ReturnType<typeof setInterval> | null>(null)

  // ===== Getters =====
  const errorRate = computed(() => {
    if (metrics.value.requestCount === 0) return 0
    return (metrics.value.errorCount / metrics.value.requestCount) * 100
  })

  const cacheEfficiency = computed(() => metrics.value.cacheHitRate)

  const poolEfficiency = computed(() => metrics.value.poolReuseRate)

  const systemHealth = computed(() => {
    const errorRateThreshold = 5 // 5% error rate threshold
    const cacheHitThreshold = 70 // 70% cache hit threshold
    const poolReuseThreshold = 50 // 50% pool reuse threshold

    const errorRateOK = errorRate.value < errorRateThreshold
    const cacheHitOK = metrics.value.cacheHitRate >= cacheHitThreshold
    const poolReuseOK = metrics.value.poolReuseRate >= poolReuseThreshold

    if (errorRateOK && cacheHitOK && poolReuseOK) {
      return 'healthy'
    } else if (errorRate > errorRateThreshold * 2) {
      return 'critical'
    } else {
      return 'warning'
    }
  })

  // ===== Actions =====

  /**
   * 开始监控性能
   */
  function startMonitoring(interval: number = 5000) {
    if (isMonitoring.value) {
      return
    }

    isMonitoring.value = true
    updateInterval.value = setInterval(() => {
      updateMetrics()
    }, interval)

    console.log(`📊 性能监控已启动 (间隔: ${interval}ms)`)
  }

  /**
   * 停止监控性能
   */
  function stopMonitoring() {
    if (!isMonitoring.value) {
      return
    }

    if (updateInterval.value) {
      clearInterval(updateInterval.value)
      updateInterval.value = null
    }

    isMonitoring.value = false
    console.log('⏸️  性能监控已停止')
  }

  /**
   * 更新性能指标
   */
  async function updateMetrics() {
    try {
      // 在实际应用中，这里会调用后端API获取性能数据
      // const response = await api.get('/api/metrics')

      // 模拟更新（实际应用中从后端获取）
      metrics.value.timestamp = Date.now()
      // 更新其他指标...

    } catch (error) {
      console.error('获取性能指标失败:', error)
    }
  }

  /**
   * 模拟更新性能指标（演示用）
   */
  function simulateMetrics() {
    metrics.value.requestCount += Math.floor(Math.random() * 10)
    metrics.value.errorCount += Math.random() > 0.95 ? 1 : 0
    metrics.value.averageLatency = Math.floor(Math.random() * 100) + 10
    metrics.value.minLatency = Math.min(metrics.value.minLatency, metrics.value.averageLatency)
    metrics.value.maxLatency = Math.max(metrics.value.maxLatency, metrics.value.averageLatency)
    metrics.value.throughput = Math.floor(Math.random() * 1000) + 500
    metrics.value.cacheHitRate = Math.min(100, Math.floor(metrics.value.cacheHitRate + Math.random() * 5))

    // 计算池复用率
    if (poolStats.value.messageAllocated > 0) {
      metrics.value.poolReuseRate = (poolStats.value.messageReused / poolStats.value.messageAllocated) * 100
    }
  }

  /**
   * 获取性能报告
   */
  function getPerformanceReport() {
    return {
      summary: getPerformanceSummary(),
      details: metrics.value,
      pools: poolStats.value,
      recommendations: getRecommendations(),
    }
  }

  /**
   * 获取性能摘要
   */
  function getPerformanceSummary() {
    const health = systemHealth.value
    const healthStatus = {
      healthy: { icon: '✅', color: 'success', text: '健康' },
      warning: { icon: '⚠️', color: 'warning', text: '警告' },
      critical: { icon: '🔴', color: 'danger', text: '严重' },
    }

    const status = healthStatus[health] || healthStatus.healthy

    return {
      status,
      health,
      errorRate: errorRate.value,
      cacheHitRate: metrics.value.cacheHitRate,
      poolReuseRate: metrics.value.poolReuseRate,
    }
  }

  /**
   * 获取优化建议
   */
  function getRecommendations() {
    const recommendations: string[] = []

    // 错误率检查
    if (errorRate.value > 5) {
      recommendations.push('错误率较高，请检查系统日志')
    }

    // 缓存命中率检查
    if (metrics.value.cacheHitRate < 70) {
      recommendations.push('缓存命中率较低，考虑增加缓存容量')
    }

    // 对象池复用率检查
    if (metrics.value.poolReuseRate < 50) {
      recommendations.push('对象池复用率较低，建议优化对象获取和释放')
    }

    // 延迟检查
    if (metrics.value.averageLatency > 100) {
      recommendations.push('平均延迟较高，建议优化关键路径')
    }

    return recommendations
  }

  /**
   * 重置性能统计
   */
  function resetStats() {
    metrics.value = {
      timestamp: Date.now(),
      requestCount: 0,
      errorCount: 0,
      averageLatency: 0,
      minLatency: 0,
      maxLatency: 0,
      throughput: 0,
      cacheHitRate: 0,
      poolReuseRate: 0,
    }
  }

  /**
   * 导出性能数据
   */
  function exportMetrics() {
    const data = {
      metrics: metrics.value,
      pools: poolStats.value,
      exportTime: new Date().toISOString(),
    }

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `performance-metrics-${Date.now()}.json`
    link.click()

    URL.revokeObjectURL(url)
    console.log('📊 性能数据已导出')
  }

  // Initialize from localStorage
  function initFromStorage() {
    const saved = localStorage.getItem('performanceConfig')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        // 应用配置...
      } catch (error) {
        console.error('Failed to load performance config:', error)
      }
    }
  }

  return {
    // State
    metrics,
    poolStats,
    isMonitoring,
    updateInterval,

    // Getters
    errorRate,
    cacheEfficiency,
    poolEfficiency,
    systemHealth,

    // Actions
    startMonitoring,
    stopMonitoring,
    updateMetrics,
    simulateMetrics,
    getPerformanceReport,
    getPerformanceSummary,
    getRecommendations,
    resetStats,
    exportMetrics,
    initFromStorage,
  }
})
