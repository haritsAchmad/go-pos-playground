import { describe, expect, it } from 'vitest'
import { isUIRouteAllowed } from './authorization'

describe('UI authorization policy', () => {
  it('allows shared read routes for every application role', () => {
    for (const role of ['admin', 'cashier', 'viewer']) {
      expect(isUIRouteAllowed('/', role)).toBe(true)
      expect(isUIRouteAllowed('/barang', role)).toBe(true)
    }
  })

  it('restricts operational routes to admin and cashier', () => {
    expect(isUIRouteAllowed('/kasir', 'admin')).toBe(true)
    expect(isUIRouteAllowed('/kasir', 'cashier')).toBe(true)
    expect(isUIRouteAllowed('/kasir', 'viewer')).toBe(false)
  })

  it('restricts administrative routes to admin', () => {
    expect(isUIRouteAllowed('/pengguna', 'admin')).toBe(true)
    expect(isUIRouteAllowed('/pengguna', 'cashier')).toBe(false)
    expect(isUIRouteAllowed('/audit', 'viewer')).toBe(false)
  })

  it('denies routes and roles absent from the policy', () => {
    expect(isUIRouteAllowed('/unknown', 'admin')).toBe(false)
    expect(isUIRouteAllowed('/', 'owner')).toBe(false)
  })
})
