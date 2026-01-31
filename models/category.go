package models

// Category adalah struct untuk data kategori
type Category struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Nama        string `json:"nama"`
	Description string `json:"deskription"`
}

// TableName untuk override nama table di database
func (Category) TableName() string {
	return "categories"
}
