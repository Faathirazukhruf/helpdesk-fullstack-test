<script setup>
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
</script>

<template>
  <div class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header Dashboard Sesi -->
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-8">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 tracking-tight">
            Selamat datang, {{ authStore.user?.name }}! 👋
          </h1>
          <p class="mt-1 text-sm text-gray-500">
            Anda berhasil masuk sebagai <span class="font-semibold text-indigo-600 capitalize">{{ authStore.user?.role }}</span> ({{ authStore.user?.email }}).
          </p>
        </div>

        <div class="flex items-center space-x-3">
          <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-800">
            ● Sesi Aktif
          </span>
        </div>
      </div>
    </div>

    <!-- Placeholder Status Fase 7 -->
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-indigo-100 text-indigo-600 mb-4">
        <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <h3 class="text-lg font-semibold text-gray-900">Autentikasi & Navigation Guard Berhasil!</h3>
      <p class="mt-2 text-sm text-gray-500 max-w-md mx-auto">
        State session pengguna telah tersimpan aman di Pinia dan LocalStorage. Pada <strong>Fase 8</strong>, halaman ini akan memuat tabel tiket lengkap dengan filter status, prioritas, pencarian, dan pagination.
      </p>
    </div>
  </div>
</template>
