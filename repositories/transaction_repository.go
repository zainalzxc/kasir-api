package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"time"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction creates a new transaction with details
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

	var totalAmount float64
	var totalDiscount float64 // Akumulasi Item Discount + Global Discount
	details := make([]models.TransactionDetail, 0)

	// 1. Process Items & Item Discounts (Product OR Category)
	for _, item := range req.Items {
		var price float64
		var stok int
		var categoryID sql.NullInt64  // Bisa NULL
		var hargaBeli sql.NullFloat64 // Harga beli/modal (bisa NULL)

		// Fetch product details including category_id dan harga_beli
		err = tx.QueryRow("SELECT harga, stok, category_id, harga_beli FROM products WHERE id = $1", item.ProductID).Scan(&price, &stok, &categoryID, &hargaBeli)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		if stok < item.Quantity {
			return nil, fmt.Errorf("stok produk ID %d tidak mencukupi", item.ProductID)
		}

		// --- ITEM DISCOUNT LOGIC (Priority: Product > Category) ---
		var itemDiscountPerUnit float64 = 0
		var discType string
		var discValue float64
		var discountFound bool = false

		// 1. Cek Diskon PRODAK (Highest Priority)
		errDisc := tx.QueryRow(`
			SELECT type, value FROM discounts 
			WHERE product_id = $1 AND is_active = TRUE 
			AND NOW() BETWEEN start_date AND end_date
			ORDER BY value DESC LIMIT 1`, item.ProductID).Scan(&discType, &discValue)

		if errDisc == nil {
			// Product Discount Found!
			discountFound = true
		} else if categoryID.Valid {
			// 2. Jika tidak ada diskon produk, Cek Diskon KATEGORI
			errCatDisc := tx.QueryRow(`
				SELECT type, value FROM discounts 
				WHERE category_id = $1 AND is_active = TRUE 
				AND NOW() BETWEEN start_date AND end_date
				ORDER BY value DESC LIMIT 1`, categoryID.Int64).Scan(&discType, &discValue)

			if errCatDisc == nil {
				// Category Discount Found!
				discountFound = true
			}
		}

		// Calculate Discount Amount if any found
		if discountFound {
			if discType == "PERCENTAGE" {
				itemDiscountPerUnit = price * (discValue / 100)
			} else {
				itemDiscountPerUnit = discValue
			}
			// Safety: Discount cannot exceed price
			if itemDiscountPerUnit > price {
				itemDiscountPerUnit = price
			}
		}

		// Calculate Subtotal (Net Price * Qty)
		priceAfterDisc := price - itemDiscountPerUnit
		subtotal := priceAfterDisc * float64(item.Quantity)

		// Accumulate totals
		totalAmount += subtotal
		totalDiscount += itemDiscountPerUnit * float64(item.Quantity)

		// Update Stock
		_, err = tx.Exec("UPDATE products SET stok = stok - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		// Tentukan harga beli untuk snapshot di transaction_details
		// Jika harga_beli NULL, gunakan harga jual sebagai fallback (profit = 0)
		var hargaBeliSnapshot float64
		if hargaBeli.Valid {
			hargaBeliSnapshot = hargaBeli.Float64
		} else {
			hargaBeliSnapshot = price // fallback: anggap modal = harga jual
		}

		// Add to details
		// Note: Subtotal is NET. Price is GROSS. HargaBeli is snapshot modal.
		details = append(details, models.TransactionDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,             // Store Original Price
			Subtotal:  subtotal,          // Store Net Subtotal
			HargaBeli: hargaBeliSnapshot, // Store harga beli snapshot untuk profit analysis
		})
	}

	// 2. Process Global Discount (if any)
	// Apply on top of totalAmount (which is already net of item discounts)
	var usedDiscountID *int
	var globalDiscountAmount float64

	if req.DiscountID != nil {
		var d models.Discount
		var isValid bool
		// Cek validitas global discount (product_id IS NULL AND category_id IS NULL)
		// Pastikan diskon ini BUKAN diskon produk/kategori
		err = tx.QueryRow(`
			SELECT id, type, value, min_order_amount, (NOW() BETWEEN start_date AND end_date) as is_valid 
			FROM discounts 
			WHERE id = $1 AND is_active = TRUE 
			AND product_id IS NULL AND category_id IS NULL`, *req.DiscountID).Scan(
			&d.ID, &d.Type, &d.Value, &d.MinOrderAmount, &isValid,
		)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("diskon global ID %d tidak valid atau (mungkin diskon produk/kategori)", *req.DiscountID)
		}
		if err != nil {
			return nil, err
		}
		if !isValid {
			return nil, fmt.Errorf("diskon global sudah kedaluwarsa")
		}

		if totalAmount < d.MinOrderAmount {
			return nil, fmt.Errorf("min order %.0f tidak terpenuhi", d.MinOrderAmount)
		}

		// Calculate global discount
		if d.Type == "PERCENTAGE" {
			globalDiscountAmount = totalAmount * (d.Value / 100)
		} else {
			globalDiscountAmount = d.Value
		}

		// Cap global discount
		if globalDiscountAmount > totalAmount {
			globalDiscountAmount = totalAmount
		}

		usedDiscountID = req.DiscountID
	}

	// Final Calculation
	finalTotal := totalAmount - globalDiscountAmount
	totalDiscount += globalDiscountAmount

	// 3. Insert Transaction Header
	var transactionID int
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount, discount_id, discount_amount) VALUES ($1, $2, $3) RETURNING id",
		finalTotal, usedDiscountID, totalDiscount,
	).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// 4. Batch Insert Details (including harga_beli snapshot)
	if len(details) > 0 {
		query := "INSERT INTO transaction_details (transaction_id, product_id, quantity, price, subtotal, harga_beli) VALUES "
		values := make([]interface{}, 0, len(details)*6)

		for i, detail := range details {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
				i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)
			values = append(values, transactionID, detail.ProductID, detail.Quantity, detail.Price, detail.Subtotal, detail.HargaBeli)
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
		ID:             transactionID,
		TotalAmount:    finalTotal,
		DiscountID:     usedDiscountID,
		DiscountAmount: totalDiscount,
	}

	return transaction, nil
}

