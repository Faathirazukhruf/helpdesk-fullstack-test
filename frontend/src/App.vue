<script setup>
import Navbar from './components/Navbar.vue'
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex flex-col font-sans">
    <Navbar />
    <main class="flex-1">
      <router-view />
    </main>
  </div>
</template>
