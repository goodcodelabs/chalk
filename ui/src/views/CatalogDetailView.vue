<template>
  <div>
    <PageHeader
      :title="catalog?.Name ?? 'Catalog'"
      :subtitle="id"
      back-to="/catalogs"
      back-label="Catalogs"
    >
      <div class="flex gap-2">
        <button
          class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
          @click="showAddAgent = true"
        >
          + Add Agent
        </button>
        <button
          class="px-3 py-2 text-sm text-red-600 border border-red-300 rounded-lg hover:bg-red-50"
          @click="deleteCatalog"
        >
          Delete
        </button>
      </div>
    </PageHeader>

    <div class="p-6">
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="h-28 bg-gray-100 rounded-xl animate-pulse" />
      </div>

      <div v-else-if="!catalog || (catalog.Agents ?? []).length === 0" class="text-center py-16 text-gray-400">
        <p class="font-medium">No agents in this catalog</p>
        <p class="text-sm mt-1">Add an agent to get started.</p>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div
          v-for="agent in (catalog.Agents ?? [])"
          :key="agent.ID"
          class="bg-white rounded-xl border border-gray-200 p-4 hover:border-indigo-300 cursor-pointer"
          @click="router.push(`/agents/${agent.ID}`)"
        >
          <div class="flex items-start justify-between mb-2">
            <div>
              <div class="font-medium text-gray-900">{{ agent.Name }}</div>
              <div class="text-xs text-gray-500 font-mono mt-0.5">{{ agent.ID }}</div>
            </div>
            <div class="flex items-center gap-1.5">
              <span
                v-if="agent.External"
                class="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded-full"
              >external</span>
              <button
                class="text-red-400 hover:text-red-600 text-xs px-2 py-0.5 rounded hover:bg-red-50"
                @click.stop="deleteAgent(agent)"
              >Delete</button>
            </div>
          </div>
          <div class="text-xs text-gray-500 space-y-0.5">
            <div><span class="font-medium">Model:</span> {{ agent.Model }}</div>
            <div v-if="(agent.Tools ?? []).length > 0">
              <span class="font-medium">Tools:</span> {{ (agent.Tools ?? []).join(', ') }}
            </div>
            <div v-if="agent.Instructions" class="mt-1 text-gray-400 truncate">
              {{ agent.Instructions.slice(0, 80) }}{{ agent.Instructions.length > 80 ? '…' : '' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add agent modal -->
    <Modal
      :open="showAddAgent"
      title="Add Agent"
      confirm-label="Add"
      :loading="addingAgent"
      @close="showAddAgent = false"
      @confirm="addAgent"
    >
      <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
      <input
        v-model="newAgentName"
        type="text"
        placeholder="agent-name"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        @keydown.enter="addAgent"
      />
      <p v-if="addError" class="mt-2 text-xs text-red-600">{{ addError }}</p>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import Modal from '../components/Modal.vue'
import { api } from '../api/client'
import type { Catalog, Agent } from '../api/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const catalog = ref<Catalog | null>(null)
const loading = ref(true)
const showAddAgent = ref(false)
const newAgentName = ref('')
const addingAgent = ref(false)
const addError = ref('')

async function load() {
  loading.value = true
  try {
    const cats = await api.listCatalogs()
    catalog.value = cats.find(c => c.ID === props.id) ?? null
  } finally {
    loading.value = false
  }
}

async function addAgent() {
  if (!newAgentName.value.trim()) { addError.value = 'Name required'; return }
  addingAgent.value = true
  addError.value = ''
  try {
    await api.addAgent(props.id, newAgentName.value.trim())
    showAddAgent.value = false
    newAgentName.value = ''
    await load()
  } catch (e: any) {
    addError.value = e.message
  } finally {
    addingAgent.value = false
  }
}

async function deleteAgent(agent: Agent) {
  if (!confirm(`Delete agent "${agent.Name}"?`)) return
  await api.delAgent(agent.ID)
  await load()
}

async function deleteCatalog() {
  if (!catalog.value) return
  if (!confirm(`Delete catalog "${catalog.value.Name}"?`)) return
  await api.deleteCatalog(catalog.value.Name)
  router.push('/catalogs')
}

onMounted(load)
</script>
