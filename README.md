# Go POS Playground

Aplikasi point of sale dan operasional koperasi dengan REST API Go, PostgreSQL, serta frontend Nuxt. Project ini menjadi playground untuk mempelajari Go Standard Library, layered architecture, transaksi database, autentikasi, dan otorisasi.

## Fitur

- Dashboard penjualan, pembelian, piutang, dan stok
- Master barang, kategori, merek, satuan, supplier, dan metode pembayaran
- Pelanggan member dan non-member; pelanggan `UMUM` dibuat otomatis
- Kasir/penjualan dengan pengurangan stok atomik
- Pembelian/penerimaan barang dengan penambahan stok atomik
- Histori transaksi, pembatalan, piutang, dan pembayaran piutang
- Import/export Excel dan laporan PDF
- Laporan Excel multi-sheet berisi ringkasan bulanan, rekap harian, histori transaksi, piutang, dan catatan opsional
- Rekap dashboard per hari berbasis zona waktu Asia/Jakarta (UTC+7)
- Login JWT dengan password bcrypt
- Otorisasi berbasis role: `admin`, `cashier`, dan `viewer`
- CRUD pengguna khusus admin
- Soft delete untuk barang, pelanggan, dan supplier

## Tech stack

- Go 1.26 dan Standard Library `net/http`
- PostgreSQL dan pgx v5
- Nuxt 3 dan Vue 3
- bcrypt dan JWT HS256
- ExcelJS dan SweetAlert2

## Persiapan

Pastikan Go, Node.js, npm, dan PostgreSQL tersedia. Buat database `pos_playground`, lalu salin konfigurasi environment:

```powershell
Copy-Item backend/.env.example backend/.env
```

Konfigurasi penting:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pos_playground
DB_SCHEMA=go_pos_playground
DB_SSLMODE=disable

JWT_SECRET=ganti-dengan-random-secret-minimal-32-karakter
JWT_ISSUER=go-pos-playground
JWT_EXPIRY_MINUTES=15
REFRESH_TOKEN_EXPIRY_DAYS=7

INITIAL_ADMIN_NAME=Administrator
INITIAL_ADMIN_EMAIL=admin@example.com
INITIAL_ADMIN_PASSWORD=ganti-dengan-password-kuat
```

`INITIAL_ADMIN_*` hanya digunakan ketika tabel `users` masih kosong. Setelah akun admin pertama berhasil dibuat, hapus `INITIAL_ADMIN_PASSWORD` dari `.env`.

Database schema dan tabel dibuat otomatis ketika backend dijalankan. Jika memakai nama database/schema versi lama, migrasikan terlebih dahulu ke `pos_playground` dan `go_pos_playground`.

## Menjalankan aplikasi

Terminal backend:

```powershell
Set-Location backend
go run ./cmd/api
```

Terminal frontend:

```powershell
npm.cmd --prefix frontend install
npm.cmd --prefix frontend run dev
```

Frontend tersedia di `http://localhost:3000`. Request `/api` diproksikan ke backend. Pastikan `APP_PORT` sama dengan target proxy di `frontend/nuxt.config.ts`.

## Authentication

Login menggunakan akun admin awal:

```http
POST /auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "password-admin"
}
```

Endpoint operasional selain health check, login, refresh, dan logout membutuhkan header berikut:

```http
Authorization: Bearer <access_token>
```

### Session dan refresh token rotation

Access token berlaku sesuai `JWT_EXPIRY_MINUTES`. Login juga membuat session server-side
dan refresh token acak yang dikirim melalui cookie `HttpOnly`, sehingga token jangka
panjang tidak dapat dibaca JavaScript:

- Aktivitas seperti perpindahan halaman, pemuatan data, dan operasi CRUD dicatat di browser tanpa request background tambahan.
- Ketika request API dilakukan dan access token mendekati atau melewati masa kedaluwarsa, frontend memanggil `POST /auth/refresh`.
- Request paralel berbagi proses refresh yang sama agar tidak menerbitkan banyak token sekaligus.
- Setiap refresh memutar token: token lama dicabut dan hanya hash token yang disimpan di tabel `auth_sessions`.
- Logout mencabut refresh session aktif dan menghapus cookie. Access token yang sudah diterbitkan tetap berumur pendek serta tidak dapat dipakai jika akun dinonaktifkan atau dihapus.
- Session berakhir setelah `REFRESH_TOKEN_EXPIRY_DAYS` atau ketika dicabut. Mengubah masa berlaku token memerlukan restart backend.

### Role dan akses

