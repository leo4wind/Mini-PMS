import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { APP_BASE, stripAppBase } from '@/config'

const router = createRouter({
  history: createWebHistory(APP_BASE),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        {
          path: 'products',
          name: 'products',
          component: () => import('@/views/product/ProductList.vue'),
          children: [
            { path: 'new', name: 'product-new', component: () => import('@/views/product/ProductForm.vue') },
            { path: ':id', name: 'product-detail', component: () => import('@/views/product/ProductDetail.vue') },
            { path: ':id/edit', name: 'product-edit', component: () => import('@/views/product/ProductForm.vue') },
          ],
        },
        {
          path: 'stories',
          name: 'stories',
          component: () => import('@/views/story/StoryList.vue'),
          children: [
            { path: 'new', name: 'story-new', component: () => import('@/views/story/StoryForm.vue') },
            { path: ':id', name: 'story-detail', component: () => import('@/views/story/StoryDetail.vue') },
            { path: ':id/edit', name: 'story-edit', component: () => import('@/views/story/StoryForm.vue') },
          ],
        },
        {
          path: 'projects',
          name: 'projects',
          component: () => import('@/views/project/ProjectList.vue'),
          children: [
            { path: 'new', name: 'project-new', component: () => import('@/views/project/ProjectForm.vue') },
            { path: ':id', name: 'project-detail', component: () => import('@/views/project/ProjectDetail.vue') },
            { path: ':id/edit', name: 'project-edit', component: () => import('@/views/project/ProjectForm.vue') },
          ],
        },
        {
          path: 'sprints',
          name: 'sprints',
          component: () => import('@/views/sprint/SprintList.vue'),
          children: [
            { path: 'new', name: 'sprint-new', component: () => import('@/views/sprint/SprintForm.vue') },
            { path: ':id', name: 'sprint-detail', component: () => import('@/views/sprint/SprintDetail.vue') },
            { path: ':id/edit', name: 'sprint-edit', component: () => import('@/views/sprint/SprintForm.vue') },
          ],
        },
        {
          path: 'bugs',
          name: 'bugs',
          component: () => import('@/views/bug/BugList.vue'),
          children: [
            { path: 'new', name: 'bug-new', component: () => import('@/views/bug/BugForm.vue') },
            { path: ':id', name: 'bug-detail', component: () => import('@/views/bug/BugDetail.vue') },
            { path: ':id/edit', name: 'bug-edit', component: () => import('@/views/bug/BugForm.vue') },
          ],
        },
        { path: 'system/users', name: 'users', component: () => import('@/views/system/UserList.vue') },
        { path: 'system/roles', name: 'roles', component: () => import('@/views/system/RoleList.vue') },
        { path: 'system/roles/:id/menus', name: 'role-menus', component: () => import('@/views/system/RoleMenus.vue') },
        { path: 'system/menus', name: 'menus', component: () => import('@/views/system/MenuList.vue') },
        { path: 'profile/password', name: 'password', component: () => import('@/views/PasswordView.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (auth.isLogin && to.name === 'login') return { name: 'dashboard' }
    return true
  }
  if (!auth.isLogin) {
    return {
      name: 'login',
      query: { redirect: stripAppBase(to.fullPath) },
    }
  }
  if (!auth.user) {
    try {
      await auth.loadMe()
    } catch {
      auth.logoutLocal()
      return { name: 'login' }
    }
  }
  return true
})

export default router
