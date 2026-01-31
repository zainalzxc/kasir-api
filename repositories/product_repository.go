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
	// SQL query untuk select semua kolom dari table products
	query := "SELECT id, nama, harga, stok FROM products"

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
		var product models.Product // Buat variable product untuk setiap row

		// Scan data dari row ke struct product
		// Urutan harus sama dengan SELECT: id, nama, harga, stok
		err := rows.Scan(&product.ID, &product.Nama, &product.Harga, &product.Stok)
		if err != nil {
			return nil, err // Kalau scan error, return error
		}

		// Tambahkan product ke slice products
		products = append(products, product)
	}

	return products, nil // Return slice products dan nil (no error)
}

// GetByID retrieves a product by ID
// Fungsi ini mengambil 1 produk berdasarkan ID
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	// SQL query dengan placeholder $1 (untuk parameter id)
	// $1 akan diganti dengan value id saat execute
	query := "SELECT id, nama, harga, stok FROM products WHERE id = $1"

	// QueryRow untuk query yang return 1 row saja
	row := r.db.QueryRow(query, id) // id akan replace $1

	var product models.Product // Buat variable untuk menampung hasil

	// Scan hasil query ke struct product
	err := row.Scan(&product.ID, &product.Nama, &product.Harga, &product.Stok)
	if err != nil {
		return nil, err // Kalau tidak ketemu atau error, return nil
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
	// Jika belum ada, akan insert produk baru
	query := `
		INSERT INTO products (nama, harga, stok) 
		VALUES ($1, $2, $3)
		ON CONFLICT (nama) 
		DO UPDATE SET 
			harga = EXCLUDED.harga,
			stok = products.stok + EXCLUDED.stok
		RETURNING id, stok
	`

	// Execute query dan scan ID + stok terbaru yang di-return
	// $1 = product.Nama, $2 = product.Harga, $3 = product.Stok
	err := r.db.QueryRow(query, product.Nama, product.Harga, product.Stok).Scan(&product.ID, &product.Stok)

	return err // Return error (nil kalau sukses)
}

// Update updates an existing product
// Fungsi ini mengupdate produk yang sudah ada
func (r *ProductRepository) Update(product *models.Product) error {
	// SQL query untuk UPDATE
	// SET untuk set nilai baru
	// WHERE untuk kondisi (update produk dengan id tertentu)
	query := "UPDATE products SET nama = $1, harga = $2, stok = $3 WHERE id = $4"

	// Execute query
	// $1 = nama, $2 = harga, $3 = stok, $4 = id
	_, err := r.db.Exec(query, product.Nama, product.Harga, product.Stok, product.ID)

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
