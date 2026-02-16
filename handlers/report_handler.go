package handlers

import (
	"encoding/json"
	"kasir-api/services"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ReportHandler handles HTTP requests for reports
// Handler untuk report/laporan
type ReportHandler struct {
	service *services.ReportService
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// parseTimezone parses timezone from query parameter, defaults to Asia/Jakarta
func parseTimezone(r *http.Request) *time.Location {
	tzStr := r.URL.Query().Get("timezone")
	if tzStr == "" {
		tzStr = "Asia/Jakarta" // Default WIB
	}

	loc, err := time.LoadLocation(tzStr)
	if err != nil {
		log.Printf("⚠️ Timezone '%s' tidak valid, menggunakan Asia/Jakarta: %v", tzStr, err)
		loc, _ = time.LoadLocation("Asia/Jakarta")
		if loc == nil {
			loc = time.FixedZone("WIB", 7*60*60)
		}
	}
	return loc
}

// GetDailySalesReport handles GET /api/report/hari-ini?timezone=Asia/Jakarta
// Fungsi ini handle request untuk laporan penjualan hari ini
func (h *ReportHandler) GetDailySalesReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse timezone dari query parameter (default: Asia/Jakarta)
	loc := parseTimezone(r)

	// Hitung "hari ini" berdasarkan timezone user
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc)

	// Panggil service dengan date range yang sudah di-timezone
	report, err := h.service.GetSalesReportByDateRange(startOfDay, endOfDay)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetSalesReportByDateRange handles GET /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&timezone=Asia/Makassar
// Fungsi ini handle request untuk laporan penjualan berdasarkan rentang tanggal
func (h *ReportHandler) GetSalesReportByDateRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date dan end_date harus diisi (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Parse timezone (default: Asia/Jakarta)
	loc := parseTimezone(r)

	// Parse tanggal
	startDateParsed, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Format start_date tidak valid (gunakan: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDateParsed, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Format end_date tidak valid (gunakan: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Buat boundary waktu berdasarkan timezone user
	// Contoh: 2026-02-17 di Asia/Makassar (UTC+8)
	// → startOfDay = 2026-02-17 00:00:00 WITA = 2026-02-16 16:00:00 UTC
	// → endOfDay   = 2026-02-17 23:59:59 WITA = 2026-02-17 15:59:59 UTC
	startDate := time.Date(startDateParsed.Year(), startDateParsed.Month(), startDateParsed.Day(), 0, 0, 0, 0, loc)
	endDate := time.Date(endDateParsed.Year(), endDateParsed.Month(), endDateParsed.Day(), 23, 59, 59, 999999999, loc)

	report, err := h.service.GetSalesReportByDateRange(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetSalesTrend handles GET /api/dashboard/sales-trend?period=day|month|year&timezone=Asia/Jakarta
func (h *ReportHandler) GetSalesTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	// Parse timezone
	loc := parseTimezone(r)

	trends, err := h.service.GetSalesTrend(period, loc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period": period,
		"data":   trends,
	})
}

// GetTopProducts handles GET /api/dashboard/top-products?limit=5&timezone=Asia/Jakarta
func (h *ReportHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 5
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse timezone
	loc := parseTimezone(r)

	topQty, topProfit, err := h.service.GetTopProducts(limit, loc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"by_quantity": topQty,
		"by_profit":   topProfit,
	})
}
