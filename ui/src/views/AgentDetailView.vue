<template>
  <div>
    <PageHeader
      :title="agent?.Name ?? 'Agent'"
      :subtitle="id"
      back-to="/catalogs"
      back-label="Catalogs"
    />

    <div v-if="loading" class="p-6 space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 bg-gray-100 rounded-xl animate-pulse" />
    </div>

    <div v-else-if="agent" class="p-6 space-y-6">
      <!-- Basic info card -->
      <div class="bg-white rounded-xl border border-gray-200 p-5">
        <h2 class="text-sm font-semibold text-gray-700 mb-4">Configuration</h2>

        <!-- Model -->
        <div class="mb-4">
          <label class="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wide">Model</label>
          <div class="flex gap-2">
            <select
              v-model="editModel"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="claude-sonnet-4-6">claude-sonnet-4-6</option>
              <option value="claude-opus-4-6">claude-opus-4-6</option>
              <option value="claude-haiku-4-5-20251001">claude-haiku-4-5-20251001</option>
            </select>
            <button
              :disabled="savingModel"
              class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              @click="saveModel"
            >
              {{ savingModel ? '…' : 'Save' }}
            </button>
          </div>
        </div>

        <!-- Instructions -->
        <div>
          <label class="block text-xs font-medium text-gray-500 mb-1 uppercase tracking-wide">Instructions (System Prompt)</label>
          <textarea
            v-model="editInstructions"
            rows="6"
            placeholder="You are a helpful assistant…"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <div class="flex justify-end mt-2">
            <button
              :disabled="savingInstructions"
              class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              @click="saveInstructions"
            >
              {{ savingInstructions ? 'Saving…' : 'Save Instructions' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Tools card -->
      <div class="bg-white rounded-xl border border-gray-200 p-5">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-gray-700">Tools</h2>
          <div class="flex gap-2 items-center">
            <select
              v-model="toolToAdd"
              class="px-2 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">— add tool —</option>
              <option v-for="t in availableTools" :key="t" :value="t" :disabled="tools.includes(t)">
                {{ t }}
              </option>
            </select>
            <button
              :disabled="!toolToAdd || addingTool"
              class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
              @click="addTool"
            >
              Add
            </button>
          </div>
        </div>

        <div v-if="tools.length === 0" class="text-sm text-gray-400">No tools attached</div>
        <div v-else class="flex flex-wrap gap-2">
          <div
            v-for="tool in tools"
            :key="tool"
            class="flex items-center gap-1.5 px-3 py-1.5 bg-gray-100 rounded-lg text-sm"
          >
            <span>{{ tool }}</span>
            <button
              class="text-gray-400 hover:text-red-500 ml-1 text-xs leading-none"
              @click="removeTool(tool)"
            >✕</button>
          </div>
        </div>
      </div>

      <!-- Agent Threads card -->
      <div class="bg-white rounded-xl border border-gray-200 p-5">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-gray-700">Threads</h2>
          <button
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700"
            @click="showNewThread = true"
          >
            + New Thread
          </button>
        </div>

        <div v-if="threadsLoading" class="space-y-2">
          <div v-for="i in 2" :key="i" class="h-12 bg-gray-100 rounded-lg animate-pulse" />
        </div>
        <div v-else-if="agentThreads.length === 0" class="text-sm text-gray-400">
          No threads yet
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="t in agentThreads"
            :key="t.id"
            class="flex items-center justify-between p-3 rounded-lg border border-gray-200 hover:border-indigo-300 cursor-pointer"
            @click="router.push(`/agents/${props.id}/threads/${t.id}`)"
          >
            <div>
              <div class="text-sm font-medium text-gray-800">{{ t.name }}</div>
              <div class="text-xs text-gray-400 mt-0.5">{{ t.messages }} messages · {{ formatDate(t.updated_at) }}</div>
            </div>
            <StatusBadge :status="t.state" />
          </div>
        </div>

        <!-- New thread modal -->
        <Modal
          :open="showNewThread"
          title="New Agent Thread"
          confirm-label="Create"
          :loading="creatingThread"
          @close="showNewThread = false"
          @confirm="createAgentThread"
        >
          <label class="block text-sm font-medium text-gray-700 mb-1">Name (optional)</label>
          <input
            v-model="newThreadName"
            type="text"
            placeholder="thread name"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            @keydown.enter="createAgentThread"
          />
        </Modal>
      </div>

      <!-- Test run card -->
      <div class="bg-white rounded-xl border border-gray-200 p-5">
        <h2 class="text-sm font-semibold text-gray-700 mb-4">Test Run</h2>
        <textarea
          v-model="runInput"
          rows="3"
          placeholder="Enter input to test this agent…"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <div class="flex justify-end mt-2">
          <button
            :disabled="!runInput || running"
            class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
            @click="runAgent"
          >
            {{ running ? 'Running…' : 'Run Agent' }}
          </button>
        </div>
        <div v-if="runResponse" class="mt-3 p-3 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-700 whitespace-pre-wrap">
          {{ runResponse }}
        </div>
        <div v-if="runError" class="mt-3 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
          {{ runError }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import Modal from '../components/Modal.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { api } from '../api/client'
import type { Agent, Thread } from '../api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()

const agent = ref<Agent | null>(null)
const agentThreads = ref<Thread[]>([])
const threadsLoading = ref(false)
const showNewThread = ref(false)
const newThreadName = ref('')
const creatingThread = ref(false)

function formatDate(s: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}
const tools = ref<string[]>([])
const loading = ref(true)

const editModel = ref('')
const editInstructions = ref('')
const savingModel = ref(false)
const savingInstructions = ref(false)

const toolToAdd = ref('')
const addingTool = ref(false)
const availableTools = ['http_fetch', 'shell', 'file', 'call_agent']

const runInput = ref('')
const running = ref(false)
const runResponse = ref('')
const runError = ref('')

async function loadThreads() {
  threadsLoading.value = true
  try { agentThreads.value = await api.listAgentThreads(props.id) } finally { threadsLoading.value = false }
}

async function createAgentThread() {
  creatingThread.value = true
  try {
    const t = await api.createAgentThread(props.id, newThreadName.value || undefined)
    showNewThread.value = false
    newThreadName.value = ''
    router.push(`/agents/${props.id}/threads/${t.id}`)
  } finally { creatingThread.value = false }
}

async function load() {
  loading.value = true
  try {
    const [cats, agentTools] = await Promise.all([
      api.listCatalogs(),
      api.listTools(props.id),
    ])
    for (const c of cats) {
      const found = (c.Agents ?? []).find(a => a.ID === props.id)
      if (found) { agent.value = found; break }
    }
    tools.value = agentTools
    if (agent.value) {
      editModel.value = agent.value.Model
      editInstructions.value = agent.value.Instructions
    }
  } finally {
    loading.value = false
  }
}

async function saveModel() {
  savingModel.value = true
  try { await api.setAgentModel(props.id, editModel.value) } finally { savingModel.value = false }
}

async function saveInstructions() {
  savingInstructions.value = true
  try { await api.setAgentInstructions(props.id, editInstructions.value) } finally { savingInstructions.value = false }
}

async function addTool() {
  if (!toolToAdd.value) return
  addingTool.value = true
  try {
    await api.addTool(props.id, toolToAdd.value)
    tools.value = await api.listTools(props.id)
    toolToAdd.value = ''
  } finally { addingTool.value = false }
}

async function removeTool(tool: string) {
  await api.removeTool(props.id, tool)
  tools.value = await api.listTools(props.id)
}

async function runAgent() {
  if (!runInput.value) return
  running.value = true
  runResponse.value = ''
  runError.value = ''
  try {
    const { job_id } = await api.runAgent(props.id, runInput.value)
    while (true) {
      await new Promise(r => setTimeout(r, 500))
      const result = await api.jobResult(job_id)
      if (result.status === 'completed') {
        runResponse.value = result.result
        break
      } else if (result.status === 'failed') {
        runError.value = result.error || 'Job failed'
        break
      }
    }
  } catch (e: any) {
    runError.value = e.message
  } finally { running.value = false }
}

onMounted(() => { load(); loadThreads() })
</script>
