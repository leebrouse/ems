import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from '@/components/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: MainLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('../views/DashboardView.vue')
        },
        {
          path: 'warehouse',
          name: 'Warehouse',
          component: () => import('../views/WarehouseView.vue'),
          meta: { roles: ['Admin', 'WarehouseManager'] }
        },
        {
          path: 'scheduling',
          name: 'Scheduling',
          component: () => import('../views/SchedulingView.vue'),
          meta: { roles: ['Admin', 'Dispatcher'] }
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('../views/UsersView.vue'),
          meta: { roles: ['Admin'] }
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const roles = to.meta.roles as string[] | undefined

  if (!to.meta.public && !authStore.isAuthenticated) {
    next('/login')
  } else if (roles && !roles.some(role => authStore.user?.roles.includes(role))) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
