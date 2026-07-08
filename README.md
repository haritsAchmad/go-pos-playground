# Go Inventory Playground

Backend playground untuk belajar dan bereksperimen menggunakan Go dan PostgreSQL.

> Tujuan project ini bukan hanya membuat CRUD, tetapi memahami bagaimana membangun backend yang rapi, terstruktur, dan mudah dikembangkan.

---

## Tech Stack

- Go
- PostgreSQL 18
- pgx v5
- Air (Hot Reload)

---

## Features

### Completed

- [x] HTTP Server
- [x] Health Endpoint
- [x] PostgreSQL Connection
- [x] Environment Configuration (.env)
- [x] GET /items

### In Progress

- [ ] POST /items
- [ ] PUT /items
- [ ] DELETE /items

### Planned

- [ ] Validation
- [ ] Pagination
- [ ] Search
- [ ] Sorting
- [ ] Transaction
- [ ] Authentication
- [ ] Unit Test
- [ ] Docker
- [ ] Redis Cache

---

## Project Structure

```text
cmd/
internal/
    config/
    database/
    handler/
    model/
    repository/
    router/
migrations/
```

---

## API

### Health Check

GET /health

Response

```json
{
    "status": "ok",
    "message": "Go Inventory Playground"
}
```

---

### Get Items

GET /items

Response

```json
[
    {
        "id": 1,
        "name": "Mouse Wireless",
        "description": "Mouse kantor untuk testing API",
        "stock": 10
    }
]
```

---

## Learning Notes

Project ini dibuat sebagai playground untuk mempelajari:

- Go Standard Library
- REST API
- PostgreSQL
- Clean Project Structure
- Git Workflow
- Backend Best Practices