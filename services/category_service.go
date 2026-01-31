package services

import (
	"kasir-api/models"       // Import models untuk struct Category
	"kasir-api/repositories" // Import repositories untuk akses database
)

// CategoryService handles business logic for categories
// Service adalah layer antara Handler dan Repository
type CategoryService struct {
	repo *repositories.CategoryRepository // Pointer ke CategoryRepository
}

// NewCategoryService creates a new CategoryService
// Fungsi ini adalah "constructor" untuk membuat instance CategoryService
func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo} // Return struct dengan repo yang sudah di-inject
}

// GetAll retrieves all categories
// Fungsi ini memanggil repository untuk ambil semua kategori
func (s *CategoryService) GetAll() ([]models.Category, error) {
	// Langsung panggil repository
	// Di sini bisa tambah logic seperti: filter, sorting, dll
	return s.repo.GetAll()
}

// GetByID retrieves a category by ID
// Fungsi ini memanggil repository untuk ambil 1 kategori by ID
func (s *CategoryService) GetByID(id int) (*models.Category, error) {
	// Langsung panggil repository
	return s.repo.GetByID(id)
}

// Create adds a new category
// Fungsi ini memanggil repository untuk tambah kategori baru
func (s *CategoryService) Create(category *models.Category) error {
	// Di sini bisa tambahkan validasi business logic:
	// - Cek apakah nama tidak kosong
	// - Cek apakah kategori dengan nama yang sama sudah ada
	// - dll

	// Panggil repository untuk save ke database
	return s.repo.Create(category)
}

// Update updates an existing category
// Fungsi ini memanggil repository untuk update kategori
func (s *CategoryService) Update(id int, category *models.Category) error {
	// Set ID untuk memastikan update kategori yang benar
	category.ID = id

	// Panggil repository untuk update di database
	return s.repo.Update(category)
}

// Delete removes a category
// Fungsi ini memanggil repository untuk hapus kategori
func (s *CategoryService) Delete(id int) error {
	// Di sini bisa tambah logic seperti:
	// - Cek apakah kategori masih punya produk
	// - Kalau masih ada produk, jangan boleh dihapus
	// - dll

	// Panggil repository untuk delete dari database
	return s.repo.Delete(id)
}
