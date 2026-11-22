# 📋 Helpdesk Fullstack Specification & Progress Tracker
> *Catatan: File ini adalah referensi kerja sementara untuk menyelaraskan pemahaman teknis spesifikasi tes. File ini akan dihapus sebelum repository dikumpulkan ke pihak reviewer.*

---

## 1. Studi Kasus & Hak Akses Peran (RBAC)

Aplikasi helpdesk internal untuk menggantikan email dan spreadsheet penanganan tiket dukungan karyawan.

| Peran (Role) | Hak Akses Utama |
|---|---|
| **Requester** (Karyawan) | • Membuat tiket baru.<br>• Melihat dan berkomentar **hanya pada tiket miliknya** (dilarang melihat tiket orang lain).<br>• Menutup tiket miliknya sendiri (`Open -> Closed` atau `Resolved -> Closed`).<br>• Membuka kembali tiket miliknya (`Resolved -> Open`). |
| **Agent** (Support) | • Melihat **semua tiket** dari seluruh karyawan.<br>• Mengambil / menetapkan (*assign*) tiket ke diri sendiri atau agent lain.<br>• Mengubah status dan prioritas tiket.<br>• Menulis komentar publik.<br>• Menulis **catatan internal** (*internal notes*) yang tersembunyi dari requester. |
| **Admin** (Administrator) | • Memiliki **seluruh hak akses Agent**.<br>• Melihat daftar seluruh pengguna (`GET /api/users`).<br>• Mengubah peran pengguna (`PATCH /api/users/{id}/role`). |

---

## 2. Aturan Baku Transisi Status Tiket

Backend **wajib menolak** perpindahan status di luar tabel ini dengan pesan error yang jelas dan kode HTTP yang tepat (misal `400 Bad Request` atau `422 Unprocessable Entity`):

| Dari Status | Ke Status | Siapa yang Boleh Memindahkan | Keterangan / Aksi |
|---|---|---|---|
| `Open` | `In Progress` | **Agent, Admin** | Agent mulai mengerjakan tiket |
| `In Progress` | `Resolved` | **Agent, Admin** | Solusi telah diberikan oleh support |
| `Resolved` | `Closed` | **Requester, Agent, Admin** | Tiket selesai diverifikasi dan ditutup |
| `Resolved` | `Open` | **Requester** | Masalah belum tuntas, requester membuka kembali tiket |
| `Open` | `Closed` | **Requester, Admin** | Requester membatalkan permintaan, atau admin menutup langsung |

---

## 3. Spesifikasi Database (PostgreSQL)

### Tabel `users`
* `id` (BIGSERIAL PRIMARY KEY)
* `name` (VARCHAR)
* `email` (VARCHAR UNIQUE)
* `password_hash` (VARCHAR)
* `role` (VARCHAR CHECK: `requester`, `agent`, `admin`)
* `created_at` (TIMESTAMPTZ)

### Tabel `tickets`
* `id` (BIGSERIAL PRIMARY KEY)
* `title` (VARCHAR)
* `description` (TEXT)
* `category` (VARCHAR)
* `priority` (VARCHAR CHECK: `low`, `medium`, `high`, `urgent`)
* `status` (VARCHAR CHECK: `Open`, `In Progress`, `Resolved`, `Closed`)
* `requester_id` (BIGINT REFERENCES users(id))
* `assignee_id` (BIGINT NULLABLE REFERENCES users(id))
* `created_at` (TIMESTAMPTZ)
* `updated_at` (TIMESTAMPTZ)

### Tabel `comments`
* `id` (BIGSERIAL PRIMARY KEY)
* `ticket_id` (BIGINT REFERENCES tickets(id))
* `author_id` (BIGINT REFERENCES users(id))
* `body` (TEXT)
* `is_internal` (BOOLEAN DEFAULT FALSE)
* `created_at` (TIMESTAMPTZ)

---

## 4. Daftar Minimal Endpoint REST API

