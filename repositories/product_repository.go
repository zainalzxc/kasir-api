package repositories

import (
	"database/sql"     // Package standard Go untuk database SQL
	"kasir-api/models" // Import models untuk struct Product
)

// ProductRepository handles database operations for products
// Struct ini menyimpan koneksi database
type ProductRepository struct {
	db *sql.DB // Pointer ke database connection
}

// NewProductRepository creates a new ProductRepository
// Fungsi ini adalah "constructor" untuk membuat instance ProductRepository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db} // Return struct dengan db yang sudah di-inject
}

// GetAll retrieves all products from database
// Fungsi ini mengambil semua produk dari table products
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	// SQL query dengan LEFT JOIN ke table categories
	// LEFT JOIN = ambil semua products, meskipun tidak punya category
	query := `
		SELECT 
			p.id, 
			p.nama, 
			p.harga, 
			p.stok, 
			p.category_id,
			c.id as category_id_full,
			c.nama as category_name,
			c.description as category_description
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
	`

	// Execute query dan dapatkan rows (banyak baris)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err // Kalau error, return nil dan error
	}
	defer rows.Close() // Pastikan rows di-close setelah selesai (penting!)

	// Buat slice kosong untuk menampung products
	var products []models.Product

	// Loop semua rows yang didapat dari database
	for rows.Next() {
		var product models.Product      // Buat variable product untuk setiap row
		var categoryID sql.NullInt64    // Untuk handle NULL dari LEFT JOIN
		var categoryName sql.NullString // Untuk handle NULL dari LEFT JOIN
		var categoryDesc sql.NullString // Untuk handle NULL dari LEFT JOIN

		// Scan data dari row ke struct product
		// Urutan harus sama dengan SELECT
		err := rows.Scan(
			&product.ID,
			&product.Nama,
			&product.Harga,
			&product.Stok,
			&product.CategoryID,
			&categoryID,
			&categoryName,
			&categoryDesc,
		)
		if err != nil {
			return nil, err // Kalau scan error, return error
		}

		// Jika ada category, populate Category struct dengan semua field
		if categoryName.Valid && categoryID.Valid {
			product.Category = &models.Category{
				ID:          int(categoryID.Int64),
				Nama:        categoryName.String,
				Description: categoryDesc.String,
			}
		}

		// Tambahkan product ke slice products
		products = append(products, product)
	}

	return products, nil // Return slice products dan nil (no error)
}

// GetByID retrieves a product by ID
// Fungsi ini mengambil 1 produk berdasarkan ID
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	// SQL query dengan LEFT JOIN ke categories untuk mendapatkan category lengkap
	// $1 akan diganti dengan value id saat execute
	query := `
		SELECT 
			p.id, 
			p.nama, 
			p.harga, 
			p.stok, 
			p.category_id,
			c.id as category_id_full,
			c.nama as category_name,
			c.description as category_description
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`

	// QueryRow untuk query yang return 1 row saja
	row := r.db.QueryRow(query, id) // id akan replace $1

	var product models.Product      // Buat variable untuk menampung hasil
	var categoryID sql.NullInt64    // Untuk handle NULL dari LEFT JOIN
	var categoryName sql.NullString // Untuk handle NULL dari LEFT JOIN
	var categoryDesc sql.NullString // Untuk handle NULL dari LEFT JOIN

	// Scan hasil query ke struct product
	err := row.Scan(
		&product.ID,
		&product.Nama,
		&product.Harga,
		&product.Stok,
		&product.CategoryID,
		&categoryID,
		&categoryName,
		&categoryDesc,
	)
	if err != nil {
		return nil, err // Kalau tidak ketemu atau error, return nil
	}

	// Jika ada category, populate Category struct dengan semua field
	if categoryName.Valid && categoryID.Valid {
		product.Category = &models.Category{
			ID:          int(categoryID.Int64),
			Nama:        categoryName.String,
			Description: categoryDesc.String,
		}
	}

	return &product, nil // Return pointer ke product
}

// Create adds a new product to database or updates stock if product exists
// Fungsi ini menambahkan produk baru ATAU menambah stok jika produk dengan nama sama sudah ada
func (r *ProductRepository) Create(product *models.Product) error {
	// SQL query dengan UPSERT logic (INSERT ... ON CONFLICT ... DO UPDATE)
	// Jika produk dengan nama yang sama sudah ada, maka:
	// - Stok akan ditambahkan (stok lama + stok baru)
	// - Harga akan diupdate dengan harga terbaru
	// - CategoryID akan diupdate jika diberikan
	// Jika belum ada, akan insert produk baru
	query := `
		INSERT INTO products (nama, harga, stok, category_id) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (nama) 
		DO UPDATE SET 
			harga = EXCLUDED.harga,
			stok = products.stok + EXCLUDED.stok,
			category_id = EXCLUDED.category_id
		RETURNING id, stok
	`

	// Execute query dan scan ID + stok terbaru yang di-return
	// $1 = product.Nama, $2 = product.Harga, $3 = product.Stok, $4 = product.CategoryID
	err := r.db.QueryRow(query, product.Nama, product.Harga, product.Stok, product.CategoryID).Scan(&product.ID, &product.Stok)

	return err // Return error (nil kalau sukses)
}

// Update updates an existing product
// Fungsi ini mengupdate produk yang sudah ada
func (r *ProductRepository) Update(product *models.Product) error {
	// SQL query untuk UPDATE
	// SET untuk set nilai baru (termasuk category_id)
	// WHERE untuk kondisi (update produk dengan id tertentu)
	query := "UPDATE products SET nama = $1, harga = $2, stok = $3, category_id = $4 WHERE id = $5"

	// Execute query
	// $1 = nama, $2 = harga, $3 = stok, $4 = category_id, $5 = id
	_, err := r.db.Exec(query, product.Nama, product.Harga, product.Stok, product.CategoryID, product.ID)

	return err // Return error (nil kalau sukses)
}

// Delete removes a product from database
// Fungsi ini menghapus produk dari database
func (r *ProductRepository) Delete(id int) error {
	// SQL query untuk DELETE
	// WHERE untuk kondisi (hapus produk dengan id tertentu)
	query := "DELETE FROM products WHERE id = $1"

	// Execute query
	// $1 = id
	_, err := r.db.Exec(query, id)

	return err // Return error (nil kalau sukses)
}