| Fitur | Admin | Cashier | Viewer |
|---|:---:|:---:|:---:|
| Dashboard dan baca barang | Ya | Ya | Ya |
| CRUD barang | Ya | Ya | Tidak |
| CRUD pelanggan | Ya | Ya | Tidak |
| CRUD supplier | Ya | Ya | Tidak |
| Membuat merek dari form barang | Ya | Ya | Tidak |
| Transaksi dan pembayaran piutang | Ya | Ya | Tidak |
| Kelola seluruh master data | Ya | Tidak | Tidak |
| CRUD pengguna | Ya | Tidak | Tidak |
| Melihat audit log | Ya | Tidak | Tidak |

Admin tidak dapat mengubah role, menonaktifkan, atau menghapus akun sendiri. Status dan role terbaru diperiksa ke database pada setiap request, sehingga perubahan akses langsung berlaku untuk token yang sudah diterbitkan.

## Endpoint utama

| Method | Endpoint | Keterangan |
|---|---|---|
| `GET` | `/health` | Health check publik |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/refresh` | Putar refresh session dan terbitkan access token baru |
| `POST` | `/auth/logout` | Cabut refresh session aktif |
| `GET` | `/auth/me` | Profil pengguna aktif |
| `GET, POST` | `/users` | Daftar dan tambah pengguna |
| `PUT, DELETE` | `/users/{id}` | Ubah dan hapus pengguna |
| `GET` | `/audit-logs` | Audit perubahan data khusus admin |
| `GET, POST` | `/items` | Daftar dan tambah barang |
| `GET, PUT, DELETE` | `/items/{id}` | Detail, ubah, dan soft delete barang |
| `GET, POST` | `/suppliers` | Daftar dan tambah supplier |
| `GET, PUT, DELETE` | `/suppliers/{id}` | Detail, ubah, dan soft delete supplier |
| `GET, POST` | `/customers` | Daftar dan tambah pelanggan |
| `GET, PUT, DELETE` | `/customers/{id}` | Detail, ubah, dan soft delete pelanggan |
| `GET` | `/dashboard` | Ringkasan operasional |
| `GET, POST` | `/transactions` | Histori dan pembuatan transaksi |
| `PUT` | `/transactions/{id}` | Ubah transaksi |
| `POST` | `/transactions/{id}/void` | Batalkan transaksi |
| `GET` | `/debts` | Daftar piutang |
| `POST` | `/debts/{id}/payments` | Catat pembayaran piutang |
| `GET, POST, PUT, DELETE` | `/masters/{name}` | Kelola master data |

### Pagination API

Endpoint daftar `/items`, `/suppliers`, `/customers`, `/transactions`, `/debts`, dan `/users` mendukung pagination opt-in:

```http
GET /items?page=1&per_page=20
GET /transactions?type=SALE&page=2&per_page=25
```

`page` dimulai dari `1`, nilai default `per_page` adalah `20`, dan batas maksimalnya `100`. Request dengan salah satu parameter pagination mengembalikan bentuk berikut:

```json
{
  "success": true,
  "data": {
    "items": [],
    "meta": {
      "page": 1,
      "per_page": 20,
      "total": 0,
      "total_pages": 0
    }
  }
}
```

Request tanpa `page` dan `per_page` tetap mengembalikan array seperti sebelumnya agar pilihan barang transaksi dan proses export tetap lengkap. Tabel frontend menerapkan pagination pada hasil filter lokal, dengan pilihan 10, 20, atau 50 baris. Migrasi tabel ke pagination server-side akan dilakukan bersama pencarian, sorting, dan filtering API agar pencarian tidak terbatas pada satu halaman.

## Struktur project

```text
.
|-- backend/
|   |-- cmd/api/                # application entry point
|   `-- internal/
|       |-- auth/               # JWT issuance dan validation
|       |-- config/             # environment configuration
|       |-- database/           # PostgreSQL dan migration
|       |-- dto/                # request DTO
|       |-- entity/             # domain entities
|       |-- handler/            # HTTP handlers
|       |-- middleware/         # authentication dan authorization
|       |-- repository/         # database access
|       `-- router/             # routes dan role policy
|-- frontend/                   # Nuxt operational console
|   |-- components/             # layout dan tampilan console
|   |-- composables/            # state dan operasi per domain
|   `-- utils/                  # utility session dan unit test
|-- docs/
|-- migrations/
|-- CHANGELOG.md
`-- README.md
```

Frontend memisahkan pemuatan data dan operasi berdasarkan domain dashboard, barang, pelanggan, supplier, master data, transaksi, piutang, dan pengguna. Setiap halaman hanya memuat data yang dibutuhkan oleh route aktif, sedangkan operasi CRUD hanya memuat ulang domain yang terdampak.

## Testing dan build

```powershell
Set-Location backend
go test ./...

