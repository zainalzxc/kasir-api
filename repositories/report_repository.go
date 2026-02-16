package repositories

import (
	"database/sql"
	"kasir-api/models"
	"log"
	"time"
)

// wibTimezone adalah timezone Asia/Jakarta (WIB, UTC+7)
// Digunakan agar report konsisten dengan waktu lokal pengguna
var wibTimezone *time.Location

func init() {
	var err error
	wibTimezone, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback ke fixed offset UTC+7 jika LoadLocation gagal (misal di minimal container)
		log.Printf("⚠️ Warning: Gagal load timezone Asia/Jakarta, menggunakan fixed UTC+7: %v", err)
		wibTimezone = time.FixedZone("WIB", 7*60*60)
	}
}

// ReportRepository handles database operations for reports
// Repository untuk report/laporan
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository creates a new ReportRepository
// Constructor untuk membuat instance ReportRepository
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetDailySalesReport retrieves sales report for today
// Fungsi ini mengambil laporan penjualan untuk hari ini (timezone WIB)
func (r *ReportRepository) GetDailySalesReport() (*models.SalesReport, error) {
	// Get today's date range (00:00:00 - 23:59:59) in WIB timezone
	now := time.Now().In(wibTimezone)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, wibTimezone)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, wibTimezone)

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// GetSalesReportByDateRange retrieves sales report for a date range
// Fungsi ini mengambil laporan penjualan untuk rentang tanggal tertentu (timezone WIB)
func (r *ReportRepository) GetSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	// Set time to start and end of day in WIB timezone
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, wibTimezone)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, wibTimezone)

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// getSalesReportByDateRange is a private helper function
// Fungsi helper untuk mengambil laporan berdasarkan rentang tanggal
func (r *ReportRepository) getSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	var report models.SalesReport

	// Query untuk mendapatkan total revenue, total transaksi, total items terjual, dan total profit
	// JOIN dengan transaction_details untuk menghitung items sold dan profit
	queryRevenue := `
		SELECT 
			COALESCE(SUM(t.total_amount), 0) as total_revenue,
			COUNT(DISTINCT t.id) as total_transaksi,
			COALESCE(SUM(td.quantity), 0) as total_items_sold,
			COALESCE(SUM(td.subtotal - (COALESCE(td.harga_beli, td.price) * td.quantity)), 0) as total_profit
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		WHERE t.created_at BETWEEN $1 AND $2
	`

	err := r.db.QueryRow(queryRevenue, startDate, endDate).Scan(
		&report.TotalRevenue,
		&report.TotalTransaksi,
		&report.TotalItemsSold,
		&report.TotalProfit,
	)
	if err != nil {
		return nil, err
	}

	// Query untuk mendapatkan semua produk terjual
	// Diurutkan berdasarkan total_sales DESC (terlaris di atas)
	queryProducts := `
		SELECT 
			p.nama as nama_produk,
			SUM(td.quantity) as jumlah,
			COALESCE(SUM(td.quantity * td.price), 0) as total_sales,
			COALESCE(SUM(td.quantity * (td.price - COALESCE(td.harga_beli, 0))), 0) as total_profit
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

	// Set array (jika kosong, tetap akan jadi [] di JSON karena bukan pointer)
	if produkTerlaris == nil {
		produkTerlaris = []models.TopProduct{} // Pastikan selalu return [] bukan null
	}
	report.ProdukTerlaris = produkTerlaris

	return &report, nil
}

// GetSalesTrend retrieves sales trend data for chart
// Fungsi untuk mengambil data grafik penjualan
// interval: 'day', 'month', 'year'
func (r *ReportRepository) GetSalesTrend(startDate, endDate time.Time, interval string) ([]models.SalesTrend, error) {
	var trends []models.SalesTrend

	// Tentukan format tanggal output berdasarkan interval
	// PostgreSQL format strings
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

	// Query Aggregation Complex
	// 1. Group by DATE_TRUNC(unit, created_at)
	// 2. Sum total_amount -> TotalSales
	// 3. Sum (subtotal - (harga_beli * quantity)) -> TotalProfit
	// 4. Count distinct transactions -> TransactionCount
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC($1, t.created_at), $2) as period,
			COALESCE(SUM(t.total_amount), 0) as total_sales,
			COALESCE(SUM(td.subtotal - (COALESCE(td.harga_beli, 0) * td.quantity)), 0) as total_profit,
			COUNT(DISTINCT t.id) as transaction_count
		FROM transactions t
		LEFT JOIN transaction_details td ON t.id = td.transaction_id
		WHERE t.created_at BETWEEN $3 AND $4
		GROUP BY 1
		ORDER BY 1 ASC
	`

	rows, err := r.db.Query(query, truncateUnit, dateFormat, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.SalesTrend
		// Scan hasil query ke struct
		// Note: total_sales & total_profit mungkin float/decimal
		err := rows.Scan(&t.Date, &t.TotalSales, &t.TotalProfit, &t.TransactionCount)
		if err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}

	return trends, nil
}

// GetTopProducts returns top selling products by quantity and by profit
// Fungsi untuk mengambil produk terlaris (by Qty) dan paling untung (by Profit)
func (r *ReportRepository) GetTopProducts(startDate, endDate time.Time, limit int) ([]models.TopProduct, []models.TopProduct, error) {
	// 1. Top by Quantity
	// Query untuk top qty
	queryQty := `
		SELECT 
			p.nama,
			COALESCE(SUM(td.quantity), 0) as jumlah,
			COALESCE(SUM(td.subtotal), 0) as total_sales,
			COALESCE(SUM(td.subtotal - (COALESCE(td.harga_beli, 0) * td.quantity)), 0) as total_profit
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
	// Query untuk top profit
	queryProfit := `
		SELECT 
			p.nama,
			COALESCE(SUM(td.quantity), 0) as jumlah,
			COALESCE(SUM(td.subtotal), 0) as total_sales,
			COALESCE(SUM(td.subtotal - (COALESCE(td.harga_beli, 0) * td.quantity)), 0) as total_profit
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
