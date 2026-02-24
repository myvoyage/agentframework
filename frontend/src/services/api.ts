/**
 * AgentFramework - Unified API Service
 * 统一 API 服务层，支持 Wails 和 HTTP 两种通信方式
 */

import { ElMessage } from 'element-plus'

// API 配置
export interface ApiConfig {
  // 使用 HTTP API（默认 false，使用 Wails）
  useHttp: boolean
  // HTTP API 基础 URL
  baseUrl: string
  // 请求超时时间（毫秒）
  timeout: number
}

// 默认配置
const defaultConfig: ApiConfig = {
  useHttp: false,
  baseUrl: 'http://localhost:8080',
  timeout: 30000,
}

let currentConfig = { ...defaultConfig }

// API 响应类型
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// WebSocket 消息类型
export interface WebSocketMessage {
  type: string
  data?: any
  timestamp: number
}

// WebSocket 客户端
class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string = ''
  private reconnectTimer: NodeJS.Timeout | null = null
  private messageHandlers: Map<string, ((data: any) => void)[]> = new Map()

  connect(url: string) {
    this.url = url
    this.disconnect()

    try {
      this.ws = new WebSocket(url)

      this.ws.onopen = () => {
        console.log('[WebSocket] Connected to', url)
        this.clearReconnectTimer()
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (err) {
          console.error('[WebSocket] Failed to parse message:', err)
        }
      }

      this.ws.onclose = () => {
        console.log('[WebSocket] Disconnected')
        this.scheduleReconnect()
      }

      this.ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error)
      }
    } catch (err) {
      console.error('[WebSocket] Failed to connect:', err)
      this.scheduleReconnect()
    }
  }

  disconnect() {
    this.clearReconnectTimer()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  private scheduleReconnect() {
    this.clearReconnectTimer()
    this.reconnectTimer = setTimeout(() => {
      console.log('[WebSocket] Attempting to reconnect...')
      this.connect(this.url)
    }, 5000)
  }

  on(messageType: string, handler: (data: any) => void) {
    if (!this.messageHandlers.has(messageType)) {
      this.messageHandlers.set(messageType, [])
    }
    this.messageHandlers.get(messageType)!.push(handler)
  }

  off(messageType: string, handler: (data: any) => void) {
    const handlers = this.messageHandlers.get(messageType)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  private handleMessage(message: WebSocketMessage) {
    const handlers = this.messageHandlers.get(message.type)
    if (handlers) {
      handlers.forEach(handler => handler(message.data))
    }
  }

  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }
}

// HTTP API 客户端
class HttpClient {
  private baseUrl: string
  private timeout: number

  constructor(config: ApiConfig) {
    this.baseUrl = config.baseUrl
    this.timeout = config.timeout
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), this.timeout)

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      })

      clearTimeout(timeoutId)

      const data = await response.json()

      if (response.ok) {
        return {
          success: true,
          data: data.data || data,
        }
      } else {
        return {
          success: false,
          error: data.message || data.error || `HTTP ${response.status}`,
        }
      }
    } catch (err) {
      clearTimeout(timeoutId)
      return {
        success: false,
        error: err instanceof Error ? err.message : 'Network error',
      }
    }
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  async post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

// Wails API 客户端（使用现有的 wailsjs 绑定）
class WailsClient {
  async importMain() {
    return import('../../wailsjs/go/main/App')
  }

  async get<T>(method: string, ...args: any[]): Promise<ApiResponse<T>> {
    try {
      const main = await this.importMain()
      // @ts-ignore - 动态调用 Wails 方法
      const result = await main[method](...args)

      if (result && result.error) {
        return {
          success: false,
          error: result.error,
        }
      }

      return {
        success: true,
        data: result,
      }
    } catch (err) {
      return {
        success: false,
        error: err instanceof Error ? err.message : 'Wails call failed',
      }
    }
  }

  async post<T>(method: string, data?: any): Promise<ApiResponse<T>> {
    try {
      const main = await this.importMain()
      // @ts-ignore - 动态调用 Wails 方法
      const result = await main[method](data)

      if (result && result.error) {
        return {
          success: false,
          error: result.error,
        }
      }

      return {
        success: true,
        data: result,
      }
    } catch (err) {
      return {
        success: false,
        error: err instanceof Error ? err.message : 'Wails call failed',
      }
    }
  }
}

// 统一 API 服务
class ApiService {
  private http: HttpClient | null = null
  private wails: WailsClient
  private ws: WebSocketClient

  constructor() {
    this.wails = new WailsClient()
    this.ws = new WebSocketClient()

    if (currentConfig.useHttp) {
      this.http = new HttpClient(currentConfig)
    }
  }

  // 配置管理
  configure(config: Partial<ApiConfig>) {
    currentConfig = { ...currentConfig, ...config }

    if (currentConfig.useHttp && !this.http) {
      this.http = new HttpClient(currentConfig)

      // 连接 WebSocket
      const wsUrl = currentConfig.baseUrl
        .replace('http://', 'ws://')
        .replace('https://', 'wss://')
      this.ws.connect(`${wsUrl}/ws`)
    }
  }

  getConfig(): ApiConfig {
    return { ...currentConfig }
  }

  // WebSocket 事件监听
  onWsEvent(eventType: string, handler: (data: any) => void) {
    this.ws.on(eventType, handler)
  }

  offWsEvent(eventType: string, handler: (data: any) => void) {
    this.ws.off(eventType, handler)
  }

