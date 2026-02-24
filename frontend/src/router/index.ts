import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { title: '控制台' }
  },
  {
    path: '/chat',
    name: 'Chat',
    component: () => import('../views/Chat.vue'),
    meta: { title: '对话' }
  },
  {
    path: '/workflow',
    name: 'Workflow',
    component: () => import('../views/WorkflowEditor.vue'),
    meta: { title: '工作流编辑器' }
  },
  {
    path: '/workflow/builder',
    name: 'WorkflowBuilder',
    component: () => import('../views/WorkflowBuilder.vue'),
    meta: { title: '工作流构建器' }
  },
  {
    path: '/workflow/:id',
    name: 'WorkflowDetail',
    component: () => import('../views/WorkflowDetail.vue'),
    meta: { title: '工作流详情' }
  },
  {
    path: '/agents',
    name: 'Agents',
    component: () => import('../views/AgentStudio.vue'),
    meta: { title: 'Agent Studio' }
  },
  {
    path: '/skills',
    name: 'Skills',
    component: () => import('../views/SkillManager.vue'),
    meta: { title: '技能管理' }
  },
  {
    path: '/config',
    name: 'Config',
    component: () => import('../views/ConfigManager.vue'),
    meta: { title: '配置管理' }
  },
  {
    path: '/files',
    name: 'Files',
    component: () => import('../views/FileExplorer.vue'),
    meta: { title: '文件浏览器' }
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('../views/Logs.vue'),
    meta: { title: '日志' }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 设置页面标题
router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} - AgentFramework`
  }
  next()
})

export default router
