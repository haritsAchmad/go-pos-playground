import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

export function usePagination<T>(source: MaybeRefOrGetter<T[]>, initialPageSize = 10) {
  const page = ref(1)
  const pageSize = ref(initialPageSize)
  const totalItems = computed(() => toValue(source).length)
  const totalPages = computed(() => Math.max(1, Math.ceil(totalItems.value / pageSize.value)))
  const pageItems = computed(() => {
    const start = (page.value - 1) * pageSize.value
    return toValue(source).slice(start, start + pageSize.value)
  })
  const rangeStart = computed(() => totalItems.value ? (page.value - 1) * pageSize.value + 1 : 0)
  const rangeEnd = computed(() => Math.min(page.value * pageSize.value, totalItems.value))

  function setPage(value: number) {
    page.value = Math.min(Math.max(1, value), totalPages.value)
  }

  watch(() => toValue(source), () => {
    page.value = 1
  })
  watch([totalItems, pageSize], () => {
    if (page.value > totalPages.value) page.value = totalPages.value
  })

  return { page, pageSize, totalItems, totalPages, pageItems, rangeStart, rangeEnd, setPage }
}
