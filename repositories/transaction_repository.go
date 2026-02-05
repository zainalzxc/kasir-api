package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
)

// TransactionRepository handles database operations for transactions
// Repository ini menangani operasi database untuk transaksi
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new TransactionRepository
// Constructor untuk membuat instance TransactionRepository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction creates a new transaction with details
// Fungsi ini membuat transaksi baru beserta detail items
// Menggunakan database transaction untuk memastikan data consistency
func (r *TransactionRepository) CreateTransaction(req *models.CheckoutRequest) (*models.Transaction, error) {
	// Begin database transaction
	// tx.Begin() memulai transaksi database
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	// Defer rollback jika terjadi error
	// Jika ada panic atau error, transaksi akan di-rollback
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Variable untuk menyimpan total amount
	var totalAmount float64

	// Loop untuk menghitung total amount dan validasi stok
	for _, item := range req.Items {
		// Ambil data produk untuk cek harga dan stok
		var price float64
		var stok int
		err = tx.QueryRow("SELECT harga, stok FROM products WHERE id = $1", item.ProductID).Scan(&price, &stok)
		if err != nil {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", item.ProductID)
		}

		// Validasi stok
		if stok < item.Quantity {
			return nil, fmt.Errorf("stok produk ID %d tidak mencukupi (tersedia: %d, diminta: %d)", item.ProductID, stok, item.Quantity)
		}

		// Hitung subtotal dan tambahkan ke total amount
		subtotal := price * float64(item.Quantity)
		totalAmount += subtotal
	}

	// Insert transaction header
	var transactionID int
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id",
		totalAmount,
	).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// Insert transaction details dan update stok produk
	for _, item := range req.Items {
		// Ambil harga produk lagi (untuk consistency)
		var price float64
		err = tx.QueryRow("SELECT harga FROM products WHERE id = $1", item.ProductID).Scan(&price)
		if err != nil {
			return nil, err
		}

		subtotal := price * float64(item.Quantity)

		// Insert transaction detail
		// BUG FIX: Loop ini sebelumnya salah menggunakan variable yang sama
		_, err = tx.Exec(
			"INSERT INTO transaction_details (transaction_id, product_id, quantity, price, subtotal) VALUES ($1, $2, $3, $4, $5)",
			transactionID,
			item.ProductID,
			item.Quantity,
			price,
			subtotal,
		)
		if err != nil {
			return nil, err
		}

		// Update stok produk (kurangi stok)
		_, err = tx.Exec(
			"UPDATE products SET stok = stok - $1 WHERE id = $2",
			item.Quantity,
			item.ProductID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	// Jika semua berhasil, commit perubahan ke database
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Return transaction yang baru dibuat
	transaction := &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
	}

	return transaction, nil
}
