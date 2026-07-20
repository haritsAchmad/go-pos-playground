# Go POS Playground

Aplikasi point of sale (POS) dan operasional koperasi hasil adaptasi modul SMS516, menggunakan REST API Go, PostgreSQL, dan frontend Nuxt.

> **Catatan rename database:** nama konfigurasi baru adalah database `pos_playground` dan schema `go_pos_playground`. Jika sebelumnya memakai `inventory_playground` atau `go_inventory_playground`, rename atau migrasikan database/schema lama sebelum menjalankan aplikasi dengan konfigurasi baru.

## Fitur koperasi

- Dashboard penjualan, pembelian, piutang, dan stok
- Master barang, kategori, merek, satuan, supplier, serta metode bayar
- Pelanggan member dan non-member; pelanggan `UMUM` tersedia otomatis
- Kasir/penjualan dengan pengurangan stok atomik
- Pembelian/penerimaan barang dengan penambahan stok atomik
- Histori transaksi, piutang, dan pembayaran piutang
- Autentikasi JWT dan otorisasi berbasis role (`admin`, `cashier`, `viewer`)

## Menjalankan aplikasi

```powershell
go run ./cmd/api
npm.cmd --prefix frontend run dev
```

Frontend tersedia di `http://localhost:3000` dan meneruskan request `/api` ke backend pada `http://localhost:8080`.

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
- [x] Authentication (JWT)
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
    "message": "Go POS Playground"
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

- Complete POS REST API

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
