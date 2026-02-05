package handlers

import (
	"encoding/json"
	"kasir-api/models"
	"kasir-api/services"
	"net/http"
)

// TransactionHandler handles HTTP requests for transactions
// Handler untuk transaction/checkout
type TransactionHandler struct {
	service *services.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler
// Constructor untuk membuat instance TransactionHandler
func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// Checkout handles POST /api/checkout
// Fungsi ini handle checkout request
func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	// Hanya terima POST method
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CheckoutRequest

	// Decode JSON dari request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Panggil service untuk proses checkout
	transaction, err := h.service.Checkout(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Set status code 201 (Created)
	w.WriteHeader(http.StatusCreated)

	// Encode transaction dan kirim ke client
	json.NewEncoder(w).Encode(transaction)
}
