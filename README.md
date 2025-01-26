Studi Kasus Penyewaan Mobil

- Package dan Import
fiber/v2: Framework untuk membuat REST API di Go.
cors: Middleware untuk mengatur Cross-Origin Resource Sharing (CORS).
jwt/v3 dan jwt/v4: Digunakan untuk membuat dan memverifikasi token JWT.
godotenv: Memuat variabel lingkungan dari file .env.
gorm: ORM untuk menghubungkan aplikasi dengan database MySQL.

- Struktur Database (Models):
User: Menyimpan data pengguna (admin dan customer).
Car: Menyimpan data mobil yang dapat disewa.
Rental: Menyimpan informasi penyewaan, seperti pengguna, mobil, tanggal mulai/akhir, harga total, dan status.
Payment: Menyimpan data pembayaran yang terkait dengan penyewaan.

- Middleware
CORS: Mengizinkan permintaan dari domain tertentu.
JWT Authentication: Melindungi rute API agar hanya pengguna yang memiliki token valid dapat mengakses.

- Rute API
1. Users:
CRUD untuk data pengguna.
Endpoint untuk registrasi (/register) dan login (/login).
Endpoint /users/me untuk mendapatkan data pengguna yang sedang login (mengambil user ID dari JWT).
2. Cars:
CRUD untuk data mobil yang disewa.
3. Rentals:
CRUD untuk data penyewaan.
Menghitung harga sewa berdasarkan lama waktu sewa.
Otomatis membuat entri pembayaran terkait dengan status "unpaid".
4. Payments:
CRUD untuk data pembayaran.
Mengubah status penyewaan menjadi "paid" setelah pembayaran berhasil.
Contoh Endpoint:
GET /users: Mendapatkan semua pengguna.
POST /rentals: Membuat data penyewaan baru dan menghitung total harga berdasarkan lama waktu.
