import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/tickets',
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/tickets',
    name: 'tickets',
    component: () => import('../views/TicketListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: () => import('../views/TicketListView.vue'), // Placeholder sebelum Fase 9
    meta: { requiresAuth: true, requiresAdmin: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation Guard: Melindungi rute yang memerlukan login atau hak akses admin
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Jika memiliki token tetapi data profil user belum dimuat ke memory, ambil profilnya
  if (authStore.token && !authStore.user) {
    await authStore.fetchMe()
  }

  // 1. Rute yang butuh login: tendang ke /login jika belum authenticated
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next({ name: 'login', query: { redirect: to.fullPath } })
  }

  // 2. Rute khusus tamu (login/register): jika sudah login, redirect ke /tickets
  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return next({ name: 'tickets' })
  }

  // 3. Rute khusus admin: jika bukan admin, cegah akses dan redirect ke /tickets
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return next({ name: 'tickets' })
  }

  next()
})

export default router
