# Changelog

Semua perubahan penting pada backend dan frontend Go POS Playground dicatat di file ini.

## Unreleased

### Added

- Sliding session ringan dengan endpoint `POST /auth/refresh` untuk memperpanjang JWT pengguna aktif.
- Pencatatan aktivitas frontend tanpa polling serta deduplikasi refresh untuk request API paralel.
- Routing Nuxt per halaman, halaman login terpisah, dan middleware navigasi berbasis autentikasi.
- Export laporan bulanan Excel multi-sheet untuk ringkasan, penjualan harian, histori penjualan, histori pembelian, piutang, dan catatan opsional.
- Tabel rekap harian dashboard lengkap dengan total bulanan.
- Agregasi laporan berbasis zona waktu Asia/Jakarta (UTC+7).
- Command seed generator configurable untuk supplier, pelanggan, barang, pembelian, penjualan, stok, dan piutang.
- Autentikasi JWT HS256 dan endpoint profil pengguna.
- Password hashing menggunakan bcrypt.
- Otorisasi berbasis role `admin`, `cashier`, dan `viewer`.
- Tabel serta migration pengguna dengan initial admin dari environment.
- CRUD pengguna khusus admin.
- Halaman login, penyimpanan access token, profil pengguna, dan logout pada frontend.
- Halaman administrasi pengguna dengan pengaturan role dan status akun.
- Validasi role dan status pengguna terkini pada setiap authenticated request.
- Unit test untuk issuance, parsing, dan validasi konfigurasi JWT.

### Changed

- Contoh masa berlaku access token diubah menjadi 15 menit; frontend memperbarui token pada lima menit terakhir saat ada aktivitas.
- Seluruh endpoint operasional kini membutuhkan autentikasi, kecuali `/health` dan `/auth/login`.
- Kasir dapat melakukan CRUD barang, pelanggan, dan supplier.
- Kasir dapat membuat merek baru dari form barang tanpa memperoleh akses penuh ke master data.
- Navigasi frontend disesuaikan berdasarkan role pengguna.
- Dokumentasi project diperbarui untuk mencakup frontend, authentication, authorization, dan setup admin.
- Changelog dipindahkan dari direktori `backend` ke root project.

### Fixed

- Memperbaiki pattern kode/SKU agar valid pada browser yang menggunakan regular expression mode `v`.
- Menyamakan dokumentasi port backend dengan proxy frontend.

### Security

- Token yang sudah kedaluwarsa tidak dapat digunakan untuk menghidupkan kembali sliding session.
- Respons `401` membersihkan token frontend dan mengarahkan pengguna kembali ke halaman login.
- Admin tidak dapat mengubah role, menonaktifkan, atau menghapus akun sendiri.
- Token milik pengguna yang dinonaktifkan atau dihapus langsung ditolak.
- Perubahan role langsung berlaku tanpa menunggu access token kedaluwarsa.

## v0.2.0

### Added

- PostgreSQL connection.
- Endpoint `GET /items` dan `POST /items`.
- Environment configuration.
- Air hot reload configuration.

### Changed

- Rename package `model` menjadi `entity`.
- Rename package `request` menjadi `dto`.
