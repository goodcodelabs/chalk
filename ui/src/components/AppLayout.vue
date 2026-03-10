<template>
  <div class="flex h-screen bg-gray-50 overflow-hidden">
    <!-- Sidebar -->
    <aside class="w-56 flex-shrink-0 bg-slate-900 flex flex-col">
      <!-- Logo -->
      <div class="px-4 py-5 border-b border-slate-700">
        <div class="flex items-center gap-2">
          <div class="w-7 h-7 bg-indigo-500 rounded-md flex items-center justify-center">
            <span class="text-white text-sm font-bold">S</span>
          </div>
          <span class="text-white font-semibold text-base">Slate Admin</span>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-2 py-4 space-y-0.5 overflow-y-auto">
        <NavItem to="/dashboard" label="Dashboard" icon="dashboard" />
        <NavItem to="/workspaces" label="Workspaces" icon="workspace" />
        <NavItem to="/catalogs" label="Catalogs" icon="catalog" />
        <NavItem to="/jobs" label="Jobs" icon="jobs" />
      </nav>

      <!-- Status indicator -->
      <div class="px-4 py-3 border-t border-slate-700">
        <div class="flex items-center gap-2">
          <div
            class="w-2 h-2 rounded-full"
            :class="connected ? 'bg-green-400' : 'bg-red-400'"
          />
          <span class="text-slate-400 text-xs">
            {{ connected ? 'Connected' : 'Disconnected' }}
          </span>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-y-auto">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import NavItem from './NavItem.vue'
import { api } from '../api/client'

const connected = ref(false)

onMounted(async () => {
  try {
    await api.health()
    connected.value = true
  } catch {
    connected.value = false
  }
  // Poll health every 10s
  setInterval(async () => {
    try {
      await api.health()
      connected.value = true
    } catch {
      connected.value = false
    }
  }, 10_000)
})
</script>
