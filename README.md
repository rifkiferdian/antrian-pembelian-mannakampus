# Antrian Pembelian Manna Kampus

Aplikasi web untuk mengelola akses dan operasional portal antrian pembelian di Manna Kampus. Dibangun dengan Go (Gin) dan MySQL, dengan autentikasi berbasis session serta manajemen user, role, dan permission.

## Tech Stack

- Go (Gin Web Framework)
- MySQL (driver `go-sql-driver/mysql`)
- HTML templates (`templates/`) dan asset statis (`assets/`)
- Tailwind CSS via CDN (dipakai di template)

## Fitur

- Login dan logout user dengan password di-hash (bcrypt)
- Autentikasi berbasis session dan middleware proteksi halaman
- Dashboard (template UI)
- Manajemen user (tambah, ubah, hapus) dan mapping store
- Manajemen role dan permission
- Proteksi akses per permission pada route admin

## Struktur Proyek

- `main.go` - entry point aplikasi, inisialisasi Gin, session, template, dan server
- `routes/web.go` - definisi route utama
- `controllers/` - handler HTTP (auth, dashboard, user, role)
- `middleware/` - middleware autentikasi dan permission
- `models/` - struktur data domain (user, role, permission, session)
- `repositories/` - akses data ke database
- `services/` - logika bisnis (user, role, permission)
- `templates/` - file HTML template
- `assets/` - file CSS, JS, dan aset frontend
- `config/` - konfigurasi koneksi database
- `gobase_app.sql` - schema dan seed data database

## Persyaratan

- Go 1.21+ terinstall
- MySQL server berjalan

## Konfigurasi Environment

Aplikasi membaca variabel environment menggunakan `github.com/joho/godotenv`. Contoh `.env`:

```env
APP_NAME=AntrianService
APP_PORT=8080
BASE_URL=http://localhost:8080
APP_SECURE_COOKIE=false

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=
DB_NAME=gobase_app
```

Catatan:
- `BASE_URL` dipakai oleh template helper `baseURL` untuk membentuk URL absolut.
- `APP_SECURE_COOKIE=true` jika aplikasi diakses via HTTPS.

## Menyiapkan Database

1. Buat database sesuai `DB_NAME`.
2. Import schema dan seed data:

```bash
mysql -u root -p gobase_app < gobase_app.sql
```

## Menjalankan Aplikasi

1. Masuk ke folder proyek:

```bash
cd antrian_pembelian_go_v1
```

2. Download dependensi:

```bash
go mod tidy
```

3. Jalankan aplikasi:

```bash
go run main.go
```

4. Buka browser dan akses:

```text
http://localhost:8080
```

## Endpoint Utama

- `GET /` atau `GET /login` - halaman login
- `POST /login` - proses login
- `POST /register` - registrasi user baru
- `GET /logout` - logout user
- `GET /dashboard` - dashboard (butuh login)
- `GET /users` - manajemen user (butuh permission)
- `GET /role` - manajemen role (butuh permission)

Detail route dan middleware dapat dilihat di `routes/web.go`.

## Session dan Keamanan

Aplikasi menggunakan session berbasis cookie dari `github.com/gin-contrib/sessions`:

- Session name: `mysession`
- Secret key diset di `main.go` dan sebaiknya dipindahkan ke environment untuk production
- Cookie `Secure` mengikuti `APP_SECURE_COOKIE`

## Lisensi

Proyek ini digunakan untuk kebutuhan internal atau pembelajaran. Silakan modifikasi sesuai kebutuhan.
