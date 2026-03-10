<template>
  <div>
    <PageHeader
      :title="workspace?.name ?? 'Workspace'"
      :subtitle="id"
      back-to="/workspaces"
      back-label="Workspaces"
    >
      <button
        class="px-3 py-2 text-sm text-red-600 border border-red-300 rounded-lg hover:bg-red-50"
        @click="confirmDelete"
      >
        Delete
      </button>
    </PageHeader>

    <!-- Tabs -->
    <div class="bg-white border-b border-gray-200 px-6">
      <nav class="flex gap-0">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="px-4 py-3 text-sm font-medium border-b-2 transition-colors"
          :class="activeTab === tab.key
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <div class="p-6">
      <!-- Overview tab -->
      <div v-if="activeTab === 'overview'">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Catalog card -->
          <div class="bg-white rounded-xl border border-gray-200 p-5">
            <h3 class="text-sm font-semibold text-gray-700 mb-3">Catalog</h3>
            <div v-if="workspace?.catalog_id" class="mb-3">
              <span class="text-xs font-mono text-gray-500">{{ workspace.catalog_id }}</span>
            </div>
            <div v-else class="text-sm text-gray-400 mb-3">No catalog assigned</div>
            <select
              v-model="selectedCatalogId"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">— select catalog —</option>
              <option v-for="c in catalogs" :key="c.ID" :value="c.ID">{{ c.Name }}</option>
            </select>
            <button
              :disabled="!selectedCatalogId || savingCatalog"
              class="mt-3 px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              @click="saveCatalog"
            >
              {{ savingCatalog ? 'Saving…' : 'Set Catalog' }}
            </button>
          </div>

          <!-- Router card -->
          <div class="bg-white rounded-xl border border-gray-200 p-5">
            <h3 class="text-sm font-semibold text-gray-700 mb-3">Router Agent</h3>
            <div v-if="workspace?.router_agent_id" class="mb-3">
              <span class="text-xs font-mono text-gray-500">{{ workspace.router_agent_id }}</span>
            </div>
            <div v-else class="text-sm text-gray-400 mb-3">No router agent assigned</div>
            <select
              v-model="selectedRouterAgentId"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">— select agent —</option>
              <template v-for="c in catalogs" :key="c.ID">
                <optgroup :label="c.Name">
                  <option v-for="a in (c.Agents ?? [])" :key="a.ID" :value="a.ID">
                    {{ a.Name }}
                  </option>
                </optgroup>
              </template>
            </select>
            <button
              :disabled="!selectedRouterAgentId || savingRouter"
              class="mt-3 px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              @click="saveRouter"
            >
              {{ savingRouter ? 'Saving…' : 'Set Router' }}
            </button>
          </div>
        </div>

      </div>

      <!-- Threads tab -->
      <div v-else-if="activeTab === 'threads'">
        <div class="flex justify-between items-center mb-4">
          <span class="text-sm text-gray-500">{{ threads.length }} thread(s)</span>
          <button
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
            @click="showNewThread = true"
          >
            + New Thread
          </button>
        </div>

        <div v-if="threadsLoading" class="space-y-3">
          <div v-for="i in 3" :key="i" class="h-16 bg-gray-100 rounded-xl animate-pulse" />
        </div>

        <div v-else-if="threads.length === 0" class="text-center py-12 text-gray-400">
          <p>No threads yet</p>
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="t in threads"
            :key="t.id"
            class="bg-white rounded-xl border border-gray-200 p-4 flex items-center justify-between hover:border-indigo-300 cursor-pointer"
            @click="router.push(`/workspaces/${id}/threads/${t.id}`)"
          >
            <div>
              <div class="font-medium text-gray-900 text-sm">{{ t.name }}</div>
              <div class="text-xs text-gray-500 mt-0.5">
                {{ t.messages }} messages · {{ formatDate(t.updated_at) }}
              </div>
            </div>
            <div class="flex items-center gap-2">
              <StatusBadge :status="t.state" />
            </div>
          </div>
        </div>
      </div>

      <!-- Pipelines tab -->
      <div v-else-if="activeTab === 'pipelines'">
        <div class="flex justify-between items-center mb-4">
          <span class="text-sm text-gray-500">{{ pipelines.length }} pipeline(s)</span>
          <button
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
            @click="showNewPipeline = true"
          >
            + New Pipeline
          </button>
        </div>

        <div v-if="pipelinesLoading" class="space-y-3">
          <div v-for="i in 3" :key="i" class="h-20 bg-gray-100 rounded-xl animate-pulse" />
        </div>

        <div v-else-if="pipelines.length === 0" class="text-center py-12 text-gray-400">
          <p>No pipelines yet</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="p in pipelines"
            :key="p.id"
            class="bg-white rounded-xl border border-gray-200 p-4"
          >
            <div class="flex items-start justify-between">
              <div>
                <div class="font-medium text-gray-900">{{ p.name }}</div>
                <div class="text-xs text-gray-500 font-mono mt-0.5">{{ p.id }}</div>
                <div class="flex flex-wrap gap-2 mt-2">
                  <span
                    v-for="(step, i) in p.steps"
                    :key="i"
                    class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-100 rounded text-xs"
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :class="step.mode === 'parallel' ? 'bg-violet-500' : 'bg-blue-500'"
                    />
                    {{ agentName(step.agent_id) }}
                    <span class="text-gray-400">({{ step.mode }})</span>
                  </span>
                </div>
              </div>
              <button
                class="px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-lg hover:bg-indigo-700"
                @click="openRunPipeline(p.id)"
              >
                Run
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Jobs tab -->
      <div v-else-if="activeTab === 'jobs'">
        <div v-if="jobsLoading" class="space-y-3">
          <div v-for="i in 3" :key="i" class="h-14 bg-gray-100 rounded-xl animate-pulse" />
        </div>
        <div v-else-if="jobs.length === 0" class="text-center py-12 text-gray-400">
          <p>No jobs for this workspace</p>
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="j in jobs"
            :key="j.id"
            class="bg-white rounded-xl border border-gray-200 p-4 flex items-center justify-between"
          >
            <div>
              <div class="text-sm font-mono text-gray-700">{{ j.id }}</div>
              <div class="text-xs text-gray-500 mt-0.5">{{ j.type }} · {{ formatDate(j.created_at) }}</div>
            </div>
            <div class="flex items-center gap-3">
              <StatusBadge :status="j.status" />
              <button
                v-if="j.status === 'running' || j.status === 'pending'"
                class="text-xs text-red-600 hover:text-red-700"
                @click="cancelJob(j.id)"
              >Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- New thread modal -->
    <Modal
      :open="showNewThread"
      title="New Thread"
      confirm-label="Create"
      :loading="creatingThread"
      @close="showNewThread = false"
      @confirm="createThread"
    >
      <label class="block text-sm font-medium text-gray-700 mb-1">Name (optional)</label>
      <input
        v-model="newThreadName"
        type="text"
        placeholder="thread name"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        @keydown.enter="createThread"
      />
    </Modal>

    <!-- New pipeline modal -->
    <Modal
      :open="showNewPipeline"
      title="New Pipeline"
      confirm-label="Create"
      :loading="creatingPipeline"
      @close="showNewPipeline = false"
      @confirm="createPipeline"
    >
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
        <input
          v-model="newPipelineName"
          type="text"
          placeholder="my-pipeline"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
    </Modal>

    <!-- Run pipeline modal -->
    <Modal
      :open="showRunPipeline"
      title="Run Pipeline"
      confirm-label="Run"
      :loading="runningPipeline"
      @close="showRunPipeline = false"
      @confirm="runPipeline"
    >
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Input</label>
        <textarea
          v-model="pipelineInput"
          rows="4"
          placeholder="Pipeline input text…"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
      <div v-if="runJobId" class="mt-3 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700">
        Job started: <span class="font-mono">{{ runJobId }}</span>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import Modal from '../components/Modal.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { api } from '../api/client'
