# Database migrations

`database.Migrate` runs automatically when the API or seeder starts.

Applied versions are recorded in `<DB_SCHEMA>.schema_migrations`. Migration
version 1 is the idempotent baseline that safely adopts databases created
before version tracking existed.

To add a schema change:

1. Append a migration with the next version to the `migrations` slice in
   `postgres.go`.
2. Keep the version immutable after it has been deployed.
3. Make the migration atomic and test it against an isolated PostgreSQL schema.
4. Never edit a migration that may already exist in `schema_migrations`; add a
   new version instead.

Inspect the current database version with:

```sql
SELECT version, name, applied_at
FROM <DB_SCHEMA>.schema_migrations
ORDER BY version;
```
