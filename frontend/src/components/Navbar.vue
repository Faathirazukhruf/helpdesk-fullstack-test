<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const roleBadgeColor = (role) => {
  switch (role) {
    case 'admin':
      return 'bg-purple-100 text-purple-800 border-purple-200'
    case 'agent':
      return 'bg-blue-100 text-blue-800 border-blue-200'
    default:
      return 'bg-emerald-100 text-emerald-800 border-emerald-200'
  }
}
</script>

<template>
  <nav class="bg-white border-b border-gray-200 shadow-sm sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <!-- Logo & Navigation Links -->
        <div class="flex items-center space-x-8">
          <router-link to="/tickets" class="flex items-center space-x-2">
            <span class="p-2 bg-indigo-600 text-white rounded-lg font-bold text-lg">HD</span>
            <span class="font-bold text-xl text-gray-900 tracking-tight">Helpdesk</span>
          </router-link>

          <div v-if="authStore.isAuthenticated" class="hidden sm:flex sm:space-x-4">
            <router-link
              to="/tickets"
              active-class="border-indigo-600 text-gray-900 font-semibold"
              class="inline-flex items-center px-1 pt-1 border-b-2 border-transparent text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300 transition-colors"
            >
              Tiket Dukungan
            </router-link>

            <router-link
              v-if="authStore.isAdmin"
              to="/admin/users"
              active-class="border-indigo-600 text-gray-900 font-semibold"
              class="inline-flex items-center px-1 pt-1 border-b-2 border-transparent text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300 transition-colors"
            >
              Kelola Pengguna
            </router-link>
          </div>
        </div>

        <!-- User Info & Logout Button -->
        <div v-if="authStore.isAuthenticated" class="flex items-center space-x-4">
          <div class="hidden md:flex flex-col text-right">
            <span class="text-sm font-semibold text-gray-900">{{ authStore.user?.name }}</span>
            <span class="text-xs text-gray-500">{{ authStore.user?.email }}</span>
          </div>

          <span
            :class="roleBadgeColor(authStore.user?.role)"
            class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize border"
          >
            {{ authStore.user?.role }}
          </span>

          <button
            @click="handleLogout"
            class="inline-flex items-center px-3 py-1.5 border border-gray-300 text-xs font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 hover:text-red-600 transition-colors"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  </nav>
</template>
