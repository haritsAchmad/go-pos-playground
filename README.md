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

Endpoint selain health check dan login membutuhkan header berikut:

```http
Authorization: Bearer <access_token>
```

### Sliding session

Access token berlaku sesuai `JWT_EXPIRY_MINUTES`. Frontend menggunakan sliding session ringan untuk pengguna aktif:

- Aktivitas seperti perpindahan halaman, pemuatan data, dan operasi CRUD dicatat di browser tanpa request background tambahan.
- Ketika request API dilakukan dan sisa umur token maksimal lima menit, frontend memanggil `POST /auth/refresh` satu kali lalu menggunakan token baru.
- Request paralel berbagi proses refresh yang sama agar tidak menerbitkan banyak token sekaligus.
- Token yang sudah kedaluwarsa tidak dapat diperbarui. Pengguna yang idle sampai batas waktu akan diarahkan kembali ke `/login`.
- Mengubah `JWT_EXPIRY_MINUTES` memerlukan restart backend dan berlaku untuk token yang diterbitkan setelah login atau refresh berikutnya.

Implementasi ini tidak menggunakan refresh token jangka panjang atau penyimpanan session server-side. Karena itu, logout tidak mencabut JWT yang sudah disalin ke tempat lain; token tersebut tetap valid sampai waktu kedaluwarsanya.

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

Admin tidak dapat mengubah role, menonaktifkan, atau menghapus akun sendiri. Status dan role terbaru diperiksa ke database pada setiap request, sehingga perubahan akses langsung berlaku untuk token yang sudah diterbitkan.

## Endpoint utama

| Method | Endpoint | Keterangan |
|---|---|---|
| `GET` | `/health` | Health check publik |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/refresh` | Perpanjang access token yang masih aktif |
| `GET` | `/auth/me` | Profil pengguna aktif |
| `GET, POST` | `/users` | Daftar dan tambah pengguna |
| `PUT, DELETE` | `/users/{id}` | Ubah dan hapus pengguna |
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
|-- docs/
|-- migrations/
|-- CHANGELOG.md
`-- README.md
```

## Testing dan build

```powershell
Set-Location backend
go test ./...

Set-Location ../frontend
npm.cmd run build
```

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

- Pagination, sorting, dan pencarian di API
- Refresh token atau session rotation
- Audit log aktivitas pengguna
- Pemulihan data soft delete
- Docker dan deployment configuration
- Redis untuk caching atau session support

Riwayat perubahan tersedia di [CHANGELOG.md](CHANGELOG.md).
