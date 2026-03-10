<template>
  <div>
    <PageHeader title="Jobs" subtitle="All asynchronous pipeline jobs">
      <button
        class="text-sm text-indigo-600 hover:text-indigo-700 font-medium"
        @click="load"
      >
        Refresh
      </button>
    </PageHeader>

    <div class="p-6">
      <div v-if="loading" class="space-y-2">
        <div v-for="i in 5" :key="i" class="h-14 bg-gray-100 rounded-xl animate-pulse" />
      </div>

      <div v-else-if="jobs.length === 0" class="text-center py-16 text-gray-400">
        <div class="text-5xl mb-3">⟳</div>
        <p class="font-medium">No jobs</p>
      </div>

      <div v-else class="bg-white rounded-xl border border-gray-200 divide-y divide-gray-100">
        <div
          v-for="job in jobs"
          :key="job.id"
          class="p-4 flex items-center justify-between hover:bg-gray-50 cursor-pointer"
          @click="openJob(job)"
        >
          <div>
            <div class="text-sm font-mono text-gray-700">{{ job.id }}</div>
            <div class="text-xs text-gray-500 mt-0.5">
              {{ job.type }} ·
              {{ job.workspace_id.slice(0, 8) }}… ·
              {{ formatDate(job.created_at) }}
            </div>
          </div>
          <div class="flex items-center gap-3">
            <StatusBadge :status="job.status" />
            <button
              v-if="job.status === 'running' || job.status === 'pending'"
              class="text-xs text-red-500 hover:text-red-700 border border-red-200 px-2 py-1 rounded"
              @click.stop="cancelJob(job.id)"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Job detail panel -->
    <Teleport to="body">
      <div v-if="selectedJob" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
        <div class="absolute inset-0 bg-black/40" @click="selectedJob = null" />
        <div class="relative bg-white rounded-t-2xl sm:rounded-2xl shadow-xl p-6 w-full sm:max-w-lg sm:mx-4 max-h-[80vh] overflow-y-auto">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold text-gray-900">Job Detail</h2>
            <button class="text-gray-400 hover:text-gray-600 text-xl leading-none" @click="selectedJob = null">✕</button>
          </div>

          <div class="space-y-3 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500">ID</span>
              <span class="font-mono text-xs text-gray-700">{{ selectedJob.id }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Status</span>
              <StatusBadge :status="jobStatus?.status ?? selectedJob.status" />
            </div>
            <div v-if="jobStatus?.created_at" class="flex justify-between">
              <span class="text-gray-500">Created</span>
              <span class="text-gray-700">{{ formatDate(jobStatus.created_at) }}</span>
            </div>
            <div v-if="jobStatus?.started_at" class="flex justify-between">
              <span class="text-gray-500">Started</span>
              <span class="text-gray-700">{{ formatDate(jobStatus.started_at) }}</span>
            </div>
            <div v-if="jobStatus?.completed_at" class="flex justify-between">
              <span class="text-gray-500">Completed</span>
              <span class="text-gray-700">{{ formatDate(jobStatus.completed_at) }}</span>
            </div>
          </div>

          <div v-if="jobResult" class="mt-4">
            <div v-if="jobResult.result" class="mt-2 p-3 bg-gray-50 rounded-lg text-sm text-gray-700 whitespace-pre-wrap max-h-48 overflow-y-auto">
              {{ jobResult.result }}
            </div>
            <div v-if="jobResult.error" class="mt-2 p-3 bg-red-50 rounded-lg text-sm text-red-700">
              {{ jobResult.error }}
            </div>
          </div>

          <div class="mt-4 flex justify-end">
            <button
              class="text-sm text-indigo-600 hover:text-indigo-700 font-medium"
              @click="refreshJobDetail"
            >Refresh</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { api } from '../api/client'
import type { Job, JobStatus, JobResult } from '../api/types'

const jobs = ref<Job[]>([])
const loading = ref(true)
const selectedJob = ref<Job | null>(null)
const jobStatus = ref<JobStatus | null>(null)
const jobResult = ref<JobResult | null>(null)

function formatDate(s: string) {
  if (!s) return '—'
  return new Date(s).toLocaleString()
}

async function load() {
  loading.value = true
  try {
    jobs.value = await api.listJobs()
  } finally {
    loading.value = false
  }
}

async function openJob(job: Job) {
  selectedJob.value = job
  jobStatus.value = null
  jobResult.value = null
  await refreshJobDetail()
}

async function refreshJobDetail() {
  if (!selectedJob.value) return
  const [status, result] = await Promise.all([
    api.jobStatus(selectedJob.value.id),
    api.jobResult(selectedJob.value.id),
  ])
  jobStatus.value = status
  jobResult.value = result
}

async function cancelJob(id: string) {
  await api.cancelJob(id)
  await load()
}

onMounted(load)
</script>
