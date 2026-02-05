package models

import "time"

// Transaction represents a transaction header
// Struct ini menyimpan informasi header transaksi
type Transaction struct {
	ID          int       `json:"id"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// TransactionDetail represents a transaction detail item
// Struct ini menyimpan detail item yang dibeli dalam transaksi
type TransactionDetail struct {
	ID            int       `json:"id"`
	TransactionID int       `json:"transaction_id"`
	ProductID     int       `json:"product_id"`
	Quantity      int       `json:"quantity"`
	Price         float64   `json:"price"`
	Subtotal      float64   `json:"subtotal"`
	CreatedAt     time.Time `json:"created_at"`
}

// CheckoutItem represents an item in checkout request
// Struct ini untuk menerima data item yang akan di-checkout dari client
type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CheckoutRequest represents the checkout request body
// Struct ini untuk menerima request checkout dari client
type CheckoutRequest struct {
	Items []CheckoutItem `json:"items"`
}
