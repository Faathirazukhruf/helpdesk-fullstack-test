<script setup>
import { ref, onMounted } from 'vue'
import apiClient from '../api/client'
import { useAuthStore } from '../stores/auth'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import AlertError from '../components/AlertError.vue'

const authStore = useAuthStore()

const users = ref([])
const isLoading = ref(true)
const errorMessage = ref('')
const successMessage = ref('')
const updatingUserId = ref(null)

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

const fetchUsers = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await apiClient.get('/users')
    users.value = response.data
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal memuat daftar pengguna.'
  } finally {
    isLoading.value = false
  }
}

const handleRoleChange = async (user, newRole) => {
  if (user.role === newRole) return

  updatingUserId.value = user.id
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const response = await apiClient.patch(`/users/${user.id}/role`, { role: newRole })
    user.role = response.data.role
    successMessage.value = `Peran pengguna "${user.name}" berhasil diubah menjadi "${newRole}".`
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengubah peran pengguna.'
    // Revert role selection in UI if failed
    await fetchUsers()
  } finally {
    updatingUserId.value = null
  }
}

const formatDate = (isoString) => {
  if (!isoString) return '-'
  const d = new Date(isoString)
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(d)
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header Page -->
    <div class="mb-6">
      <div class="flex items-center space-x-3">
        <span class="p-2 bg-purple-100 text-purple-700 rounded-lg">
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
        </span>
        <div>
          <h1 class="text-2xl font-bold text-gray-900 tracking-tight">Kelola Pengguna & Peran</h1>
          <p class="mt-0.5 text-sm text-gray-500">
            Halaman khusus Administrator untuk melihat seluruh akun dan mengatur hak akses (role).
          </p>
        </div>
      </div>
    </div>

    <!-- Alert Notifikasi -->
    <div v-if="successMessage" class="bg-emerald-50 border border-emerald-200 text-emerald-800 px-4 py-3 rounded-xl text-sm flex justify-between items-center mb-6 shadow-sm">
      <div class="flex items-center space-x-2">
        <span>✓</span>
        <span>{{ successMessage }}</span>
      </div>
      <button @click="successMessage = ''" class="text-emerald-600 hover:text-emerald-800 font-bold">✕</button>
    </div>

    <AlertError :message="errorMessage" />

    <!-- Loading Spinner -->
    <LoadingSpinner v-if="isLoading" text="Mengambil daftar pengguna..." />

    <!-- Tabel Pengguna -->
    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
          <thead class="bg-gray-50 text-gray-600 uppercase text-xs font-semibold">
            <tr>
              <th scope="col" class="px-6 py-3.5">ID</th>
              <th scope="col" class="px-6 py-3.5">Nama & Email</th>
              <th scope="col" class="px-6 py-3.5">Peran Aktif</th>
              <th scope="col" class="px-6 py-3.5">Terdaftar Sejak</th>
              <th scope="col" class="px-6 py-3.5 text-right">Ubah Peran Hak Akses</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr
              v-for="u in users"
              :key="u.id"
              class="hover:bg-gray-50/80 transition-colors"
            >
              <td class="px-6 py-4 whitespace-nowrap font-mono text-xs text-gray-400">
                #{{ u.id }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <span class="font-bold text-gray-900">{{ u.name }}</span>
                  <span v-if="u.id === authStore.user?.id" class="text-xs px-2 py-0.5 rounded bg-gray-100 text-gray-600 font-medium border">
                    (Akun Anda)
                  </span>
                </div>
                <div class="text-xs text-gray-500">{{ u.email }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  :class="roleBadgeColor(u.role)"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize border shadow-sm"
                >
                  {{ u.role }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                {{ formatDate(u.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <!-- Dropdown Pemilihan Peran -->
                <div class="inline-flex items-center space-x-2">
                  <select
                    :value="u.role"
                    :disabled="updatingUserId === u.id"
                    @change="handleRoleChange(u, $event.target.value)"
                    class="text-xs font-medium py-1.5 px-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-purple-500 focus:border-purple-500 bg-white disabled:opacity-50"
                  >
                    <option value="requester">Requester (Karyawan)</option>
                    <option value="agent">Agent (Support)</option>
                    <option value="admin">Admin (Administrator)</option>
                  </select>

                  <span v-if="updatingUserId === u.id" class="text-xs text-purple-600 animate-pulse font-medium">
                    Menyimpan...
                  </span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer Info -->
      <div class="bg-gray-50 px-6 py-3 border-t border-gray-200 text-xs text-gray-500">
        Total Pengguna: <span class="font-semibold text-gray-700">{{ users.length }}</span> akun terdaftar.
      </div>
    </div>
  </div>
</template>
