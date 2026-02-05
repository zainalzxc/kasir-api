package services

import (
	"fmt"
	"kasir-api/models"
	"kasir-api/repositories"
	"time"
)

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
