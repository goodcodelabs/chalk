<template>
  <div>
    <PageHeader title="Catalogs" subtitle="Agent registries">
      <button
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700"
        @click="showCreate = true"
      >
        + New Catalog
      </button>
    </PageHeader>

    <div class="p-6">
      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="h-24 bg-gray-100 rounded-xl animate-pulse" />
      </div>

      <div v-else-if="catalogs.length === 0" class="text-center py-16 text-gray-400">
        <div class="text-5xl mb-3">☰</div>
        <p class="font-medium">No catalogs yet</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="cat in catalogs"
          :key="cat.ID"
          class="bg-white rounded-xl border border-gray-200 p-4 flex items-center justify-between hover:border-indigo-300 cursor-pointer"
          @click="router.push(`/catalogs/${cat.ID}`)"
        >
          <div>
            <div class="font-medium text-gray-900">{{ cat.Name }}</div>
            <div class="text-xs text-gray-500 font-mono mt-0.5">{{ cat.ID }}</div>
            <div class="text-xs text-gray-400 mt-1">
              {{ (cat.Agents ?? []).length }} agent(s)
            </div>
          </div>
          <span class="text-gray-300 text-sm">→</span>
        </div>
      </div>
    </div>

    <Modal
      :open="showCreate"
      title="New Catalog"
      confirm-label="Create"
      :loading="creating"
      @close="showCreate = false"
      @confirm="createCatalog"
    >
      <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
      <input
        v-model="newName"
        type="text"
        placeholder="my-catalog"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        @keydown.enter="createCatalog"
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
import type { Catalog } from '../api/types'

const router = useRouter()
const catalogs = ref<Catalog[]>([])
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
    catalogs.value = await api.listCatalogs()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createCatalog() {
  if (!newName.value.trim()) { createError.value = 'Name is required'; return }
  creating.value = true
  createError.value = ''
  try {
    await api.createCatalog(newName.value.trim())
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
