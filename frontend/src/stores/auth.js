import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '../api/client'

// Pinia Store untuk state management sesi autentikasi pengguna
export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const isLoading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isAgent = computed(() => user.value?.role === 'agent')
  const isRequester = computed(() => user.value?.role === 'requester')

  // Login action
  async function login(email, password) {
    isLoading.value = true
    error.value = null
    try {
      const response = await apiClient.post('/auth/login', { email, password })
      token.value = response.data.token
      user.value = response.data.user

      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      return response.data
    } catch (err) {
      const message = err.response?.data?.error || 'Gagal login, periksa kembali email dan password.'
      error.value = message
      throw new Error(message)
    } finally {
      isLoading.value = false
    }
  }

  // Register action (pendaftaran requester)
  async function register(name, email, password) {
    isLoading.value = true
    error.value = null
    try {
      const response = await apiClient.post('/auth/register', { name, email, password })
      token.value = response.data.token
      user.value = response.data.user

      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      return response.data
    } catch (err) {
      const message = err.response?.data?.error || 'Gagal mendaftar akun baru.'
      error.value = message
      throw new Error(message)
    } finally {
      isLoading.value = false
    }
  }

  // Fetch current user profile (/api/me)
  async function fetchMe() {
    if (!token.value) return null
    try {
      const response = await apiClient.get('/me')
      user.value = response.data
      localStorage.setItem('user', JSON.stringify(response.data))
      return response.data
    } catch (err) {
      logout()
      return null
    }
  }

  // Logout action
  function logout() {
    token.value = ''
    user.value = null
    error.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    user,
    isLoading,
    error,
    isAuthenticated,
    isAdmin,
    isAgent,
    isRequester,
    login,
    register,
    fetchMe,
    logout,
  }
})
