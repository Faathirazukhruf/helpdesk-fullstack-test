<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '../api/client'
import { useAuthStore } from '../stores/auth'
import StatusBadge from '../components/StatusBadge.vue'
import PriorityBadge from '../components/PriorityBadge.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import EmptyState from '../components/EmptyState.vue'
import AlertError from '../components/AlertError.vue'

const router = useRouter()
const authStore = useAuthStore()

const tickets = ref([])
const pagination = ref({ page: 1, limit: 10, total: 0, total_pages: 1 })
const isLoading = ref(false)
const errorMessage = ref('')

// Filter & Search states
const search = ref('')
const statusFilter = ref('')
const priorityFilter = ref('')
const sortBy = ref('created_at_desc')

const fetchTickets = async (page = 1) => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const params = {
      page,
      limit: pagination.value.limit,
      sort: sortBy.value,
    }
    if (search.value.trim()) params.search = search.value.trim()
    if (statusFilter.value) params.status = statusFilter.value
    if (priorityFilter.value) params.priority = priorityFilter.value

    const response = await apiClient.get('/tickets', { params })
    tickets.value = response.data.data
    pagination.value = response.data.pagination
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal memuat daftar tiket.'
  } finally {
    isLoading.value = false
  }
}

// Watch filters: reset ke halaman 1 setiap kali filter berubah
watch([statusFilter, priorityFilter, sortBy], () => {
  fetchTickets(1)
})

let searchTimeout = null
const handleSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchTickets(1)
  }, 400)
}

const formatDate = (isoString) => {
  if (!isoString) return '-'
  const d = new Date(isoString)
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d)
}

onMounted(() => {
  fetchTickets(1)
})
</script>

<template>
  <div class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header Page -->
    <div class="sm:flex sm:items-center sm:justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 tracking-tight">Daftar Tiket Dukungan</h1>
        <p class="mt-1 text-sm text-gray-500">
          <span v-if="authStore.isRequester">Menampilkan seluruh tiket bantuan yang Anda buat.</span>
          <span v-else>Memantau dan menangani seluruh permohonan tiket bantuan karyawan.</span>
        </p>
      </div>
      <div class="mt-4 sm:mt-0">
        <router-link
          to="/tickets/create"
          class="inline-flex items-center justify-center px-4 py-2.5 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors"
        >
          <svg class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Buat Tiket Baru
        </router-link>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Pencarian Judul -->
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">Cari Judul Tiket</label>
          <div class="relative">
            <input
              v-model="search"
              @input="handleSearchInput"
              type="text"
              placeholder="Ketik kata kunci..."
              class="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500"
            />
            <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-gray-400">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>

        <!-- Filter Status -->
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">Status</label>
          <select
            v-model="statusFilter"
            class="w-full py-2 px-3 text-sm border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 bg-white"
          >
            <option value="">Semua Status</option>
            <option value="Open">Open</option>
            <option value="In Progress">In Progress</option>
            <option value="Resolved">Resolved</option>
            <option value="Closed">Closed</option>
          </select>
        </div>

        <!-- Filter Prioritas -->
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">Prioritas</label>
          <select
            v-model="priorityFilter"
            class="w-full py-2 px-3 text-sm border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 bg-white"
          >
            <option value="">Semua Prioritas</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>
        </div>

        <!-- Pengurutan (Sort) -->
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">Urutkan</label>
          <select
            v-model="sortBy"
            class="w-full py-2 px-3 text-sm border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 bg-white"
          >
            <option value="created_at_desc">Terbaru Dibuat</option>
            <option value="created_at_asc">Terlama Dibuat</option>
            <option value="priority_desc">Prioritas Tertinggi</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Tampilan Pesan Error -->
    <AlertError :message="errorMessage" />

    <!-- Tampilan Loading -->
    <LoadingSpinner v-if="isLoading" text="Mengambil data tiket..." />

    <!-- Tampilan Kosong (Empty State) -->
    <EmptyState
      v-else-if="tickets.length === 0"
      title="Tidak ada tiket ditemukan"
      description="Coba sesuaikan kata kunci pencarian atau ubah kriteria filter status Anda."
    >
      <template #action>
        <router-link
          to="/tickets/create"
          class="inline-flex items-center px-4 py-2 text-sm font-medium text-indigo-700 bg-indigo-50 hover:bg-indigo-100 rounded-lg border border-indigo-200"
        >
          Buat Tiket Pertama
        </router-link>
      </template>
    </EmptyState>

    <!-- Tabel Daftar Tiket -->
    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
          <thead class="bg-gray-50 text-gray-600 uppercase text-xs font-semibold">
            <tr>
              <th scope="col" class="px-6 py-3.5">ID</th>
              <th scope="col" class="px-6 py-3.5">Judul & Kategori</th>
              <th scope="col" class="px-6 py-3.5">Prioritas</th>
              <th scope="col" class="px-6 py-3.5">Status</th>
              <th scope="col" class="px-6 py-3.5">Pembuat</th>
              <th scope="col" class="px-6 py-3.5">Assignee</th>
              <th scope="col" class="px-6 py-3.5">Dibuat Pada</th>
              <th scope="col" class="px-6 py-3.5 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr
              v-for="ticket in tickets"
              :key="ticket.id"
              class="hover:bg-gray-50/80 transition-colors cursor-pointer"
              @click="router.push(`/tickets/${ticket.id}`)"
            >
              <td class="px-6 py-4 whitespace-nowrap font-mono text-xs text-gray-500">
                #{{ ticket.id }}
              </td>
              <td class="px-6 py-4">
                <div class="font-semibold text-gray-900 line-clamp-1 hover:text-indigo-600">
                  {{ ticket.title }}
                </div>
                <span class="inline-block mt-0.5 text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded">
                  {{ ticket.category }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <PriorityBadge :priority="ticket.priority" />
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <StatusBadge :status="ticket.status" />
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900">{{ ticket.requester_name }}</div>
                <div class="text-xs text-gray-500">{{ ticket.requester_email }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div v-if="ticket.assignee_name" class="text-sm text-gray-900 font-medium">
                  {{ ticket.assignee_name }}
                </div>
                <span v-else class="text-xs text-gray-400 italic">Belum ditugaskan</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                {{ formatDate(ticket.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium" @click.stop>
                <router-link
                  :to="`/tickets/${ticket.id}`"
                  class="text-indigo-600 hover:text-indigo-900 bg-indigo-50 hover:bg-indigo-100 px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors"
                >
                  Detail
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="bg-gray-50 px-6 py-3.5 border-t border-gray-200 flex items-center justify-between">
        <div class="text-xs text-gray-600">
          Menampilkan <span class="font-semibold">{{ tickets.length }}</span> dari
          <span class="font-semibold">{{ pagination.total }}</span> tiket
        </div>
        <div class="flex space-x-2">
          <button
            :disabled="pagination.page <= 1 || isLoading"
            @click="fetchTickets(pagination.page - 1)"
            class="px-3 py-1.5 text-xs font-medium bg-white border border-gray-300 rounded-lg hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Sebelumnya
          </button>
          <span class="inline-flex items-center px-2 text-xs font-medium text-gray-600">
            Hal {{ pagination.page }} dari {{ pagination.total_pages }}
          </span>
          <button
            :disabled="pagination.page >= pagination.total_pages || isLoading"
            @click="fetchTickets(pagination.page + 1)"
            class="px-3 py-1.5 text-xs font-medium bg-white border border-gray-300 rounded-lg hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Selanjutnya
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