# Integration test transaksi memakai schema PostgreSQL sementara yang
# otomatis dihapus setelah test selesai.
$env:GO_POS_INTEGRATION_TESTS='1'
go test ./internal/repository -run Integration -count=1
Remove-Item Env:GO_POS_INTEGRATION_TESTS

Set-Location ../frontend
npm.cmd run typecheck
npm.cmd test
npm.cmd run build
```

Integration test hanya menerima database bernama `playground` atau `pos_playground`, lalu membuat schema acak dengan prefix `go_pos_test_`. Schema aplikasi utama dan database lain tidak digunakan sebagai target test maupun cleanup.

## Seed data demo

Seed generator membuat supplier, pelanggan, barang, transaksi pembelian, transaksi penjualan, stok, dan piutang demo. Generator tidak dijalankan otomatis saat API startup.

```powershell
Set-Location backend
go run ./cmd/seed
```

Jumlah data dan rentang transaksi dapat diatur:

```powershell
go run ./cmd/seed `
  -suppliers 8 `
  -customers 30 `
  -items 20 `
  -purchases 60 `
  -sales 150 `
  -months 6 `
  -debt-rate 0.20 `
  -seed 20260720
```

`-debt-rate 0.20` membuat sekitar 20% penjualan menjadi piutang. Nilai `-seed` membuat hasil acak dapat direproduksi. Invoice dan kode seed menggunakan prefix `SEED-`; menjalankan ulang dengan seed yang sama tidak menduplikasi invoice.

## Roadmap

Pengembangan berikutnya berfokus pada pencarian, sorting, dan filtering API, session management, observability, serta deployment. Status lengkap dan prioritas terkini tersedia di [ROADMAP.md](ROADMAP.md).

### Query katalog barang

`GET /items` mendukung pencarian case-insensitive pada nama, SKU, dan deskripsi melalui
`search`. Hasil dapat difilter dengan `supplier_id`, `category_id`, `brand_id`,
`unit_id`, `min_stock`, dan `max_stock`, lalu diurutkan dengan `sort` dan `order`.
Nilai `sort` yang tersedia adalah `id`, `sku`, `name`, `stock`, `price`, `cost`,
`created_at`, dan `updated_at`; `order` menerima `asc` atau `desc`.

Semua parameter tersebut dapat digabungkan dengan pagination opt-in, misalnya:

```text
GET /items?search=kopi&category_id=2&min_stock=1&sort=stock&order=desc&page=1&per_page=20
```

Endpoint `GET /suppliers` menerima `search` pada kode, nama, telepon, dan alamat.
Sorting tersedia untuk `id`, `code`, `name`, `phone`, `created_at`, dan `updated_at`.

Endpoint `GET /customers` menerima `search` pada kode, nama, telepon, dan alamat,
serta filter `customer_type=MEMBER|NON_MEMBER`. Sorting tersedia untuk `id`, `code`,
`name`, `customer_type`, dan `created_at`. Parameter `search`, `sort`, `order`, dan
filter tersebut juga dapat digabungkan dengan `page` dan `per_page`.

Endpoint admin `GET /users` menerima `search` pada nama dan email, filter
`role=admin|cashier|viewer`, filter `active=true|false`, serta sorting berdasarkan
`id`, `name`, `email`, `role`, atau `active`.

Endpoint `GET /transactions` menerima:

- `search` pada nomor invoice, pelanggan, supplier, dan catatan
- `type=SALE|PURCHASE`
- `payment_status=PAID|UNPAID|PARTIAL`
- `status=ACTIVE|VOID`
- `date_from` dan `date_to` dalam format `YYYY-MM-DD` berdasarkan zona waktu Asia/Jakarta
- sorting `id`, `invoice_no`, `transaction_date`, `grand_total`, `payment_status`, atau `status`

Endpoint `GET /debts` menerima `search` pada invoice dan pelanggan, filter
`status=OPEN|PAID`, rentang `min_remaining`/`max_remaining`, serta sorting `id`,
`invoice_no`, `customer_name`, `original_amount`, `remaining_amount`, `status`,
atau `created_at`.

## AI-assisted development

Project ini dirancang, diarahkan, diuji, dan direview oleh Harits Achmad Fauzan. ChatGPT dan Codex digunakan sebagai alat bantu untuk brainstorming, diskusi arsitektur, review kode, refactoring, dan implementasi. Keputusan akhir, validasi fitur, smoke testing, dan persetujuan perubahan tetap dilakukan oleh pemilik project.

## License

Copyright (c) 2026 Harits Achmad Fauzan. Project ini dilisensikan menggunakan [MIT License](LICENSE).

Riwayat perubahan tersedia di [CHANGELOG.md](CHANGELOG.md), sedangkan arah pengembangan tersedia di [ROADMAP.md](ROADMAP.md).
