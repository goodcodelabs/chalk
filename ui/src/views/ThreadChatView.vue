<template>
  <div class="flex flex-col h-full">
    <PageHeader
      :title="thread?.name ?? 'Thread'"
      :subtitle="threadId"
      :back-to="`/workspaces/${workspaceId}`"
      back-label="Workspace"
    />

    <!-- Messages -->
    <div ref="messagesEl" class="flex-1 overflow-y-auto p-6 space-y-4 bg-gray-50">
      <div v-if="loading" class="flex justify-center py-8">
        <div class="text-gray-400 text-sm">Loading…</div>
      </div>

      <div v-else-if="messages.length === 0" class="text-center text-gray-400 py-12">
        <p class="font-medium">No messages yet</p>
        <p class="text-sm mt-1">Send a message to start the conversation.</p>
      </div>

      <div
        v-for="(msg, i) in messages"
        :key="i"
        class="flex"
        :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <div
          class="max-w-[70%] rounded-2xl px-4 py-3 text-sm"
          :class="msg.role === 'user'
            ? 'bg-indigo-600 text-white rounded-br-sm'
            : 'bg-white border border-gray-200 text-gray-800 rounded-bl-sm'"
        >
          <div class="text-xs font-medium mb-1 opacity-70">{{ msg.role }}</div>
          <div class="whitespace-pre-wrap">{{ messageText(msg) }}</div>
        </div>
      </div>

      <!-- Sending indicator -->
      <div v-if="sending" class="flex justify-start">
        <div class="bg-white border border-gray-200 rounded-2xl rounded-bl-sm px-4 py-3">
          <div class="flex gap-1">
            <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:0ms" />
            <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:150ms" />
            <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:300ms" />
          </div>
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="bg-white border-t border-gray-200 p-4">
      <div v-if="error" class="mb-2 text-xs text-red-600">{{ error }}</div>
      <div class="flex gap-3">
        <textarea
          v-model="input"
          rows="2"
          placeholder="Type a message…"
          class="flex-1 px-3 py-2 border border-gray-300 rounded-xl text-sm resize-none focus:outline-none focus:ring-2 focus:ring-indigo-500"
          @keydown.enter.exact.prevent="send"
        />
        <button
          :disabled="!input.trim() || sending"
          class="px-5 py-2 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 disabled:opacity-50 self-end"
          @click="send"
        >
          Send
        </button>
      </div>
      <div class="text-xs text-gray-400 mt-1">Enter to send</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import { api } from '../api/client'
import type { Message, Thread } from '../api/types'

const props = defineProps<{
  workspaceId: string
  threadId: string
}>()

const messages = ref<Message[]>([])
const thread = ref<Thread | null>(null)
const loading = ref(true)
const sending = ref(false)
const input = ref('')
const error = ref('')
const messagesEl = ref<HTMLElement | null>(null)

function messageText(msg: Message): string {
  if (typeof msg.content === 'string') return msg.content
  if (Array.isArray(msg.content)) {
    return msg.content
      .filter(b => b.type === 'text')
      .map(b => b.text ?? '')
      .join('')
  }
  return ''
}

async function scrollToBottom() {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

async function load() {
  loading.value = true
  try {
    const [threads, history] = await Promise.all([
      api.listThreads(props.workspaceId),
      api.threadHistory(props.threadId),
    ])
    thread.value = threads.find(t => t.id === props.threadId) ?? null
    messages.value = history
    await scrollToBottom()
  } finally {
    loading.value = false
  }
}

async function send() {
  const msg = input.value.trim()
  if (!msg || sending.value) return

  input.value = ''
  error.value = ''
  messages.value.push({ role: 'user', content: msg })
  await scrollToBottom()
  sending.value = true

  try {
    const { job_id } = await api.chat(props.threadId, msg)
    while (true) {
      await new Promise(r => setTimeout(r, 500))
      const result = await api.jobResult(job_id)
      if (result.status === 'completed') {
        messages.value.push({ role: 'assistant', content: result.result })
        await scrollToBottom()
        break
      } else if (result.status === 'failed') {
        error.value = result.error || 'Job failed'
        break
      }
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    sending.value = false
  }
}

onMounted(load)
</script>
