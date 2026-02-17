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

// CreateTransaction creates a new transaction with details (OPTIMIZED - batch queries)
// API contract tidak berubah - request dan response format tetap sama
func (r *TransactionRepository) CreateTransaction(req *models.CheckoutRequest) (*models.Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// ─── STEP 1: Batch fetch ALL products in 1 query (bukan per-item) ───
	productIDPlaceholders := ""
	productIDArgs := make([]interface{}, len(req.Items))
	for i, item := range req.Items {
		productIDArgs[i] = item.ProductID
		if i > 0 {
			productIDPlaceholders += ", "
		}
		productIDPlaceholders += fmt.Sprintf("$%d", i+1)
	}

	productRows, err := tx.Query(
		fmt.Sprintf("SELECT id, harga, stok, category_id, harga_beli FROM products WHERE id IN (%s)", productIDPlaceholders),
		productIDArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data produk: %w", err)
	}

	type productInfo struct {
		Price      float64
		Stok       int
		CategoryID sql.NullInt64
		HargaBeli  sql.NullFloat64
	}
	productMap := make(map[int]*productInfo)
	for productRows.Next() {
		var id int
		var p productInfo
		if scanErr := productRows.Scan(&id, &p.Price, &p.Stok, &p.CategoryID, &p.HargaBeli); scanErr != nil {
			productRows.Close()
			return nil, scanErr
		}
		productMap[id] = &p
	}
	productRows.Close()

	// Validasi
	for _, item := range req.Items {
		p, exists := productMap[item.ProductID]
		if !exists {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", item.ProductID)
		}
		if p.Stok < item.Quantity {
			return nil, fmt.Errorf("stok produk ID %d tidak mencukupi (sisa: %d, diminta: %d)", item.ProductID, p.Stok, item.Quantity)
		}
	}

	// ─── STEP 2: Batch fetch ALL active item discounts in 1 query ───
	discountRows, err := tx.Query(`
		SELECT id, type, value, product_id, category_id FROM discounts 
		WHERE is_active = TRUE AND NOW() BETWEEN start_date AND end_date
		AND (product_id IS NOT NULL OR category_id IS NOT NULL)
		ORDER BY value DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data diskon: %w", err)
	}

	type discInfo struct {
		Type  string
		Value float64
	}
	productDiscounts := make(map[int]*discInfo)
	categoryDiscounts := make(map[int]*discInfo)
	for discountRows.Next() {
		var id int
		var dType string
		var dValue float64
		var productID, categoryID sql.NullInt64
		if scanErr := discountRows.Scan(&id, &dType, &dValue, &productID, &categoryID); scanErr != nil {
			discountRows.Close()
			return nil, scanErr
		}
		if productID.Valid {
			pid := int(productID.Int64)
			if _, exists := productDiscounts[pid]; !exists {
				productDiscounts[pid] = &discInfo{Type: dType, Value: dValue}
			}
		}
		if categoryID.Valid {
			cid := int(categoryID.Int64)
			if _, exists := categoryDiscounts[cid]; !exists {
				categoryDiscounts[cid] = &discInfo{Type: dType, Value: dValue}
			}
		}
	}
	discountRows.Close()

	// ─── STEP 3: Process items in-memory (TANPA query tambahan) ───
	var totalAmount float64
	var totalDiscount float64
	details := make([]models.TransactionDetail, 0, len(req.Items))

	for _, item := range req.Items {
		p := productMap[item.ProductID]

		var itemDiscountPerUnit float64
		disc, found := productDiscounts[item.ProductID]
		if !found && p.CategoryID.Valid {
			disc, _ = categoryDiscounts[int(p.CategoryID.Int64)]
		}
		if disc != nil {
			if disc.Type == "PERCENTAGE" {
				itemDiscountPerUnit = p.Price * (disc.Value / 100)
			} else {
				itemDiscountPerUnit = disc.Value
			}
			if itemDiscountPerUnit > p.Price {
				itemDiscountPerUnit = p.Price
			}
		}

		priceAfterDisc := p.Price - itemDiscountPerUnit
		subtotal := priceAfterDisc * float64(item.Quantity)
		totalAmount += subtotal
		totalDiscount += itemDiscountPerUnit * float64(item.Quantity)

		var hargaBeliSnapshot float64
		if p.HargaBeli.Valid {
			hargaBeliSnapshot = p.HargaBeli.Float64
		} else {
			hargaBeliSnapshot = p.Price
		}

		details = append(details, models.TransactionDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     p.Price,
			Subtotal:  subtotal,
			HargaBeli: hargaBeliSnapshot,
		})
	}

	// ─── STEP 4: Batch UPDATE stock in 1 query ───
	stockQuery := "UPDATE products SET stok = CASE "
	stockIDs := ""
	stockArgs := make([]interface{}, 0, len(req.Items)*2)
	argIdx := 1
	for i, item := range req.Items {
		stockQuery += fmt.Sprintf("WHEN id = $%d THEN stok - $%d ", argIdx, argIdx+1)
		stockArgs = append(stockArgs, item.ProductID, item.Quantity)
		if i > 0 {
			stockIDs += ", "
		}
		stockIDs += fmt.Sprintf("$%d", argIdx)
		argIdx += 2
	}
	stockQuery += fmt.Sprintf("END WHERE id IN (%s)", stockIDs)
	_, err = tx.Exec(stockQuery, stockArgs...)
	if err != nil {
		return nil, fmt.Errorf("gagal update stok: %w", err)
	}

	// ─── STEP 5: Process Global Discount (if any) ───
	var usedDiscountID *int
	var globalDiscountAmount float64

	if req.DiscountID != nil {
		var d models.Discount
		var isValid bool
		err = tx.QueryRow(`
			SELECT id, type, value, min_order_amount, (NOW() BETWEEN start_date AND end_date) as is_valid 
			FROM discounts WHERE id = $1 AND is_active = TRUE 
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
		if d.Type == "PERCENTAGE" {
			globalDiscountAmount = totalAmount * (d.Value / 100)
		} else {
			globalDiscountAmount = d.Value
		}
		if globalDiscountAmount > totalAmount {
			globalDiscountAmount = totalAmount
		}
		usedDiscountID = req.DiscountID
	}

	finalTotal := totalAmount - globalDiscountAmount
	totalDiscount += globalDiscountAmount

	paymentAmount := req.PaymentAmount
	changeAmount := 0.0
	if paymentAmount > 0 {
		changeAmount = paymentAmount - finalTotal
		if changeAmount < 0 {
			changeAmount = 0
		}
	}

	// ─── STEP 6: Insert transaction header ───
	var transactionID int
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount, discount_id, discount_amount, payment_amount, change_amount) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		finalTotal, usedDiscountID, totalDiscount, paymentAmount, changeAmount,
	).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// ─── STEP 7: Batch insert details ───
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
		_, err = tx.Exec(query, values...)
		if err != nil {
			return nil, err
		}
	}

	// ─── STEP 8: Commit ───
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:             transactionID,
		TotalAmount:    finalTotal,
		DiscountID:     usedDiscountID,
		DiscountAmount: totalDiscount,
		PaymentAmount:  paymentAmount,
		ChangeAmount:   changeAmount,
	}, nil
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
			COALESCE(t.payment_amount, 0) as payment_amount,
			COALESCE(t.change_amount, 0) as change_amount,
			t.created_at,
			COALESCE(SUM(td.quantity), 0) as total_items,
			t.total_amount - COALESCE(SUM(COALESCE(td.harga_beli, td.price) * td.quantity), 0) as profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		GROUP BY t.id, t.total_amount, t.discount_id, t.discount_amount, t.payment_amount, t.change_amount, t.created_at
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

		err := rows.Scan(&t.ID, &t.TotalAmount, &discountID, &t.DiscountAmount, &t.PaymentAmount, &t.ChangeAmount, &t.CreatedAt, &t.TotalItems, &t.Profit)
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
			COALESCE(t.payment_amount, 0) as payment_amount,
			COALESCE(t.change_amount, 0) as change_amount,
			t.created_at,
			COALESCE(SUM(td.quantity), 0) as total_items,
			t.total_amount - COALESCE(SUM(COALESCE(td.harga_beli, td.price) * td.quantity), 0) as profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY t.id, t.total_amount, t.discount_id, t.discount_amount, t.payment_amount, t.change_amount, t.created_at
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

		err := rows.Scan(&t.ID, &t.TotalAmount, &discountID, &t.DiscountAmount, &t.PaymentAmount, &t.ChangeAmount, &t.CreatedAt, &t.TotalItems, &t.Profit)
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

