package repositories

import (
	"database/sql"
	"kasir-api/models"
	"time"
)

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
// Fungsi ini mengambil laporan penjualan untuk hari ini
func (r *ReportRepository) GetDailySalesReport() (*models.SalesReport, error) {
	// Get today's date range (00:00:00 - 23:59:59)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// GetSalesReportByDateRange retrieves sales report for a date range
// Fungsi ini mengambil laporan penjualan untuk rentang tanggal tertentu
func (r *ReportRepository) GetSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	// Set time to start and end of day
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	return r.getSalesReportByDateRange(startOfDay, endOfDay)
}

// getSalesReportByDateRange is a private helper function
// Fungsi helper untuk mengambil laporan berdasarkan rentang tanggal
func (r *ReportRepository) getSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	var report models.SalesReport

	// Query untuk mendapatkan total revenue dan total transaksi
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

	// Query untuk mendapatkan produk terlaris
	// Join transaction_details dengan products untuk mendapatkan nama produk
	// Group by product_id dan nama produk
	// Order by total quantity DESC untuk mendapatkan yang terlaris
	// LIMIT 1 untuk ambil yang paling top
	queryTopProduct := `
		SELECT 
			p.nama as nama_produk,
			SUM(td.quantity) as jumlah
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE t.created_at BETWEEN $1 AND $2
		GROUP BY p.id, p.nama
		ORDER BY jumlah DESC
		LIMIT 1
	`

	var topProduct models.TopProduct
	err = r.db.QueryRow(queryTopProduct, startDate, endDate).Scan(
		&topProduct.NamaProduk,
		&topProduct.Jumlah,
	)

	// Jika ada produk terlaris, tambahkan ke report
	if err == nil {
		report.ProdukTerlaris = &topProduct
	} else if err != sql.ErrNoRows {
		// Jika error bukan "no rows", return error
		return nil, err
	}
	// Jika sql.ErrNoRows (tidak ada transaksi), ProdukTerlaris akan tetap nil

	return &report, nil
}
