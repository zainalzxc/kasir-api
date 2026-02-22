package repositories

import (
	"database/sql"
	"kasir-api/models"
	"log"
	"time"
)

// defaultTimezone digunakan hanya untuk fallback GetDailySalesReport tanpa parameter
var defaultTimezone *time.Location

func init() {
	var err error
	defaultTimezone, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("⚠️ Warning: Gagal load timezone Asia/Jakarta, menggunakan fixed UTC+7: %v", err)
		defaultTimezone = time.FixedZone("WIB", 7*60*60)
	}
}

// ReportRepository handles database operations for reports
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository creates a new ReportRepository
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetDailySalesReport retrieves sales report for today (fallback, uses default timezone)
func (r *ReportRepository) GetDailySalesReport() (*models.SalesReport, error) {
	now := time.Now().In(defaultTimezone)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, defaultTimezone)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, defaultTimezone)

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// GetSalesReportByDateRange retrieves sales report for a date range
// startDate dan endDate sudah mengandung timezone yang benar dari handler/caller
func (r *ReportRepository) GetSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	// Gunakan timezone yang sudah embedded di startDate/endDate (dari handler)
	loc := startDate.Location()
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// getSalesReportByDateRange is a private helper function
func (r *ReportRepository) getSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	var report models.SalesReport

	// Query 1A: Total revenue dan total transaksi (dari transactions saja, TANPA JOIN)
	// Ini menghindari duplikasi row akibat JOIN ke transaction_details
	queryRevenue := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE created_at BETWEEN $1 AND $2
	`
	err := r.db.QueryRow(queryRevenue, startDate, endDate).Scan(
		&report.TotalRevenue,
		&report.TotalTransaksi,
	)
	if err != nil {
		return nil, err
	}

	// Query 1B: Total items terjual dan profit
	// Profit = t.total_amount (revenue nett setelah SEMUA diskon) - HPP
	// Subquery: hitung HPP & qty per transaksi dulu untuk hindari duplikasi
	queryItems := `
		SELECT 
			COALESCE(SUM(hpp.total_qty), 0) as total_items_sold,
			COALESCE(SUM(t.total_amount) - SUM(hpp.total_hpp), 0) as total_profit
		FROM transactions t
		JOIN (
			SELECT 
				td.transaction_id,
				SUM(td.quantity) as total_qty,
				SUM(COALESCE(td.harga_beli, td.price) * td.quantity) as total_hpp
			FROM transaction_details td
			GROUP BY td.transaction_id
		) hpp ON hpp.transaction_id = t.id
		WHERE t.created_at BETWEEN $1 AND $2
	`
	err = r.db.QueryRow(queryItems, startDate, endDate).Scan(
		&report.TotalItemsSold,
		&report.TotalProfit,
	)
	if err != nil {
		return nil, err
	}

	// Query 2: Total pengeluaran (pembelian) dalam periode yang sama
	queryPengeluaran := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_pengeluaran,
			COUNT(*) as total_pembelian
		FROM purchases
		WHERE created_at BETWEEN $1 AND $2
	`

	err = r.db.QueryRow(queryPengeluaran, startDate, endDate).Scan(
		&report.TotalPengeluaran,
		&report.TotalPembelian,
	)
	if err != nil {
		return nil, err
	}

	// Hitung laba bersih = revenue - pengeluaran
	report.LabaBersih = report.TotalRevenue - report.TotalPengeluaran

	// Query 3: Semua produk terjual (sorted by total_sales DESC)
	// Profit per produk dihitung dengan distribusi proporsional tx-level discount:
	//   item_share = (td.subtotal / SUM(subtotal per transaksi)) × tx.discount_amount
	//   item_profit = td.subtotal - (harga_beli × qty) - item_share
	queryProducts := `
		SELECT 
			p.nama as nama_produk,
			SUM(td.quantity) as jumlah,
			COALESCE(SUM(td.subtotal), 0) as total_sales,
			COALESCE(SUM(
				td.subtotal
				- (COALESCE(td.harga_beli, 0) * td.quantity)
				- (
					td.subtotal
					/ NULLIF((SELECT SUM(s.subtotal) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
					* COALESCE(t.discount_amount - (SELECT COALESCE(SUM(s.discount_amount),0) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
				)
			), 0) as total_profit
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY p.id, p.nama
		ORDER BY total_sales DESC
	`

	rows, err := r.db.Query(queryProducts, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var produkTerlaris []models.TopProduct
	for rows.Next() {
		var p models.TopProduct
		if err := rows.Scan(&p.NamaProduk, &p.Jumlah, &p.TotalSales, &p.TotalProfit); err != nil {
			return nil, err
		}
		produkTerlaris = append(produkTerlaris, p)
	}

	if produkTerlaris == nil {
		produkTerlaris = []models.TopProduct{}
	}
	report.ProdukTerlaris = produkTerlaris

	return &report, nil
}

