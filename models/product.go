package models

// Product adalah struct untuk data produk
type Product struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Nama  string `json:"nama"`
	Harga int    `json:"harga"`
	Stok  int    `json:"stok"`
}

// TableName untuk override nama table di database
func (Product) TableName() string {
	return "products"
}
