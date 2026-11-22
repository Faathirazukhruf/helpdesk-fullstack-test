<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AlertError from '../components/AlertError.vue'

const authStore = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const clientError = ref('')
const serverError = ref('')

const handleRegister = async () => {
  clientError.value = ''
  serverError.value = ''

  // 1. Validasi Sisi Klien
  if (!name.value.trim()) {
    clientError.value = 'Nama lengkap wajib diisi.'
    return
  }
  if (!email.value.trim() || !email.value.includes('@')) {
    clientError.value = 'Alamat email tidak valid.'
    return
  }
  if (password.value.length < 6) {
    clientError.value = 'Password minimal harus 6 karakter.'
    return
  }
  if (password.value !== confirmPassword.value) {
    clientError.value = 'Konfirmasi password tidak cocok.'
    return
  }

  // 2. Eksekusi Registrasi via Pinia Store
  try {
    await authStore.register(name.value.trim(), email.value.trim(), password.value)
    router.push('/tickets')
  } catch (err) {
    serverError.value = err.message
  }
}
</script>

<template>
  <div class="min-h-[calc(100vh-4rem)] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <div class="text-center">
        <span class="inline-block p-3 bg-indigo-600 text-white rounded-xl font-bold text-2xl shadow-md">HD</span>
        <h2 class="mt-4 text-3xl font-extrabold text-gray-900 tracking-tight">Daftar Akun Karyawan</h2>
        <p class="mt-2 text-sm text-gray-600">
          Sudah punya akun?
          <router-link to="/login" class="font-medium text-indigo-600 hover:text-indigo-500">
            Masuk di sini
          </router-link>
        </p>
      </div>

      <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div class="bg-white py-8 px-4 shadow sm:rounded-xl sm:px-10 border border-gray-100">
          <AlertError :message="clientError || serverError" />

          <form class="space-y-5" @submit.prevent="handleRegister">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700">Nama Lengkap</label>
              <div class="mt-1">
                <input
                  id="name"
                  v-model="name"
                  type="text"
                  placeholder="Misal: Ahmad Fauzi"
                  class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>

            <div>
              <label for="email" class="block text-sm font-medium text-gray-700">Alamat Email</label>
              <div class="mt-1">
                <input
                  id="email"
                  v-model="email"
                  type="email"
                  placeholder="ahmad@perusahaan.com"
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
                  placeholder="Minimal 6 karakter"
                  class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>

            <div>
              <label for="confirmPassword" class="block text-sm font-medium text-gray-700">Konfirmasi Password</label>
              <div class="mt-1">
                <input
                  id="confirmPassword"
                  v-model="confirmPassword"
                  type="password"
                  placeholder="Ketik ulang password"
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
                <span v-if="authStore.isLoading">Mendaftarkan Akun...</span>
                <span v-else>Daftar Sekarang</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
