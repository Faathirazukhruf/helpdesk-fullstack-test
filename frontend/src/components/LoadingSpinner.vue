<script setup>
defineProps({
  text: {
    type: String,
    default: 'Memuat data...',
  },
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12">
    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mb-3"></div>
    <span class="text-sm font-medium text-gray-500">{{ text }}</span>
  </div>
</template>
