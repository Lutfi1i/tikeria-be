# Tikeria ERP (Frappe Custom App)

Modul ERP, Admin Panel, dan Source of Truth untuk platform **Tikeria** (Platform Tiketing Multiplatform untuk Event Sekolah, Kampus, dan Pensi).

Repo ini dirancang sebagai **Frappe Custom App** yang siap dipasang ke dalam **Frappe Bench v14 / v15**.

---

## 1. Arsitektur Sistem & Peran ERP

```text
┌────────────────────────────────────────────────────────┐
│                   FRONTEND CLIENTS                     │
│  - Buyer Web (React / PWA)                             │
│  - Buyer Mobile (Flutter)                              │
│  - Organizer Web (React + /scan QR)                   │
└──────────────────────────┬─────────────────────────────┘
                           │ HTTP / JSON (JWT)
                           ▼
┌────────────────────────────────────────────────────────┐
│             BACKEND API GATEWAY (Golang)               │
│  Repo: https://github.com/Lutfi1i/tikeria-be           │
│  - Validasi bisnis, checkout, hold stok                │
│  - Enforcement limit paket Organizer (Basic / Pro)     │
│  - Simulasi pembayaran & verifikasi QR check-in        │
└──────────────────────────┬─────────────────────────────┘
                           │ REST API (/api/resource/...)
                           │ Header: token <api_key>:<api_secret>
                           ▼
┌────────────────────────────────────────────────────────┐
│                  TIKERIA ERP (Frappe)                  │
│               [ REPOSITORI INI ]                       │
│  - Admin Desk UI (Dashboard, Moderasi, Approval)       │
│  - Master Data & 17 DocType Schema                     │
│  - Audit Log & Scheduled Jobs (Cron hold expiry)       │
└──────────────────────────┬─────────────────────────────┘
                           │ SQL
                           ▼
┌────────────────────────────────────────────────────────┐
│                   DATABASE (MariaDB)                   │
│             Pusat Penyimpanan Data Utama               │
└────────────────────────────────────────────────────────┘
```

---

## 2. Bentuk Frappe ERP: Bagaimana Cara Kerjanya?

Bagi tim developer yang baru mengenal Frappe, berikut adalah bentuk dan wujud dari ERP ini:

### A. Dalam Repositori (Codebase Ini)
Frappe mengadopsi pendekatan **Schema-as-Code** berbasis metadata:
* **Folder `tikeria/tikeria/doctype/`**: Berisi 17 DocType. Masing-masing terdiri dari file skema `.json` (field, tipe data, validasi, relasi, permission) dan controller `.py` (business logic, masking NIK, cron job).
* **Folder `tikeria/fixtures/`**: Seed data otomatis untuk master data awal (Role, Workflow status, Event Type default, dan Paket Organizer).
* **`hooks.py`**: Konfigurasi integrasi lifecycle, permission fixtures, dan scheduled background jobs.

### B. Saat Dijalankan di Browser (Frappe Desk)
Ketika app ini di-install ke Frappe Bench, Frappe otomatis menyediakan antarmuka web admin lengkap tanpa perlu coding frontend admin:
1. **Admin Workspace / Desk**: Dashboard modern di browser (`http://localhost:8000/app`) untuk tim Admin Tikeria.
2. **Form & List View Otomatis**: Setiap DocType (Event, Order, Ticket, Scan Log) memiliki tabel, fitur pencarian, filter, dan form edit bawaan.
3. **Workflow Board (Kanban)**: Memoderasi pengajuan event dari panitia (`Draft` ➔ `Waiting Approval` ➔ `Published` / `Reject`).
4. **Auto-generated REST API**: Frappe otomatis menyediakan endpoint REST API untuk setiap DocType di `/api/resource/<DocType>` yang langsung dapat dikonsumsi oleh backend Golang.

---

## 3. Daftar 17 DocType & Naming Convention

