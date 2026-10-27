# Credits

go-pos-playground dibangun di atas karya orang lain. File ini mencatat
library dan program pihak lain yang dipakai, beserta pembuat dan
lisensinya. Kode proyek ini sendiri berlisensi MIT (lihat [LICENSE](LICENSE)).

Setiap kali menambah library, font, ikon, atau gambar dari pihak lain,
tambahkan juga ke file ini di PR yang sama.

Terakhir diperiksa: 27 Oktober 2026.

## Backend (Go)

Daftar dari `go list -deps ./...` di `backend/` (tanpa dependensi khusus
tes); lisensi dibaca dari file LICENSE di module cache. Toolchain:
[Go](https://go.dev) (BSD-3-Clause).

| Modul | Versi | Lisensi |
| --- | --- | --- |
| [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) | v5.10.0 | MIT |
| [github.com/jackc/pgpassfile](https://github.com/jackc/pgpassfile) | v1.0.0 | MIT |
| [github.com/jackc/pgservicefile](https://github.com/jackc/pgservicefile) | v0.0.0-20240606120523-5a60cdf6a761 | MIT |
| [github.com/jackc/puddle/v2](https://github.com/jackc/puddle) | v2.2.2 | MIT |
| [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) | v10.30.3 | MIT |
| [github.com/go-playground/locales](https://github.com/go-playground/locales) | v0.14.1 | MIT |
| [github.com/go-playground/universal-translator](https://github.com/go-playground/universal-translator) | v0.18.1 | MIT |
| [github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) | v1.4.13 | MIT |
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | v1.5.1 | MIT |
| [github.com/leodido/go-urn](https://github.com/leodido/go-urn) | v1.4.0 | MIT |
| `golang.org/x/crypto`, `x/sync`, `x/sys`, `x/text` | -- | BSD-3-Clause |

## Frontend (npm)

| Paket | Lisensi | Dipakai untuk |
| --- | --- | --- |
| [Nuxt](https://nuxt.com/) 3, [Vue](https://vuejs.org/) 3 | MIT | UI |
| [ExcelJS](https://github.com/exceljs/exceljs) | MIT | Ekspor laporan Excel |
| [SweetAlert2](https://sweetalert2.github.io/) | MIT | Dialog konfirmasi |

Alat pengembangan (Playwright Apache-2.0, TypeScript Apache-2.0, Vitest
MIT, vue-tsc MIT) tidak ikut di hasil build.

## Program

- **[PostgreSQL](https://www.postgresql.org/)** -- PostgreSQL License.

Ada yang kurang atau salah? Buka issue atau PR.
