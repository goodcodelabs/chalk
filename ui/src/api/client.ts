import type {
  Workspace, Catalog, Thread, Message,
  Pipeline, Job, JobStatus, JobResult, SystemStats,
} from './types'

const BASE = '/api'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = { method, headers: { 'Content-Type': 'application/json' } }
  if (body !== undefined) opts.body = JSON.stringify(body)
  const res = await fetch(`${BASE}${path}`, opts)
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data as T
}

const get  = <T>(path: string) => request<T>('GET', path)
const post = <T>(path: string, body: unknown) => request<T>('POST', path, body)
const put  = <T>(path: string, body: unknown) => request<T>('PUT', path, body)
const del  = <T>(path: string) => request<T>('DELETE', path)

// ---- Health / Stats ---------------------------------------------------------

export const api = {
  health: () => get<{ status: string }>('/health'),
  stats:  () => get<SystemStats>('/stats'),

  // ---- Workspaces -----------------------------------------------------------

  listWorkspaces: () =>
    get<{ workspaces: Workspace[] }>('/workspaces').then(r => r.workspaces ?? []),

  createWorkspace: (name: string) =>
    post<{ status: string }>('/workspaces', { name }),

  deleteWorkspace: (name: string) =>
    del<{ status: string }>(`/workspaces/${encodeURIComponent(name)}`),

  setWorkspaceCatalog: (workspaceId: string, catalogId: string) =>
    put<{ status: string }>(`/workspaces/${workspaceId}/catalog`, { catalog_id: catalogId }),

  setWorkspaceRouter: (workspaceId: string, agentId: string) =>
    put<{ status: string }>(`/workspaces/${workspaceId}/router`, { agent_id: agentId }),

  // ---- Catalogs -------------------------------------------------------------

  listCatalogs: () =>
    get<{ catalogs: Catalog[] }>('/catalogs').then(r => r.catalogs ?? []),

  createCatalog: (name: string) =>
    post<{ status: string }>('/catalogs', { name }),

  deleteCatalog: (name: string) =>
    del<{ status: string }>(`/catalogs/${encodeURIComponent(name)}`),

  addAgent: (catalogId: string, name: string) =>
    post<{ id: string; name: string }>(`/catalogs/${catalogId}/agents`, { name }),

  // ---- Agents ---------------------------------------------------------------

  setAgentInstructions: (agentId: string, instructions: string) =>
    put<{ status: string }>(`/agents/${agentId}/instructions`, { instructions }),

  setAgentModel: (agentId: string, model: string) =>
    put<{ status: string }>(`/agents/${agentId}/model`, { model }),

  runAgent: (agentId: string, input: string) =>
    post<{ job_id: string }>(`/agents/${agentId}/run`, { input }),

  listTools: (agentId: string) =>
    get<{ tools: string[] }>(`/agents/${agentId}/tools`).then(r => r.tools ?? []),

  addTool: (agentId: string, tool: string) =>
    post<{ status: string }>(`/agents/${agentId}/tools`, { tool }),

  removeTool: (agentId: string, tool: string) =>
    del<{ status: string }>(`/agents/${agentId}/tools/${encodeURIComponent(tool)}`),

  // ---- Threads --------------------------------------------------------------

  listThreads: (workspaceId: string) =>
    get<{ threads: Thread[] }>(`/workspaces/${workspaceId}/threads`).then(r => r.threads ?? []),

  createThread: (workspaceId: string, name?: string) =>
    post<{ id: string; name: string }>(`/workspaces/${workspaceId}/threads`, { name }),

  chat: (threadId: string, message: string) =>
    post<{ job_id: string }>(`/threads/${threadId}/chat`, { message }),

  threadHistory: (threadId: string) =>
    get<{ messages: Message[] }>(`/threads/${threadId}/history`).then(r => r.messages ?? []),

  // ---- Agent Threads --------------------------------------------------------

  listAgentThreads: (agentId: string) =>
    get<{ threads: Thread[] }>(`/agents/${agentId}/threads`).then(r => r.threads ?? []),

  createAgentThread: (agentId: string, name?: string) =>
    post<{ id: string; name: string }>(`/agents/${agentId}/threads`, { name }),

  agentChat: (threadId: string, message: string) =>
    post<{ job_id: string }>(`/agent-threads/${threadId}/chat`, { message }),

  agentThreadHistory: (threadId: string) =>
    get<{ messages: Message[] }>(`/agent-threads/${threadId}/history`).then(r => r.messages ?? []),

  // ---- Pipelines ------------------------------------------------------------

  listPipelines: (workspaceId: string) =>
    get<{ pipelines: Pipeline[] }>(`/workspaces/${workspaceId}/pipelines`).then(r => r.pipelines ?? []),

  createPipeline: (workspaceId: string, name: string) =>
    post<{ pipeline_id: string }>(`/workspaces/${workspaceId}/pipelines`, { name }),

  addPipelineStep: (pipelineId: string, agentId: string, mode: string) =>
    post<{ status: string }>(`/pipelines/${pipelineId}/steps`, { agent_id: agentId, mode }),

  runPipeline: (pipelineId: string, input: string) =>
    post<{ job_id: string }>(`/pipelines/${pipelineId}/run`, { input }),

  // ---- Jobs -----------------------------------------------------------------

  listJobs: (workspaceId?: string) =>
    workspaceId
      ? get<Job[]>(`/workspaces/${workspaceId}/jobs`)
      : get<Job[]>('/jobs'),

  jobStatus: (jobId: string) =>
    get<JobStatus>(`/jobs/${jobId}/status`),

  jobResult: (jobId: string) =>
    get<JobResult>(`/jobs/${jobId}/result`),

  cancelJob: (jobId: string) =>
    post<{ status: string }>(`/jobs/${jobId}/cancel`, {}),
}
