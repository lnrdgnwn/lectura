# 📚 Lectura API

Lectura API adalah backend untuk platform baca web novel dengan fitur:

- ✍️ Author bisa membuat & mengelola novel
- 📖 Reader bisa membaca chapter yang sudah rilis
- 🔖 Bookmark novel favorit
- 🏷️ Genre & Tag untuk pengelompokan konten
- 🛡️ Role-based Access (User & Admin)
- 📜 Dokumentasi API via Swagger

Semua respons sudah dikemas dalam format JSON yang konsisten:
- `success`
- `message`
- `data`
- `meta` (untuk list / pagination)

---

## 🧱 Tech Stack

**Backend**

- ⚙️ Go (Golang)
- 🚀 Fiber v2 — Web framework cepat & minimalis
- 🗄️ GORM — ORM untuk komunikasi dengan database
- 🐬 MySQL / MariaDB — Database utama

**Auth & Security**

- 🔐 JWT (`github.com/golang-jwt/jwt/v4`)
- 🍪 Encrypted Cookie `access_token` (pakai util `utils.Decrypt`)
- 🔑 `bcrypt` untuk hashing password

**Struktur Project (contoh)**

- `main.go`
- `database/`
- `models/`
- `controllers/`
- `utils/`
- `docs/` (Swagger, jika pakai swag)

---

## ⚙️ Environment Variables

Buat file `.env` di root:

```env
DB_HOST=your_db_host
DB_PORT=your_db_port
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
JWT_SECRET=your_jwt_secret
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
ENCRYPT_KEY=dw9/+WzKosiNV+1XKgSVUsJ0GRuTGw/z4kBbmp/2UNs=
BASE_URL=http://127.0.0.1:3000

---

## 🐬 Setup Database