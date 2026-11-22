# Helpdesk Application (Fullstack)

Proyek ini adalah implementasi sistem helpdesk internal berbasis **Go 1.22+** (`net/http`), **Vue 3** (Composition API), dan **PostgreSQL**, dirancang berdasarkan spesifikasi teknis *Shift Engineer - Fullstack*.

---

## Struktur Monorepo

```text
.
├── backend/               # Backend Go 1.22+
│   ├── cmd/server/        # Entrypoint aplikasi (main.go)
│   ├── internal/
│   │   ├── config/        # Loader environment variable
│   │   └── database/      # Koneksi database PostgreSQL & pooling
│   ├── Dockerfile         # Multi-stage Docker build untuk Go backend
│   ├── go.mod             # Go module definition
│   └── go.sum             # Go module checksums
├── frontend/              # Frontend Vue 3 (akan dikonfigurasi di fase frontend)
├── docs/                  # Dokumentasi API dan arsitektur
├── docker-compose.yml     # Konfigurasi container service DB & backend
├── .env.example           # Template environment variable (tanpa secret riil)
├── .gitignore             # Berkas yang diabaikan oleh Git
├── NOTES.md               # Catatan keputusan arsitektur & hal teknis
└── README.md              # Petunjuk menjalankan dan dokumentasi proyek
```

---

## Asumsi Teknis (Spesifikasi)

1. **Format Role & Status**:
   - Role pengguna disimpan dalam format string huruf kecil konsisten: `requester`, `agent`, `admin`.
   - Status tiket dinormalisasi dalam format string: `Open`, `In Progress`, `Resolved`, `Closed` sesuai tabel transisi spesifikasi.
2. **Koneksi Database & Port**:
   - Port default backend di host adalah `8080`, dan database PostgreSQL di port `5432`.
   - Service database menggunakan container internal bernama `db` dengan network default dari Docker Compose.

---

## Cara Menjalankan (Fase 1)

### 1. Salin Environment File
Pastikan file konfigurasi `.env` telah disalin dari `.env.example`:

```bash
cp .env.example .env
```
*(Di Windows PowerShell: `Copy-Item .env.example .env`)*

### 2. Jalankan dengan Docker Compose
Pastikan Docker Desktop sudah aktif, lalu jalankan:

```bash
docker compose up --build
```
*(Atau `docker-compose up --build` jika menggunakan versi Compose standalone)*

### 3. Verifikasi Health Check
Setelah container `db` dan `backend` berjalan (health status database `healthy`), buka terminal baru dan uji endpoint health check:

```bash
curl http://localhost:8080/api/health
```

Respons yang diharapkan (HTTP 200 OK):
```json
{"status":"ok","database":"connected"}
```
