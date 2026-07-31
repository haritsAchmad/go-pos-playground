# Authorization audit

Date: 2026-07-31

Scope: static review of the Go router/middleware and Nuxt route behavior, plus
the automated Playwright suite in `frontend/e2e/authorization/`. No runtime
account was created and no business data was touched because the available
local schema was not established as disposable test data.

## Policy inventory

The complete method/path/role inventory is executable in
`frontend/e2e/authorization/policy.ts`. In summary:

- `admin` can use all protected operations.
- `cashier` can read operational data and mutate items, suppliers, customers,
  transactions, debts, brands, and supported bulk operations.
- `viewer` can read dashboard, items, suppliers, customers, masters,
  transactions, debts, stock movements, and payment history.
- Only `admin` can manage users, read audit logs, and mutate general master data.

The requested personas are represented without changing the application
policy: `superadmin -> admin`, `manager -> viewer`, and `cashier -> cashier`.

## Static findings

### AUTHZ-001 — access-token replay after logout

| Field | Value |
|---|---|
| Endpoint | `POST /auth/logout`, followed by any protected endpoint such as `GET /auth/me` |
| Account | Any authenticated account; automated case uses cashier |
| Expected | Reusing the logged-out session/token is rejected with `401` |
| Actual | Logout revokes the refresh token only. The existing stateless access token remains accepted until its configured expiry. |
| Risk | High when a token has been copied or stolen; exposure is bounded by `JWT_EXPIRY_MINUTES` |
| Likely root cause | Access JWTs have no server-side session identifier/revocation check; authentication checks signature, expiry, and current user state only |

This is also described in the current README, so it is an architectural
limitation rather than a newly introduced regression. The Playwright session
test deliberately fails and records trace/screenshot/report artifacts while
this behavior remains.

### AUTHZ-002 — hidden UI routes have no role-level navigation guard

| Field | Value |
|---|---|
| Endpoint | UI routes `/pelanggan`, `/supplier`, `/kasir`, `/pembelian`, `/histori`, `/piutang`, `/pengaturan`, `/pengguna`, and `/audit` |
| Account | manager persona (`viewer` application role); cashier on `/pengguna` and `/audit` |
| Expected | Direct navigation to a route hidden for that role redirects away or renders a forbidden page |
| Actual | `auth.global.ts` checks token presence only. Direct URL navigation remains on the hidden route. Backend API authorization still protects privileged responses. |
| Risk | Medium for UI policy inconsistency; backend RBAC limits the direct security impact |
| Likely root cause | Navigation visibility is implemented in `KoperasiConsole.vue`, but the same role policy is not enforced by route middleware |

For several read endpoints the backend intentionally allows `viewer`, so direct
navigation can reveal data that the navigation hides. For admin-only users and
audit APIs, the page remains addressable but its API request receives `403`.

### AUTHZ-003 — no owner/tenant boundary exists for object-level checks

| Field | Value |
|---|---|
| Endpoint | All `{id}` resource endpoints |
| Account | All authenticated roles allowed by each endpoint's RBAC policy |
| Expected | A user cannot substitute an ID outside their authorized ownership/tenant scope |
| Actual | Resources have no owner or tenant attribute; authorization is role-only and grants the allowed role access to every resource ID |
| Risk | Informational for a deliberately shared POS dataset; High if per-branch, per-store, or per-user isolation is expected later |
| Likely root cause | Current domain and middleware model implement RBAC but not relationship/tenant-based authorization |

The automated ID-substitution cases verify that changing IDs cannot bypass
RBAC. They cannot assert ownership isolation until the product defines such a
boundary.

## Runtime artifacts

After running against an isolated test database, generated findings are written
to `frontend/test-results/authz/authorization-audit.md`. Failure screenshots,
videos, and traces are retained below `frontend/test-results/authz/artifacts/`;
HTML and JSON reports are generated alongside them.