  // 通用请求方法
  private async request<T>(
    wailsMethod: string,
    httpEndpoint: string,
    method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET',
    ...args: any[]
  ): Promise<ApiResponse<T>> {
    if (currentConfig.useHttp && this.http) {
      // 使用 HTTP API
      switch (method) {
        case 'GET':
          return this.http.get<T>(httpEndpoint)
        case 'POST':
          return this.http.post<T>(httpEndpoint, args[0])
        case 'PUT':
          return this.http.put<T>(httpEndpoint, args[0])
        case 'DELETE':
          return this.http.delete<T>(httpEndpoint)
      }
    } else {
      // 使用 Wails
      if (method === 'GET') {
        return this.wails.get<T>(wailsMethod, ...args)
      } else {
        return this.wails.post<T>(wailsMethod, args[0])
      }
    }

    return {
      success: false,
      error: 'Invalid request configuration',
    }
  }

  // ========== 工作流 API ==========

  async listWorkflows() {
    return this.request('GetWorkflows', '/api/workflows', 'GET')
  }

  async getWorkflow(id: string) {
    return this.request('GetWorkflow', `/api/workflows/${id}`, 'GET', id)
  }

  async createWorkflow(data: { name: string; description: string; definition: string }) {
    return this.request('CreateWorkflow', '/api/workflows', 'POST', data.name, data.description, data.definition)
  }

  async updateWorkflow(id: string, data: { name?: string; description?: string; definition?: string }) {
    return this.request('UpdateWorkflow', `/api/workflows/${id}`, 'PUT', id, data.name || '', data.description || '', data.definition || '')
  }

  async deleteWorkflow(id: string) {
    return this.request('DeleteWorkflow', `/api/workflows/${id}`, 'DELETE', id)
  }

  async executeWorkflow(id: string, input: string) {
    return this.request('ExecuteWorkflow', `/api/workflows/${id}/execute`, 'POST', { input })
  }

  async getWorkflowVersions(workflowId: string) {
    return this.request('GetWorkflowVersions', `/api/workflows/${workflowId}/versions`, 'GET', workflowId)
  }

  async getWorkflowExecutionResult(executionId: string) {
    return this.request('GetWorkflowExecutionResult', `/api/workflows/executions/${executionId}`, 'GET', executionId)
  }

  // ========== 技能 API ==========

  async listSkills() {
    return this.request('ListSkills', '/api/skills', 'GET')
  }

  async getSkill(id: string) {
    return this.request('GetSkillInfo', `/api/skills/${id}`, 'GET', id)
  }

  async enableSkill(id: string) {
    return this.request('EnableSkill', `/api/skills/${id}/enable`, 'PUT', id)
  }

  async disableSkill(id: string) {
    return this.request('DisableSkill', `/api/skills/${id}/disable`, 'PUT', id)
  }

  async deleteSkill(id: string) {
    return this.request('DeleteSkill', `/api/skills/${id}`, 'DELETE', id)
  }

  async registerSkill(skillData: any) {
    return this.request('RegisterSkillFromMap', '/api/skills', 'POST', skillData)
  }

  // ========== Agent API ==========

  async listAgents() {
    return this.request('ListAgents', '/api/agents', 'GET')
  }

  async getAgent(id: string) {
    return this.request('GetAgentInfo', `/api/agents/${id}`, 'GET', id)
  }

  async chatWithAgent(agentId: string, message: string) {
    return this.request('ChatWithAgent', `/api/agents/${agentId}/chat`, 'POST', { message })
  }

  // ========== 文件系统 API ==========

  async listFiles(path: string, depth: number = 1) {
    return this.request('ListFiles', `/api/files?path=${encodeURIComponent(path)}&depth=${depth}`, 'GET', path, depth)
  }

  async readFile(path: string) {
    return this.request('ReadFile', `/api/files?path=${encodeURIComponent(path)}`, 'GET', path)
  }

  async writeFile(path: string, content: string, mode: string = 'overwrite') {
    return this.request('CreateFile', '/api/files', 'POST', { path, content, mode })
  }

  async deleteFile(path: string) {
    return this.request('DeleteFile', `/api/files?path=${encodeURIComponent(path)}`, 'DELETE', path)
  }

  async copyFile(src: string, dst: string) {
    return this.request('CopyFile', '/api/files/copy', 'POST', { src, dst })
  }

  async createDirectory(path: string) {
    return this.request('CreateDirectory', '/api/files/directory', 'POST', { path })
  }

  // ========== 配置 API ==========

  async getConfig() {
    return this.request('GetConfig', '/api/config', 'GET')
  }

  async updateConfig(config: any) {
    return this.request('UpdateConfig', '/api/config', 'PUT', config)
  }

  // ========== 健康检查 ==========

  async healthCheck() {
    if (currentConfig.useHttp && this.http) {
      return this.http.get('/health')
    }
    return { success: true, data: { status: 'healthy' } }
  }
}

// 单例实例
export const apiService = new ApiService()

// 配置快捷方法
export function configureApi(config: Partial<ApiConfig>) {
  apiService.configure(config)
}

export function getApiConfig(): ApiConfig {
  return apiService.getConfig()
}

// 错误处理辅助函数
export function handleApiError(error: any, context: string = '操作') {
  const message = error?.error || error?.message || `${context}失败`
  ElMessage.error(message)
  console.error(`[API] ${context} error:`, error)
}
