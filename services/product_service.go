package services

import (
	"kasir-api/models"       // Import models untuk struct Product
	"kasir-api/repositories" // Import repositories untuk akses database
)

// ProductService handles business logic for products
// Service adalah layer antara Handler dan Repository
// Di sini kita bisa tambahkan validasi, business rules, dll
type ProductService struct {
	repo *repositories.ProductRepository // Pointer ke ProductRepository
}

// NewProductService creates a new ProductService
// Fungsi ini adalah "constructor" untuk membuat instance ProductService
func NewProductService(repo *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo} // Return struct dengan repo yang sudah di-inject
}

// GetAll retrieves all products
// Fungsi ini memanggil repository untuk ambil semua produk
func (s *ProductService) GetAll() ([]models.Product, error) {
	// Langsung panggil repository
	// Di sini bisa tambah logic seperti: filter, sorting, dll
	return s.repo.GetAll()
}

// GetByID retrieves a product by ID
// Fungsi ini memanggil repository untuk ambil 1 produk by ID
func (s *ProductService) GetByID(id int) (*models.Product, error) {
	// Langsung panggil repository
	// Di sini bisa tambah logic seperti: cek permission, logging, dll
	return s.repo.GetByID(id)
}

// Create adds a new product
// Fungsi ini memanggil repository untuk tambah produk baru
func (s *ProductService) Create(product *models.Product) error {
	// Di sini bisa tambahkan validasi business logic:
	// - Cek apakah harga > 0
	// - Cek apakah stok >= 0
	// - Cek apakah nama tidak kosong
	// - dll

	// Contoh validasi sederhana (opsional):
	// if product.Harga <= 0 {
	//     return errors.New("harga harus lebih dari 0")
	// }

	// Panggil repository untuk save ke database
	return s.repo.Create(product)
}

// Update updates an existing product
// Fungsi ini memanggil repository untuk update produk
func (s *ProductService) Update(id int, product *models.Product) error {
	// Set ID untuk memastikan update produk yang benar
	product.ID = id

	// Di sini bisa tambah validasi seperti di Create

	// Panggil repository untuk update di database
	return s.repo.Update(product)
}

// Delete removes a product
// Fungsi ini memanggil repository untuk hapus produk
func (s *ProductService) Delete(id int) error {
	// Di sini bisa tambah logic seperti:
	// - Cek apakah produk sedang dipakai di transaksi
	// - Soft delete (update status jadi "deleted" instead of hapus permanent)
	// - Logging
	// - dll

	// Panggil repository untuk delete dari database
	return s.repo.Delete(id)
}