| Method | Path | Fungsi | Hak Akses |
|---|---|---|---|
| `POST` | `/api/auth/register` | Mendaftar akun baru (default role: requester) | Publik |
| `POST` | `/api/auth/login` | Login, verifikasi bcrypt, mengembalikan token JWT | Publik |
| `GET` | `/api/me` | Mengambil profil user yang sedang login | Authenticated (Semua role) |
| `GET` | `/api/tickets` | Mengambil daftar tiket (filter, search, pagination) | Authenticated (Requester dibatasi miliknya) |
| `POST` | `/api/tickets` | Membuat tiket baru | Authenticated (Semua role) |
| `GET` | `/api/tickets/{id}` | Mengambil detail tiket beserta riwayat komentar | Pemilik tiket, Agent, Admin |
| `PATCH` | `/api/tickets/{id}` | Mengubah status, prioritas, atau assignee | Sesuai tabel transisi status & hak peran |
| `POST` | `/api/tickets/{id}/comments` | Menambahkan komentar publik atau catatan internal | Pemilik tiket, Agent, Admin |
| `GET` | `/api/users` | Mengambil daftar pengguna | Khusus Admin |
| `PATCH` | `/api/users/{id}/role` | Mengubah peran pengguna lain | Khusus Admin |

---

## 5. Rencana Fase & Sesi Pengerjaan (Roadmap)

Total terdapat **10 Fase terukur**:

- [x] **Fase 1: Project Setup & Health Endpoint**
  - Setup monorepo (`backend/`, `frontend/`, `docs/`, `docker-compose.yml`, `.env.example`, `.gitignore`).
  - Init Git & commit bertahap.
  - Go HTTP Server dengan Go 1.22 `net/http` ServeMux (`GET /api/health`).
  - Docker Compose untuk DB PostgreSQL & Backend.
- [x] **Fase 2: Skema Database, Reversible Migrations & Seed Data**
  - Berkas migrasi `.up.sql` dan `.down.sql` untuk tabel `users`, `tickets`, `comments`.
  - Embedded migration runner di Go (`//go:embed`).
  - Seed 3 akun awal (`admin@example.com`, `agent@example.com`, `requester@example.com` / `Password123!`).
- [x] **Fase 3: Autentikasi & Authorization (JWT & RBAC Middleware)**
  - Register (`POST /api/auth/register`) dengan hashing password bcrypt.
  - Login (`POST /api/auth/login`) mengembalikan JWT token.
  - Profil (`GET /api/me`).
  - Middleware JWT extractor & Middleware RBAC checker.
- [x] **Fase 4: Core Ticketing API (CRUD, Query Filters & State Transition)**
  - Buat tiket (`POST /api/tickets`).
  - List tiket (`GET /api/tickets`) dengan pagination, status/priority/assignee filter, search, & pembatasan peran requester.
  - Detail tiket (`GET /api/tickets/{id}`).
  - Update tiket (`PATCH /api/tickets/{id}`) dengan validasi ketat status transition matrix.
- [x] **Fase 5: Comments & User Management (Admin)**
  - Tambah komentar publik / internal note (`POST /api/tickets/{id}/comments`).
  - Proteksi visibilitas *internal note* (hanya Agent & Admin).
  - Admin users endpoint: `GET /api/users` dan `PATCH /api/users/{id}/role`.
- [x] **Fase 6: Automated Testing Backend (Min. 5 Meaningful Tests)**
  - Test validasi aturan transisi status (sukses & tolak transisi ilegal).
  - Test pembatasan hak akses (requester dilarang baca tiket requester lain).
  - Test otentikasi & proteksi role.
- [x] **Fase 7: Frontend Setup (Vue 3 + Vite + Pinia + Vue Router + Tailwind CSS)**
  - Inisialisasi Vue 3 (`<script setup>` + Composition API).
  - Konfigurasi Pinia Auth Store & Axios HTTP client terpusat (auto token & handle 401).
  - Vue Router Navigation Guards (guest only, auth required, admin only).
  - Halaman Login & Register dengan validasi client-side.
- [x] **Fase 8: Frontend Ticket Pages (List, Filter, Create, Detail, Comments & Status Action)**
  - Halaman Ticket List (tabel, search, filter, badge warna status/prioritas, pagination).
  - Halaman Create Ticket.
  - Halaman Ticket Detail: riwayat komentar, badge catatan internal, form komentar, tombol aksi perubahan status berbasis peran.
- [ ] **Fase 9: Frontend Admin User Management & Full Docker Integration**
  - Halaman Kelola Pengguna khusus Admin (list user & ganti role).
  - Dockerfile frontend & integrasi service frontend ke `docker-compose.yml` (`docker compose up` langsung menjalankan DB, Backend, dan Frontend).
- [ ] **Fase 10: Finalisasi, OpenAPI/Swagger Documentation, & Submission Prep**
  - Dokumentasi API (OpenAPI 3.0 / Postman Collection di folder `docs/`).
  - Pembaruan final `README.md` dan `NOTES.md` (arsitektur & batasan).
  - Hapus file tracker sementara (`docs/SPEC_TRACKER.md`).
  - Final sanity check sebelum siap kumpul.
