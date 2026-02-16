export namespace agent {
	
	export class AgentSpec {
	    Name: string;
	    Kind: string;
	    Model: string;
	    Instructions: string;
	    Tools: string[];
	    Middlewares: string[];
	    MaxMessages: number;
	    MaxMessageSize: number;
	    TrimRatio: number;
	    EnableTrimming: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Kind = source["Kind"];
	        this.Model = source["Model"];
	        this.Instructions = source["Instructions"];
	        this.Tools = source["Tools"];
	        this.Middlewares = source["Middlewares"];
	        this.MaxMessages = source["MaxMessages"];
	        this.MaxMessageSize = source["MaxMessageSize"];
	        this.TrimRatio = source["TrimRatio"];
	        this.EnableTrimming = source["EnableTrimming"];
	    }
	}
	export class AsyncTaskSpec {
	    Enabled: boolean;
	    MaxTasks: number;
	    TaskTimeout: number;
	
	    static createFrom(source: any = {}) {
	        return new AsyncTaskSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.MaxTasks = source["MaxTasks"];
	        this.TaskTimeout = source["TaskTimeout"];
	    }
	}
	export class ChannelConfigSpec {
	    Type: string;
	    Enabled: boolean;
	    Config: Record<string, any>;
	    BufferSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ChannelConfigSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Enabled = source["Enabled"];
	        this.Config = source["Config"];
	        this.BufferSize = source["BufferSize"];
	    }
	}
	export class OpenVikingStoreSpec {
	    Endpoint: string;
	    APIKey: string;
	    Workspace: string;
	    Timeout: number;
	    MaxRetries: number;
	    AutoSync: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OpenVikingStoreSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Endpoint = source["Endpoint"];
	        this.APIKey = source["APIKey"];
	        this.Workspace = source["Workspace"];
	        this.Timeout = source["Timeout"];
	        this.MaxRetries = source["MaxRetries"];
	        this.AutoSync = source["AutoSync"];
	    }
	}
	export class ContextStoreSpec {
	    Enabled: boolean;
	    Type: string;
	    OpenViking: OpenVikingStoreSpec;
	
	    static createFrom(source: any = {}) {
	        return new ContextStoreSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.Type = source["Type"];
	        this.OpenViking = this.convertValues(source["OpenViking"], OpenVikingStoreSpec);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    type: string;
	    size: number;
	    // Go type: time
	    modified: any;
	    // Go type: time
	    created: any;
	    mode: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.modified = this.convertValues(source["modified"], null);
	        this.created = this.convertValues(source["created"], null);
	        this.mode = source["mode"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HeartbeatSpec {
	    Enabled: boolean;
	    Interval: number;
	    Timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new HeartbeatSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.Interval = source["Interval"];
	        this.Timeout = source["Timeout"];
	    }
	}
	export class TokenCompressionSpec {
	    Enabled: boolean;
	    Strategy: string;
	    TargetTokens: number;
	    MinTokens: number;
	    MaxTokens: number;
	    PreserveSystemMessages: boolean;
	    SummaryModelName: string;
	    SummaryMaxTokens: number;
	    Temperature: number;
	    CheckInterval: number;
	    SessionTokenLimit: number;
	    DailyTokenLimit: number;
	    LongTermTokenLimit: number;
	
	    static createFrom(source: any = {}) {
	        return new TokenCompressionSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.Strategy = source["Strategy"];
	        this.TargetTokens = source["TargetTokens"];
	        this.MinTokens = source["MinTokens"];
	        this.MaxTokens = source["MaxTokens"];
	        this.PreserveSystemMessages = source["PreserveSystemMessages"];
	        this.SummaryModelName = source["SummaryModelName"];
	        this.SummaryMaxTokens = source["SummaryMaxTokens"];
	        this.Temperature = source["Temperature"];
	        this.CheckInterval = source["CheckInterval"];
	        this.SessionTokenLimit = source["SessionTokenLimit"];
	        this.DailyTokenLimit = source["DailyTokenLimit"];
	        this.LongTermTokenLimit = source["LongTermTokenLimit"];
	    }
	}
	export class SchedulerSpec {
	    Enabled: boolean;
	    Timezone: string;
	    MaxJobs: number;
	    JobTimeout: number;
	
	    static createFrom(source: any = {}) {
	        return new SchedulerSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.Timezone = source["Timezone"];
	        this.MaxJobs = source["MaxJobs"];
	        this.JobTimeout = source["JobTimeout"];
	    }
	}
	export class MessagingConfig {
	    Enabled: boolean;
	    EnableMetrics: boolean;
	    Channels: Record<string, ChannelConfigSpec>;
	    DefaultChannel: string;
	
	    static createFrom(source: any = {}) {
	        return new MessagingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.EnableMetrics = source["EnableMetrics"];
	        this.Channels = this.convertValues(source["Channels"], ChannelConfigSpec, true);
	        this.DefaultChannel = source["DefaultChannel"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NodeSpec {
	    ID: string;
	    Kind: string;
	    AgentName: string;
	    InlineName: string;
	    InlineKind: string;
	    InlineModel: string;
	    InlineInstructions: string;
	    InlineTools: string[];
	    InlineMiddlewares: string[];
	
	    static createFrom(source: any = {}) {
	        return new NodeSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Kind = source["Kind"];
	        this.AgentName = source["AgentName"];
	        this.InlineName = source["InlineName"];
	        this.InlineKind = source["InlineKind"];
	        this.InlineModel = source["InlineModel"];
	        this.InlineInstructions = source["InlineInstructions"];
	        this.InlineTools = source["InlineTools"];
	        this.InlineMiddlewares = source["InlineMiddlewares"];
	    }
	}
	export class WorkflowSpec {
	    Name: string;
	    Kind: string;
	    Model: string;
	    Steps: string[];
	    Agents: string[];
	    Aggregator: string;
	    Routes: Record<string, string>;
	    Nodes: NodeSpec[];
	    Edges: Record<string, string>;
	    EdgesList: Record<string, Array<string>>;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Kind = source["Kind"];
	        this.Model = source["Model"];
	        this.Steps = source["Steps"];
	        this.Agents = source["Agents"];
	        this.Aggregator = source["Aggregator"];
	        this.Routes = source["Routes"];
	        this.Nodes = this.convertValues(source["Nodes"], NodeSpec);
	        this.Edges = source["Edges"];
	        this.EdgesList = source["EdgesList"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MemoryMonitorSpec {
	    Enabled: boolean;
	    Interval: number;
	    HistorySize: number;
	    AlertThreshold: number;
	    AlertInterval: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryMonitorSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.Interval = source["Interval"];
	        this.HistorySize = source["HistorySize"];
	        this.AlertThreshold = source["AlertThreshold"];
	        this.AlertInterval = source["AlertInterval"];
	    }
	}
	export class ModelCacheSpec {
	    Enabled: boolean;
	    MaxSize: number;
	    TTL: number;
	    CleanupInterval: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelCacheSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Enabled = source["Enabled"];
	        this.MaxSize = source["MaxSize"];
	        this.TTL = source["TTL"];
	        this.CleanupInterval = source["CleanupInterval"];
	    }
	}
	export class MemoryManagementSpec {
	    ModelCache: ModelCacheSpec;
	    MemoryMonitor: MemoryMonitorSpec;
	    // Go type: struct { MaxCheckpoints int "yaml:\"maxCheckpoints,omitempty\""; TTL int "yaml:\"ttl,omitempty\""; CleanupInterval int "yaml:\"cleanupInterval,omitempty\"" }
	    Checkpoint: any;
	
	    static createFrom(source: any = {}) {
	        return new MemoryManagementSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ModelCache = this.convertValues(source["ModelCache"], ModelCacheSpec);
	        this.MemoryMonitor = this.convertValues(source["MemoryMonitor"], MemoryMonitorSpec);
	        this.Checkpoint = this.convertValues(source["Checkpoint"], Object);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ThreadStoreSpec {
	    Type: string;
	    Dir: string;
	    RedisAddr: string;
	    RedisPrefix: string;
	    DriverName: string;
	    DSN: string;
	    Table: string;
	    MaxMessages: number;
	    MaxMessageSize: number;
	    TTL: number;
	
	    static createFrom(source: any = {}) {
	        return new ThreadStoreSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Dir = source["Dir"];
	        this.RedisAddr = source["RedisAddr"];
	        this.RedisPrefix = source["RedisPrefix"];
	        this.DriverName = source["DriverName"];
	        this.DSN = source["DSN"];
	        this.Table = source["Table"];
	        this.MaxMessages = source["MaxMessages"];
	        this.MaxMessageSize = source["MaxMessageSize"];
	        this.TTL = source["TTL"];
	    }
	}
	export class ModelConfig {
	    type: string;
	    model: string;
	    api_key: string;
	    base_url: string;
	    options: Record<string, any>;
	    timeout: number;
	    max_retries: number;
	    retry_interval: number;
	    temperature: number;
	    max_tokens: number;
	    top_p: number;
	    top_k: number;
	    log_level: string;
	    priority: number;
	    enabled: boolean;
	    headers: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ModelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.model = source["model"];
	        this.api_key = source["api_key"];
	        this.base_url = source["base_url"];
	        this.options = source["options"];
	        this.timeout = source["timeout"];
	        this.max_retries = source["max_retries"];
	        this.retry_interval = source["retry_interval"];
	        this.temperature = source["temperature"];
	        this.max_tokens = source["max_tokens"];
	        this.top_p = source["top_p"];
	        this.top_k = source["top_k"];
	        this.log_level = source["log_level"];
	        this.priority = source["priority"];
	        this.enabled = source["enabled"];
	        this.headers = source["headers"];
	    }
	}
	export class HostConfig {
	    Name: string;
	    Version: string;
	    DefaultModel: string;
	    Models: Record<string, ModelConfig>;
	    ThreadStore: ThreadStoreSpec;
	    Memory: MemoryManagementSpec;
	    Agents: AgentSpec[];
	    Workflows: WorkflowSpec[];
	    SkillSystemDir: string;
	    ContextStore: ContextStoreSpec;
	    Messaging?: MessagingConfig;
	    Scheduler?: SchedulerSpec;
	    Heartbeat?: HeartbeatSpec;
	    AsyncTask?: AsyncTaskSpec;
	    TokenCompression?: TokenCompressionSpec;
	    Extensions: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new HostConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Version = source["Version"];
	        this.DefaultModel = source["DefaultModel"];
	        this.Models = this.convertValues(source["Models"], ModelConfig, true);
	        this.ThreadStore = this.convertValues(source["ThreadStore"], ThreadStoreSpec);
	        this.Memory = this.convertValues(source["Memory"], MemoryManagementSpec);
	        this.Agents = this.convertValues(source["Agents"], AgentSpec);
	        this.Workflows = this.convertValues(source["Workflows"], WorkflowSpec);
	        this.SkillSystemDir = source["SkillSystemDir"];
	        this.ContextStore = this.convertValues(source["ContextStore"], ContextStoreSpec);
	        this.Messaging = this.convertValues(source["Messaging"], MessagingConfig);
	        this.Scheduler = this.convertValues(source["Scheduler"], SchedulerSpec);
	        this.Heartbeat = this.convertValues(source["Heartbeat"], HeartbeatSpec);
	        this.AsyncTask = this.convertValues(source["AsyncTask"], AsyncTaskSpec);
	        this.TokenCompression = this.convertValues(source["TokenCompression"], TokenCompressionSpec);
	        this.Extensions = source["Extensions"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	export class NodeExecutionResult {
	    node_id: string;
	    status: string;
	    input?: string;
	    output?: string;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    end_time: any;
	    error?: string;
	    retry_count?: number;
	
	    static createFrom(source: any = {}) {
	        return new NodeExecutionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.node_id = source["node_id"];
	        this.status = source["status"];
	        this.input = source["input"];
	        this.output = source["output"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.end_time = this.convertValues(source["end_time"], null);
	        this.error = source["error"];
	        this.retry_count = source["retry_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class SkillMetadata {
	    name: string;
	    version: string;
	    author: string;
	    description: string;
	    category: string;
	    tags: string[];
	    dependencies: string[];
	    license: string;
	    homepage: string;
	    repository: string;
	    keywords: string[];
	    config: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new SkillMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.dependencies = source["dependencies"];
	        this.license = source["license"];
	        this.homepage = source["homepage"];
	        this.repository = source["repository"];
	        this.keywords = source["keywords"];
	        this.config = source["config"];
	    }
	}
	
	
	export class WorkflowExecutionResult {
	    workflow_id: string;
	    execution_id: string;
	    status: string;
	    input?: string;
	    output?: string;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    end_time?: any;
	    error?: string;
	    node_results?: NodeExecutionResult[];
	
	    static createFrom(source: any = {}) {
	        return new WorkflowExecutionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workflow_id = source["workflow_id"];
	        this.execution_id = source["execution_id"];
	        this.status = source["status"];
	        this.input = source["input"];
	        this.output = source["output"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.end_time = this.convertValues(source["end_time"], null);
	        this.error = source["error"];
	        this.node_results = this.convertValues(source["node_results"], NodeExecutionResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkflowInfo {
	    id: string;
	    name: string;
	    description: string;
	    definition: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.definition = source["definition"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	
	export class WorkflowVersion {
	    id: string;
	    workflow_id: string;
	    version: number;
	    definition: string;
	    name: string;
	    description: string;
	    // Go type: time
	    created_at: any;
	    created_by?: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workflow_id = source["workflow_id"];
	        this.version = source["version"];
	        this.definition = source["definition"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.created_by = source["created_by"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class CacheStats {
	    hitRate: number;
	    hits: number;
	    misses: number;
	    totalEntries: number;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hitRate = source["hitRate"];
	        this.hits = source["hits"];
	        this.misses = source["misses"];
	        this.totalEntries = source["totalEntries"];
	        this.size = source["size"];
	    }
	}
	export class ExecuteSkillInput {
	    skillName: string;
	    input: string;
	    workspace: string;
	    parameters?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteSkillInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.input = source["input"];
	        this.workspace = source["workspace"];
	        this.parameters = source["parameters"];
	    }
	}
	export class ExecuteSkillOutput {
	    success: boolean;
	    result?: any;
	    error?: string;
	    stats?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteSkillOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.result = source["result"];
	        this.error = source["error"];
	        this.stats = source["stats"];
	    }
	}
	export class ImportSkillOptions {
	    skillId?: string;
	    overwrite: boolean;
	    autoEnable: boolean;
	    validate: boolean;
	    workspace?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportSkillOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.overwrite = source["overwrite"];
	        this.autoEnable = source["autoEnable"];
	        this.validate = source["validate"];
	        this.workspace = source["workspace"];
	    }
	}
	export class ImportSkillResult {
	    success: boolean;
	    skillId: string;
	    skillName: string;
	    message: string;
	    warnings?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportSkillResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.skillId = source["skillId"];
	        this.skillName = source["skillName"];
	        this.message = source["message"];
	        this.warnings = source["warnings"];
	    }
	}
	export class PoolStats {
	    activeConnections: number;
	    idleConnections: number;
	    maxConnections: number;
	    minConnections: number;
	    utilizationRate: number;
	
	    static createFrom(source: any = {}) {
	        return new PoolStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeConnections = source["activeConnections"];
	        this.idleConnections = source["idleConnections"];
	        this.maxConnections = source["maxConnections"];
	        this.minConnections = source["minConnections"];
	        this.utilizationRate = source["utilizationRate"];
	    }
	}
	export class SkillDefinitionInfo {
	    id: string;
	    name: string;
	    description: string;
	    version: string;
	    category: string;
	    author: string;
	    license: string;
	    workflow: Record<string, any>;
	    config: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new SkillDefinitionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.category = source["category"];
	        this.author = source["author"];
	        this.license = source["license"];
	        this.workflow = source["workflow"];
	        this.config = source["config"];
	    }
	}
	export class SkillListItem {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    tags: string[];
	    version: string;
	    enabled: boolean;
	    useCount: number;
	    lastUsed: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.version = source["version"];
	        this.enabled = source["enabled"];
	        this.useCount = source["useCount"];
	        this.lastUsed = source["lastUsed"];
	    }
	}
	export class SkillSystemInfo {
	    initialized: boolean;
	    baseDir: string;
	    totalSkills: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillSystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.initialized = source["initialized"];
	        this.baseDir = source["baseDir"];
	        this.totalSkills = source["totalSkills"];
	    }
	}
	export class SkillSystemStats {
	    totalSkills: number;
	    totalUses: number;
	    categories: Record<string, number>;
	    mostUsedSkills: any[];
	
	    static createFrom(source: any = {}) {
	        return new SkillSystemStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSkills = source["totalSkills"];
	        this.totalUses = source["totalUses"];
	        this.categories = source["categories"];
	        this.mostUsedSkills = source["mostUsedSkills"];
	    }
	}

}

