package handlers

import (
	"encoding/json"
	"kasir-api/services"
	"net/http"
	"time"
)

// ReportHandler handles HTTP requests for reports
// Handler untuk report/laporan
type ReportHandler struct {
	service *services.ReportService
}

// NewReportHandler creates a new ReportHandler
// Constructor untuk membuat instance ReportHandler
func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// GetDailySalesReport handles GET /api/report/hari-ini
// Fungsi ini handle request untuk laporan penjualan hari ini
func (h *ReportHandler) GetDailySalesReport(w http.ResponseWriter, r *http.Request) {
	// Hanya terima GET method
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Panggil service untuk ambil laporan hari ini
	report, err := h.service.GetDailySalesReport()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode report dan kirim ke client
	json.NewEncoder(w).Encode(report)
}

// GetSalesReportByDateRange handles GET /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
// Fungsi ini handle request untuk laporan penjualan berdasarkan rentang tanggal
func (h *ReportHandler) GetSalesReportByDateRange(w http.ResponseWriter, r *http.Request) {
	// Hanya terima GET method
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Validasi query parameters
	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date dan end_date harus diisi (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Parse string ke time.Time
	// Format: 2006-01-02 adalah format date di Go (YYYY-MM-DD)
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Format start_date tidak valid (gunakan: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Format end_date tidak valid (gunakan: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Panggil service untuk ambil laporan berdasarkan date range
	report, err := h.service.GetSalesReportByDateRange(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode report dan kirim ke client
	json.NewEncoder(w).Encode(report)
}
