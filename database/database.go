package database

import (
	"database/sql" // Package standard Go untuk database SQL
	"fmt"          // Package untuk print/format string
	"log"          // Package untuk logging error

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver pgx (better support untuk pooler)
)

// InitDB menginisialisasi koneksi database menggunakan database/sql
func InitDB(dsn string) *sql.DB {
	// Log connection attempt
	log.Println("🔌 Connecting to database...")

	// Connect ke database PostgreSQL menggunakan pgx driver
	// sql.Open tidak langsung connect, hanya prepare connection
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// Kalau gagal, stop aplikasi dan tampilkan error
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Test koneksi dengan Ping (ini yang benar-benar connect)
	err = db.Ping()
	if err != nil {
		// Kalau ping gagal, berarti database tidak bisa diakses
		log.Fatal("❌ Failed to ping database:", err)
	}

	// Buat table jika belum ada
	createTables(db)

	fmt.Println("✅ Database connected successfully")
	return db // Return pointer ke database connection
}

// createTables membuat table products, categories, transactions, dan transaction_details jika belum ada
func createTables(db *sql.DB) {
	// SQL untuk create table products
	// SERIAL = auto-increment integer
	// PRIMARY KEY = unique identifier
	// VARCHAR(255) = string maksimal 255 karakter
	// INTEGER = angka bulat
	// NOT NULL = field wajib diisi
	productsTable := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		nama VARCHAR(255) NOT NULL,
		harga INTEGER NOT NULL,
		stok INTEGER NOT NULL
	)`

	// Execute SQL query untuk create table
	_, err := db.Exec(productsTable)
	if err != nil {
		// Kalau gagal create table, stop aplikasi
		log.Fatal("❌ Failed to create products table:", err)
	}

	// SQL untuk create table categories
	// TEXT = string panjang (untuk description)
	categoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		nama VARCHAR(255) NOT NULL,
		description TEXT
	)`

	// Execute SQL query untuk create table categories
	_, err = db.Exec(categoriesTable)
	if err != nil {
		// Kalau gagal create table, stop aplikasi
		log.Fatal("❌ Failed to create categories table:", err)
	}

	// SQL untuk create table transactions
	// DECIMAL(10, 2) = angka desimal dengan 10 digit total, 2 digit di belakang koma
	// TIMESTAMP = tanggal dan waktu
	// DEFAULT CURRENT_TIMESTAMP = otomatis diisi dengan waktu sekarang
	transactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		total_amount DECIMAL(10, 2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	// Execute SQL query untuk create table transactions
	_, err = db.Exec(transactionsTable)
	if err != nil {
		// Kalau gagal create table, stop aplikasi
		log.Fatal("❌ Failed to create transactions table:", err)
	}

	// SQL untuk create table transaction_details
	// REFERENCES = foreign key ke table lain
	// ON DELETE CASCADE = jika transaction dihapus, detail juga ikut terhapus
	transactionDetailsTable := `
	CREATE TABLE IF NOT EXISTS transaction_details (
		id SERIAL PRIMARY KEY,
		transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id),
		quantity INTEGER NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		subtotal DECIMAL(10, 2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	// Execute SQL query untuk create table transaction_details
	_, err = db.Exec(transactionDetailsTable)
	if err != nil {
		// Kalau gagal create table, stop aplikasi
		log.Fatal("❌ Failed to create transaction_details table:", err)
	}

	// Create indexes untuk mempercepat query
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_transaction_details_transaction_id ON transaction_details(transaction_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_details_product_id ON transaction_details(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at)`,
	}

	for _, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			log.Fatal("❌ Failed to create index:", err)
		}
	}

	fmt.Println("✅ Tables created/verified successfully")
}
