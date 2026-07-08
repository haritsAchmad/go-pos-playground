# Go Inventory Playground

Backend playground untuk belajar dan bereksperimen menggunakan Go dan PostgreSQL.

> Project ini dibuat sebagai laboratorium backend untuk mempelajari Go dari dasar menggunakan Standard Library tanpa framework terlebih dahulu, kemudian berkembang secara bertahap mengikuti kebutuhan.

---

# Tech Stack

- Go
- PostgreSQL 18
- pgx v5
- Air (Hot Reload)

---

# Current Features

## Completed

- [x] HTTP Server
- [x] Health Endpoint
- [x] PostgreSQL Connection
- [x] Environment Configuration
- [x] GET /items
- [x] POST /items

## Next Milestones

- [ ] Validation
- [ ] PUT /items
- [ ] DELETE /items
- [ ] Search
- [ ] Pagination
- [ ] Sorting
- [ ] Transaction
- [ ] Middleware
- [ ] Authentication (JWT)
- [ ] Unit Test
- [ ] Docker
- [ ] Redis

---

# Project Structure

```text
cmd/
└── api/

internal/
├── config/
├── database/
├── dto/
├── entity/
├── handler/
├── repository/
└── router/

migrations/
```

---

# API Endpoints

## Health Check

GET /health

Response

```json
{
    "status": "ok",
    "message": "Go Inventory Playground"
}
```

---

## Get Items

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

## Create Item

POST /items

Request

```json
{
    "name": "Laptop",
    "description": "ThinkPad T14",
    "stock": 8
}
```

Response

```json
{
    "message": "item created successfully"
}
```

---

# Roadmap

## v0.1

- HTTP Server
- Health Endpoint

## v0.2

- PostgreSQL Connection
- GET /items
- POST /items

## v0.3

- Request Validation

## v0.4

- Update Item

## v0.5

- Delete Item

## v0.6

- Search & Pagination

## v0.7

- JWT Authentication

## v1.0

- Complete Inventory REST API

---

# Learning Goals

Project ini dibuat untuk mempelajari:

- Go Standard Library
- REST API
- PostgreSQL
- Clean Project Structure
- Git Workflow
- Backend Best Practices
- Layered Architecture