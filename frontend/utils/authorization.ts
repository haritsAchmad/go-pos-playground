export type ApplicationRole = 'admin' | 'cashier' | 'viewer'

const allRoles: ApplicationRole[] = ['admin', 'cashier', 'viewer']
const operators: ApplicationRole[] = ['admin', 'cashier']
const administrators: ApplicationRole[] = ['admin']

export const uiRouteRoles: Readonly<Record<string, readonly ApplicationRole[]>> = {
  '/': allRoles,
  '/barang': allRoles,
  '/pelanggan': operators,
  '/supplier': operators,
  '/kasir': operators,
  '/pembelian': operators,
  '/histori': operators,
  '/piutang': operators,
  '/pengaturan': operators,
  '/pengguna': administrators,
  '/audit': administrators,
}

export function isUIRouteAllowed(path: string, role: string): boolean {
  const allowed = uiRouteRoles[path]
  return Boolean(allowed?.includes(role as ApplicationRole))
}
