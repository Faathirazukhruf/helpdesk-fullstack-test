<script setup>
defineProps({
  priority: {
    type: String,
    required: true,
  },
})

const badgeClass = (priority) => {
  switch (priority) {
    case 'urgent':
      return 'bg-red-100 text-red-800 border-red-200'
    case 'high':
      return 'bg-orange-100 text-orange-800 border-orange-200'
    case 'medium':
      return 'bg-blue-100 text-blue-800 border-blue-200'
    case 'low':
      return 'bg-gray-100 text-gray-700 border-gray-200'
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200'
  }
}
</script>

<template>
  <span
    :class="badgeClass(priority)"
    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize border"
  >
    {{ priority }}
  </span>
</template>
