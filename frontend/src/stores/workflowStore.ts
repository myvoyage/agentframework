/**
 * AgentFramework - Workflow Store
 * 工作流状态管理
 * 重构以使用统一的 API 服务层
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { apiService, handleApiError } from '@/services/api'

export interface Workflow {
  id: string
  name: string
  description: string
  type: string
  status: 'running' | 'completed' | 'failed' | 'pending'
  createdAt: Date
  updatedAt: Date
}

export interface WorkflowExecution {
  id: string
  workflowId: string
  status: 'running' | 'completed' | 'failed'
  startTime: Date
  endTime?: Date
  result?: string
  error?: string
}

export const useWorkflowStore = defineStore('workflow', () => {
  // ===== State =====
  const workflows = ref<Workflow[]>([])
  const executions = ref<WorkflowExecution[]>([])
  const selectedWorkflow = ref<Workflow | undefined>(undefined)
  const loading = ref(false)
  const executing = ref(false)
  const error = ref<string | undefined>(undefined)

  // ===== Getters =====
  const activeWorkflows = computed(() => {
    return workflows.value.filter((w) => w.status === 'running')
  })

  const completedWorkflows = computed(() => {
    return workflows.value.filter((w) => w.status === 'completed')
  })

  const failedWorkflows = computed(() => {
    return workflows.value.filter((w) => w.status === 'failed')
  })

  const recentExecutions = computed(() => {
    return executions.value
      .sort((a, b) => b.startTime.getTime() - a.startTime.getTime())
      .slice(0, 10)
  })

  const isReady = computed(() => {
    return !loading.value && workflows.value.length > 0
  })

  // ===== Actions =====

  /**
   * Load all workflows
   */
  async function loadWorkflows() {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.listWorkflows()
      if (!result.success) {
        throw new Error(result.error || 'Failed to load workflows')
      }

      // Transform API response to local format
      workflows.value = (result.data || []).map((wf: any) => ({
        id: wf.id,
        name: wf.name,
        description: wf.description || '',
        type: wf.type || 'sequential',
        status: wf.status || 'pending',
        createdAt: new Date(wf.createdAt || Date.now()),
        updatedAt: new Date(wf.updatedAt || Date.now()),
      }))
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load workflows'
      handleApiError(err, '加载工作流')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Get a single workflow by ID
   */
  async function getWorkflow(id: string) {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.getWorkflow(id)
      if (!result.success) {
        throw new Error(result.error || 'Failed to get workflow')
      }

      const wf = result.data
      return {
        id: wf.id,
        name: wf.name,
        description: wf.description || '',
        type: wf.type || 'sequential',
        status: wf.status || 'pending',
        createdAt: new Date(wf.createdAt || Date.now()),
        updatedAt: new Date(wf.updatedAt || Date.now()),
      } as Workflow
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get workflow'
      handleApiError(err, '获取工作流')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Create a new workflow
   */
  async function createWorkflow(data: {
    name: string
    description: string
    definition: string
  }) {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.createWorkflow(data)
      if (!result.success) {
        throw new Error(result.error || 'Failed to create workflow')
      }

      ElMessage.success('工作流创建成功')
      await loadWorkflows() // Refresh list
      return result.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create workflow'
      handleApiError(err, '创建工作流')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Update an existing workflow
   */
  async function updateWorkflow(
    id: string,
    data: {
      name?: string
      description?: string
      definition?: string
    }
  ) {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.updateWorkflow(id, data)
      if (!result.success) {
        throw new Error(result.error || 'Failed to update workflow')
      }

      ElMessage.success('工作流更新成功')
      await loadWorkflows() // Refresh list
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update workflow'
      handleApiError(err, '更新工作流')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete a workflow
   */
  async function deleteWorkflow(id: string) {
    loading.value = true
    error.value = undefined
    try {
      const result = await apiService.deleteWorkflow(id)
      if (!result.success) {
        throw new Error(result.error || 'Failed to delete workflow')
      }

      ElMessage.success('工作流删除成功')
      await loadWorkflows() // Refresh list
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete workflow'
      handleApiError(err, '删除工作流')
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Execute a workflow
   */
  async function executeWorkflow(workflowId: string, input: string) {
    executing.value = true
    error.value = undefined
    try {
      const result = await apiService.executeWorkflow(workflowId, input)
      if (!result.success) {
        throw new Error(result.error || 'Failed to execute workflow')
      }

      // Create execution record
      const execution: WorkflowExecution = {
        id: result.data.executionId || Date.now().toString(),
        workflowId,
        status: 'running',
        startTime: new Date(),
        result: result.data.output,
      }
      executions.value.unshift(execution)

      ElMessage.success('工作流执行已启动')
      return execution
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to execute workflow'
      handleApiError(err, '执行工作流')
      throw err
    } finally {
      executing.value = false
    }
  }

  /**
   * Get workflow versions
   */
  async function getWorkflowVersions(workflowId: string) {
    try {
      const result = await apiService.getWorkflowVersions(workflowId)
      if (!result.success) {
        throw new Error(result.error || 'Failed to get workflow versions')
      }
      return result.data
    } catch (err) {
      handleApiError(err, '获取工作流版本')
      throw err
    }
  }

  /**
   * Get workflow execution result
   */
  async function getExecutionResult(executionId: string) {
    try {
      const result = await apiService.getWorkflowExecutionResult(executionId)
      if (!result.success) {
        throw new Error(result.error || 'Failed to get execution result')
      }

      // Update execution record
      const index = executions.value.findIndex((e) => e.id === executionId)
      if (index > -1 && result.data) {
        executions.value[index] = {
          ...executions.value[index],
          status: result.data.status || 'completed',
          endTime: new Date(result.data.endTime || Date.now()),
          result: result.data.output,
          error: result.data.error,
        }
      }

      return result.data
    } catch (err) {
      handleApiError(err, '获取执行结果')
      throw err
    }
  }

  /**
   * Select a workflow
   */
  function selectWorkflow(workflow: Workflow | undefined) {
    selectedWorkflow.value = workflow
  }

  /**
   * Reset store state
   */
  function reset() {
    workflows.value = []
    executions.value = []
    selectedWorkflow.value = undefined
    loading.value = false
    executing.value = false
    error.value = undefined
  }

  // ===== WebSocket Event Handling =====

  /**
   * Set up WebSocket event listeners for real-time updates
   */
  function setupWebSocketListeners() {
    // Listen for workflow execution events
    apiService.onWsEvent('workflow_executed', (data: any) => {
      const { id, result } = data
      // Update execution record
      const execution = executions.value.find((e) => e.workflowId === id && e.status === 'running')
      if (execution) {
        execution.status = 'completed'
        execution.result = result
        execution.endTime = new Date()
      }
    })

    // Listen for workflow status changes
    apiService.onWsEvent('workflow_status_changed', (data: any) => {
      const { id, status } = data
      const workflow = workflows.value.find((w) => w.id === id)
      if (workflow) {
        workflow.status = status
      }
    })
  }

  /**
   * Clean up WebSocket listeners
   */
  function cleanupWebSocketListeners() {
    // Remove all workflow-related listeners
    apiService.offWsEvent('workflow_executed', () => {})
    apiService.offWsEvent('workflow_status_changed', () => {})
  }

  return {
    // State
    workflows,
    executions,
    selectedWorkflow,
    loading,
    executing,
    error,

    // Getters
    activeWorkflows,
    completedWorkflows,
    failedWorkflows,
    recentExecutions,
    isReady,

    // Actions
    loadWorkflows,
    getWorkflow,
    createWorkflow,
    updateWorkflow,
    deleteWorkflow,
    executeWorkflow,
    getWorkflowVersions,
    getExecutionResult,
    selectWorkflow,
    reset,
    setupWebSocketListeners,
    cleanupWebSocketListeners,
  }
})
