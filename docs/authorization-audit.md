# Authorization audit

Date: 2026-08-01

Scope: static review plus a complete runtime Playwright authorization run
against a disposable PostgreSQL 17 instance. The temporary server listened on
`127.0.0.1:55432`, used database `go_pos_authz_e2e_20260801_fix` and schema
`authz_e2e`, and was stopped and deleted after the run. The application,
development, staging, and production databases/schemas were not used. No
non-test business data was changed.

## Runtime summary

| Browser/project | Registered and run | Passed | Failed | Skipped |
|---|---:|---:|---:|---:|
| Setup | 1 | 1 | 0 | 0 |
| Playwright Chromium | 122 | 122 | 0 | 0 |
| Google Chrome (installed) | 122 | 122 | 0 | 0 |
| Microsoft Edge (installed) | 122 | 122 | 0 | 0 |
| Playwright Firefox | 122 | 122 | 0 | 0 |
| **Total** | **489** | **489** | **0** | **0** |

All requested browsers were available and executed; none were skipped. The
run took approximately 8.1 minutes. All direct API authorization matrices,
anonymous checks, role allow/deny checks, sentinel resource-ID substitutions,
expired-token checks, logout replay checks, post-logout UI checks, and direct
UI route navigation checks passed.

Runtime artifacts are retained in `frontend/test-results/authz/`: the generated
`authorization-audit.md`, `results.json`, and the HTML report. No failure
screenshot, video, or trace was produced by the passing final run.

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

## Resolved findings

### AUTHZ-001 — access-token replay after logout

| Field | Value |
|---|---|
| Endpoint | `POST /auth/logout`, followed by any protected endpoint such as `GET /auth/me` |
| Account | Any authenticated account; automated case uses cashier |
| Expected | Reusing the logged-out session/token is rejected with `401` |
| Actual | Logout returns `200`; both refresh replay and `GET /auth/me` with the old access token now return `401`. Verified in all four browsers/projects. |
| Risk | High when a token has been copied or stolen; exposure is bounded by `JWT_EXPIRY_MINUTES` |
| Resolution | Access JWTs now carry `auth_sessions.id` as `sid`; authentication requires that session to match the user and remain unexpired and unrevoked. Logout and refresh revoke the prior session. |

Status: **resolved**. A token without `sid`, with a foreign `sid`, or with a
revoked/expired session is rejected before the protected handler.

### AUTHZ-002 — hidden UI routes have no role-level navigation guard

| Field | Value |
|---|---|
| Endpoint | UI routes `/pelanggan`, `/supplier`, `/kasir`, `/pembelian`, `/histori`, `/piutang`, `/pengaturan`, `/pengguna`, and `/audit` |
| Account | manager persona (`viewer` application role); cashier on `/pengguna` and `/audit` |
| Expected | Direct navigation to a route hidden for that role redirects away or renders a forbidden page |
| Actual | Direct navigation validates the current session and role, then redirects disallowed roles to `/`. Verified in all four browsers: manager on nine routes, plus cashier on `/pengguna` and `/audit`. |
| Risk | Medium for UI policy inconsistency; backend RBAC limits the direct security impact |
| Resolution | `frontend/utils/authorization.ts` is shared by `auth.global.ts` and `KoperasiConsole.vue`; menu visibility and direct navigation enforce the same roles. |

Status: **resolved**. Backend API policy remains unchanged; this enforces the
existing UI policy rather than granting or removing API permissions.

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

Generated findings are in `frontend/test-results/authz/authorization-audit.md`.
Failure screenshots, videos, and traces are retained below
`frontend/test-results/authz/artifacts/`; HTML and JSON reports are generated
alongside them. Each generated failure row includes browser, endpoint or UI
route, persona, expected/actual result, risk, root-cause hint, and artifact
paths.
