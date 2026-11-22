<script setup>
defineProps({
  status: {
    type: String,
    required: true,
  },
})

const badgeClass = (status) => {
  switch (status) {
    case 'Open':
      return 'bg-amber-100 text-amber-800 border-amber-200'
    case 'In Progress':
      return 'bg-blue-100 text-blue-800 border-blue-200'
    case 'Resolved':
      return 'bg-emerald-100 text-emerald-800 border-emerald-200'
    case 'Closed':
      return 'bg-gray-100 text-gray-700 border-gray-300'
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200'
  }
}
</script>

<template>
  <span
    :class="badgeClass(status)"
    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border shadow-sm"
  >
    <span class="w-1.5 h-1.5 mr-1.5 rounded-full bg-current opacity-70"></span>
    {{ status }}
  </span>
</template>
