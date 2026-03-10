<template>
  <div>
    <PageHeader title="Dashboard" subtitle="System overview" />

    <div class="p-6">
      <!-- Connection error -->
      <div v-if="error" class="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <!-- Stats grid -->
      <div v-if="stats" class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="LLM Calls" :value="stats.llm_calls" color="indigo" />
        <StatCard label="Tool Calls" :value="stats.tool_calls" color="violet" />
        <StatCard label="Active Connections" :value="stats.active_connections" color="blue" />
        <StatCard label="Errors" :value="stats.errors" color="red" />
        <StatCard label="Input Tokens" :value="stats.input_tokens_total.toLocaleString()" color="gray" />
        <StatCard label="Output Tokens" :value="stats.output_tokens_total.toLocaleString()" color="gray" />
        <StatCard label="Scheduler Queue" :value="stats.scheduler_queue" color="amber" />
        <StatCard label="Jobs Running" :value="stats.jobs?.running ?? 0" color="blue" />
      </div>

      <!-- Loading -->
      <div v-else-if="loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <div v-for="i in 8" :key="i" class="h-24 bg-gray-100 rounded-xl animate-pulse" />
      </div>

      <!-- Job breakdown -->
      <div v-if="stats?.jobs" class="bg-white rounded-xl border border-gray-200 p-5">
        <h2 class="text-sm font-semibold text-gray-700 mb-4">Job Status Breakdown</h2>
        <div class="grid grid-cols-4 gap-4">
          <div v-for="(count, status) in stats.jobs" :key="status" class="text-center">
            <div class="text-2xl font-bold text-gray-900">{{ count }}</div>
            <div class="text-xs text-gray-500 mt-1 capitalize">{{ status }}</div>
          </div>
        </div>
      </div>

      <div class="mt-4 text-right">
        <button
          class="text-sm text-indigo-600 hover:text-indigo-700 font-medium"
          @click="load"
        >Refresh</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import { api } from '../api/client'
import type { SystemStats } from '../api/types'

const stats = ref<SystemStats | null>(null)
const loading = ref(true)
const error = ref('')

const StatCard = {
  props: ['label', 'value', 'color'],
  template: `
    <div class="bg-white rounded-xl border border-gray-200 p-4">
      <div class="text-xs font-medium text-gray-500 mb-1">{{ label }}</div>
      <div class="text-2xl font-bold text-gray-900">{{ value }}</div>
    </div>
  `,
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    stats.value = await api.stats()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
