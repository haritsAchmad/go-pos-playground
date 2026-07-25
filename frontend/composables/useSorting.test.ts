import { describe, expect, it } from 'vitest'
import { sortRows } from './useSorting'

describe('sortRows', () => {
  const rows = [
    { id: 1, name: 'Zebra 10', stock: 2, created_at: '2026-01-01T00:00:00Z' },
    { id: 2, name: 'apel', stock: 20, created_at: '2026-03-01T00:00:00Z' },
    { id: 3, name: 'Zebra 2', stock: 5, created_at: '2026-02-01T00:00:00Z' },
  ]

  it('sorts text naturally and case-insensitively', () => {
    expect(sortRows(rows, 'name:asc').map(row => row.id)).toEqual([2, 3, 1])
  })

  it('sorts numbers descending', () => {
    expect(sortRows(rows, 'stock:desc').map(row => row.id)).toEqual([2, 3, 1])
  })

  it('sorts ISO dates from newest to oldest', () => {
    expect(sortRows(rows, 'created_at:desc').map(row => row.id)).toEqual([2, 3, 1])
  })

  it('does not mutate the source array', () => {
    const result = sortRows(rows, 'name:asc')
    expect(result).not.toBe(rows)
    expect(rows.map(row => row.id)).toEqual([1, 2, 3])
  })
})
