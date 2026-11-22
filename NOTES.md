# Architectural Notes & Decisions

Dokumen ini mencatat keputusan arsitektur, asumsi teknis, dan catatan penting yang terus diperbarui seiring berjalannya fase pengembangan.

---

## 1. Keputusan Arsitektur

### Backend (Go 1.22+)
- **Routing Standar (`net/http`)**: Menggunakan router bawaan Go 1.22+ (`http.ServeMux`) dengan method and path pattern (`"GET /path"`), menghindari dependensi eksternal berlebih seperti Gin/Echo sesuai prinsip kode eksplisit dan minimal.
- **Struktur Modular Sederhana**:
  - `cmd/server/`: Entrypoint binary utama aplikasi.
  - `internal/config/`: Parsing dan validasi konfigurasi dari environment variables dengan fallback default yang aman.
  - `internal/database/`: Pengelolaan koneksi PostgreSQL menggunakan driver `github.com/lib/pq` dan connection pool (`database/sql`).
- **Graceful Shutdown**: Menggunakan signal OS (`SIGINT`, `SIGTERM`) untuk menutup server dan koneksi DB secara bersih tanpa memutus request yang sedang berjalan mendadak.

### Containerization (Docker)
- **Multi-Stage Build**: Image Docker backend di-build menggunakan `golang:1.22-alpine` untuk proses kompilasi, kemudian binary statis disalin ke runtime image `alpine:3.20` yang sangat kecil dan aman.
- **Healthcheck & Dependency Order**: Service `backend` bergantung pada `db` dengan kondisi `service_healthy` (`pg_isready`) sehingga backend hanya menyala saat PostgreSQL sudah siap menerima query.

---

## 2. Catatan Fase 1
- Base monorepo diinisialisasi (`backend/`, `frontend/`, `docs/`, `docker-compose.yml`, `README.md`, `NOTES.md`).
- Endpoint `GET /api/health` memverifikasi kesehatan koneksi database dan mengembalikan status dalam format JSON.
