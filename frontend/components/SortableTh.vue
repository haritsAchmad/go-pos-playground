<script setup lang="ts">
const props = withDefaults(defineProps<{
  field: string
  defaultDirection?: 'asc' | 'desc'
}>(), {
  defaultDirection: 'asc',
})

const sorting = defineModel<string>({ required: true })
const active = computed(() => sorting.value.startsWith(`${props.field}:`))
const direction = computed(() => active.value && sorting.value.endsWith(':desc') ? 'desc' : 'asc')
const ariaSort = computed(() => !active.value ? 'none' : direction.value === 'asc' ? 'ascending' : 'descending')

function toggle() {
  if (!active.value) {
    sorting.value = `${props.field}:${props.defaultDirection}`
    return
  }
  sorting.value = `${props.field}:${direction.value === 'asc' ? 'desc' : 'asc'}`
}
</script>

<template>
  <th :aria-sort="ariaSort">
    <button type="button" class="sortable-heading" :class="{ active }" @click="toggle">
      <span><slot /></span>
      <span class="sort-indicator" aria-hidden="true">{{ active ? (direction === 'asc' ? '▲' : '▼') : '↕' }}</span>
    </button>
  </th>
</template>
