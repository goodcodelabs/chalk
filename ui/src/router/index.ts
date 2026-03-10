import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    {
      path: '/dashboard',
      component: () => import('../views/DashboardView.vue'),
    },
    {
      path: '/workspaces',
      component: () => import('../views/WorkspacesView.vue'),
    },
    {
      path: '/workspaces/:id',
      component: () => import('../views/WorkspaceDetailView.vue'),
      props: true,
    },
    {
      path: '/workspaces/:workspaceId/threads/:threadId',
      component: () => import('../views/ThreadChatView.vue'),
      props: true,
    },
    {
      path: '/catalogs',
      component: () => import('../views/CatalogsView.vue'),
    },
    {
      path: '/catalogs/:id',
      component: () => import('../views/CatalogDetailView.vue'),
      props: true,
    },
    {
      path: '/agents/:id',
      component: () => import('../views/AgentDetailView.vue'),
      props: true,
    },
    {
      path: '/agents/:agentId/threads/:threadId',
      component: () => import('../views/AgentThreadChatView.vue'),
      props: true,
    },
    {
      path: '/jobs',
      component: () => import('../views/JobsView.vue'),
    },
  ],
})

export default router
