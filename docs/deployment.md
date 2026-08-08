# Deployment

Konfigurasi container menyediakan tiga service: PostgreSQL, API Go, dan Nuxt production server. Hanya port web `3000` yang dipublikasikan; browser mengakses API melalui proxy `/api` milik Nuxt.

## Menjalankan

```powershell
$env:JWT_SECRET = 'secret-acak-minimal-32-karakter'
$env:POSTGRES_PASSWORD = 'password-database-yang-kuat'
$env:INITIAL_ADMIN_PASSWORD = 'password-admin-yang-kuat'
docker compose up --build -d
docker compose ps
```

API menunggu database sehat dan web menunggu API sehat. Endpoint health API adalah `/health`. Data PostgreSQL persisten pada named volume `postgres-data`.

## Konfigurasi penting

- `JWT_SECRET`, `POSTGRES_PASSWORD`, dan `INITIAL_ADMIN_PASSWORD` wajib dioverride di luar demo lokal.
- `DUMMY_PAYMENT_EXPIRY_MINUTES` menerima 1–1440 dan default ke 15.
- `INITIAL_ADMIN_EMAIL` dan `INITIAL_ADMIN_NAME` dapat dioverride saat bootstrap pertama.
- `NUXT_API_PROXY_TARGET` ditanam saat image web dibangun; Compose mengarahkannya ke `http://api:8082`.

Jangan commit file `.env`. Gunakan secret manager platform deployment untuk production. Setelah admin pertama dibuat, rotasi atau hapus nilai bootstrap password dari konfigurasi runtime.

## Operasi

```powershell
docker compose logs -f api web
docker compose restart api web
docker compose down
```

`docker compose down` mempertahankan volume database. Menghapus volume akan menghapus data dan harus dilakukan hanya dengan keputusan eksplisit serta backup yang sesuai.