// GetAll retrieves all transactions ordered by date descending
// Fungsi ini mengambil semua data transaksi untuk history, termasuk profit per transaksi
func (r *TransactionRepository) GetAll() ([]models.Transaction, error) {
	// Query yang menghitung profit dan total items per transaksi
	// Profit = total_amount - SUM(harga_beli * quantity)
	// Jika harga_beli NULL, fallback ke price (profit = 0)
	query := `
		SELECT 
			t.id, 
			t.total_amount, 
			t.discount_id, 
			t.discount_amount, 
			t.created_at,
			COALESCE(SUM(td.quantity), 0) as total_items,
			t.total_amount - COALESCE(SUM(COALESCE(td.harga_beli, td.price) * td.quantity), 0) as profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		GROUP BY t.id, t.total_amount, t.discount_id, t.discount_amount, t.created_at
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	for rows.Next() {
		var t models.Transaction
		var discountID sql.NullInt64

		err := rows.Scan(&t.ID, &t.TotalAmount, &discountID, &t.DiscountAmount, &t.CreatedAt, &t.TotalItems, &t.Profit)
		if err != nil {
			return nil, err
		}

		if discountID.Valid {
			id := int(discountID.Int64)
			t.DiscountID = &id
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByDateRange retrieves transactions within a date range
// startDate dan endDate sudah mengandung timezone yang benar dari handler
func (r *TransactionRepository) GetByDateRange(startDate, endDate time.Time) ([]models.Transaction, error) {
	query := `
		SELECT 
			t.id, 
			t.total_amount, 
			t.discount_id, 
			t.discount_amount, 
			t.created_at,
			COALESCE(SUM(td.quantity), 0) as total_items,
			t.total_amount - COALESCE(SUM(COALESCE(td.harga_beli, td.price) * td.quantity), 0) as profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY t.id, t.total_amount, t.discount_id, t.discount_amount, t.created_at
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	for rows.Next() {
		var t models.Transaction
		var discountID sql.NullInt64

		err := rows.Scan(&t.ID, &t.TotalAmount, &discountID, &t.DiscountAmount, &t.CreatedAt, &t.TotalItems, &t.Profit)
		if err != nil {
			return nil, err
		}

		if discountID.Valid {
			id := int(discountID.Int64)
			t.DiscountID = &id
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}
