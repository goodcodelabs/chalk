<template>
  <RouterLink
    :to="to"
    class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors"
    :class="isActive
      ? 'bg-indigo-600 text-white'
      : 'text-slate-300 hover:bg-slate-800 hover:text-white'"
  >
    <span class="text-base leading-none">{{ iconChar }}</span>
    {{ label }}
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const props = defineProps<{
  to: string
  label: string
  icon: string
}>()

const route = useRoute()
const isActive = computed(() => route.path.startsWith(props.to))

const iconChar = computed(() => {
  const icons: Record<string, string> = {
    dashboard: '◈',
    workspace: '⬡',
    catalog:   '☰',
    jobs:      '⟳',
  }
  return icons[props.icon] ?? '○'
})
</script>
