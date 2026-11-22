# Architectural Notes & Decisions

Dokumen ini mencatat keputusan arsitektur, asumsi teknis, dan catatan penting yang terus diperbarui seiring berjalannya fase pengembangan.

---

## 1. Keputusan Arsitektur

### Backend (Go 1.22+)
- **Routing Standar (`net/http`)**: Menggunakan router bawaan Go 1.22+ (`http.ServeMux`) dengan method and path pattern (`"GET /path"`), menghindari dependensi eksternal berlebih sesuai prinsip kode eksplisit dan minimal.
- **Struktur Modular Sederhana**:
  - `cmd/server/`: Entrypoint binary utama aplikasi.
  - `internal/config/`: Parsing dan validasi konfigurasi dari environment variables dengan fallback default yang aman.
  - `internal/database/`: Pengelolaan koneksi PostgreSQL menggunakan driver `github.com/lib/pq`, connection pool, dan migration runner.
  - `migrations/`: Berkas SQL migrasi murni yang di-embed langsung ke binary dengan `embed.FS`.
- **Embedded Migrations Tanpa Tool Eksternal**:
  - Alih-alih menambahkan dependency CLI migrasi pihak ketiga (seperti `golang-migrate` CLI), kami mengimplementasikan runner migrasi eksplisit menggunakan fitur standar Go `//go:embed`.
  - Berkas SQL (.up.sql dan .down.sql) menyatu dalam binary Go, menjamin migrasi selalu sinkron dengan versi aplikasi dan otomatis dieksekusi saat container menyala tanpa langkah instalasi terpisah.
- **Tabel Pelacak `schema_migrations`**:
  - Setiap migrasi yang dieksekusi dicatat ke tabel `schema_migrations` di dalam database transaction (`BEGIN ... COMMIT`).
  - Jika terjadi kegagalan eksekusi pada salah satu file migrasi, perubahan di-rollback secara otomatis (`tx.Rollback()`).

### Containerization (Docker)
- **Multi-Stage Build**: Image Docker backend di-build menggunakan `golang:1.22-alpine` untuk kompilasi, kemudian binary disalin ke runtime image `alpine:3.20`.
- **Healthcheck & Dependency Order**: Service `backend` bergantung pada `db` dengan kondisi `service_healthy` (`pg_isready`) sehingga backend hanya menyala saat PostgreSQL sudah siap menerima query.

---

## 2. Catatan Fase 1 & 2
- **Fase 1**: Base monorepo diinisialisasi (`backend/`, `frontend/`, `docs/`, `docker-compose.yml`, `README.md`, `NOTES.md`) dengan endpoint `GET /api/health`.
- **Fase 2**: 
  - Skema database relasional lengkap (`users`, `tickets`, `comments`) dibuat dengan *foreign key constraints*, *check constraints*, dan *indexing*.
  - Migrasi reversibel (.up.sql dan .down.sql) diimplementasikan dan dijalankan otomatis pada saat startup server.
  - Akun awal (seed data) untuk ketiga peran (`admin`, `agent`, `requester`) telah terisi dengan password ter-hash bcrypt.
