// ==================== KASIR API ====================
// Aplikasi backend sederhana untuk sistem kasir
// Fitur: CRUD Produk dan CRUD Category
// Port: 8080
// ===================================================

package main // Package utama untuk aplikasi Go

// Import package yang dibutuhkan
import (
	"encoding/json" // Package untuk encode/decode JSON
	"fmt"           // Package untuk print ke console
	"net/http"      // Package untuk membuat HTTP server
	"os"
	// Package untuk akses environment variables
	"strconv" // Package untuk convert string ke integer (dan sebaliknya)
	"strings" // Package untuk manipulasi string (misal: TrimPrefix)
)

// ==================== STRUCT & DATA ====================

// Struct Produk - Blueprint/template untuk data produk
// Struct ini mendefinisikan field-field apa saja yang dimiliki produk
type Produk struct {
	ID    int    `json:"id"`    // ID produk (integer), akan jadi "id" di JSON
	Nama  string `json:"nama"`  // Nama produk (string), akan jadi "nama" di JSON
	Harga int    `json:"harga"` // Harga produk (integer), akan jadi "harga" di JSON
	Stok  int    `json:"stok"`  // Stok produk (integer), akan jadi "stok" di JSON
}

// Data produk awal (slice/array dari Produk)
// Ini adalah data dummy untuk testing, nanti bisa diganti dengan database
var produk = []Produk{
	{ID: 1, Nama: "Indomie goreng", Harga: 3500, Stok: 10},
	{ID: 2, Nama: "Vit 600ml", Harga: 3000, Stok: 40},
	{ID: 3, Nama: "kecap ABC", Harga: 12000, Stok: 20},
}

// Struct Category - Blueprint/template untuk data kategori
// Struct ini mendefinisikan field-field apa saja yang dimiliki kategori
type Category struct {
	ID          int    `json:"id"`          // ID kategori (integer)
	Nama        string `json:"nama"`        // Nama kategori (string)
	Description string `json:"deskription"` // Deskripsi kategori (string)
}

// Data kategori awal (slice/array dari Category)
var categories = []Category{
	{ID: 1, Nama: "Makanan", Description: "Semua produk makanan"},
	{ID: 2, Nama: "Minuman", Description: "Semua produk minuman"},
	{ID: 3, Nama: "Snack", Description: "Semua produk snack"},
}

// ==================== CRUD PRODUK ====================

// GET /api/produk/{id} - Ambil satu produk berdasarkan ID
func getProdukByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/") // Ambil ID dari URL, misal: /api/produk/1 -> "1"
	id, err := strconv.Atoi(idStr)                          // Ubah string "1" menjadi integer 1
	if err != nil {                                         // Kalau gagal convert (misal: /api/produk/abc)
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest) // Kirim error 400
		return
	}

	for _, p := range produk { // Loop semua produk
		if p.ID == id { // Kalau ketemu produk dengan ID yang dicari
			w.Header().Set("Content-Type", "application/json") // Set header response jadi JSON
			json.NewEncoder(w).Encode(p)                       // Kirim data produk dalam format JSON
			return
		}
	}

	http.Error(w, "Produk belum ada", http.StatusNotFound) // Kalau tidak ketemu, kirim error 404
}

// PUT /api/produk/{id} - Update produk berdasarkan ID
func updateProduk(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/") // Ambil ID dari URL
	id, err := strconv.Atoi(idStr)                          // Ubah string ke integer
	if err != nil {
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
		return
	}

	var updateProduk Produk                             // Buat variable untuk menampung data baru dari request
	err = json.NewDecoder(r.Body).Decode(&updateProduk) // Baca data JSON dari request body
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for i := range produk { // Loop semua produk dengan index
		if produk[i].ID == id { // Kalau ketemu produk dengan ID yang dicari
			updateProduk.ID = id     // Pastikan ID tetap sama (tidak berubah)
			produk[i] = updateProduk // Replace produk lama dengan data baru

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateProduk) // Kirim response produk yang sudah diupdate
			return
		}
	}
	http.Error(w, "Produk belum ada", http.StatusNotFound)
}

