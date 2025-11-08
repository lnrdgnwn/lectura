
# Backened For Lectura
Lectura adalah platform web novel berbasis komunitas, di mana pengguna dapat membaca, menulis, dan membagikan karya mereka.
Backend ini dibangun menggunakan Golang (Fiber + GORM) dengan sistem autentikasi berbasis JWT Cookie Encryption, serta mendukung fitur:

- CRUD Novel, Chapter, Genre, Tag, dan Bookmark
- Upload gambar untuk cover dan profil pengguna
- Dokumentasi otomatis menggunakan Swagger
- Role-based access control (User dan Admin)

## 🚀Tech Stack 
Lectura Backend dibangun dengan teknologi modern untuk performa cepat, arsitektur modular, dan pengelolaan data yang efisien.

| Teknologi        | Kegunaan                                              |
| :--------------- | :---------------------------------------------------- |
| **Go (Golang)**  | Bahasa utama pengembangan backend.                    |
| **Fiber**        | Web framework cepat dan ringan.                       |
| **GORM**         | ORM untuk mempermudah interaksi dengan database.      |
| **MySQL**        | Database relasional untuk menyimpan data utama.       |
| **JWT (golang-jwt)** | Autentikasi berbasis token (stateless).          |
| **bcrypt**       | Hashing password dengan aman.                         |
| **dotenv**       | Manajemen environment variabel melalui file `.env`.   |
| **Swagger**      | Dokumentasi API otomatis dan interaktif.              |
| **Postman**      | Dokumentasi API interaktif.              |

## ⚙️ Instalasi & Konfigurasi

### 1️⃣ Clone Repository
```bash
git clone https://github.com/lnrdgnwn/lectura.git
cd lectura-backend
```

2️⃣ Install Dependencies
```bash
go mod tidy
```

### 3️⃣ Siapkan File .env
```bash
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
```

4️⃣ Jalankan Server
```bash
go run main.go
```
✔️Server akan berjalan di : 
```bash
👉 http://localhost:3000
```

## 🧩 Swagger API Documentation
Swagger digunakan untuk menampilkan dokumentasi API interaktif.
Jika sudah diaktifkan, kamu bisa membukanya melalui browser di alamat berikut:
```bash
👉 http://localhost:3000/swagger/index.html
```

## 🖥️ Menjalankan Database (XAMPP)
1. Buka XAMPP Control Panel.
2. Jalankan Apache dan MySQL.
3. Buka phpMyAdmin melalui browser:
```bash
👉 http://localhost/phpmyadmin
```
4. Buat database baru dengan nama:
```bash
👉 your_db_name
```

## 🐳 Menjalankan Docker (Redis)
1. Jalankan container baru dari image Redis.
```bash
👉 docker run --name container-name -p 6379:6379 -d redis:lates
```
Untuk --name container-name itu kasih nama "lectura" atau yang lain.

2. Untuk mengecek daftar container yang berjalan.
```bash
👉 docker ps 
```
3. Menjalankan container yang sudah ada atau container yang lagi stop.
```bash
👉 docker start container-name
```
4. Masuk ke dalam container dan buka Redis CLI, dan mengecek seluruh User yang ada di dalam database redis
```bash
👉 docker exec -it container-name redis-cli
```
5. Mengecek detail dan perubahan yang ada di bagian users
```bash
👉 HGETALL refresh:<userID>
```

## 🪄 Fitur Utama
- 🔐 Autentikasi JWT (login, register, logout, refresh token)
- 🧑‍💻 User Management (update profil, ubah password)
- 📚 Novel Management (CRUD novel & chapter)
- 🏷️ Genre & Tag Management
- 📖 Bookmark & Rating System
- 🧩 Relasi Many-to-Many (Novel–Genre, Novel–Tag)
- 🗂️ Upload Gambar (Cover Novel/Profile Picture)
- ⚙️ Middleware Autentikasi & Role-based Access
- 🧹 Cascade delete otomatis untuk data terkait

## ⚠️ Catatan Penting
- Gunakan Go 1.22+ agar semua fitur berjalan optimal.
- Pastikan XAMPP aktif sebelum menjalankan server.
- Refresh token disimpan per-device agar login multi-perangkat tetap aman.
- Untuk keamanan, selalu gunakan .env yang tidak dipublikasikan ke GitHub.
- Jalankan go mod tidy setiap kali menambahkan package baru.

## 🧠 Deskripsi Singkat
Lectura dikembangkan sebagai platform digital untuk membaca dan menerbitkan novel secara online.
Dengan sistem many-to-many antara novel, genre, dan tag, serta fitur chapter management dan bookmark, aplikasi ini menyediakan pengalaman membaca interaktif bagi pengguna dan fleksibilitas penuh bagi penulis untuk mengelola karya mereka.

## 👨‍🚀 Live API Documentation
Dibawah ini live API documentation dari postman
```bash
👉 https://documenter.getpostman.com/view/40551641/2sB3Wqufcm
```
