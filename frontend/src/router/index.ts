import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/workflow',
    name: 'Workflow',
    component: () => import('../views/WorkflowEditor.vue')
  },
  {
    path: '/skills',
    name: 'Skills',
    component: () => import('../views/SkillManager.vue')
  },
  {
    path: '/config',
    name: 'Config',
    component: () => import('../views/ConfigManager.vue')
  },
  {
    path: '/files',
    name: 'Files',
    component: () => import('../views/FileExplorer.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
