# Changelog

Semua perubahan penting pada backend dan frontend Go POS Playground dicatat di file ini.

## Unreleased

### Added

- Roadmap project terpisah untuk melacak pekerjaan yang selesai, sedang berjalan, direncanakan, dan belum diprioritaskan.
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
- Quality gate `npm run typecheck` menggunakan TypeScript dan `vue-tsc`.
- Unit test frontend untuk parsing expiry JWT, keputusan refresh session, dan throttle activity tracker.
- Integration test PostgreSQL terisolasi untuk penjualan, pembelian, rollback stok, edit, pembatalan transaksi, dan pembayaran piutang.
- Guard integration test yang hanya menerima database playground serta membuat dan membersihkan schema `go_pos_test_*`.
- Pagination backend opt-in untuk barang, supplier, pelanggan, transaksi, piutang, dan pengguna.
- Metadata pagination reusable dengan validasi halaman, default 20 baris, dan batas maksimal 100 baris.
- Unit test parser pagination dan integration test PostgreSQL untuk batas halaman serta kombinasi filter tipe transaksi.
- Kontrol pagination frontend reusable untuk barang, pelanggan, supplier, histori, piutang, dan pengguna.
- Unit test pagination frontend untuk slicing, batas halaman, perubahan filter, dan penyesuaian jumlah data.
- Query katalog barang untuk pencarian nama/SKU/deskripsi, filter relasi dan rentang stok, serta sorting aman yang dapat digabungkan dengan pagination.
- Parser query list reusable dengan validasi panjang pencarian, allowlist sorting, arah urutan, dan nilai filter numerik.
- Pencarian dan sorting supplier pada kode, nama, telepon, alamat, serta kolom operasional.
- Pencarian, filter tipe, dan sorting pelanggan dengan urutan default pelanggan umum tetap dipertahankan.
- Pencarian nama/email, filter role/status aktif, dan sorting untuk daftar pengguna admin.
- Pencarian histori transaksi, filter tipe/status/tanggal, dan sorting dengan batas tanggal berbasis zona waktu Asia/Jakarta.
- Pencarian invoice/pelanggan, filter status/rentang sisa, dan sorting untuk daftar piutang.
- Sorting reusable melalui klik header kolom dengan indikator arah pada tabel barang, pelanggan, supplier, histori, dan pengguna.
- Session server-side dengan refresh token acak, penyimpanan hash, rotasi sekali pakai, masa berlaku configurable, dan endpoint logout.
- Unit test sorting frontend dan refresh-token manager backend.
- Audit log append-only untuk mutasi pengguna dengan snapshot identitas, aksi, target, status HTTP, alamat IP, dan user-agent tanpa menyimpan request body.
- Endpoint dan halaman audit khusus admin dengan search, filter aksi, pagination, dan sorting melalui header tabel.

### Changed

- Kontrol jumlah baris dipindahkan ke area filter, pagination memakai nomor halaman, dan tombol refresh ditempatkan dekat data yang diperbarui.
- Expiry JWT dan activity throttle diekstrak menjadi utility session yang dapat diuji.
- Activity tracker membersihkan DOM listener dan page hook ketika plugin mengalami hot reload.
- Frontend Vue/Vite lama dihapus setelah seluruh entrypoint aktif dipastikan menggunakan Nuxt.
- Pemuatan data frontend dipisahkan per domain sehingga route aktif dan operasi CRUD hanya memuat data yang diperlukan tanpa menambah jumlah request API.
- State dan operasi dashboard, barang, pelanggan, supplier, master data, transaksi, piutang, dan pengguna diekstrak dari `KoperasiConsole.vue` ke composable masing-masing.
- Perataan kolom tabel dan input harga barang diperbaiki agar tampilan serta validasi form lebih konsisten.
- Contoh masa berlaku access token diubah menjadi 15 menit; frontend memperbarui token pada lima menit terakhir saat ada aktivitas.
- Sliding session berbasis access token diganti dengan refresh-token rotation melalui cookie `HttpOnly`.
- Seluruh endpoint operasional kini membutuhkan autentikasi, kecuali `/health` dan `/auth/login`.
- Kasir dapat melakukan CRUD barang, pelanggan, dan supplier.
- Kasir dapat membuat merek baru dari form barang tanpa memperoleh akses penuh ke master data.
- Navigasi frontend disesuaikan berdasarkan role pengguna.
- Dokumentasi project diperbarui untuk mencakup frontend, authentication, authorization, dan setup admin.
- Changelog dipindahkan dari direktori `backend` ke root project.

### Fixed

- Migrasi upgrade audit log kini menambahkan snapshot identitas yang belum ada pada tabel versi awal dan melepas foreign key yang dapat menghambat penghapusan pengguna.
- Heading konten pada halaman master data dibedakan dari judul route agar hierarki halaman tidak terlihat berulang.
- Mempertahankan data transaksi yang akan diubah ketika berpindah dari halaman histori ke halaman kasir atau pembelian.
- Memperbaiki pattern kode/SKU agar valid pada browser yang menggunakan regular expression mode `v`.
- Menyamakan dokumentasi port backend dengan proxy frontend.

### Security

- File environment lokal dihapus dari seluruh histori Git dan histori bersih diverifikasi menggunakan Gitleaks.
- Refresh token lama langsung dicabut setelah rotasi dan tidak dapat digunakan ulang.
- Logout mencabut refresh session server-side.
- Audit log tidak menyimpan password, token, atau isi request sensitif dan tidak menghalangi penghapusan akun pengguna.
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
