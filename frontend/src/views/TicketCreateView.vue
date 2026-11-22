<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '../api/client'
import AlertError from '../components/AlertError.vue'

const router = useRouter()

const title = ref('')
const category = ref('')
const priority = ref('medium')
const description = ref('')

const clientError = ref('')
const serverError = ref('')
const isSubmitting = ref(false)

const handleSubmit = async () => {
  clientError.value = ''
  serverError.value = ''

  // 1. Validasi Sisi Klien
  if (!title.value.trim()) {
    clientError.value = 'Judul tiket wajib diisi.'
    return
  }
  if (!category.value.trim()) {
    clientError.value = 'Kategori permohonan wajib dipilih.'
    return
  }
  if (!description.value.trim()) {
    clientError.value = 'Deskripsi kendala atau permintaan wajib diisi.'
    return
  }

  // 2. Kirim Permintaan ke Backend
  isSubmitting.value = true
  try {
    const payload = {
      title: title.value.trim(),
      category: category.value.trim(),
      priority: priority.value,
      description: description.value.trim(),
    }

    const response = await apiClient.post('/tickets', payload)
    // Berhasil: Arahkan langsung ke halaman detail tiket yang baru dibuat
    router.push(`/tickets/${response.data.id}`)
  } catch (err) {
    serverError.value = err.response?.data?.error || 'Gagal membuat tiket bantuan.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="max-w-3xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Breadcrumb / Tombol Kembali -->
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

    <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="p-6 border-b border-gray-200">
        <h1 class="text-xl font-bold text-gray-900 tracking-tight">Buat Permohonan Tiket Baru</h1>
        <p class="mt-1 text-sm text-gray-500">
          Sampaikan kendala teknis atau permohonan bantuan Anda kepada tim support.
        </p>
      </div>

      <div class="p-6">
        <AlertError :message="clientError || serverError" />

        <form class="space-y-6" @submit.prevent="handleSubmit">
          <!-- Judul Tiket -->
          <div>
            <label for="title" class="block text-sm font-semibold text-gray-700">
              Judul Tiket <span class="text-red-500">*</span>
            </label>
            <div class="mt-1">
              <input
                id="title"
                v-model="title"
                type="text"
                placeholder="Ringkasan singkat kendala (misal: Monitor tidak menyala)"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 text-sm"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <!-- Kategori -->
            <div>
              <label for="category" class="block text-sm font-semibold text-gray-700">
                Kategori <span class="text-red-500">*</span>
              </label>
              <div class="mt-1">
                <select
                  id="category"
                  v-model="category"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 text-sm bg-white"
                >
                  <option value="" disabled>Pilih Kategori</option>
                  <option value="Hardware">Perangkat Keras (Hardware)</option>
                  <option value="Software">Perangkat Lunak (Software)</option>
                  <option value="Jaringan">Jaringan & Internet</option>
                  <option value="Akses Akun">Akses Akun & Password</option>
                  <option value="Lainnya">Lainnya</option>
                </select>
              </div>
            </div>

            <!-- Prioritas -->
            <div>
              <label for="priority" class="block text-sm font-semibold text-gray-700">
                Tingkat Prioritas <span class="text-red-500">*</span>
              </label>
              <div class="mt-1">
                <select
                  id="priority"
                  v-model="priority"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 text-sm bg-white"
                >
                  <option value="low">Rendah (Low)</option>
                  <option value="medium">Sedang (Medium)</option>
                  <option value="high">Tinggi (High)</option>
                  <option value="urgent">Mendesak (Urgent)</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Deskripsi -->
          <div>
            <label for="description" class="block text-sm font-semibold text-gray-700">
              Deskripsi Lengkap <span class="text-red-500">*</span>
            </label>
            <div class="mt-1">
              <textarea
                id="description"
                v-model="description"
                rows="5"
                placeholder="Jelaskan kronologi kendala, lokasi meja, atau langkah yang sudah dicoba..."
                class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 text-sm"
              ></textarea>
            </div>
          </div>

          <div class="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
            <router-link
              to="/tickets"
              class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              Batal
            </router-link>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="inline-flex justify-center px-5 py-2 text-sm font-semibold text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 shadow-sm disabled:opacity-50 transition-colors"
            >
              <span v-if="isSubmitting">Menyimpan Tiket...</span>
              <span v-else>Kirim Tiket</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
