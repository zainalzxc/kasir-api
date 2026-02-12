package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
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
		var categoryID sql.NullInt64 // Bisa NULL

		// Fetch product details including category_id
		err = tx.QueryRow("SELECT harga, stok, category_id FROM products WHERE id = $1", item.ProductID).Scan(&price, &stok, &categoryID)
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

		// Add to details
		// Note: Subtotal is NET. Price is GROSS.
		details = append(details, models.TransactionDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,    // Store Original Price
			Subtotal:  subtotal, // Store Net Subtotal
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

	// 4. Batch Insert Details
	if len(details) > 0 {
		query := "INSERT INTO transaction_details (transaction_id, product_id, quantity, price, subtotal) VALUES "
		values := make([]interface{}, 0, len(details)*5)

		for i, detail := range details {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
				i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
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
		ID:             transactionID,
		TotalAmount:    finalTotal,
		DiscountID:     usedDiscountID,
		DiscountAmount: totalDiscount,
	}

	return transaction, nil
}
