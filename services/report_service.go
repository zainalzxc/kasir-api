package services

import (
	"fmt"
	"kasir-api/models"
	"kasir-api/repositories"
	"log"
	"time"
)

// wibLocation adalah timezone Asia/Jakarta (WIB, UTC+7)
var wibLocation *time.Location

func init() {
	var err error
	wibLocation, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("⚠️ Warning: Gagal load timezone Asia/Jakarta di service, menggunakan fixed UTC+7: %v", err)
		wibLocation = time.FixedZone("WIB", 7*60*60)
	}
}

// ReportService handles business logic for reports
// Service layer untuk report
type ReportService struct {
	repo *repositories.ReportRepository
}

// NewReportService creates a new ReportService
// Constructor untuk membuat instance ReportService
func NewReportService(repo *repositories.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

// GetDailySalesReport retrieves sales report for today
// Fungsi ini mengambil laporan penjualan hari ini
func (s *ReportService) GetDailySalesReport() (*models.SalesReport, error) {
	return s.repo.GetDailySalesReport()
}

// GetSalesReportByDateRange retrieves sales report for a date range
// Fungsi ini mengambil laporan penjualan untuk rentang tanggal
func (s *ReportService) GetSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	// Validasi: startDate harus sebelum atau sama dengan endDate
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date harus sebelum atau sama dengan end_date")
	}

	return s.repo.GetSalesReportByDateRange(startDate, endDate)
}

// GetSalesTrend retrieves sales trend data based on period type
// periodType: 'daily' (7 days), 'monthly' (12 months), 'yearly' (5 years)
func (s *ReportService) GetSalesTrend(periodType string) ([]models.SalesTrend, error) {
	now := time.Now().In(wibLocation)
	var startDate, endDate time.Time
	var interval string // "day", "month", "year"

	// Set endDate to now (WIB)
	endDate = now

	// Calculate startDate and interval based on periodType
	switch periodType {
	case "monthly":
		// 11 bulan ke belakang + bulan ini = 12 bulan
		// Menggunakan AddDate(years, months, days)
		startDate = now.AddDate(0, -11, 0)
		// Set ke awal bulan (tanggal 1) agar grafik rapi
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, wibLocation)
		interval = "month"
	case "yearly":
		// 4 tahun ke belakang + tahun ini = 5 tahun
		startDate = now.AddDate(-4, 0, 0)
		// Set ke awal tahun (1 Januari)
		startDate = time.Date(startDate.Year(), 1, 1, 0, 0, 0, 0, wibLocation)
		interval = "year"
	default: // "daily" (default)
		// 6 hari ke belakang + hari ini = 7 hari
		startDate = now.AddDate(0, 0, -6)
		// Set ke awal hari (00:00:00)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, wibLocation)
		interval = "day"
	}

	return s.repo.GetSalesTrend(startDate, endDate, interval)
}

// GetTopProducts retrieves top selling products (last 30 days)
// Fungsi ini mengambil produk terlaris dalam 30 hari terakhir
func (s *ReportService) GetTopProducts(limit int) ([]models.TopProduct, []models.TopProduct, error) {
	if limit <= 0 {
		limit = 5 // Default limit
	}

	now := time.Now().In(wibLocation)
	// Default: 30 hari terakhir
	startDate := now.AddDate(0, 0, -30)

	// Set start date to beginning of day (WIB)
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, wibLocation)
	// Set end date to end of day (WIB)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, wibLocation)

	return s.repo.GetTopProducts(startOfDay, endOfDay, limit)
}
