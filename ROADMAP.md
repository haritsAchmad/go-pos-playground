# Roadmap

Roadmap ini menggambarkan arah pengembangan Go POS Playground saat ini. Prioritas dapat berubah berdasarkan hasil pembelajaran, kebutuhan operasional, dan temuan selama pengembangan.

## ✅ Completed

- Autentikasi JWT dengan sliding session
- Otorisasi berbasis role dan manajemen pengguna
- Manajemen barang, pelanggan, supplier, dan master data
- Transaksi penjualan, pembelian, pembatalan, dan pengelolaan stok atomik
- Piutang dan pembayaran piutang
- Dashboard dan rekap harian berbasis zona waktu Asia/Jakarta
- Export laporan bulanan Excel multi-sheet dan laporan PDF
- Frontend loader refactor berbasis route aktif
- Pemisahan state dan operasi ke domain composables
- TypeScript quality gate untuk frontend Nuxt
- Cleanup activity tracker dan lifecycle sliding session
- Unit test untuk expiry, refresh window, dan activity throttle
- Integration test PostgreSQL untuk konsistensi stok penjualan dan pembelian
- Integration test untuk rollback edit, pembatalan transaksi, dan pembayaran piutang
- Backend pagination opt-in dengan metadata dan batas ukuran halaman
- Penghapusan frontend Vue/Vite lama setelah migrasi Nuxt
- Seed generator untuk data demo yang dapat direproduksi

## 🟨 In Progress

- Belum ada item aktif; prioritas berikutnya dipilih dari bagian Planned

## ⬜ Planned

- Pencarian, sorting, dan filtering di API
- Refresh token rotation atau session server-side
- Audit log aktivitas pengguna
- Pemulihan data yang menggunakan soft delete
- Benchmark dan performance baseline
- Docker dan deployment configuration
- Redis untuk caching atau session support
- Background job atau queue untuk proses berat

## Not Planned Yet

Elasticsearch dan komponen infrastruktur lain belum menjadi prioritas sampai kebutuhan pencarian, skala data, atau hasil benchmark menunjukkan alasan yang jelas untuk menambahkannya.