// DELETE /api/produk/{id} - Hapus produk berdasarkan ID
func deleteProduk(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/") // Ambil ID dari URL
	id, err := strconv.Atoi(idStr)                          // Ubah string ke integer
	if err != nil {
		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
		return
	}

	for i, p := range produk { // Loop semua produk dengan index dan value
		if p.ID == id { // Kalau ketemu produk dengan ID yang dicari
			produk = append(produk[:i], produk[i+1:]...) // Hapus produk: gabungkan slice sebelum dan sesudah index i
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{ // Kirim response sukses
				"message": "sukses delete",
			})

			return
		}
	}

	http.Error(w, "Produk belum ada", http.StatusNotFound)
}

// ==================== CRUD CATEGORY ====================

// GET /api/categories/{id} - Ambil satu kategori berdasarkan ID
func getCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/") // Ambil ID dari URL, misal: /api/categories/1 -> "1"
	id, err := strconv.Atoi(idStr)                              // Ubah string "1" menjadi integer 1
	if err != nil {                                             // Kalau gagal convert (misal: /categories/abc)
		http.Error(w, "Invalid Category ID", http.StatusBadRequest) // Kirim error 400
		return
	}

	for _, c := range categories { // Loop semua kategori
		if c.ID == id { // Kalau ketemu kategori dengan ID yang dicari
			w.Header().Set("Content-Type", "application/json") // Set header response jadi JSON
			json.NewEncoder(w).Encode(c)                       // Kirim data kategori dalam format JSON
			return
		}
	}

	http.Error(w, "Category tidak ditemukan", http.StatusNotFound) // Kalau tidak ketemu, kirim error 404
}

// PUT /api/categories/{id} - Update kategori berdasarkan ID
func updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/") // Ambil ID dari URL
	id, err := strconv.Atoi(idStr)                              // Ubah string ke integer
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	var updatedCategory Category                           // Buat variable untuk menampung data baru dari request
	err = json.NewDecoder(r.Body).Decode(&updatedCategory) // Baca data JSON dari request body
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i := range categories { // Loop semua kategori dengan index
		if categories[i].ID == id { // Kalau ketemu kategori dengan ID yang dicari
			updatedCategory.ID = id         // Pastikan ID tetap sama (tidak berubah)
			categories[i] = updatedCategory // Replace kategori lama dengan data baru
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedCategory) // Kirim response kategori yang sudah diupdate
			return
		}
	}

	http.Error(w, "Category tidak ditemukan", http.StatusNotFound)
}

// DELETE /api/categories/{id} - Hapus kategori berdasarkan ID
func deleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/") // Ambil ID dari URL
	id, err := strconv.Atoi(idStr)                              // Ubah string ke integer
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	for i, c := range categories { // Loop semua kategori dengan index dan value
		if c.ID == id { // Kalau ketemu kategori dengan ID yang dicari
			categories = append(categories[:i], categories[i+1:]...) // Hapus kategori: gabungkan slice sebelum dan sesudah index i
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{ // Kirim response sukses
				"message": "Category berhasil dihapus",
			})
			return
		}
	}

	http.Error(w, "Category tidak ditemukan", http.StatusNotFound)
}

// GET /api/categories - Ambil semua kategori
func getAllCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Set header response jadi JSON
	json.NewEncoder(w).Encode(categories)              // Kirim semua data categories dalam format JSON
}

// POST /api/categories - Tambah kategori baru
func createCategory(w http.ResponseWriter, r *http.Request) {
	var newCategory Category                            // Buat variable untuk menampung data kategori baru
	err := json.NewDecoder(r.Body).Decode(&newCategory) // Baca data JSON dari request body
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newCategory.ID = len(categories) + 1         // Generate ID baru (jumlah kategori + 1)
	categories = append(categories, newCategory) // Tambahkan kategori baru ke slice categories

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)      // Set status code 201 (Created)
	json.NewEncoder(w).Encode(newCategory) // Kirim response kategori yang baru dibuat
}

// ==================== MAIN FUNCTION ====================
// Fungsi main adalah entry point (titik awal) aplikasi
// Semua routing endpoint didefinisikan di sini
// Server akan start dan listen di port 8080

