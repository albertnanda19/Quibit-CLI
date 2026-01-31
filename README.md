# Quibit CLI

Quibit adalah CLI untuk **menghasilkan ide proyek portfolio** yang terstruktur, lengkap, dan siap dieksekusi—dengan dukungan **multi AI provider + fallback otomatis**.

Repo ini: [Quibit-CLI](https://github.com/albertnanda19/Quibit-CLI)

**Author**: Albert Mangiri

## Kenapa Quibit?

- **Ide proyek “portfolio-ready”**: output terstruktur (bukan sekadar ide singkat).
- **Prompt Contract ketat (JSON strict)**: input/output AI dipaksa format JSON—mudah diparsing, disimpan, dan diproses ulang.
- **Multi AI provider (transparent fallback)**:
  - Primary: **Gemini**
  - Fallback: **Hugging Face Router** (OpenAI-compatible API)
- **Stabil saat rate limit / error**: ketika Gemini rate-limited/timeout/error, sistem otomatis fallback tanpa mengubah UX.
- **Anti-duplikasi ide**:
  - DNA hash + similarity scoring untuk mendeteksi ide terlalu mirip dan auto-regenerate.
- **Persistence**: semua hasil disimpan di Postgres:
  - `projects` (ringkasan + raw JSON + metadata AI)
  - `project_features` (fitur/benefit/learning outcome, dll)
  - `project_meta` (target users, tech stack, raw JSON)
  - `project_evolutions` (history evolusi yang di-accept)

## Fitur Utama (Output)

Saat generate project baru, Quibit menampilkan (contoh):
- Title/Tagline, Summary, Detailed Explanation
- Problem Statement + Why It Matters + Current Gaps
- Target Users + Use Cases
- Value Proposition
- MVP (goal + must-have + nice-to-have + out-of-scope)
- Recommended Tech Stack + Justification
- Complexity + Estimated Duration + Assumptions
- Future Extensions + Learning Outcomes

## Prasyarat

- **Go**: `go 1.25.5` (lihat `go.mod`)
- **PostgreSQL** (karena penyimpanan menggunakan Postgres via `DATABASE_URL`)
- API key:
  - `GEMINI_API_KEY` (wajib untuk primary provider)
  - `HF_TOKEN` (wajib agar fallback Hugging Face bisa dipakai)

## Instalasi & Menjalankan

### 1) Clone repository

**SSH**

```bash
git clone git@github.com:albertnanda19/Quibit-CLI.git
```

**HTTPS**

```bash
git clone https://github.com/albertnanda19/Quibit-CLI.git
```

Lalu masuk folder:

```bash
cd Quibit-CLI
```

### 2) Install dependencies Go

```bash
go mod download
```

### 3) Setup environment variables

Quibit membaca env dari environment atau `.env` (opsional).

Buat `.env`:

```bash
cp .env.example .env
```

Isi minimal:

```bash
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable
GEMINI_API_KEY=YOUR_GEMINI_API_KEY
HF_TOKEN=YOUR_HUGGINGFACE_TOKEN
```

### 4) Jalankan database migration

```bash
go run . --migrate
```

### 5) Jalankan CLI

```bash
go run .
```

Atau build binary:

```bash
go build -o quibit .
./quibit
```

## Cara Pakai

Jalankan:

```bash
go run . generate
```

### Menu utama

- **Generate New Project**: generate ide baru
- **Continue Existing Project**: pilih project lama lalu generate evolusi (accept akan tersimpan)
- **View Saved Projects**: lihat project yang pernah disimpan + list evolusi yang pernah di-accept
- **Exit**: satu-satunya cara keluar dari CLI

### Input yang tersedia (Generate New Project)

Quibit akan menanyakan beberapa input (sebagian bisa Custom):
- Application Type (web/cli/mobile/desktop/ml/backend-api, dll)
- Project Category (Optional) — LMS/ERP/CRM/SCM, dsb (bisa skip)
- Complexity
- Technology Stack (framework/bahasa)
- Database (Custom / No database / PostgreSQL/MySQL/SQLite/MongoDB/Redis)
- Project Goal
- Estimated Timeframe

### Regenerate

Setelah output muncul, tersedia opsi:
- **Accept**: simpan ke DB
- **Regenerate**: generate ulang
- **Regenerate (higher complexity)**: generate ulang dengan complexity dinaikkan (beginner→intermediate→advanced)
- **Back**: kembali ke menu awal

## AI Providers

### Primary: Gemini
- Env: `GEMINI_API_KEY`

### Fallback: Hugging Face Router (OpenAI-compatible)
- Base URL: `https://router.huggingface.co/v1`
- Model default: `moonshotai/Kimi-K2-Instruct-0905`
- Env: `HF_TOKEN`

## Troubleshooting

### `DATABASE_URL is required`
- Pastikan `DATABASE_URL` ter-set di environment atau `.env`

### `GEMINI_API_KEY is required` / `HF_TOKEN is required`
- Pastikan env sudah terisi. Quibit butuh `HF_TOKEN` agar fallback bisa bekerja.

### Migrasi / schema berubah
- Jalankan ulang:

```bash
go run . --migrate
```

## Lisensi

Belum ditentukan.