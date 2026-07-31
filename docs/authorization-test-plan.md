# Authorization test plan

The executable policy inventory lives in `frontend/e2e/authorization/policy.ts`.
It is derived from `backend/internal/router/router.go`; tests do not modify the
application's authorization rules.

## Test identities

| Requested identity | Application role | Test email |
|---|---|---|
| superadmin | `admin` | `e2e-authz-<run>-superadmin@example.test` |
| manager | `viewer` | `e2e-authz-<run>-manager@example.test` |
| cashier | `cashier` | `e2e-authz-<run>-cashier@example.test` |
| unauthenticated | none | none |

There is no `superadmin` or `manager` role in the application. The mapping above
preserves the existing policy: superadmin exercises the administrative role,
while manager exercises the read-only role. This mismatch is deliberately
visible rather than changing production authorization to make tests pass.

## Safety

- Run only against a disposable local or isolated test database.
- Non-local targets are refused unless `E2E_ALLOW_REMOTE_TEST_TARGET=true`.
- Hostnames containing `prod` or `production` are always refused.
- The setup project creates or resets only deterministic `e2e-authz-*` accounts.
- Mutation probes use invalid payloads and sentinel IDs `2147483646` and
  `2147483647`; they do not create, edit, delete, settle, or void business data.

## Run

Start the backend and frontend against a test database. Then:

```powershell
Set-Location frontend
$env:E2E_BASE_URL='http://127.0.0.1:3000'
$env:E2E_TEST_RUN_ID='local'
$env:E2E_BOOTSTRAP_EMAIL='admin account in the test database'
$env:E2E_BOOTSTRAP_PASSWORD='test database password'
$env:E2E_ACCOUNT_PASSWORD='dedicated test-account password'
$env:E2E_JWT_SECRET='JWT secret used only by the isolated test backend'
npm run test:authz
```

Do not put these credentials in Git. `E2E_JWT_SECRET` lets the suite construct a
genuinely expired but correctly signed access token; it must be the secret of
the isolated test backend, never production. The bootstrap account must be an `admin`
in the isolated database. Reusing a run ID makes setup reset the same three
test accounts instead of accumulating accounts.

## Artifacts

Playwright writes the following ignored runtime artifacts:

- `frontend/test-results/authz/authorization-audit.md`: security findings with
  endpoint, account, expected result, actual result, risk, likely root cause,
  and artifact paths.
- `frontend/test-results/authz/html/`: interactive report.
- `frontend/test-results/authz/artifacts/`: screenshots, video, and traces for
  failures.
- `frontend/test-results/authz/results.json`: machine-readable result.

The API matrix checks authentication before handler validation. An allowed role
may receive a successful response or a validation/not-found response from a
safe probe, but it must never receive `401` or `403`. A disallowed role must
receive `403`, and an unauthenticated request must receive `401`.

Resource-ID replacement confirms that the RBAC decision cannot be bypassed by
changing an ID. True owner/tenant isolation is not currently testable because
resources do not carry an owner or tenant identifier.