import type { Workspace, Catalog, Thread, Pipeline, Job } from '../api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()

const workspace = ref<Workspace | null>(null)
const catalogs = ref<Catalog[]>([])
const threads = ref<Thread[]>([])
const pipelines = ref<Pipeline[]>([])
const jobs = ref<Job[]>([])

const activeTab = ref('overview')
const tabs = [
  { key: 'overview',  label: 'Overview'  },
  { key: 'threads',   label: 'Threads'   },
  { key: 'pipelines', label: 'Pipelines' },
  { key: 'jobs',      label: 'Jobs'      },
]

const selectedCatalogId = ref('')
const selectedRouterAgentId = ref('')
const savingCatalog = ref(false)
const savingRouter = ref(false)

const threadsLoading = ref(false)
const pipelinesLoading = ref(false)
const jobsLoading = ref(false)

const showNewThread = ref(false)
const newThreadName = ref('')
const creatingThread = ref(false)

const showNewPipeline = ref(false)
const newPipelineName = ref('')
const creatingPipeline = ref(false)

const showRunPipeline = ref(false)
const runPipelineId = ref('')
const pipelineInput = ref('')
const runningPipeline = ref(false)
const runJobId = ref('')

function formatDate(s: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

function agentName(agentId: string) {
  for (const c of catalogs.value) {
    for (const a of (c.Agents ?? [])) {
      if (a.ID === agentId) return a.Name
    }
  }
  return agentId.slice(0, 8)
}

async function load() {
  const [ws, cats] = await Promise.all([
    api.listWorkspaces(),
    api.listCatalogs(),
  ])
  workspace.value = ws.find(w => w.id === props.id) ?? null
  catalogs.value = cats
  if (workspace.value?.catalog_id) selectedCatalogId.value = workspace.value.catalog_id
  if (workspace.value?.router_agent_id) selectedRouterAgentId.value = workspace.value.router_agent_id
}

async function loadThreads() {
  threadsLoading.value = true
  try { threads.value = await api.listThreads(props.id) } finally { threadsLoading.value = false }
}

async function loadPipelines() {
  pipelinesLoading.value = true
  try { pipelines.value = await api.listPipelines(props.id) } finally { pipelinesLoading.value = false }
}

async function loadJobs() {
  jobsLoading.value = true
  try { jobs.value = await api.listJobs(props.id) } finally { jobsLoading.value = false }
}

watch(activeTab, tab => {
  if (tab === 'threads') loadThreads()
  if (tab === 'pipelines') loadPipelines()
  if (tab === 'jobs') loadJobs()
})

async function saveCatalog() {
  savingCatalog.value = true
  try {
    await api.setWorkspaceCatalog(props.id, selectedCatalogId.value)
    await load()
  } finally { savingCatalog.value = false }
}

async function saveRouter() {
  savingRouter.value = true
  try {
    await api.setWorkspaceRouter(props.id, selectedRouterAgentId.value)
    await load()
  } finally { savingRouter.value = false }
}

async function createThread() {
  creatingThread.value = true
  try {
    await api.createThread(props.id, newThreadName.value || undefined)
    showNewThread.value = false
    newThreadName.value = ''
    await loadThreads()
  } finally { creatingThread.value = false }
}

async function createPipeline() {
  if (!newPipelineName.value) return
  creatingPipeline.value = true
  try {
    await api.createPipeline(props.id, newPipelineName.value)
    showNewPipeline.value = false
    newPipelineName.value = ''
    await loadPipelines()
  } finally { creatingPipeline.value = false }
}

function openRunPipeline(pipelineId: string) {
  runPipelineId.value = pipelineId
  pipelineInput.value = ''
  runJobId.value = ''
  showRunPipeline.value = true
}

async function runPipeline() {
  if (!pipelineInput.value) return
  runningPipeline.value = true
  try {
    const r = await api.runPipeline(runPipelineId.value, pipelineInput.value)
    runJobId.value = r.job_id
  } finally { runningPipeline.value = false }
}

async function cancelJob(jobId: string) {
  await api.cancelJob(jobId)
  await loadJobs()
}

async function confirmDelete() {
  if (!workspace.value) return
  if (!confirm(`Delete workspace "${workspace.value.name}"?`)) return
  await api.deleteWorkspace(workspace.value.name)
  router.push('/workspaces')
}

onMounted(load)
</script>
