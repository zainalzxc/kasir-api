package main

import (
	"encoding/json"          // Package untuk encode/decode JSON
	"fmt"                    // Package untuk print ke console
	"kasir-api/config"       // Import package config untuk configuration management
	"kasir-api/database"     // Import package database untuk koneksi DB
	"kasir-api/handlers"     // Import package handlers untuk HTTP handlers
	"kasir-api/repositories" // Import package repositories untuk database operations
	"kasir-api/services"     // Import package services untuk business logic
	"log"                    // Package untuk logging
	"net/http"               // Package untuk HTTP server
)

func main() {
	// ==================== LOAD CONFIGURATION ====================
	// Load config dari .env file dan environment variables menggunakan Viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}

	// ==================== INITIALIZE DATABASE ====================
	// Panggil InitDB untuk connect ke database PostgreSQL
	// Gunakan connection string dari config
	db := database.InitDB(cfg.GetDatabaseURL())

	// defer = pastikan db.Close() dipanggil saat program selesai
	// Ini penting untuk tutup koneksi database dengan benar
	defer db.Close()

	// ==================== INITIALIZE REDIS ====================
	// Initialize Redis connection untuk caching
	config.InitRedis()
	// defer = pastikan Redis connection ditutup saat program selesai
	defer config.CloseRedis()

	// ==================== DEPENDENCY INJECTION ====================
	// Dependency Injection = "inject" dependency ke layer yang membutuhkan
	// Flow: Database -> Repository -> Service -> Handler

	// Cache service (shared across all services)
	cacheService := services.NewCacheService()

	// Product layers
	productRepo := repositories.NewProductRepository(db)                    // Inject db ke repository
	productService := services.NewProductService(productRepo, cacheService) // Inject repo dan cache ke service
	productHandler := handlers.NewProductHandler(productService)            // Inject service ke handler

	// Category layers
	categoryRepo := repositories.NewCategoryRepository(db)                     // Inject db ke repository
	categoryService := services.NewCategoryService(categoryRepo, cacheService) // Inject repo dan cache ke service
	categoryHandler := handlers.NewCategoryHandler(categoryService)            // Inject service ke handler

	// Transaction layers
	transactionRepo := repositories.NewTransactionRepository(db)             // Inject db ke repository
	transactionService := services.NewTransactionService(transactionRepo)    // Inject repo ke service
	transactionHandler := handlers.NewTransactionHandler(transactionService) // Inject service ke handler

	// Report layers
	reportRepo := repositories.NewReportRepository(db)        // Inject db ke repository
	reportService := services.NewReportService(reportRepo)    // Inject repo ke service
	reportHandler := handlers.NewReportHandler(reportService) // Inject service ke handler

	// ==================== ROUTING ====================
	// Routing = mapping URL ke handler function

	// Health check endpoint
	// GET /health -> return status OK
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json") // Set response type jadi JSON
		json.NewEncoder(w).Encode(map[string]string{       // Encode map jadi JSON
			"status":  "OK",
			"message": "API Running",
		})
	})

	// Product routes
	// /api/produk/ (dengan slash) -> untuk endpoint dengan ID (/api/produk/1)
	http.HandleFunc("/api/produk/", productHandler.HandleProductByID)

	// /api/produk (tanpa slash) -> untuk endpoint tanpa ID (/api/produk)
	http.HandleFunc("/api/produk", productHandler.HandleProducts)

	// Category routes
	// /api/categories/ (dengan slash) -> untuk endpoint dengan ID
	http.HandleFunc("/api/categories/", categoryHandler.HandleCategoryByID)

	// /api/categories (tanpa slash) -> untuk endpoint tanpa ID
	http.HandleFunc("/api/categories", categoryHandler.HandleCategories)

	// Transaction routes
	// POST /api/checkout -> untuk checkout/create transaction
	http.HandleFunc("/api/checkout", transactionHandler.Checkout)

	// Report routes
	// GET /api/report/hari-ini -> untuk laporan penjualan hari ini
	http.HandleFunc("/api/report/hari-ini", reportHandler.GetDailySalesReport)

	// GET /api/report -> untuk laporan penjualan berdasarkan rentang tanggal
	http.HandleFunc("/api/report", reportHandler.GetSalesReportByDateRange)

	// ==================== START SERVER ====================
	// Get port from config
	port := cfg.Port

	// Print informasi server
	fmt.Println("🚀 Server running on port:", port)
	fmt.Println("📚 Endpoints:")
	fmt.Println("  - GET    /health")
	fmt.Println("  - GET    /api/produk")
	fmt.Println("  - POST   /api/produk")
	fmt.Println("  - GET    /api/produk/{id}")
	fmt.Println("  - PUT    /api/produk/{id}")
	fmt.Println("  - DELETE /api/produk/{id}")
	fmt.Println("  - GET    /api/categories")
	fmt.Println("  - POST   /api/categories")
	fmt.Println("  - GET    /api/categories/{id}")
	fmt.Println("  - PUT    /api/categories/{id}")
	fmt.Println("  - DELETE /api/categories/{id}")
	fmt.Println("  - POST   /api/checkout")
	fmt.Println("  - GET    /api/report/hari-ini")
	fmt.Println("  - GET    /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD")

	// Start HTTP server
	// ListenAndServe akan block (program tidak akan lanjut ke baris berikutnya)
	// Server akan terus running sampai di-stop (Ctrl+C)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		// Kalau gagal start server (misal: port sudah dipakai), print error
		fmt.Println("❌ Failed to start server:", err)
	}
}
