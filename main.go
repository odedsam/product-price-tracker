package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite" // Import pure Go SQLite driver
)

func main() {
    // Initialize database
    db, err := NewDatabase("prices.db")
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // Create tracker
    tracker := NewPriceTracker(db)

    // Add some sample products to track
    sampleProducts := []Product{
        {ID: "laptop-1", Name: "Gaming Laptop", URL: "https://example.com/laptop-1"},
        {ID: "phone-1", Name: "Smartphone X", URL: "https://example.com/phone-1"},
        {ID: "tablet-1", Name: "Tablet Pro", URL: "https://example.com/tablet-1"},
    }

    for _, product := range sampleProducts {
        if err := tracker.AddProduct(product); err != nil {
            log.Printf("Failed to add product %s: %v", product.ID, err)
        }
    }

    // start price tracking in background
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go tracker.StartTracking(ctx, 30*time.Second) // check prices every 30 seconds

    // create and start HTTP server
    server := NewAPIServer(tracker)
    httpServer := &http.Server{
        Addr:    ":8080",
        Handler: server.router,
    }

    // start server in goroutine
    go func() {
        log.Println("Starting HTTP server on :8080")
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("HTTP server failed:", err)
        }
    }()

    // wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()

    if err := httpServer.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    log.Println("Server stopped")
}

