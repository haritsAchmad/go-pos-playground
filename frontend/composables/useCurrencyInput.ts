export function useCurrencyInput() {
  const formatNumber = (value: unknown) => new Intl.NumberFormat('id-ID').format(Number(value) || 0)

  const setCurrency = (event: Event, target: Record<string, any>, key: string) => {
    target[key] = Number((event.target as HTMLInputElement).value.replace(/\D/g, '')) || 0
  }

  return { formatNumber, setCurrency }
}
