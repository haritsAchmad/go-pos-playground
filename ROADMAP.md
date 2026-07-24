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
- Pagination frontend responsif untuk tabel operasional
- Penghapusan frontend Vue/Vite lama setelah migrasi Nuxt
- Seed generator untuk data demo yang dapat direproduksi
- Pencarian, sorting, dan filtering API untuk barang, supplier, pelanggan, transaksi, piutang, dan pengguna
- Sorting frontend melalui header kolom tabel operasional
- Refresh token rotation dan session server-side dengan logout/revocation
- Audit log aktivitas mutasi pengguna dengan halaman monitoring admin

## 🟨 In Progress

- Belum ada item aktif; prioritas berikutnya dipilih dari bagian Planned

## ⬜ Planned

- Pemulihan data yang menggunakan soft delete
- Bulk soft delete dan bulk restore
- Bulk payment/settlement untuk beberapa piutang
- Bulk reset stok terpilih ke 0 dengan konfirmasi
- Snackbar undo untuk soft delete
- Audit log untuk seluruh bulk action
- Benchmark dan performance baseline
- Docker dan deployment configuration
- Redis untuk caching atau session support
- Background job atau queue untuk proses berat

### Catatan fitur bulk

- Bulk payment/settlement wajib membuat histori pembayaran untuk setiap piutang dan tidak boleh hanya mengubah status menjadi `PAID`.
- Bulk reset stok wajib dicatat sebagai stock movement untuk setiap barang agar perubahan stok tetap dapat ditelusuri.
- Seluruh bulk action harus memiliki konfirmasi, hasil yang jelas untuk setiap item, dan audit log yang mengidentifikasi pengguna serta target perubahan.

## Not Planned Yet

Elasticsearch dan komponen infrastruktur lain belum menjadi prioritas sampai kebutuhan pencarian, skala data, atau hasil benchmark menunjukkan alasan yang jelas untuk menambahkannya.
