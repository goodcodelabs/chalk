export interface Workspace {
  id: string
  name: string
  catalog_id?: string
  router_agent_id?: string
}

export interface Agent {
  ID: string
  Name: string
  Instructions: string
  Model: string
  MaxTokens: number
  Temperature: number
  Tools: string[]
  Metadata: Record<string, string> | null
  External: boolean
}

export interface Catalog {
  ID: string
  Name: string
  Agents: Agent[] | null
}

export interface Thread {
  id: string
  name: string
  state: string
  messages: number
  created_at: string
  updated_at: string
}

export interface Message {
  role: string
  content: string | MessageContent[]
}

export interface MessageContent {
  type: string
  text?: string
}

export interface Pipeline {
  id: string
  name: string
  steps: PipelineStep[]
}

export interface PipelineStep {
  agent_id: string
  mode: 'sequential' | 'parallel'
}

export interface Job {
  id: string
  type: string
  workspace_id: string
  pipeline_id?: string
  thread_id?: string
  status: string
  created_at: string
}

export interface JobStatus {
  status: string
  created_at: string
  started_at: string
  completed_at: string
}

export interface JobResult {
  status: string
  result: string
  error: string
}

export interface TraceEvent {
  type: string
  agent_id?: string
  tool?: string
  input?: string
  output?: string
  timestamp: string
}

export interface SystemStats {
  jobs: Record<string, number>
  scheduler_queue: number
  llm_calls: number
  tool_calls: number
  errors: number
  input_tokens_total: number
  output_tokens_total: number
  active_connections: number
}