| No | Nama DocType | Jenis | Naming Rule | Deskripsi & Fungsi |
|:---:|---|:---:|:---:|---|
| 1 | **Tikeria Event Type** | Master | `field:event_type_name` | Master tipe event: Pensi, Konser, Festival Musik, Seminar, dll |
| 2 | **Tikeria Organizer Package** | Master | `field:package_code` | Master limit paket panitia: `basic` (1 kategori, max 1k tiket) & `pro` |
| 3 | **Tikeria Organizer Profile** | Master | `ORG-{name}-{#####}` | Profil panitia pensi/kampus, institusi, WhatsApp, logo branding |
| 4 | **Tikeria Buyer Profile** | Master | `BYR-{username}` | Profil pembeli terdaftar & status kelengkapan data diri |
| 5 | **Tikeria Auth Session** | Security | Hash UUID | Tracking session login per perangkat & hash refresh token |
| 6 | **Tikeria Platform Setting** | Single | *Single record* | Konfigurasi biaya layanan flat (*service fee*) per transaksi |
| 7 | **Tikeria Event** | Transaksi | `EVT-.YYYY.-.#####` | Entitas utama event, PIC, banner, venue, tanggal, dan status approval |
| 8 | **Tikeria Ticket Category** | Transaksi | `CAT-.YYYY.-.#####` | Kategori tiket dalam event (Harga, kuota stok, layout peta kursi) |
| 9 | **Tikeria Ticket Hold** | Transaksi | `HOLD-.YYYY.-.#####` | Temporary hold stok tiket selama 15 menit saat proses checkout |
| 10 | **Tikeria Order** | Transaksi | `ORD-.YYYY.-.#####` | Transaksi pesanan, nomor invoice `INV-`, subtotal, biaya layanan, total |
| 11 | **Tikeria Order Item** | Child Table | Auto | Detail snapshot harga & kategori tiket per order |
| 12 | **Tikeria Participant Detail** | Transaksi | `PART-.YYYY.-.#####` | 1 data peserta per order. Otomatis masking NIK (`3174********1234`) |
| 13 | **Tikeria Payment** | Transaksi | `PAY-.YYYY.-.#####` | Status pembayaran (Waiting, Success, Failed, Expired) & simulasi |
| 14 | **Tikeria Ticket** | Transaksi | `TKT-.YYYY.-.#####` | E-ticket penerbitan, hash token QR acak, status Issued ➔ Used ➔ Expired |
| 15 | **Tikeria Scan Log** | Audit | `SCAN-.YYYY.-.#####` | Riwayat audit trail setiap pemindaian tiket gate di hari H (append-only) |
| 16 | **Tikeria WebView Token** | Security | Hash UUID | Token sekali pakai (60s) untuk handoff login Flutter ke web checkout |
| 17 | **Tikeria Scanner Assignment**| Support | `SCN-{org}-{#####}` | Akun staf scanner gate tambahan untuk Organizer paket Pro |

---

## 4. Alur Approval Workflow Event

Event yang dibuat oleh panitia pensi melewati siklus moderasi Frappe:

```text
[ Organizer ]                     [ Tikeria Admin ]
     │                                   │
   Draft ──(Submit Approval)──► Waiting Approval
     ▲                                   │
     │                           ┌───────┴───────┐
     │                           ▼               ▼
(Revisi & Resubmit)           Published       Reject
     │                           │               │
     └───────────────────────────┴────────(Isi Reject Reason)
```

* **Draft**: Panitia bebas mengisi/mengubah deskripsi dan kategori tiket.
* **Waiting Approval**: Event terkunci dari pengeditan, masuk ke antrean review Admin di Frappe Desk.
* **Published**: Disetujui Admin, tanggal publish tercatat, tampil di aplikasi pembeli.
* **Reject**: Ditolak Admin dengan alasan penolakan wajib (`reject_reason`). Panitia dapat merevisi dan mengajukan kembali.

---

## 5. Cara Instalasi & Menjalankan (Untuk Tim Developer)

Frappe berjalan di atas Linux/WSL2 atau Docker.

### Prasyarat
* Linux (Ubuntu 22.04+) atau Windows dengan WSL2 Ubuntu
* Python 3.10+, Node.js 18+, MariaDB 10.6+, Redis Server
* Frappe Bench CLI (`pip install frappe-bench`)

### Langkah Instalasi
```bash
# 1. Masuk ke direktori bench Anda
cd ~/frappe-bench

# 2. Ambil repository aplikasi ini
bench get-app https://github.com/Lutfi1i/tikeria-erp.git

# 3. Install aplikasi ke site tujuan
bench --site tikeria.local install-app tikeria

# 4. Migrasi schema dan load fixtures (Roles, Packages, Workflows)
bench --site tikeria.local migrate

# 5. Jalankan server ERP
bench start
```

Setelah berjalan:
* Buka browser: `http://localhost:8000`
* Login sebagai `Administrator`
* Buka menu **Tikeria** di Desk untuk mengelola Event, Moderasi, Paket, dan Laporan.

---

## 6. Integrasi dengan Backend Golang

Backend Golang (`tikeria-be`) berkomunikasi ke Frappe menggunakan API Key & Secret:
1. Di Frappe Desk, buat User baru dengan role `Tikeria Admin` atau generate API Key di profil Administrator.
2. Simpan key di `.env` Golang:
   ```env
   FRAPPE_BASE_URL=http://localhost:8000
   FRAPPE_API_KEY=your_api_key
   FRAPPE_API_SECRET=your_api_secret
   ```
3. Golang melakukan panggilan HTTP standard:
   ```http
   GET /api/resource/Tikeria Event?filters=[["status","=","Published"]]
   Authorization: token <api_key>:<api_secret>
   ```

---

## Lisensi
MIT License - Tim Pengembang Tikeria
