/**
 * AgentFramework - Workflow Store
 * 工作流状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as main from '../../wailsjs/go/main/App'

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

  // ===== Actions =====
  async function loadWorkflows() {
    loading.value = true
    error.value = undefined
    try {
      // TODO: Implement API call when available
      // const result = await main.ListWorkflows()
      workflows.value = []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load workflows'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function executeWorkflow(workflowId: string, input: string) {
    executing.value = true
    error.value = undefined
    try {
      // TODO: Implement API call when available
      // const result = await main.ExecuteWorkflow(workflowId, input)
      const execution: WorkflowExecution = {
        id: Date.now().toString(),
        workflowId,
        status: 'running',
        startTime: new Date(),
      }
      executions.value.unshift(execution)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to execute workflow'
      throw err
    } finally {
      executing.value = false
    }
  }

  function selectWorkflow(workflow: Workflow | undefined) {
    selectedWorkflow.value = workflow
  }

  function reset() {
    workflows.value = []
    executions.value = []
    selectedWorkflow.value = undefined
    loading.value = false
    executing.value = false
    error.value = undefined
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
    
    // Actions
    loadWorkflows,
    executeWorkflow,
    selectWorkflow,
    reset,
  }
})
