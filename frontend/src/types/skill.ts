/**
 * AgentFramework - Skill System Type Definitions
 * 技能系统的 TypeScript 类型定义
 */

// ===== 基础类型 =====

/**
 * 技能系统信息
 */
export interface SkillSystemInfo {
  initialized: boolean
  baseDir: string
  totalSkills: number
}

/**
 * 技能列表项
 */
export interface SkillListItem {
  id: string
  name: string
  description: string
  category: string
  tags: string[]
  version: string
  enabled: boolean
  useCount: number
  lastUsed: string
}

/**
 * 技能定义信息
 */
export interface SkillDefinitionInfo {
  id: string
  name: string
  description: string
  version: string
  category: string
  author: string
  license: string
  workflow: Record<string, any>
  config: Record<string, any>
}

/**
 * 工作流步骤信息
 */
export interface WorkflowStepInfo {
  name: string
  action: string
  timeout?: string
  skip_if?: string
}

/**
 * 技能配置
 */
export interface SkillConfig {
  cache_enabled: boolean
  cache_ttl: string
  default_timeout: string
  max_retries: number
}

/**
 * 执行技能输入
 */
export interface ExecuteSkillInput {
  skillName: string
  input: string
  workspace: string
  parameters?: Record<string, any>
}

/**
 * 执行技能输出
 */
export interface ExecuteSkillOutput {
  success: boolean
  result?: any
  error?: string
  stats?: Record<string, any>
}

/**
 * 技能系统统计
 */
export interface SkillSystemStats {
  totalSkills: number
  totalUses: number
  categories: Record<string, number>
  mostUsedSkills: Array<{
    id: string
    name: string
    count: number
  }>
}

/**
 * 缓存统计
 */
export interface CacheStats {
  hitRate: number
  hits: number
  misses: number
  totalEntries: number
  size: number
}

/**
 * 连接池统计
 */
export interface PoolStats {
  activeConnections: number
  idleConnections: number
  maxConnections: number
  minConnections: number
  utilizationRate: number
}

// ===== 执行历史相关 =====

/**
 * 执行历史记录
 */
export interface ExecutionHistoryItem {
  id: string
  skillName: string
  input: string
  output?: any
  error?: string
  timestamp: string
  duration: number
  success: boolean
}

// ===== 性能监控相关 =====

/**
 * 性能指标卡片
 */
export interface PerformanceCard {
  title: string
  value: string | number
  unit?: string
  trend?: 'up' | 'down' | 'stable'
  trendValue?: number
  icon: string
  color: string
}

/**
 * 系统性能概览
 */
export interface PerformanceOverview {
  cacheHitRate: number
  poolUtilization: number
  totalExecutions: number
  avgExecutionTime: number
}

// ===== 表单相关 =====

/**
 * 表单字段定义
 */
export interface FormField {
  name: string
  label: string
  type: 'string' | 'number' | 'boolean' | 'array' | 'object'
  required: boolean
  description?: string
  default?: any
  options?: Array<{ label: string; value: any }>
}

/**
 * 表单数据
 */
export type FormData = Record<string, any>

// ===== UI 状态相关 =====

/**
 * 技能系统 UI 状态
 */
export interface SkillSystemUIState {
  // 加载状态
  loading: boolean
  refreshing: boolean

  // 数据
  skills: SkillListItem[]
  definitions: SkillDefinitionInfo[]
  executionHistory: ExecutionHistoryItem[]

  // 统计
  systemStats?: SkillSystemStats
  cacheStats?: CacheStats
  poolStats?: PoolStats

  // 选中的技能
  selectedSkill?: SkillListItem
  selectedDefinition?: SkillDefinitionInfo

  // 执行状态
  executing: boolean
  executionResult?: ExecuteSkillOutput

  // 错误
  error?: string
}

/**
 * 技能系统操作
 */
export interface SkillSystemActions {
  // 加载数据
  loadSkills: () => Promise<void>
  loadDefinitions: () => Promise<void>
  loadStats: () => Promise<void>

  // 执行技能
  executeSkill: (input: ExecuteSkillInput) => Promise<ExecuteSkillOutput>

  // 选择
  selectSkill: (skill: SkillListItem) => void
  selectDefinition: (definition: SkillDefinitionInfo) => void

  // 清空
  clearCache: () => Promise<void>
  clearHistory: () => void

  // 刷新
  refresh: () => Promise<void>
}

// ===== 常量 =====

/**
 * 技能分类
 */
export const SKILL_CATEGORIES = {
  HTTP: 'http',
  API: 'api',
  FILE: 'file',
  DATA: 'data',
  CODE: 'code',
  CUSTOM: 'custom',
} as const

/**
 * 工作流动作类型
 */
export const WORKFLOW_ACTIONS = {
  VALIDATE: 'validate',
  PREPARE: 'prepare',
  EXECUTE: 'execute',
  CHECK_EXISTS: 'check_exists',
  GENERATE_CODE: 'generate_code',
  CLEANUP: 'cleanup',
} as const

/**
 * 技能状态
 */
export const SKILL_STATUS = {
  ENABLED: 'enabled',
  DISABLED: 'disabled',
  LOADING: 'loading',
  ERROR: 'error',
} as const

// ===== 辅助类型 =====

/**
 * 分页参数
 */
export interface PaginationParams {
  page: number
  pageSize: number
}

/**
 * 分页结果
 */
export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

/**
 * 搜索过滤参数
 */
export interface SearchFilters {
  keyword?: string
  category?: string
  tags?: string[]
  enabled?: boolean
}

/**
 * 排序参数
 */
export interface SortParams {
  field: keyof SkillListItem
  order: 'asc' | 'desc'
}

// ===== 技能导入相关 =====

/**
 * 技能导入来源类型
 */
export type ImportSourceType = 'file' | 'url' | 'paste' | 'git'

/**
 * 技能导入输入
 */
export interface ImportSkillInput {
  sourceType: ImportSourceType
  data?: ArrayBuffer
  url?: string
  content?: string
  authToken?: string
  branch?: string
  path?: string
  username?: string
  password?: string
  recursive?: boolean
  depth?: number
}

/**
 * 技能导入选项
 */
export interface ImportSkillOptions {
  skillId?: string
  overwrite: boolean
  autoEnable: boolean
  validate: boolean
  workspace?: string
}

/**
 * 技能导入结果
 */
export interface ImportSkillResult {
  success: boolean
  skillId: string
  skillName: string
  message: string
  warnings?: string[]
}

/**
 * 技能元数据（YAML frontmatter）
 */
export interface SkillMetadata {
  name: string
  description: string
  version: string
  category: string
  author?: string
  license?: string
  tags: string[]
}

/**
 * 技能包结构
 */
export interface SkillPackage {
  metadata: SkillMetadata
  content: string
  documents: string[]
  scripts: string[]
}

/**
 * 技能导入状态
 */
export interface ImportState {
  importing: boolean
  currentStep: number
  totalSteps: number
  statusMessage: string
}

