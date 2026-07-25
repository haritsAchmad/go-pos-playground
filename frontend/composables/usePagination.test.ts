import { ref } from 'vue'
import { describe, expect, it } from 'vitest'
import { usePagination } from './usePagination'

describe('usePagination', () => {
  it('slices rows and exposes the visible range', () => {
    const rows = ref([1, 2, 3, 4, 5])
    const pagination = usePagination(rows, 2)
    pagination.setPage(2)
    expect(pagination.pageItems.value).toEqual([3, 4])
    expect(pagination.rangeStart.value).toBe(3)
    expect(pagination.rangeEnd.value).toBe(4)
    expect(pagination.totalPages.value).toBe(3)
  })

  it('clamps invalid page changes', () => {
    const pagination = usePagination(() => [1, 2, 3], 2)
    pagination.setPage(99)
    expect(pagination.page.value).toBe(2)
    pagination.setPage(0)
    expect(pagination.page.value).toBe(1)
  })

  it('returns to the final valid page when filtered rows shrink', async () => {
    const rows = ref([1, 2, 3, 4, 5])
    const pagination = usePagination(rows, 2)
    pagination.setPage(3)
    rows.value = [1]
    await Promise.resolve()
    expect(pagination.page.value).toBe(1)
    expect(pagination.pageItems.value).toEqual([1])
  })

  it('returns to page one when filter results change at the same size', async () => {
    const rows = ref([1, 2, 3, 4])
    const pagination = usePagination(rows, 2)
    pagination.setPage(2)
    rows.value = [5, 6, 7, 8]
    await Promise.resolve()
    expect(pagination.page.value).toBe(1)
  })
})
