package services

import (
	"fmt"
	"kasir-api/models"
	"kasir-api/repositories"
	"time"
)

// ReportService handles business logic for reports
type ReportService struct {
	repo *repositories.ReportRepository
}

// NewReportService creates a new ReportService
func NewReportService(repo *repositories.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

// GetDailySalesReport retrieves sales report for today (kept for backward compat)
func (s *ReportService) GetDailySalesReport() (*models.SalesReport, error) {
	return s.repo.GetDailySalesReport()
}

// GetSalesReportByDateRange retrieves sales report for a date range
// startDate dan endDate sudah mengandung timezone yang benar dari handler
func (s *ReportService) GetSalesReportByDateRange(startDate, endDate time.Time) (*models.SalesReport, error) {
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date harus sebelum atau sama dengan end_date")
	}

	return s.repo.GetSalesReportByDateRange(startDate, endDate)
}

// GetSalesTrend retrieves sales trend data based on period type
// loc = timezone dari user (dikirim dari handler)
func (s *ReportService) GetSalesTrend(periodType string, loc *time.Location) ([]models.SalesTrend, error) {
	now := time.Now().In(loc)
	var startDate, endDate time.Time
	var interval string

	endDate = now

	switch periodType {
	case "monthly":
		startDate = now.AddDate(0, -11, 0)
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, loc)
		interval = "month"
	case "yearly":
		startDate = now.AddDate(-4, 0, 0)
		startDate = time.Date(startDate.Year(), 1, 1, 0, 0, 0, 0, loc)
		interval = "year"
	default: // "daily"
		startDate = now.AddDate(0, 0, -6)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
		interval = "day"
	}

	return s.repo.GetSalesTrend(startDate, endDate, interval)
}

// GetTopProducts retrieves top selling products (last 30 days)
// loc = timezone dari user (dikirim dari handler)
func (s *ReportService) GetTopProducts(limit int, loc *time.Location) ([]models.TopProduct, []models.TopProduct, error) {
	if limit <= 0 {
		limit = 5
	}

	now := time.Now().In(loc)
	startDate := now.AddDate(0, 0, -30)

	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc)

	return s.repo.GetTopProducts(startOfDay, endOfDay, limit)
}