// GetSalesTrend retrieves sales trend data for chart
func (r *ReportRepository) GetSalesTrend(startDate, endDate time.Time, interval string) ([]models.SalesTrend, error) {
	var trends []models.SalesTrend

	var dateFormat string
	var truncateUnit string

	switch interval {
	case "day":
		truncateUnit = "day"
		dateFormat = "YYYY-MM-DD"
	case "month":
		truncateUnit = "month"
		dateFormat = "YYYY-MM"
	case "year":
		truncateUnit = "year"
		dateFormat = "YYYY"
	default:
		truncateUnit = "day"
		dateFormat = "YYYY-MM-DD"
	}

	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC($1, t.created_at), $2) as period,
			COALESCE(SUM(t.total_amount), 0) as total_sales,
			COALESCE(SUM(t.total_amount) - SUM(hpp.total_hpp), 0) as total_profit,
			COUNT(DISTINCT t.id) as transaction_count
		FROM transactions t
		JOIN (
			SELECT 
				td.transaction_id,
				SUM(COALESCE(td.harga_beli, 0) * td.quantity) as total_hpp
			FROM transaction_details td
			GROUP BY td.transaction_id
		) hpp ON hpp.transaction_id = t.id
		WHERE t.created_at BETWEEN $3 AND $4
		GROUP BY DATE_TRUNC($1, t.created_at)
		ORDER BY DATE_TRUNC($1, t.created_at) ASC
	`

	rows, err := r.db.Query(query, truncateUnit, dateFormat, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.SalesTrend
		err := rows.Scan(&t.Date, &t.TotalSales, &t.TotalProfit, &t.TransactionCount)
		if err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}

	return trends, nil
}

// GetTopProducts returns top selling products by quantity and by profit
func (r *ReportRepository) GetTopProducts(startDate, endDate time.Time, limit int) ([]models.TopProduct, []models.TopProduct, error) {
	// 1. Top by Quantity
	// Profit per produk dihitung dengan distribusi proporsional tx-level discount:
	//   item_profit = td.subtotal - (harga_beli × qty) - bagian_proporsional_tx_discount
	queryQty := `
		SELECT 
			p.nama,
			COALESCE(SUM(td.quantity), 0) as jumlah,
			COALESCE(SUM(td.subtotal), 0) as total_sales,
			COALESCE(SUM(
				td.subtotal
				- (COALESCE(td.harga_beli, 0) * td.quantity)
				- (
					td.subtotal
					/ NULLIF((SELECT SUM(s.subtotal) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
					* COALESCE(t.discount_amount - (SELECT COALESCE(SUM(s.discount_amount),0) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
				)
			), 0) as total_profit
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY p.id, p.nama
		ORDER BY jumlah DESC
		LIMIT $3
	`

	rowsQty, err := r.db.Query(queryQty, startDate, endDate, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rowsQty.Close()

	var topQty []models.TopProduct
	for rowsQty.Next() {
		var p models.TopProduct
		if err := rowsQty.Scan(&p.NamaProduk, &p.Jumlah, &p.TotalSales, &p.TotalProfit); err != nil {
			return nil, nil, err
		}
		topQty = append(topQty, p)
	}

	// 2. Top by Profit
	queryProfit := `
		SELECT 
			p.nama,
			COALESCE(SUM(td.quantity), 0) as jumlah,
			COALESCE(SUM(td.subtotal), 0) as total_sales,
			COALESCE(SUM(
				td.subtotal
				- (COALESCE(td.harga_beli, 0) * td.quantity)
				- (
					td.subtotal
					/ NULLIF((SELECT SUM(s.subtotal) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
					* COALESCE(t.discount_amount - (SELECT COALESCE(SUM(s.discount_amount),0) FROM transaction_details s WHERE s.transaction_id = td.transaction_id), 0)
				)
			), 0) as total_profit
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY p.id, p.nama
		ORDER BY total_profit DESC
		LIMIT $3
	`

	rowsProfit, err := r.db.Query(queryProfit, startDate, endDate, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rowsProfit.Close()

	var topProfit []models.TopProduct
	for rowsProfit.Next() {
		var p models.TopProduct
		if err := rowsProfit.Scan(&p.NamaProduk, &p.Jumlah, &p.TotalSales, &p.TotalProfit); err != nil {
			return nil, nil, err
		}
		topProfit = append(topProfit, p)
	}

	return topQty, topProfit, nil
}
