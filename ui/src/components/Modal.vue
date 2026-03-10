<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center"
    >
      <div class="absolute inset-0 bg-black/40" @click="$emit('close')" />
      <div class="relative bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">{{ title }}</h2>
        <slot />
        <div class="flex justify-end gap-3 mt-6">
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            type="button"
            :disabled="loading"
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 disabled:opacity-50"
            @click="$emit('confirm')"
          >
            {{ loading ? 'Saving…' : confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  confirmLabel?: string
  loading?: boolean
}>()
defineEmits<{
  close: []
  confirm: []
}>()
</script>
