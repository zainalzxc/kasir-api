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

// createTables membuat table products dan categories jika belum ada
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

	fmt.Println("✅ Tables created/verified successfully")
}
