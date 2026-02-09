package models

// Category adalah struct untuk data kategori
type Category struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Nama        string    `json:"nama"`
	Description string    `json:"description"`
	Products    []Product `json:"products,omitempty"` // List products dalam category ini (untuk GET by ID)
}

// TableName untuk override nama table di database
func (Category) TableName() string {
	return "categories"
}
