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
│   │   └── database/      # Koneksi PostgreSQL, connection pool & migration runner
│   ├── migrations/        # SQL migration files (.up.sql & .down.sql) & embed.go
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

## Akun Bawaan (Seed Data)

Database secara otomatis dimigrasikan dan diisi data akun awal saat backend pertama kali dijalankan:

| Peran (Role) | Nama | Email | Password |
|---|---|---|---|
| **Admin** | System Admin | `admin@example.com` | `Password123!` |
| **Agent** | Support Agent | `agent@example.com` | `Password123!` |
| **Requester** | Requester User | `requester@example.com` | `Password123!` |

*Catatan: Seluruh password disimpan dalam bentuk hash bcrypt (cost 10).*

---

## Asumsi Teknis (Spesifikasi)

1. **Format Role & Status**:
   - Role pengguna disimpan dalam format string huruf kecil konsisten: `requester`, `agent`, `admin`.
   - Status tiket dinormalisasi dalam format string: `Open`, `In Progress`, `Resolved`, `Closed` sesuai tabel transisi spesifikasi.
   - Prioritas tiket: `low`, `medium`, `high`, `urgent`.
2. **Koneksi Database & Port**:
   - Port default backend di host dikonfigurasi via `.env` (`PORT=8080` atau `8081` jika port 8080 terpakai oleh service lain di host).
   - Service database menggunakan container internal bernama `db` dengan network default dari Docker Compose.
3. **Mekanisme Migrasi**:
   - Migrasi menggunakan SQL murni dengan skema `.up.sql` dan `.down.sql`, dikompilasi langsung ke binary via standard library `embed.FS`.
   - Dilengkapi tabel pelacak `schema_migrations` sehingga eksekusi bersifat idempoten dan aman dijalankan berulang kali.

---

## Cara Menjalankan

### 1. Siapkan Environment File
Pastikan file konfigurasi `.env` telah disalin dari `.env.example`:

```bash
cp .env.example .env
```
*(Di Windows PowerShell: `Copy-Item .env.example .env`)*

### 2. Jalankan dengan Docker Compose
Pastikan Docker Desktop sudah aktif, lalu jalankan:

```bash
docker compose up -d --build
```

### 3. Verifikasi Health Check
Setelah container `db` (status healthy) dan `backend` berjalan, periksa endpoint health check:

```bash
curl http://localhost:8081/api/health
```
*(Sesuaikan port dengan nilai `PORT` di file `.env` Anda)*

Respons yang diharapkan (HTTP 200 OK):
```json
{"status":"ok","database":"connected"}
```

### 4. Verifikasi Data Seed di Database
Untuk memastikan tabel dan data seed pengguna telah terisi:

```bash
docker exec helpdesk-db psql -U helpdesk_user -d helpdesk_db -c "SELECT id, name, email, role FROM users;"
```
