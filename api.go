package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
    tracker *PriceTracker
    router  *mux.Router
}

func NewAPIServer(tracker *PriceTracker) *APIServer {
    server := &APIServer{
        tracker: tracker,
        router:  mux.NewRouter(),
    }

    server.setupRoutes()
    return server
}

func (s *APIServer) setupRoutes() {
    api := s.router.PathPrefix("/api/v1").Subrouter()

    api.HandleFunc("/products", s.handleGetProducts).Methods("GET")
    api.HandleFunc("/products/{id}/history", s.handleGetPriceHistory).Methods("GET")
    api.HandleFunc("/health", s.handleHealth).Methods("GET")

    // serve a simple HTML page at root
    s.router.HandleFunc("/", s.handleRoot).Methods("GET")

    // add middleware
    s.router.Use(s.loggingMiddleware)
    s.router.Use(s.corsMiddleware)
}

func (s *APIServer) handleGetProducts(w http.ResponseWriter, r *http.Request) {
    products := s.tracker.GetProducts()
    s.writeJSON(w, http.StatusOK, products)
}

func (s *APIServer) handleGetPriceHistory(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    productID := vars["id"]

    if productID == "" {
        s.writeError(w, http.StatusBadRequest, "Product ID is required")
        return
    }

    // parse limit parameter
    limit := 50 // default
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }

    history, err := s.tracker.GetPriceHistory(productID, limit)
    if err != nil {
        s.writeError(w, http.StatusNotFound, err.Error())
        return
    }

    s.writeJSON(w, http.StatusOK, map[string]interface{}{
        "product_id": productID,
        "history":    history,
        "count":      len(history),
    })
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
    s.writeJSON(w, http.StatusOK, map[string]string{
        "status": "ok",
        "time":   time.Now().Format(time.RFC3339),
    })
}

func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
    html := `<!DOCTYPE html>
<html>
<head>
    <title>Price Tracker</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 20px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
        code { background: #e9e9e9; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Product Price Tracker API</h1>
    <p>Welcome to the Price Tracker API. Available endpoints:</p>

    <div class="endpoint">
        <h3>GET /api/v1/products</h3>
        <p>Get all tracked products with their latest prices</p>
        <p><a href="/api/v1/products">Try it</a></p>
    </div>

    <div class="endpoint">
        <h3>GET /api/v1/products/{id}/history</h3>
        <p>Get price history for a specific product</p>
        <p>Parameters: <code>?limit=N</code> (default: 50)</p>
        <p>Examples:</p>
        <ul>
            <li><a href="/api/v1/products/laptop-1/history">laptop-1 history</a></li>
            <li><a href="/api/v1/products/phone-1/history?limit=10">phone-1 history (limit 10)</a></li>
            <li><a href="/api/v1/products/tablet-1/history">tablet-1 history</a></li>
        </ul>
    </div>

    <div class="endpoint">
        <h3>GET /api/v1/health</h3>
        <p>Health check endpoint</p>
        <p><a href="/api/v1/health">Try it</a></p>
    </div>
</body>
</html>`
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}

func (s *APIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Failed to encode JSON: %v", err)
    }
}

func (s *APIServer) writeError(w http.ResponseWriter, status int, message string) {
    s.writeJSON(w, status, map[string]string{"error": message})
}

func (s *APIServer) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
