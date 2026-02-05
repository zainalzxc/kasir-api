package models

// SalesReport represents daily sales summary
// Struct untuk laporan penjualan harian
type SalesReport struct {
	TotalRevenue   float64     `json:"total_revenue"`
	TotalTransaksi int         `json:"total_transaksi"`
	ProdukTerlaris *TopProduct `json:"produk_terlaris,omitempty"`
}

// TopProduct represents the best selling product
// Struct untuk produk terlaris
type TopProduct struct {
	NamaProduk string `json:"nama_produk"`
	Jumlah     int    `json:"jumlah"`
}