func main() {

	// ==================== ROUTING PRODUK ====================

	// GET localhost:8080/api/produk/{id}
	// PUT localhost:8080/api/produk/{id}
	// DELETE localhost:8080/api/produk/{id}
	http.HandleFunc("/api/produk/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { // Kalau method GET, panggil fungsi getProdukByID
			getProdukByID(w, r)
		} else if r.Method == "PUT" { // Kalau method PUT, panggil fungsi updateProduk
			updateProduk(w, r)
		} else if r.Method == "DELETE" { // Kalau method DELETE, panggil fungsi deleteProduk
			deleteProduk(w, r)
		}

	})

	// GET localhost:8080/api/produk
	// POST localhost:8080/api/produk
	http.HandleFunc("/api/produk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { // Kalau method GET, ambil semua produk
			w.Header().Set("Content-Type", "application/json") // Set header response jadi JSON
			json.NewEncoder(w).Encode(produk)                  // Kirim semua data produk dalam format JSON
		} else if r.Method == "POST" { // Kalau method POST, tambah produk baru
			var produkBaru Produk                              // Buat variable untuk menampung data produk baru
			err := json.NewDecoder(r.Body).Decode(&produkBaru) // Baca data JSON dari request body
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			produkBaru.ID = len(produk) + 1     // Generate ID baru (jumlah produk + 1)
			produk = append(produk, produkBaru) // Tambahkan produk baru ke slice produk

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)     // Set status code 201 (Created)
			json.NewEncoder(w).Encode(produkBaru) // Kirim response produk yang baru dibuat
		}

	})

	// ==================== ROUTING CATEGORY ====================

	// GET localhost:8080/api/categories/{id}
	// PUT localhost:8080/api/categories/{id}
	// DELETE localhost:8080/api/categories/{id}
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { // Kalau method GET, panggil fungsi getCategoryByID
			getCategoryByID(w, r)
		} else if r.Method == "PUT" { // Kalau method PUT, panggil fungsi updateCategory
			updateCategory(w, r)
		} else if r.Method == "DELETE" { // Kalau method DELETE, panggil fungsi deleteCategory
			deleteCategory(w, r)
		}
	})

	// GET localhost:8080/api/categories
	// POST localhost:8080/api/categories
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { // Kalau method GET, panggil fungsi getAllCategories
			getAllCategories(w, r)
		} else if r.Method == "POST" { // Kalau method POST, panggil fungsi createCategory
			createCategory(w, r)
		}
	})

	// ==================== HEALTH CHECK ====================
	// Endpoint untuk cek apakah API masih berjalan atau tidak
	// Biasanya digunakan untuk monitoring atau testing
	// GET localhost:8080/health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// w = ResponseWriter (untuk kirim response ke client)
		// r = Request (berisi data request dari client)

		w.Header().Set("Content-Type", "application/json") // Set header response jadi JSON
		json.NewEncoder(w).Encode(map[string]string{       // Buat dan kirim JSON response
			"status":  "OK",          // Field status dengan value "OK"
			"message": "API Running", // Field message dengan value "API Running"
		})
		// Response JSON yang dikirim: {"status":"OK","message":"API Running"}
	})

	// ==================== START SERVER ====================
	// Get port from environment variable (untuk Railway) atau default 8080 (untuk local)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port untuk development local
	}

	fmt.Println("Server running di port:", port)
	fmt.Println("Tekan Ctrl+C untuk stop server")
	fmt.Println("")
	fmt.Println("Endpoint yang tersedia:")
	fmt.Println("- GET    /health")
	fmt.Println("- GET    /api/produk")
	fmt.Println("- POST   /api/produk")
	fmt.Println("- GET    /api/produk/{id}")
	fmt.Println("- PUT    /api/produk/{id}")
	fmt.Println("- DELETE /api/produk/{id}")
	fmt.Println("- GET    /api/categories")
	fmt.Println("- POST   /api/categories")
	fmt.Println("- GET    /api/categories/{id}")
	fmt.Println("- PUT    /api/categories/{id}")
	fmt.Println("- DELETE /api/categories/{id}")

	err := http.ListenAndServe(":"+port, nil) // Start server di port dari environment atau 8080
	if err != nil {                           // Kalau ada error saat start server
		fmt.Println("Gagal running server:", err) // Tampilkan error detail
	}
}
