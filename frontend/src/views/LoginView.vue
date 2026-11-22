<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AlertError from '../components/AlertError.vue'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const clientError = ref('')
const serverError = ref('')

// Jika redirect karena sesi habis
if (route.query.expired) {
  serverError.value = 'Sesi Anda telah berakhir. Silakan login kembali.'
}

const handleSubmit = async () => {
  clientError.value = ''
  serverError.value = ''

  // 1. Validasi Sisi Klien
  if (!email.value.trim()) {
    clientError.value = 'Email wajib diisi.'
    return
  }
  if (!email.value.includes('@')) {
    clientError.value = 'Format email tidak valid.'
    return
  }
  if (!password.value) {
    clientError.value = 'Password wajib diisi.'
    return
  }

  // 2. Eksekusi Login via Pinia Store
  try {
    await authStore.login(email.value.trim(), password.value)
    const redirectPath = route.query.redirect || '/tickets'
    router.push(redirectPath)
  } catch (err) {
    serverError.value = err.message
  }
}

// Helper untuk mengisi cepat akun seed saat testing
const fillCredentials = (role) => {
  if (role === 'admin') {
    email.value = 'admin@example.com'
    password.value = 'Password123!'
  } else if (role === 'agent') {
    email.value = 'agent@example.com'
    password.value = 'Password123!'
  } else if (role === 'requester') {
    email.value = 'requester@example.com'
    password.value = 'Password123!'
  }
}
</script>

<template>
  <div class="min-h-[calc(100vh-4rem)] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <div class="text-center">
        <span class="inline-block p-3 bg-indigo-600 text-white rounded-xl font-bold text-2xl shadow-md">HD</span>
        <h2 class="mt-4 text-3xl font-extrabold text-gray-900 tracking-tight">Masuk ke Helpdesk</h2>
        <p class="mt-2 text-sm text-gray-600">
          Atau
          <router-link to="/register" class="font-medium text-indigo-600 hover:text-indigo-500">
            daftar akun karyawan baru
          </router-link>
        </p>
      </div>

      <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div class="bg-white py-8 px-4 shadow sm:rounded-xl sm:px-10 border border-gray-100">
          <!-- Tampilan Pesan Error -->
          <AlertError :message="clientError || serverError" />

          <form class="space-y-6" @submit.prevent="handleSubmit">
            <div>
              <label for="email" class="block text-sm font-medium text-gray-700">Alamat Email</label>
              <div class="mt-1">
                <input
                  id="email"
                  v-model="email"
                  type="email"
                  autocomplete="email"
                  placeholder="nama@perusahaan.com"
                  class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>

            <div>
              <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
              <div class="mt-1">
                <input
                  id="password"
                  v-model="password"
                  type="password"
                  autocomplete="current-password"
                  placeholder="••••••••"
                  class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>

            <div>
              <button
                type="submit"
                :disabled="authStore.isLoading"
                class="w-full flex justify-center py-2.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 transition-colors"
              >
                <span v-if="authStore.isLoading">Memproses Masuk...</span>
                <span v-else>Masuk</span>
              </button>
            </div>
          </form>

          <!-- Quick Test Credentials -->
          <div class="mt-8 pt-6 border-t border-gray-200">
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider text-center mb-3">
              Uji Coba Cepat (Akun Bawaan):
            </p>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="fillCredentials('requester')"
                class="py-1 px-2 text-xs font-medium text-emerald-700 bg-emerald-50 hover:bg-emerald-100 rounded border border-emerald-200"
              >
                Requester
              </button>
              <button
                type="button"
                @click="fillCredentials('agent')"
                class="py-1 px-2 text-xs font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded border border-blue-200"
              >
                Agent
              </button>
              <button
                type="button"
                @click="fillCredentials('admin')"
                class="py-1 px-2 text-xs font-medium text-purple-700 bg-purple-50 hover:bg-purple-100 rounded border border-purple-200"
              >
                Admin
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
