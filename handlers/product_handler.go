package handlers

import (
	"encoding/json"      // Package untuk encode/decode JSON
	"kasir-api/models"   // Import models untuk struct Product
	"kasir-api/services" // Import services untuk business logic
	"net/http"           // Package untuk HTTP server
	"strconv"            // Package untuk convert string ke int
	"strings"            // Package untuk manipulasi string
)

// ProductHandler handles HTTP requests for products
// Handler adalah layer yang berhadapan langsung dengan HTTP request/response
type ProductHandler struct {
	service *services.ProductService // Pointer ke ProductService
}

// NewProductHandler creates a new ProductHandler
// Fungsi ini adalah "constructor" untuk membuat instance ProductHandler
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service} // Return struct dengan service yang sudah di-inject
}

// HandleProducts handles /api/produk (GET all, POST new)
// Fungsi ini handle 2 method: GET (ambil semua) dan POST (tambah baru)
func (h *ProductHandler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	// Switch berdasarkan HTTP method
	switch r.Method {
	case "GET":
		h.GetAll(w, r) // Kalau GET, panggil GetAll
	case "POST":
		h.Create(w, r) // Kalau POST, panggil Create
	default:
		// Kalau method lain (PUT, DELETE, dll), return error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProductByID handles /api/produk/{id} (GET, PUT, DELETE)
// Fungsi ini handle 3 method: GET (by ID), PUT (update), DELETE (hapus)
func (h *ProductHandler) HandleProductByID(w http.ResponseWriter, r *http.Request) {
	// Switch berdasarkan HTTP method
	switch r.Method {
	case "GET":
		h.GetByID(w, r) // Kalau GET, panggil GetByID
	case "PUT":
		h.Update(w, r) // Kalau PUT, panggil Update
	case "DELETE":
		h.Delete(w, r) // Kalau DELETE, panggil Delete
	default:
		// Kalau method lain, return error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetAll retrieves all products
// Fungsi ini handle GET /api/produk
// Support query parameter: ?name=xxx untuk search by name
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Ambil query parameter 'name' dari URL
	// Contoh: /api/produk?name=indom -> searchName = "indom"
	searchName := r.URL.Query().Get("name")

	// Panggil service untuk ambil produk (dengan atau tanpa filter)
	products, err := h.service.GetAll(searchName)
	if err != nil {
		// Kalau error, kirim HTTP error 500 (Internal Server Error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode products jadi JSON dan kirim ke client
	json.NewEncoder(w).Encode(products)
}

// GetByID retrieves a product by ID
// Fungsi ini handle GET /api/produk/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID dari URL
	// Misal URL: /api/produk/5 -> idStr = "5"
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")

	// Convert string "5" jadi integer 5
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Kalau gagal convert (misal: /api/produk/abc), return error 400 (Bad Request)
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	// Panggil service untuk ambil produk by ID
	product, err := h.service.GetByID(id)
	if err != nil {
		// Kalau tidak ketemu, return error 404 (Not Found)
		http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode product jadi JSON dan kirim ke client
	json.NewEncoder(w).Encode(product)
}

// Create adds a new product
// Fungsi ini handle POST /api/produk
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var product models.Product // Buat variable untuk menampung data dari request

	// Decode JSON dari request body ke struct product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		// Kalau JSON invalid, return error 400 (Bad Request)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Panggil service untuk create produk baru
	err = h.service.Create(&product)
	if err != nil {
		// Kalau error saat create, return error 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Set status code 201 (Created)
	w.WriteHeader(http.StatusCreated)

	// Encode product yang baru dibuat (sudah ada ID) dan kirim ke client
	json.NewEncoder(w).Encode(product)
}

// Update updates an existing product
// Fungsi ini handle PUT /api/produk/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract ID dari URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")

	// Convert string ke integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Kalau invalid ID, return error 400
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	var product models.Product // Buat variable untuk menampung data update

	// Decode JSON dari request body
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		// Kalau JSON invalid, return error 400
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Panggil service untuk update produk
	err = h.service.Update(id, &product)
	if err != nil {
		// Kalau error saat update, return error 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode product yang sudah di-update dan kirim ke client
	json.NewEncoder(w).Encode(product)
}

// Delete removes a product
// Fungsi ini handle DELETE /api/produk/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract ID dari URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")

	// Convert string ke integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Kalau invalid ID, return error 400
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	// Panggil service untuk delete produk
	err = h.service.Delete(id)
	if err != nil {
		// Kalau error saat delete, return error 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header Content-Type jadi application/json
	w.Header().Set("Content-Type", "application/json")

	// Kirim response sukses delete
	json.NewEncoder(w).Encode(map[string]string{"message": "sukses delete"})
}
