export function useKoperasiApi() {
  const baseURL = useRuntimeConfig().public.apiBase
  const request = async <T>(path: string, options: Record<string, unknown> = {}) => {
    const payload = await $fetch<{ success: boolean; message: string; data: T }>(path, { baseURL, ...options })
    return payload.data
  }
  return {
    dashboard: (year = new Date().getFullYear(), month = new Date().getMonth() + 1) => request<any>(`/dashboard?year=${year}&month=${month}`),
    items: () => request<any[]>('/items'),
    suppliers: () => request<any[]>('/suppliers'),
    customers: () => request<any[]>('/customers'),
    masters: (name: string) => request<any[]>(`/masters/${name}`),
    transactions: (type = '') => request<any[]>(`/transactions${type ? `?type=${type}` : ''}`),
    debts: () => request<any[]>('/debts'),
    createItem: (body: any) => request('/items', { method: 'POST', body }),
    updateItem: (id: number, body: any) => request(`/items/${id}`, { method: 'PUT', body }),
    deleteItem: (id: number) => request(`/items/${id}`, { method: 'DELETE' }),
    createSupplier: (body: any) => request('/suppliers', { method: 'POST', body }),
    updateSupplier: (id: number, body: any) => request(`/suppliers/${id}`, { method: 'PUT', body }),
    deleteSupplier: (id: number) => request(`/suppliers/${id}`, { method: 'DELETE' }),
    createCustomer: (body: any) => request('/customers', { method: 'POST', body }),
    updateCustomer: (id: number, body: any) => request(`/customers/${id}`, { method: 'PUT', body }),
    deleteCustomer: (id: number) => request(`/customers/${id}`, { method: 'DELETE' }),
    createMaster: (name: string, body: any) => request(`/masters/${name}`, { method: 'POST', body }),
    updateMaster: (table: string, id: number, body: any) => request(`/masters/${table}/${id}`, { method: 'PUT', body }),
    deleteMaster: (table: string, id: number) => request(`/masters/${table}/${id}`, { method: 'DELETE' }),
    createTransaction: (body: any) => request('/transactions', { method: 'POST', body }),
    updateTransaction: (id: number, body: any) => request(`/transactions/${id}`, { method: 'PUT', body }),
    voidTransaction: (id: number, reason: string) => request(`/transactions/${id}/void`, { method: 'POST', body: { reason } }),
    payDebt: (id: number, body: any) => request(`/debts/${id}/payments`, { method: 'POST', body }),
  }
}
