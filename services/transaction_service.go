package services

import (
	"fmt"
	"kasir-api/models"
	"kasir-api/repositories"
)

// TransactionService handles business logic for transactions
// Service layer untuk transaction
type TransactionService struct {
	repo *repositories.TransactionRepository
}

// NewTransactionService creates a new TransactionService
// Constructor untuk membuat instance TransactionService
func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// Checkout processes a checkout request
// Fungsi ini memproses checkout (membuat transaksi baru)
func (s *TransactionService) Checkout(req *models.CheckoutRequest) (*models.Transaction, error) {
	// Validasi request
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("items tidak boleh kosong")
	}

	// Validasi setiap item
	for _, item := range req.Items {
		if item.ProductID <= 0 {
			return nil, fmt.Errorf("product_id tidak valid")
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity harus lebih dari 0")
		}
	}

	// Panggil repository untuk create transaction
	return s.repo.CreateTransaction(req)
}
