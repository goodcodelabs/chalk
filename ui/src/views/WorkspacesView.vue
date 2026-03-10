<template>
  <div>
    <PageHeader title="Workspaces" subtitle="Manage your agent workspaces">
      <button
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
        @click="showCreate = true"
      >
        + New Workspace
      </button>
    </PageHeader>

    <div class="p-6">
      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="h-20 bg-gray-100 rounded-xl animate-pulse" />
      </div>

      <div v-else-if="workspaces.length === 0" class="text-center py-16 text-gray-400">
        <div class="text-5xl mb-3">⬡</div>
        <p class="font-medium">No workspaces yet</p>
        <p class="text-sm mt-1">Create one to get started.</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="ws in workspaces"
          :key="ws.id"
          class="bg-white rounded-xl border border-gray-200 p-4 flex items-center justify-between hover:border-indigo-300 transition-colors cursor-pointer"
          @click="router.push(`/workspaces/${ws.id}`)"
        >
          <div>
            <div class="font-medium text-gray-900">{{ ws.name }}</div>
            <div class="text-xs text-gray-500 mt-0.5 font-mono">{{ ws.id }}</div>
            <div class="flex gap-4 mt-2">
              <span v-if="ws.catalog_id" class="text-xs text-indigo-600">
                ☰ Catalog set
              </span>
              <span v-if="ws.router_agent_id" class="text-xs text-violet-600">
                ⟶ Router set
              </span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-gray-300 text-sm">→</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <Modal
      :open="showCreate"
      title="New Workspace"
      confirm-label="Create"
      :loading="creating"
      @close="showCreate = false"
      @confirm="createWorkspace"
    >
      <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
      <input
        v-model="newName"
        type="text"
        placeholder="my-workspace"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        @keydown.enter="createWorkspace"
      />
      <p v-if="createError" class="mt-2 text-xs text-red-600">{{ createError }}</p>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import Modal from '../components/Modal.vue'
import { api } from '../api/client'
import type { Workspace } from '../api/types'

const router = useRouter()
const workspaces = ref<Workspace[]>([])
const loading = ref(true)
const error = ref('')
const showCreate = ref(false)
const newName = ref('')
const creating = ref(false)
const createError = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    workspaces.value = await api.listWorkspaces()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createWorkspace() {
  if (!newName.value.trim()) {
    createError.value = 'Name is required'
    return
  }
  creating.value = true
  createError.value = ''
  try {
    await api.createWorkspace(newName.value.trim())
    showCreate.value = false
    newName.value = ''
    await load()
  } catch (e: any) {
    createError.value = e.message
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>
