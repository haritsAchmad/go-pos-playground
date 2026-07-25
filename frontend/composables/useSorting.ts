import { computed, toValue, type MaybeRefOrGetter } from 'vue'

export type SortOption = {
  value: string
  label: string
}

function compareValues(left: unknown, right: unknown) {
  if (typeof left === 'number' && typeof right === 'number') return left - right
  if (typeof left === 'boolean' && typeof right === 'boolean') return Number(left) - Number(right)
  return String(left ?? '').localeCompare(String(right ?? ''), 'id-ID', {
    numeric: true,
    sensitivity: 'base',
  })
}

export function sortRows<T extends Record<string, any>>(rows: T[], selection: string): T[] {
  const separator = selection.lastIndexOf(':')
  const field = separator === -1 ? selection : selection.slice(0, separator)
  const direction = separator === -1 ? 'asc' : selection.slice(separator + 1)
  const multiplier = direction === 'desc' ? -1 : 1

  return rows
    .map((row, index) => ({ row, index }))
    .sort((left, right) => {
      const compared = compareValues(left.row[field], right.row[field])
      return compared === 0 ? left.index - right.index : compared * multiplier
    })
    .map(({ row }) => row)
}

export function useSorting<T extends Record<string, any>>(
  source: MaybeRefOrGetter<T[]>,
  selection: MaybeRefOrGetter<string>,
) {
  return computed(() => sortRows(toValue(source), toValue(selection)))
}
