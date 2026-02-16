package models

import "time"

// Transaction represents a transaction header
type Transaction struct {
	ID             int       `json:"id" db:"id"`
	TotalAmount    float64   `json:"total_amount" db:"total_amount"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	DiscountID     *int      `json:"discount_id,omitempty" db:"discount_id"`
	DiscountAmount float64   `json:"discount_amount" db:"discount_amount"`
	PaymentAmount  float64   `json:"payment_amount" db:"payment_amount"` // Uang bayar customer
	ChangeAmount   float64   `json:"change_amount" db:"change_amount"`   // Uang kembalian
	TotalItems     int       `json:"total_items"`                        // Computed: total items
	Profit         float64   `json:"profit"`                             // Computed: keuntungan
}

// TransactionDetail represents a transaction detail item
type TransactionDetail struct {
	ID            int       `json:"id"`
	TransactionID int       `json:"transaction_id"`
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name"` // Nama produk (dari JOIN)
	Quantity      int       `json:"quantity"`
	Price         float64   `json:"price"`
	Subtotal      float64   `json:"subtotal"`
	HargaBeli     float64   `json:"harga_beli,omitempty"` // Snapshot harga beli
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

// TransactionWithItems represents full transaction detail with items
// Response struct untuk GET /api/transactions/{id}
type TransactionWithItems struct {
	ID             int                 `json:"id"`
	TotalAmount    float64             `json:"total_amount"`
	DiscountAmount float64             `json:"discount_amount"`
	PaymentAmount  float64             `json:"payment_amount"`
	ChangeAmount   float64             `json:"change_amount"`
	Profit         float64             `json:"profit"`
	TotalItems     int                 `json:"total_items"`
	CreatedAt      time.Time           `json:"created_at"`
	Items          []TransactionDetail `json:"items"`
}

// CheckoutItem represents an item in checkout request
type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CheckoutRequest represents the checkout request body
type CheckoutRequest struct {
	Items         []CheckoutItem `json:"items"`
	DiscountID    *int           `json:"discount_id"`    // Optional: ID diskon
	PaymentAmount float64        `json:"payment_amount"` // Uang bayar customer
}
