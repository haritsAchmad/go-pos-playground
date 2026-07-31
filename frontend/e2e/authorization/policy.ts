export const personas = {
  superadmin: { appRole: 'admin', label: 'superadmin' },
  manager: { appRole: 'viewer', label: 'manager' },
  cashier: { appRole: 'cashier', label: 'cashier' },
} as const

export type Persona = keyof typeof personas
export type AppRole = (typeof personas)[Persona]['appRole']

export type ProtectedRoute = {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  path: string
  allowed: Persona[]
  risk: 'low' | 'medium' | 'high' | 'critical'
  objectLevel?: boolean
  body?: Record<string, unknown>
}

const all: Persona[] = ['superadmin', 'manager', 'cashier']
const operators: Persona[] = ['superadmin', 'cashier']
const rootOnly: Persona[] = ['superadmin']

// This is the executable authorization inventory derived from backend/internal/router/router.go.
// "manager" is a test persona mapped to the application's read-only "viewer" role.
export const protectedRoutes: ProtectedRoute[] = [
  { method: 'GET', path: '/auth/me', allowed: all, risk: 'medium' },
  { method: 'GET', path: '/users', allowed: rootOnly, risk: 'high' },
  { method: 'POST', path: '/users', allowed: rootOnly, risk: 'critical', body: {} },
  { method: 'PUT', path: '/users/{id}', allowed: rootOnly, risk: 'critical', objectLevel: true, body: {} },
  { method: 'DELETE', path: '/users/{id}', allowed: rootOnly, risk: 'critical', objectLevel: true },
  { method: 'GET', path: '/audit-logs', allowed: rootOnly, risk: 'high' },
  { method: 'GET', path: '/items', allowed: all, risk: 'low' },
  { method: 'POST', path: '/items', allowed: operators, risk: 'high', body: {} },
  { method: 'GET', path: '/items/{id}', allowed: all, risk: 'low', objectLevel: true },
  { method: 'PUT', path: '/items/{id}', allowed: operators, risk: 'high', objectLevel: true, body: {} },
  { method: 'DELETE', path: '/items/{id}', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'POST', path: '/items/{id}/restore', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'GET', path: '/items/{id}/stock-movements', allowed: all, risk: 'medium', objectLevel: true },
  { method: 'GET', path: '/deleted/items', allowed: operators, risk: 'medium' },
  { method: 'GET', path: '/suppliers', allowed: all, risk: 'low' },
  { method: 'POST', path: '/suppliers', allowed: operators, risk: 'high', body: {} },
  { method: 'GET', path: '/suppliers/{id}', allowed: all, risk: 'low', objectLevel: true },
  { method: 'PUT', path: '/suppliers/{id}', allowed: operators, risk: 'high', objectLevel: true, body: {} },
  { method: 'DELETE', path: '/suppliers/{id}', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'POST', path: '/suppliers/{id}/restore', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'GET', path: '/deleted/suppliers', allowed: operators, risk: 'medium' },
  { method: 'GET', path: '/customers', allowed: all, risk: 'low' },
  { method: 'POST', path: '/customers', allowed: operators, risk: 'high', body: {} },
  { method: 'GET', path: '/customers/{id}', allowed: all, risk: 'low', objectLevel: true },
  { method: 'PUT', path: '/customers/{id}', allowed: operators, risk: 'high', objectLevel: true, body: {} },
  { method: 'DELETE', path: '/customers/{id}', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'POST', path: '/customers/{id}/restore', allowed: operators, risk: 'high', objectLevel: true },
  { method: 'GET', path: '/deleted/customers', allowed: operators, risk: 'medium' },
  { method: 'GET', path: '/dashboard', allowed: all, risk: 'low' },
  { method: 'GET', path: '/masters/categories', allowed: all, risk: 'low' },
  { method: 'POST', path: '/masters/brands', allowed: operators, risk: 'high', body: {} },
  { method: 'POST', path: '/masters/categories', allowed: rootOnly, risk: 'high', body: {} },
  { method: 'PUT', path: '/masters/categories/{id}', allowed: rootOnly, risk: 'high', objectLevel: true, body: {} },
  { method: 'DELETE', path: '/masters/categories/{id}', allowed: rootOnly, risk: 'high', objectLevel: true },
  { method: 'GET', path: '/transactions', allowed: all, risk: 'medium' },
  { method: 'POST', path: '/transactions', allowed: operators, risk: 'critical', body: {} },
  { method: 'GET', path: '/transactions/{id}', allowed: all, risk: 'medium', objectLevel: true },
  { method: 'PUT', path: '/transactions/{id}', allowed: operators, risk: 'critical', objectLevel: true, body: {} },
  { method: 'POST', path: '/transactions/{id}/void', allowed: operators, risk: 'critical', objectLevel: true, body: {} },
  { method: 'GET', path: '/debts', allowed: all, risk: 'high' },
  { method: 'GET', path: '/debts/{id}/payments', allowed: all, risk: 'high', objectLevel: true },
  { method: 'POST', path: '/debts/{id}/payments', allowed: operators, risk: 'critical', objectLevel: true, body: {} },
  { method: 'POST', path: '/debts/{id}/payments/{paymentId}/reverse', allowed: operators, risk: 'critical', objectLevel: true, body: {} },
  { method: 'POST', path: '/bulk/items/delete', allowed: operators, risk: 'critical', body: { ids: [] } },
  { method: 'POST', path: '/bulk/items/restore', allowed: operators, risk: 'high', body: { ids: [] } },
  { method: 'POST', path: '/bulk/items/reset-stock', allowed: operators, risk: 'critical', body: { ids: [] } },
  { method: 'POST', path: '/bulk/customers/delete', allowed: operators, risk: 'high', body: { ids: [] } },
  { method: 'POST', path: '/bulk/customers/restore', allowed: operators, risk: 'high', body: { ids: [] } },
  { method: 'POST', path: '/bulk/suppliers/delete', allowed: operators, risk: 'high', body: { ids: [] } },
  { method: 'POST', path: '/bulk/suppliers/restore', allowed: operators, risk: 'high', body: { ids: [] } },
  { method: 'POST', path: '/bulk/debts/settle', allowed: operators, risk: 'critical', body: { ids: [] } },
]

export const uiRoutes = [
  { path: '/', allowed: all },
  { path: '/barang', allowed: all },
  { path: '/pelanggan', allowed: operators },
  { path: '/supplier', allowed: operators },
  { path: '/kasir', allowed: operators },
  { path: '/pembelian', allowed: operators },
  { path: '/histori', allowed: operators },
  { path: '/piutang', allowed: operators },
  { path: '/pengaturan', allowed: operators },
  { path: '/pengguna', allowed: rootOnly },
  { path: '/audit', allowed: rootOnly },
] satisfies Array<{ path: string; allowed: Persona[] }>

export const replaceIds = (path: string, id: number, paymentId = id) =>
  path.replace('{id}', String(id)).replace('{paymentId}', String(paymentId))
