import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',        redirect: '/login' },
    { path: '/login',   component: () => import('@/views/LoginView.vue'),   meta: { guest: true } },
    { path: '/dashboard', component: () => import('@/views/DashboardView.vue'), meta: { auth: true } },
    { path: '/tickets/new', component: () => import('@/views/NewTicketView.vue'), meta: { auth: true } },
    { path: '/tickets/:id/edit', component: () => import('@/views/EditTicketView.vue'), meta: { auth: true } },
    { path: '/profile', component: () => import('@/views/ProfileView.vue'), meta: { auth: true } },
    { path: '/admin',   component: () => import('@/views/AdminView.vue'),   meta: { auth: true, admin: true } },
    { path: '/:pathMatch(.*)*', redirect: '/login' },
  ]
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isAuth)      return next('/login')
  if (to.meta.admin && !auth.isAdmin)    return next('/dashboard')
  if (to.meta.guest && auth.isAuth) {
    return next(auth.isAdmin ? '/admin' : '/dashboard')
  }
  next()
})

export default router
