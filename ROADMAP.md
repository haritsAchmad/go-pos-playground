# Roadmap

Roadmap ini menggambarkan arah pengembangan Go POS Playground berdasarkan implementasi dan test yang benar-benar tersedia di repository. Project ini adalah portfolio/playground; pembayaran non-tunai tetap berupa simulasi dan tidak akan terhubung ke payment gateway atau uang nyata.

## Completed

- Autentikasi JWT dengan refresh-token rotation, session server-side, logout, dan revocation
- Otorisasi berbasis role dan manajemen pengguna
- Manajemen barang, pelanggan, supplier, dan master data
- Transaksi penjualan, pembelian, pembatalan, dan pengelolaan stok atomik
- Piutang, pembayaran piutang, reversal, dan bulk settlement dengan histori per piutang
- Dashboard dan rekap harian berbasis zona waktu Asia/Jakarta
- Export laporan bulanan Excel multi-sheet dan laporan PDF
- Frontend Nuxt berbasis route dengan domain composables, pagination, filtering, dan sorting
- TypeScript quality gate serta unit test utility frontend
- Integration test PostgreSQL untuk stok, transaksi, pembatalan, pembayaran piutang, dan rollback
- Soft delete, restore, bulk action, snackbar undo, dan audit log
- Seed generator data demo yang dapat direproduksi
- Benchmark PostgreSQL untuk checkout dan histori transaksi berhalaman
- Nomor invoice berbasis sequence PostgreSQL dengan test concurrent checkout
- Idempotency key checkout dengan perlindungan retry paralel dan deteksi payload berbeda

### Simulated non-cash payment

- Checkout `QRIS Dummy` membuat transaksi dan payment berstatus `PENDING`
- Tabel `payments` dan `stock_reservations` menyimpan expiry 15 menit
- Stok yang masih direservasi diperhitungkan saat checkout atau edit transaksi lain
- Endpoint simulator dapat mengubah payment menjadi `PAID`, `FAILED`, atau `EXPIRED`
- Transisi status yang sama bersifat idempotent dan callback `PAID` paralel diuji agar stok hanya berkurang sekali
- Status `PAID` memfinalisasi stok; `FAILED` dan `EXPIRED` melepas reservasi serta membatalkan transaksi
- Frontend kasir dapat membuat payment dummy dan memilih simulasi berhasil, gagal, atau membiarkannya pending
- Histori transaksi menampilkan payment dan menyediakan aksi simulasi ulang untuk payment pending
- Endpoint status payment mempersist expiry secara idempotent dan menjadi sumber polling frontend
- Frontend menampilkan countdown, melakukan polling tiap tiga detik, pulih setelah refresh, dan membersihkan polling saat komponen dilepas
- Pembatalan payment pending menggunakan endpoint lifecycle khusus yang melepas reservasi
- Callback `PAID` setelah expiry ditolak tanpa menghidupkan kembali transaksi
- Integration test PostgreSQL mencakup double callback, polling paralel, callback terlambat, cancel idempotent, dan race antara callback `PAID` dengan cancel
- Unit test frontend mencakup countdown, status terminal, interval polling, cleanup, dan proteksi request overlap

Lifecycle simulator sudah lengkap untuk scope portfolio/playground dan tidak terhubung ke payment gateway atau uang nyata.

## Completed — hardening v1.0.0-rc.1

### Must-have

- Verifikasi manual alur kasir dari `PENDING` menuju `PAID`, `FAILED`, dan `EXPIRED`
- Verifikasi recovery setelah refresh dan setelah halaman histori ditutup lalu dibuka kembali
- Jalankan seluruh quality gate sebelum tag release candidate

### Should-have terselesaikan

- Tampilkan stok tersedia setelah memperhitungkan reservasi aktif pada pengalaman kasir
- Berikan pesan error yang konsisten untuk konflik state dan payment yang sudah terminal
- Tambahkan audit trail untuk perubahan status payment yang dipicu simulator atau expiry

### Nice-to-have terselesaikan

- Kontrol durasi expiry melalui konfigurasi khusus development/demo
- Skenario demo deterministik untuk `PAID`, `FAILED`, dan `EXPIRED`
- Indikator payment pending pada dashboard

Implementasi hardening mencakup stok tersedia setelah reservasi, konflik terminal berbasis HTTP `409`, aksi audit `PAYMENT_SIMULATION`/`PAYMENT_CANCEL`/`PAYMENT_EXPIRY`, expiry configurable, pemilih skenario deterministik, dan counter pending di dashboard. Seluruh Go test, integration test PostgreSQL, unit test frontend, typecheck, production build, serta 489 skenario authorization E2E lintas Chromium, Chrome, Edge, dan Firefox terakhir lulus pada 19 Agustus 2026 tanpa Docker.

## Completed untuk kesiapan deployment

- Performance baseline checkout dan histori berhalaman tersedia melalui benchmark PostgreSQL terisolasi
- Image production terpisah untuk API dan web serta Docker Compose dengan PostgreSQL, healthcheck, volume, dan urutan startup

## Conditional, belum dibutuhkan

- Redis hanya akan ditambahkan jika pengukuran menunjukkan bottleneck caching atau session yang nyata
- Background job atau queue hanya akan ditambahkan jika ada proses berat yang tidak layak ditangani sinkron

Kedua item conditional bukan pekerjaan release yang tertunda: arsitektur saat ini belum memiliki kebutuhan terukur yang membenarkan kompleksitas operasional tersebut.

## Not Planned

- Integrasi payment gateway, QRIS nyata, kartu debit/kredit, atau perpindahan uang nyata
- Elasticsearch dan komponen infrastruktur lain tanpa kebutuhan yang dibuktikan oleh skala data atau benchmark

## Release direction

Simulated non-cash payment dan hardening teknisnya telah selesai untuk `v1.0.0-rc.1`. Quality gate otomatis non-Docker telah lulus; validasi manual end-to-end dan smoke test deployment tetap menjadi syarat sebelum promosi ke `v1.0.0` stabil.
