// 技能类型定义
export interface Skill {
  name: string;
  description: string;
  category: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
  version?: string;
  author?: string;
  dependencies: string[];
  config: any;
}

// 工作流节点类型定义
export interface WorkflowNode {
  id: string;
  label: string;
  type: 'start' | 'end' | 'task' | 'condition' | 'loop' | 'fork' | 'join';
  x: number;
  y: number;
  priority?: number;
  maxRetries?: number;
  retryDelay?: string;
  timeout?: string;
  conditionExpression?: string;
  loopType?: 'while' | 'for';
  maxIterations?: number;
  iterationVariable?: string;
  config?: any;
}

// 工作流边类型定义
export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  condition?: string;
  config?: any;
}

// 工作流定义类型
export interface WorkflowDefinition {
  type: 'graph';
  name: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
}

// 工作流类型定义
export interface Workflow {
  id: string;
  name: string;
  description: string;
  definition: WorkflowDefinition;
  createdAt: string;
  updatedAt: string;
  currentVersion: number;
}

// 工作流版本类型定义
export interface WorkflowVersion {
  workflowID: string;
  version: number;
  name: string;
  description: string;
  definition: WorkflowDefinition;
  createdAt: string;
  createdBy: string;
}

// 配置项类型定义
export interface ConfigItem {
  key: string;
  value: any;
  type: string;
  description: string;
  required: boolean;
  defaultValue?: any;
}

// 文件系统条目类型定义
export interface FileSystemItem {
  path: string;
  name: string;
  isDirectory: boolean;
  size?: number;
  createdAt?: string;
  updatedAt?: string;
}
