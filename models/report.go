package models

// SalesReport represents daily sales summary
// Struct untuk laporan penjualan harian
type SalesReport struct {
	TotalRevenue   float64      `json:"total_revenue"`
	TotalTransaksi int          `json:"total_transaksi"`
	TotalItemsSold int          `json:"total_items_sold"` // Total items terjual
	TotalProfit    float64      `json:"total_profit"`     // Total keuntungan (revenue - modal)
	ProdukTerlaris []TopProduct `json:"produk_terlaris"`  // Array semua produk terjual (sorted by total_sales DESC)
}

// TopProduct represents the best selling product
// Struct untuk produk terlaris
type TopProduct struct {
	NamaProduk  string  `json:"nama_produk"`
	Jumlah      int     `json:"jumlah"`       // Quantity terjual
	TotalSales  float64 `json:"total_sales"`  // Total omzet
	TotalProfit float64 `json:"total_profit"` // Total keuntungan (profit)
}

// SalesTrend represents sales data over a period
// Struct untuk grafik trend penjualan
type SalesTrend struct {
	Date             string  `json:"date"`              // Tanggal/Bulan/Tahun
	TotalSales       float64 `json:"total_sales"`       // Total penjualan (omzet)
	TotalProfit      float64 `json:"total_profit"`      // Total keuntungan (profit)
	TransactionCount int     `json:"transaction_count"` // Jumlah transaksi
}
