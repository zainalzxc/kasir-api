package repositories

import (
	"database/sql"     // Package standard Go untuk database SQL
	"kasir-api/models" // Import models untuk struct Category
)

// CategoryRepository handles database operations for categories
// Struct ini menyimpan koneksi database
type CategoryRepository struct {
	db *sql.DB // Pointer ke database connection
}

// NewCategoryRepository creates a new CategoryRepository
// Fungsi ini adalah "constructor" untuk membuat instance CategoryRepository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db} // Return struct dengan db yang sudah di-inject
}

// GetAll retrieves all categories from database
// Fungsi ini mengambil semua kategori dari table categories
func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	// SQL query untuk select semua kolom dari table categories
	query := "SELECT id, nama, description FROM categories"

	// Execute query dan dapatkan rows (banyak baris)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err // Kalau error, return nil dan error
	}
	defer rows.Close() // Pastikan rows di-close setelah selesai

	// Buat slice kosong untuk menampung categories
	var categories []models.Category

	// Loop semua rows yang didapat dari database
	for rows.Next() {
		var category models.Category // Buat variable category untuk setiap row

		// Scan data dari row ke struct category
		// Urutan harus sama dengan SELECT: id, nama, description
		err := rows.Scan(&category.ID, &category.Nama, &category.Description)
		if err != nil {
			return nil, err // Kalau scan error, return error
		}

		// Tambahkan category ke slice categories
		categories = append(categories, category)
	}

	return categories, nil // Return slice categories dan nil (no error)
}

// GetByID retrieves a category by ID
// Fungsi ini mengambil 1 kategori berdasarkan ID
func (r *CategoryRepository) GetByID(id int) (*models.Category, error) {
	// SQL query dengan placeholder $1 (untuk parameter id)
	query := "SELECT id, nama, description FROM categories WHERE id = $1"

	// QueryRow untuk query yang return 1 row saja
	row := r.db.QueryRow(query, id) // id akan replace $1

	var category models.Category // Buat variable untuk menampung hasil

	// Scan hasil query ke struct category
	err := row.Scan(&category.ID, &category.Nama, &category.Description)
	if err != nil {
		return nil, err // Kalau tidak ketemu atau error, return nil
	}

	return &category, nil // Return pointer ke category
}

// Create adds a new category to database
// Fungsi ini menambahkan kategori baru ke database
func (r *CategoryRepository) Create(category *models.Category) error {
	// SQL query untuk INSERT
	// RETURNING id = return ID yang baru dibuat (auto-increment)
	query := "INSERT INTO categories (nama, description) VALUES ($1, $2) RETURNING id"

	// Execute query dan langsung scan ID yang di-return
	// $1 = category.Nama, $2 = category.Description
	err := r.db.QueryRow(query, category.Nama, category.Description).Scan(&category.ID)

	return err // Return error (nil kalau sukses)
}

// Update updates an existing category
// Fungsi ini mengupdate kategori yang sudah ada
func (r *CategoryRepository) Update(category *models.Category) error {
	// SQL query untuk UPDATE
	// SET untuk set nilai baru
	// WHERE untuk kondisi (update kategori dengan id tertentu)
	query := "UPDATE categories SET nama = $1, description = $2 WHERE id = $3"

	// Execute query
	// $1 = nama, $2 = description, $3 = id
	_, err := r.db.Exec(query, category.Nama, category.Description, category.ID)

	return err // Return error (nil kalau sukses)
}

// Delete removes a category from database
// Fungsi ini menghapus kategori dari database
func (r *CategoryRepository) Delete(id int) error {
	// SQL query untuk DELETE
	// WHERE untuk kondisi (hapus kategori dengan id tertentu)
	query := "DELETE FROM categories WHERE id = $1"

	// Execute query
	// $1 = id
	_, err := r.db.Exec(query, id)

	return err // Return error (nil kalau sukses)
}
