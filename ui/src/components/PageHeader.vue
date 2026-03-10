<template>
  <div class="bg-white border-b border-gray-200 px-6 py-4">
    <div class="flex items-center justify-between">
      <div>
        <div v-if="backTo" class="mb-1">
          <RouterLink :to="backTo" class="text-xs text-indigo-600 hover:text-indigo-700 font-medium">
            ← {{ backLabel }}
          </RouterLink>
        </div>
        <h1 class="text-xl font-semibold text-gray-900">{{ title }}</h1>
        <p v-if="subtitle" class="text-sm text-gray-500 mt-0.5">{{ subtitle }}</p>
      </div>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

defineProps<{
  title: string
  subtitle?: string
  backTo?: string
  backLabel?: string
}>()
</script>
