<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import apiClient from '../api/client'
import { useAuthStore } from '../stores/auth'
import StatusBadge from '../components/StatusBadge.vue'
import PriorityBadge from '../components/PriorityBadge.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import AlertError from '../components/AlertError.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const ticket = ref(null)
const isLoading = ref(true)
const errorMessage = ref('')
const actionMessage = ref('')

// State form komentar
const newComment = ref('')
const isInternal = ref(false)
const isSubmittingComment = ref(false)
const commentError = ref('')

// State penugasan (assignee) & prioritas untuk Agent/Admin
const isUpdating = ref(false)

const ticketId = computed(() => route.params.id)

// Evaluasi tombol aksi perubahan status yang legal berdasarkan peran dan status tiket saat ini
const availableActions = computed(() => {
  if (!ticket.value || !authStore.user) return []

  const currentStatus = ticket.value.status
  const role = authStore.user.role
  const actions = []

  // Matriks Aturan Spesifikasi:
  // 1. Status 'Open'
  if (currentStatus === 'Open') {
    if (role === 'agent' || role === 'admin') {
      actions.push({
        label: 'Mulai Kerjakan (In Progress)',
        toStatus: 'In Progress',
        color: 'bg-blue-600 hover:bg-blue-700 text-white',
      })
    }
    if (role === 'requester' || role === 'admin') {
      actions.push({
        label: 'Batalkan Tiket (Closed)',
        toStatus: 'Closed',
        color: 'bg-gray-600 hover:bg-gray-700 text-white',
      })
    }
  }

  // 2. Status 'In Progress'
  if (currentStatus === 'In Progress') {
    if (role === 'agent' || role === 'admin') {
      actions.push({
        label: 'Selesaikan Tiket (Resolved)',
        toStatus: 'Resolved',
        color: 'bg-emerald-600 hover:bg-emerald-700 text-white',
      })
    }
  }

  // 3. Status 'Resolved'
  if (currentStatus === 'Resolved') {
    if (role === 'requester') {
      actions.push({
        label: 'Buka Kembali Tiket (Open)',
        toStatus: 'Open',
        color: 'bg-amber-600 hover:bg-amber-700 text-white',
      })
    }
    // Semua peran boleh menutup tiket yang sudah resolved
    actions.push({
      label: 'Tutup Tiket (Closed)',
      toStatus: 'Closed',
      color: 'bg-gray-800 hover:bg-gray-900 text-white',
    })
  }

  return actions
})

const fetchTicketDetail = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await apiClient.get(`/tickets/${ticketId.value}`)
    ticket.value = response.data
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal memuat detail tiket.'
  } finally {
    isLoading.value = false
  }
}

// Handler perubahan status via tombol aksi
const handleStatusChange = async (newStatus) => {
  isUpdating.value = true
  actionMessage.value = ''
  errorMessage.value = ''
  try {
    const payload = { status: newStatus }

    // Jika agent mengambil tiket open dan belum ada assignee, otomatis assign ke dirinya
    if (newStatus === 'In Progress' && !ticket.value.assignee_id && authStore.isAgent) {
      payload.assignee_id = authStore.user.id
    }

    const response = await apiClient.patch(`/tickets/${ticketId.value}`, payload)
    const prevComments = ticket.value?.comments || []
    ticket.value = {
      ...response.data,
      comments: response.data.comments || prevComments,
    }
    actionMessage.value = `Status tiket berhasil diubah menjadi "${newStatus}".`
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengubah status tiket.'
  } finally {
    isUpdating.value = false
  }
}

// Handler ambil tiket oleh Agent/Admin
const handleAssignToMe = async () => {
  isUpdating.value = true
  try {
    const response = await apiClient.patch(`/tickets/${ticketId.value}`, {
      assignee_id: authStore.user.id,
    })
    const prevComments = ticket.value?.comments || []
    ticket.value = {
      ...response.data,
      comments: response.data.comments || prevComments,
    }
    actionMessage.value = 'Tiket berhasil Anda ambil alih.'
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal menetapkan tiket.'
  } finally {
    isUpdating.value = false
  }
}

// Handler ubah prioritas oleh Agent/Admin
const handlePriorityChange = async (newPriority) => {
  isUpdating.value = true
  actionMessage.value = ''
  errorMessage.value = ''
  try {
    const response = await apiClient.patch(`/tickets/${ticketId.value}`, {
      priority: newPriority,
    })
    const prevComments = ticket.value?.comments || []
    ticket.value = {
      ...response.data,
      comments: response.data.comments || prevComments,
    }
    actionMessage.value = `Prioritas tiket berhasil diperbarui ke "${newPriority}".`
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengubah prioritas tiket.'
  } finally {
    isUpdating.value = false
  }
}