// GetByID retrieves a transaction by ID with all its items
func (r *TransactionRepository) GetByID(id int) (*models.TransactionWithItems, error) {
	queryHeader := `
		SELECT 
			t.id, 
			t.total_amount, 
			t.discount_amount, 
			COALESCE(t.payment_amount, 0) as payment_amount,
			COALESCE(t.change_amount, 0) as change_amount,
			t.created_at,
			COALESCE(SUM(td.quantity), 0) as total_items,
			t.total_amount - COALESCE(SUM(COALESCE(td.harga_beli, td.price) * td.quantity), 0) as profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		WHERE t.id = $1
		GROUP BY t.id, t.total_amount, t.discount_amount, t.payment_amount, t.change_amount, t.created_at
	`

	var result models.TransactionWithItems
	err := r.db.QueryRow(queryHeader, id).Scan(
		&result.ID,
		&result.TotalAmount,
		&result.DiscountAmount,
		&result.PaymentAmount,
		&result.ChangeAmount,
		&result.CreatedAt,
		&result.TotalItems,
		&result.Profit,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaksi dengan ID %d tidak ditemukan", id)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data transaksi: %w", err)
	}

	queryItems := `
		SELECT 
			td.id,
			td.product_id,
			COALESCE(p.nama, 'Produk Dihapus') as product_name,
			td.quantity,
			td.price,
			td.subtotal
		FROM transaction_details td
		LEFT JOIN products p ON td.product_id = p.id
		WHERE td.transaction_id = $1
		ORDER BY td.id
	`

	rows, err := r.db.Query(queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail items: %w", err)
	}
	defer rows.Close()

	var items []models.TransactionDetail
	for rows.Next() {
		var item models.TransactionDetail
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.Quantity, &item.Price, &item.Subtotal)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca detail item: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.TransactionDetail{}
	}
	result.Items = items

	return &result, nil
}
