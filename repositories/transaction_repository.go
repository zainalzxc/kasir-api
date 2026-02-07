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
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	// Defer rollback jika terjadi error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Inisialisasi total amount dan slice untuk menyimpan transaction details
	var totalAmount float64
	details := make([]models.TransactionDetail, 0)

	// Loop pertama: validasi, hitung total, update stock, dan simpan ke slice
	for _, item := range req.Items {
		// Get product data (harga dan stok)
		var price float64
		var stok int
		err = tx.QueryRow("SELECT harga, stok FROM products WHERE id = $1", item.ProductID).Scan(&price, &stok)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		// Validasi stok
		if stok < item.Quantity {
			return nil, fmt.Errorf("stok produk ID %d tidak mencukupi (tersedia: %d, diminta: %d)", item.ProductID, stok, item.Quantity)
		}

		// Hitung subtotal dan tambahkan ke total amount
		subtotal := price * float64(item.Quantity)
		totalAmount += subtotal

		// Update stok produk (kurangi stok)
		_, err = tx.Exec(
			"UPDATE products SET stok = stok - $1 WHERE id = $2",
			item.Quantity,
			item.ProductID,
		)
		if err != nil {
			return nil, err
		}

		// Simpan detail ke slice (akan di-insert nanti)
		details = append(details, models.TransactionDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
			Subtotal:  subtotal,
		})
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

	// Batch insert transaction details (1 query untuk semua items - OPTIMASI!)
	if len(details) > 0 {
		// Build query string dengan multiple VALUES
		query := "INSERT INTO transaction_details (transaction_id, product_id, quantity, price, subtotal) VALUES "
		values := make([]interface{}, 0, len(details)*5)

		for i, detail := range details {
			// Tambahkan placeholder untuk setiap row
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
				i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)

			// Tambahkan values
			values = append(values, transactionID, detail.ProductID, detail.Quantity, detail.Price, detail.Subtotal)
		}

		// Execute batch insert - 1x query untuk semua items!
		_, err = tx.Exec(query, values...)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
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
