package main

import (
	"encoding/json"          // Package untuk encode/decode JSON
	"fmt"                    // Package untuk print ke console
	"kasir-api/config"       // Import package config untuk configuration management
	"kasir-api/database"     // Import package database untuk koneksi DB
	"kasir-api/handlers"     // Import package handlers untuk HTTP handlers
	"kasir-api/middleware"   // Import package middleware untuk auth, logging, CORS
	"kasir-api/repositories" // Import package repositories untuk database operations
	"kasir-api/services"     // Import package services untuk business logic
	"log"                    // Package untuk logging
	"log/slog"               // Package untuk structured logging
	"net/http"               // Package untuk HTTP server
	"os"                     // Package untuk environment variables
)

func main() {
	// ==================== SETUP LOGGING ====================
	// Setup structured logging dengan slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// ==================== LOAD CONFIGURATION ====================
	// Load config dari .env file dan environment variables menggunakan Viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}

	// Verify JWT_SECRET is set
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("❌ JWT_SECRET is not set in environment variables!")
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

	// User & Auth layers (NEW!)
	userRepo := repositories.NewUserRepository(db)      // Inject db ke repository
	authService := services.NewAuthService(userRepo)    // Inject repo ke service
	authHandler := handlers.NewAuthHandler(authService) // Inject service ke handler

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

	// Discount layers (NEW!)
	discountRepo := repositories.NewDiscountRepository(db)
	discountHandler := handlers.NewDiscountHandler(discountRepo)

	// ==================== SETUP ROUTER WITH MIDDLEWARE ====================
	// Create a new ServeMux for better routing
	mux := http.NewServeMux()

	// ... (routes existing) ...

	// Dashboard routes
	// /api/dashboard/sales-trend -> GET (Admin Only) ?period=day|month|year
	mux.Handle("/api/dashboard/sales-trend", middleware.AuthMiddleware(middleware.RequireAdmin(http.HandlerFunc(reportHandler.GetSalesTrend))))
	// /api/dashboard/top-products -> GET (Admin Only) ?limit=5
	mux.Handle("/api/dashboard/top-products", middleware.AuthMiddleware(middleware.RequireAdmin(http.HandlerFunc(reportHandler.GetTopProducts))))

	// Discount routes
	// /api/discounts/active -> GET (Public/Kasir)
	mux.Handle("/api/discounts/active", middleware.AuthMiddleware(http.HandlerFunc(discountHandler.GetActive)))

	// /api/discounts -> GET (Admin), POST (Admin)
	mux.Handle("/api/discounts", middleware.AuthMiddleware(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			discountHandler.GetAll(w, r)
		} else if r.Method == http.MethodPost {
			discountHandler.Create(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// /api/discounts/ -> PUT, DELETE (Admin)
	mux.Handle("/api/discounts/", middleware.AuthMiddleware(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler ini akan menangkap /api/discounts/{id}
		// Cek method
		if r.Method == http.MethodPut {
			discountHandler.Update(w, r)
		} else if r.Method == http.MethodDelete {
			discountHandler.Delete(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// ==================== PUBLIC ROUTES (No Auth Required) ====================

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "Kasir API Running - Session 4 with Authentication",
			"version": "1.0.0",
		})
	})

	// Auth routes (public - no auth required)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// ==================== PROTECTED ROUTES (Auth Required) ====================

	// Middleware for authentication
	// We wrap the handlers with AuthMiddleware to ensure only authenticated users can access

	// Product routes
	// /api/produk/ -> GET (by ID), PUT, DELETE
	mux.Handle("/api/produk/", middleware.AuthMiddleware(http.HandlerFunc(productHandler.HandleProductByID)))

	// /api/produk -> GET (all), POST
	mux.Handle("/api/produk", middleware.AuthMiddleware(http.HandlerFunc(productHandler.HandleProducts)))

	// Category routes
	mux.Handle("/api/categories/", middleware.AuthMiddleware(http.HandlerFunc(categoryHandler.HandleCategoryByID)))
	mux.Handle("/api/categories", middleware.AuthMiddleware(http.HandlerFunc(categoryHandler.HandleCategories)))

	// Transaction routes
	mux.Handle("/api/checkout", middleware.AuthMiddleware(http.HandlerFunc(transactionHandler.Checkout)))

	// Report routes
	mux.Handle("/api/report/hari-ini", middleware.AuthMiddleware(http.HandlerFunc(reportHandler.GetDailySalesReport)))
	mux.Handle("/api/report", middleware.AuthMiddleware(http.HandlerFunc(reportHandler.GetSalesReportByDateRange)))

	// ==================== APPLY GLOBAL MIDDLEWARE ====================
	// Middleware chain: CORS -> Logging -> Handler
	handler := middleware.CORSMiddleware(
		middleware.LoggingMiddleware(mux),
	)

	// ==================== START SERVER ====================
	port := cfg.Port

	// Print informasi server
	fmt.Println("🚀 ========================================")
	fmt.Println("🚀 Kasir API - Session 4 (Authentication)")
	fmt.Println("🚀 ========================================")
	fmt.Println("📡 Server running on port:", port)
	fmt.Println("🔐 Authentication: ENABLED")
	fmt.Println("📝 Logging: ENABLED (structured JSON)")
	fmt.Println("🌐 CORS: ENABLED")
	fmt.Println("")
	fmt.Println("📚 Public Endpoints:")
	fmt.Println("  - GET    /health")
	fmt.Println("  - POST   /api/auth/login")
	fmt.Println("")
	fmt.Println("📚 Product Endpoints:")
	fmt.Println("  - GET    /api/produk")
	fmt.Println("  - POST   /api/produk")
	fmt.Println("  - GET    /api/produk/{id}")
	fmt.Println("  - PUT    /api/produk/{id}")
	fmt.Println("  - DELETE /api/produk/{id}")
	fmt.Println("")
	fmt.Println("📚 Category Endpoints:")
	fmt.Println("  - GET    /api/categories")
	fmt.Println("  - POST   /api/categories")
	fmt.Println("  - GET    /api/categories/{id}")
	fmt.Println("  - PUT    /api/categories/{id}")
	fmt.Println("  - DELETE /api/categories/{id}")
	fmt.Println("")
	fmt.Println("📚 Transaction Endpoints:")
	fmt.Println("  - POST   /api/checkout")
	fmt.Println("")
	fmt.Println("📚 Report Endpoints:")
	fmt.Println("  - GET    /api/report/hari-ini")
	fmt.Println("  - GET    /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD")
	fmt.Println("")
	fmt.Println("🔑 Default Credentials:")
	fmt.Println("  - admin / admin123 (role: admin)")
	fmt.Println("  - kasir1 / kasir123 (role: kasir)")
	fmt.Println("")
	fmt.Println("✅ Ready to accept requests!")
	fmt.Println("========================================")

	// Start HTTP server
	slog.Info("Starting server", "port", port)
	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		fmt.Println("❌ Failed to start server:", err)
	}
}
