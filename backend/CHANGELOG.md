# Changelog

## Unreleased

### Added

- PostgreSQL-backed transaction invoice sequence and concurrent checkout integration coverage.
- Idempotent transaction creation with replay detection, payload conflict protection, and PostgreSQL advisory locking.
- Versioned PostgreSQL migrations with an atomic `schema_migrations` ledger and startup lock.

## v0.2.0

### Added

- PostgreSQL Connection
- GET /items
- POST /items
- Environment Configuration
- Air Hot Reload

### Refactor

- Rename model -> entity
- Rename request -> dto
