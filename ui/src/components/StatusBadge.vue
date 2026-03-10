<template>
  <span
    class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
    :class="classes"
  >{{ status }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const classes = computed(() => {
  switch (props.status) {
    case 'completed': return 'bg-green-100 text-green-800'
    case 'running':   return 'bg-blue-100 text-blue-800'
    case 'pending':   return 'bg-yellow-100 text-yellow-800'
    case 'failed':    return 'bg-red-100 text-red-800'
    case 'active':    return 'bg-green-100 text-green-800'
    case 'error':     return 'bg-red-100 text-red-800'
    default:          return 'bg-gray-100 text-gray-700'
  }
})
</script>
