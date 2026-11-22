import axios from 'axios'

// HTTP Client terpusat berbasis Axios
const apiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request Interceptor: Menyuntikkan Bearer token JWT secara otomatis ke setiap request
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response Interceptor: Menangani sesi expired / respons 401 Unauthorized secara terpusat
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // Hapus data sesi yang usang
      localStorage.removeItem('token')
      localStorage.removeItem('user')

      // Arahkan ke halaman login jika saat ini tidak di halaman login atau register
      if (
        !window.location.pathname.includes('/login') &&
        !window.location.pathname.includes('/register')
      ) {
        window.location.href = '/login?expired=1'
      }
    }
    return Promise.reject(error)
  }
)

export default apiClient
