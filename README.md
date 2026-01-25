# Kasir API

Aplikasi backend sederhana untuk sistem kasir menggunakan Golang.

## 📋 Deskripsi

Kasir API adalah REST API sederhana yang menyediakan fitur CRUD (Create, Read, Update, Delete) untuk manajemen produk dan kategori dalam sistem kasir.

## 🚀 Fitur

- ✅ CRUD Produk (Create, Read, Update, Delete)
- ✅ CRUD Kategori (Create, Read, Update, Delete)
- ✅ Health Check Endpoint
- ✅ Response dalam format JSON
- ✅ Dokumentasi lengkap dengan komentar

## 🛠️ Teknologi

- **Bahasa**: Go (Golang) 1.21
- **Framework**: Standard Library (`net/http`)
- **Format Data**: JSON

## 📦 Instalasi

### Prerequisites

- Go 1.21 atau lebih baru
- Git

### Langkah Instalasi

1. Clone repository ini:
```bash
git clone https://github.com/username/kasir-api.git
cd kasir-api
```

2. Jalankan aplikasi:
```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## 📚 API Endpoints

### Health Check
- **GET** `/health` - Cek status API

### Produk
- **GET** `/api/produk` - Ambil semua produk
- **POST** `/api/produk` - Tambah produk baru
- **GET** `/api/produk/{id}` - Ambil produk berdasarkan ID
- **PUT** `/api/produk/{id}` - Update produk
- **DELETE** `/api/produk/{id}` - Hapus produk

### Kategori
- **GET** `/api/categories` - Ambil semua kategori
- **POST** `/api/categories` - Tambah kategori baru
- **GET** `/api/categories/{id}` - Ambil kategori berdasarkan ID
- **PUT** `/api/categories/{id}` - Update kategori
- **DELETE** `/api/categories/{id}` - Hapus kategori

## 📖 Contoh Penggunaan

### Ambil Semua Produk
```bash
curl http://localhost:8080/api/produk
```

**Response:**
```json
[
  {
    "id": 1,
    "nama": "Indomie goreng",
    "harga": 3500,
    "stok": 10
  },
  {
    "id": 2,
    "nama": "Vit 600ml",
    "harga": 3000,
    "stok": 40
  }
]
```

### Tambah Produk Baru
```bash
curl -X POST http://localhost:8080/api/produk \
  -H "Content-Type: application/json" \
  -d '{
    "nama": "Teh Botol",
    "harga": 4000,
    "stok": 25
  }'
```

### Ambil Semua Kategori
```bash
curl http://localhost:8080/api/categories
```

**Response:**
```json
[
  {
    "id": 1,
    "nama": "Makanan",
    "deskription": "Semua produk makanan"
  },
  {
    "id": 2,
    "nama": "Minuman",
    "deskription": "Semua produk minuman"
  }
]
```

## 📁 Struktur Project

```
kasir-api/
├── main.go          # File utama aplikasi
├── go.mod           # Go module file
├── .gitignore       # Git ignore file
└── README.md        # Dokumentasi project
```

## 🔧 Pengembangan

### Menjalankan Server
```bash
go run main.go
```

### Build Aplikasi
```bash
go build -o kasir-api
```

Executable akan dibuat dengan nama `kasir-api` (atau `kasir-api.exe` di Windows)

## 📝 Catatan

- Data saat ini disimpan di memory (in-memory), sehingga akan hilang saat server di-restart
- Untuk production, disarankan menggunakan database (PostgreSQL, MySQL, MongoDB, dll)
- Port default: 8080

## 👨‍💻 Author

Dibuat sebagai tugas bootcamp Golang

## 📄 License

MIT License