// Handler kirim komentar / internal note
const handleSubmitComment = async () => {
  if (!newComment.value.trim()) return

  isSubmittingComment.value = true
  commentError.value = ''
  try {
    const payload = {
      body: newComment.value.trim(),
      is_internal: isInternal.value,
    }

    await apiClient.post(`/tickets/${ticketId.value}/comments`, payload)
    newComment.value = ''
    isInternal.value = false

    // Refresh tiket dan riwayat komentar
    await fetchTicketDetail()
  } catch (err) {
    commentError.value = err.response?.data?.error || 'Gagal mengirim komentar.'
  } finally {
    isSubmittingComment.value = false
  }
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
  fetchTicketDetail()
})
</script>

<template>
  <div class="max-w-5xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Breadcrumb -->
    <div class="mb-6">
      <router-link
        to="/tickets"
        class="inline-flex items-center text-sm font-medium text-gray-500 hover:text-gray-700 transition-colors"
      >
        <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        Kembali ke Daftar Tiket
      </router-link>
    </div>

    <!-- Tampilan Loading -->
    <LoadingSpinner v-if="isLoading" text="Mengambil informasi tiket..." />

    <!-- Tampilan Error Utama -->
    <AlertError v-else-if="errorMessage && !ticket" :message="errorMessage" />

    <!-- Konten Utama Tiket -->
    <div v-else-if="ticket" class="space-y-6">
      <!-- Notifikasi Aksi -->
      <div v-if="actionMessage" class="bg-emerald-50 border border-emerald-200 text-emerald-800 px-4 py-3 rounded-lg text-sm flex justify-between items-center">
        <span>{{ actionMessage }}</span>
        <button @click="actionMessage = ''" class="text-emerald-600 hover:text-emerald-800 font-bold">✕</button>
      </div>

      <AlertError v-if="errorMessage" :message="errorMessage" />

      <!-- Card Informasi Tiket -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <!-- Header Tiket -->
        <div class="p-6 border-b border-gray-200">
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <div class="flex items-center space-x-3 mb-2">
                <span class="text-xs font-mono font-bold text-gray-400">#{{ ticket.id }}</span>
                <span class="text-xs font-semibold px-2 py-0.5 rounded bg-gray-100 text-gray-600">
                  {{ ticket.category }}
                </span>
                <div class="inline-flex items-center space-x-1.5">
                  <PriorityBadge :priority="ticket.priority" />
                  <select
                    v-if="!authStore.isRequester && ticket.status !== 'Closed'"
                    :value="ticket.priority"
                    @change="handlePriorityChange($event.target.value)"
                    :disabled="isUpdating"
                    class="text-xs border-gray-300 rounded px-1.5 py-0.5 bg-white border text-gray-700 hover:border-gray-400 focus:ring-1 focus:ring-indigo-500 cursor-pointer"
                    title="Ubah prioritas tiket"
                  >
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </div>
                <StatusBadge :status="ticket.status" />
              </div>
              <h1 class="text-2xl font-bold text-gray-900 tracking-tight">{{ ticket.title }}</h1>
            </div>

            <!-- Tombol Transisi Status Dinamis Sesuai Matriks Peran -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="act in availableActions"
                :key="act.toStatus"
                :disabled="isUpdating"
                @click="handleStatusChange(act.toStatus)"
                :class="act.color"
                class="px-4 py-2 rounded-lg text-xs font-bold shadow-sm transition-colors disabled:opacity-50"
              >
                {{ act.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Meta Info & Assignee -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 p-6 bg-gray-50 border-b border-gray-200 text-xs">
          <div>
            <span class="text-gray-500 font-medium block">Pemohon (Requester):</span>
            <span class="font-semibold text-gray-900 text-sm">{{ ticket.requester_name }}</span>
            <span class="text-gray-500 block">{{ ticket.requester_email }}</span>
          </div>

          <div>
            <span class="text-gray-500 font-medium block">Ditugaskan Kepada:</span>
            <div v-if="ticket.assignee_name">
              <span class="font-semibold text-gray-900 text-sm">{{ ticket.assignee_name }}</span>
              <span class="text-gray-500 block">{{ ticket.assignee_email }}</span>
              <button
                v-if="!authStore.isRequester && ticket.assignee_id !== authStore.user.id && ticket.status !== 'Closed'"
                @click="handleAssignToMe"
                :disabled="isUpdating"
                class="text-indigo-600 hover:text-indigo-800 font-medium underline text-xs mt-1 inline-block"
              >
                Tugaskan ke Saya
              </button>
            </div>
            <div v-else class="flex items-center space-x-2 mt-1">
              <span class="text-gray-400 italic">Belum ada</span>
              <button
                v-if="!authStore.isRequester"
                @click="handleAssignToMe"
                :disabled="isUpdating"
                class="text-indigo-600 hover:text-indigo-800 font-semibold underline text-xs"
              >
                Ambil Tiket Ini
              </button>
            </div>
          </div>

          <div>
            <span class="text-gray-500 font-medium block">Waktu Dibuat:</span>
            <span class="font-semibold text-gray-900">{{ formatDate(ticket.created_at) }}</span>
            <span class="text-gray-400 block">Update: {{ formatDate(ticket.updated_at) }}</span>
          </div>
        </div>

        <!-- Deskripsi Permohonan -->
        <div class="p-6">
          <h3 class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-2">Deskripsi Masalah:</h3>
          <div class="prose max-w-none text-gray-800 text-sm leading-relaxed whitespace-pre-wrap bg-gray-50 p-4 rounded-lg border border-gray-100">
            {{ ticket.description }}
          </div>
        </div>
      </div>

      <!-- Riwayat Komentar -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-bold text-gray-900 tracking-tight mb-4 flex items-center justify-between">
          <span>Riwayat Diskusi & Tanggapan</span>
          <span class="text-xs font-normal text-gray-500">
            {{ ticket.comments?.length || 0 }} Komentar
          </span>
        </h2>

        <!-- Daftar Komentar -->
        <div v-if="ticket.comments && ticket.comments.length > 0" class="space-y-4 mb-8">
          <div
            v-for="c in ticket.comments"
            :key="c.id"
            :class="c.is_internal ? 'bg-amber-50/70 border-amber-200' : 'bg-gray-50 border-gray-200'"
            class="p-4 rounded-xl border transition-all"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center space-x-2">
                <span class="font-bold text-sm text-gray-900">{{ c.author_name }}</span>
                <span class="text-xs px-2 py-0.5 rounded capitalize bg-white text-gray-600 border border-gray-200 font-medium">
                  {{ c.author_role }}
                </span>
                <!-- Label Catatan Internal Khusus Staf -->
                <span
                  v-if="c.is_internal"
                  class="bg-amber-200 text-amber-900 text-xs px-2 py-0.5 rounded-full font-bold flex items-center"
                >
                  🔒 Catatan Internal Staf
                </span>
              </div>
              <span class="text-xs text-gray-400">{{ formatDate(c.created_at) }}</span>
            </div>

            <p class="text-sm text-gray-800 whitespace-pre-wrap leading-relaxed">{{ c.body }}</p>
          </div>
        </div>

        <div v-else class="text-center py-6 text-gray-400 text-sm italic mb-6">
          Belum ada komentar pada tiket ini.
        </div>

        <!-- Form Tambah Komentar -->
        <div class="pt-6 border-t border-gray-200">
          <h3 class="text-sm font-semibold text-gray-900 mb-3">Tulis Komentar / Tanggapan:</h3>
          <AlertError :message="commentError" />

          <form @submit.prevent="handleSubmitComment" class="space-y-3">
            <div>
              <textarea
                v-model="newComment"
                rows="3"
                placeholder="Tulis tanggapan atau solusi untuk tiket ini..."
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-indigo-500 focus:border-indigo-500 shadow-sm"
              ></textarea>
            </div>

            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <!-- Checkbox Catatan Internal (HANYA DITAMPILKAN UNTUK AGENT & ADMIN) -->
              <div v-if="!authStore.isRequester" class="flex items-center space-x-2">
                <input
                  id="internalCheck"
                  v-model="isInternal"
                  type="checkbox"
                  class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                />
                <label for="internalCheck" class="text-xs font-semibold text-amber-800 cursor-pointer flex items-center">
                  <span>Jadikan Catatan Internal (Hanya terlihat oleh Agent & Admin)</span>
                </label>
              </div>
              <div v-else></div>

              <button
                type="submit"
                :disabled="isSubmittingComment || !newComment.trim()"
                class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-xs font-bold rounded-lg shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 transition-colors"
              >
                <span v-if="isSubmittingComment">Mengirim...</span>
                <span v-else>Kirim Komentar</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
