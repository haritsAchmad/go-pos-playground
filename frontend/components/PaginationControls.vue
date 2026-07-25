<script setup lang="ts">
import { computed } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  page: number
  totalItems: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
}>()

defineEmits<{
  'update:page': [value: number]
}>()

const visiblePages = computed(() => {
  if (props.totalPages <= 5) return Array.from({ length: props.totalPages }, (_, index) => index + 1)
  const pages = new Set([1, props.totalPages, props.page - 1, props.page, props.page + 1])
  return [...pages].filter(page => page >= 1 && page <= props.totalPages).sort((a, b) => a - b)
})
</script>

<template>
  <nav v-if="totalItems" class="pagination" aria-label="Navigasi halaman">
    <span class="pagination-summary">Menampilkan {{ rangeStart }}–{{ rangeEnd }} dari {{ totalItems }}</span>
    <div class="pagination-actions">
      <button class="soft" :disabled="page <= 1" aria-label="Halaman sebelumnya" @click="$emit('update:page', page - 1)">←</button>
      <template v-for="(value, index) in visiblePages" :key="value">
        <span v-if="index && value-(visiblePages[index-1] || 0)>1" class="pagination-ellipsis">…</span>
        <button class="page-number" :class="{active:value===page}" :aria-current="value===page?'page':undefined" @click="$emit('update:page', value)">{{ value }}</button>
      </template>
      <button class="soft" :disabled="page >= totalPages" aria-label="Halaman berikutnya" @click="$emit('update:page', page + 1)">→</button>
    </div>
  </nav>
</template>
