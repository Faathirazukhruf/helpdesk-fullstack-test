# Architectural Notes & Technical Decisions

Dokumen ini disusun sebagai ringkasan keputusan teknis, batasan sengaja (*trade-offs*), serta rencana peningkatan sistem helpdesk internal ini (sesuai instruksi pada berkas spesifikasi teknis).

---

## 1. Keputusan Arsitektur (*Architectural Decisions*)

### A. Backend: Standard Library Go 1.22+ (`net/http`) & Layered Architecture
- **Router Standar `http.ServeMux`**: Memanfaatkan fitur *enhanced routing* bawaan Go 1.22+ dengan pencocokan metode HTTP dan *path variables* (misal: `"PATCH /api/tickets/{id}"` dan `r.PathValue("id")`). Keputusan ini meniadakan kebutuhan dependensi eksternal pihak ketiga (seperti Gin atau Fiber), menghasilkan binary yang sangat ramping, cepat di-compile, dan minim celah keamanan (*zero framework lock-in*).
- **Arsitektur Berlapis Bersih (Clean Layered Pattern)**:
  - `model/`: Definisi entitas data, struct tags JSON, dan fungsi murni validasi domain (seperti matriks transisi status).
  - `repository/`: Akses data relasional PostgreSQL dengan raw SQL terparameterisasi (`$1, $2`) menggunakan `database/sql` standar. Menjamin query transparan dan bebas dari SQL Injection.
  - `handler/`: HTTP controller yang bertugas membaca request, memvalidasi input payload, memanggil repository, dan menulis response JSON seragam.
  - `middleware/`: Penanganan autentikasi JWT terpusat dan *Role-Based Access Control* (RBAC) deklaratif (`RequireRole`).

### B. Database: Embedded Migrations & Idempotency
- Menggunakan Go standard library `//go:embed` untuk menyatukan seluruh berkas migrasi SQL (`.up.sql` dan `.down.sql`) ke dalam binary server.
- Database PostgreSQL otomatis bermigrasi dan mengisi *seed data* saat container pertama kali dinyalakan. Setiap versi migrasi dicatat dalam tabel `schema_migrations` dengan transaksi atomik (`BEGIN ... COMMIT`), menjamin eksekusi idempoten dan aman terhadap *concurrent startup*.

### C. Frontend: Vue 3 (Composition API `<script setup>`) + Pinia + Tailwind CSS
- Menggunakan standar modern Vue 3 dengan `<script setup>` dan TypeScript-like reactive primitives (`ref`, `computed`, `watch`).
- **Sentralisasi State & Auth**: Pinia store mengelola token dan profil pengguna dengan persistensi `localStorage`.
- **Axios Interceptor**: Secara otomatis menyisipkan header `Authorization: Bearer <token>` dan menangani respons `401 Unauthorized` secara global (logout otomatis dan redirect ke rute `/login`).
- **Vue Router Navigation Guards**: Memproteksi rute sebelum komponen di-mount berdasarkan meta `requiresAuth`, `guestOnly`, dan `requiresAdmin`.

### D. Deployment: Multi-Stage Docker & Nginx Reverse Proxy
- Container frontend menggunakan *multi-stage build*: build bundle statis dengan Node.js 22 Alpine, kemudian menyajikannya menggunakan Nginx Alpine (~20MB).
- Nginx bertindak sebagai web server SPA (`try_files $uri $uri/ /index.html;` untuk mencegah error 404 pada HTML5 history routing) sekaligus **Reverse Proxy** untuk path `/api/` ke container backend (`http://backend:8080`). Arsitektur ini meniadakan isu CORS (*Cross-Origin Resource Sharing*) di sisi browser karena seluruh request berasal dari origin yang sama (`localhost:3000`).

---

## 2. Hal yang Sengaja Ditinggalkan (*Deliberate Non-Goals & Trade-offs*)

1. **Penggunaan ORM Berat (seperti GORM)**:
   - *Alasan*: Menghindari query tersembunyi (*N+1 problem*), overhead memori, dan komputasi refleksi runtime. Penggunaan SQL murni dengan `database/sql` jauh lebih transparan, mudah dioptimasi dengan `EXPLAIN ANALYZE`, dan mudah diaudit.
2. **Framework UI Monolitik (seperti Vuetify atau Quasar)**:
   - *Alasan*: Framework UI siap pakai sering kali membawa bundle size yang besar dan styling yang sulit dikustomisasi. Memilih Tailwind CSS murni memberikan kontrol desain penuh, responsivitas tinggi, dan ukuran bundle JavaScript akhir yang sangat kecil (<200KB).
3. **Fitur Bonus (WebSocket, Attachment File, SLA Timer)**:
   - *Alasan*: Sesuai instruksi spesifikasi untuk fokus pada ketepatan fungsionalitas inti (validasi alur transisi status, isolasi data requester, keamanan autentikasi, serta stabilitas Docker Compose) agar kode tetap sederhana, eksplisit, dan minim kompleksitas yang tidak diminta.

---

## 3. Perbaikan yang Akan Dilakukan Jika Ada Waktu Lebih (*Future Improvements*)

1. **Pengujian Frontend Otomatis (Vitest & Vue Test Utils)**:
   - Menambahkan unit test komponen Vue (seperti pengujian tombol aksi transisi status pada `TicketDetailView.vue` dan validasi login form).
2. **Pembaruan Real-Time (Server-Sent Events / SSE)**:
   - Mengimplementasikan SSE pada endpoint tiket sehingga agen dan admin mendapatkan notifikasi instan saat ada tiket baru atau status tiket diubah tanpa perlu me-refresh halaman.
3. **Unggah Lampiran Tiket (*File Attachment*)**:
   - Menambahkan dukungan attachment gambar/screenshot error dengan validasi MIME-type ketat, batasan ukuran file (maksimal 5MB), dan penyimpanan berbasis Object Storage (AWS S3 / MinIO) menggunakan *Presigned URL*.
4. **Export Laporan & Dashboard Analitik**:
   - Menambahkan fitur ekspor data tiket ke format CSV/Excel untuk kebutuhan rekapitulasi tim operasional, serta visualisasi grafik waktu penyelesaian tiket rata-rata (*Mean Time to Resolution* / MTTR).
5. **Rate Limiting & Security Hardening**:
   - Menambahkan middleware rate limiting pada endpoint autentikasi (`POST /api/auth/login`) menggunakan token-bucket algorithm untuk mencegah serangan *brute force*, serta header keamanan HTTP standar (`Content-Security-Policy`, `X-Frame-Options`).
