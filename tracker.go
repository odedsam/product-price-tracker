package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type PriceTracker struct {
    db       *Database
    products map[string]Product
    mu       sync.RWMutex
}

func NewPriceTracker(db *Database) *PriceTracker {
    tracker := &PriceTracker{
        db:       db,
        products: make(map[string]Product),
    }

    // load existing products from database
    if err := tracker.loadProducts(); err != nil {
        log.Printf("Failed to load products: %v", err)
    }

    return tracker
}

func (pt *PriceTracker) loadProducts() error {
    products, err := pt.db.GetAllProducts()
    if err != nil {
        return err
    }

    pt.mu.Lock()
    defer pt.mu.Unlock()

    for _, product := range products {
        pt.products[product.ID] = product
    }

    log.Printf("Loaded %d products from database", len(products))
    return nil
}

func (pt *PriceTracker) AddProduct(product Product) error {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    // save to database
    if err := pt.db.InsertProduct(product); err != nil {
        return err
    }

    // add to in-memory map
    pt.products[product.ID] = product
    log.Printf("Added product: %s (%s)", product.Name, product.ID)

    return nil
}

func (pt *PriceTracker) GetProducts() []ProductWithLatestPrice {
    products, err := pt.db.GetProductsWithLatestPrices()
    if err != nil {
        log.Printf("Failed to get products with prices: %v", err)
        return []ProductWithLatestPrice{}
    }
    return products
}

func (pt *PriceTracker) GetPriceHistory(productID string, limit int) ([]PriceEntry, error) {
    // check if product exists
    exists, err := pt.db.ProductExists(productID)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, fmt.Errorf("product not found: %s", productID)
    }

    return pt.db.GetPriceHistory(productID, limit)
}

func (pt *PriceTracker) StartTracking(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    log.Printf("Starting price tracking with interval: %v", interval)

    for {
        select {
        case <-ctx.Done():
            log.Println("Price tracking stopped")
            return
        case <-ticker.C:
            pt.trackAllProducts()
        }
    }
}

func (pt *PriceTracker) trackAllProducts() {
    pt.mu.RLock()
    products := make([]Product, 0, len(pt.products))
    for _, product := range pt.products {
        products = append(products, product)
    }
    pt.mu.RUnlock()

    if len(products) == 0 {
        return
    }

    log.Printf("Tracking prices for %d products", len(products))

    // use worker pool pattern with goroutines
    const numWorkers = 5
    productChan := make(chan Product, len(products))
    resultChan := make(chan PriceEntry, len(products))

    // start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go pt.priceWorker(&wg, productChan, resultChan)
    }

    // send products to workers
    go func() {
        for _, product := range products {
            productChan <- product
        }
        close(productChan)
    }()

    // wait for workers to finish
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // collect results and save to database
    for entry := range resultChan {
        if err := pt.db.InsertPriceEntry(entry.ProductID, entry.Price, entry.Timestamp); err != nil {
            log.Printf("Failed to save price entry for %s: %v", entry.ProductID, err)
        } else {
            log.Printf("Saved price for %s: $%.2f", entry.ProductID, entry.Price)
        }
    }
}

func (pt *PriceTracker) priceWorker(wg *sync.WaitGroup, productChan <-chan Product, resultChan chan<- PriceEntry) {
    defer wg.Done()

    for product := range productChan {
        price := pt.fetchPrice(product)
        if price > 0 {
            entry := PriceEntry{
                ProductID: product.ID,
                Price:     price,
                Timestamp: time.Now(),
            }
            resultChan <- entry
        }
    }
}

// fetchPrice simulates fetching price from a URL
// in a real implementation, this would make HTTP requests to scrape or call APIs
func (pt *PriceTracker) fetchPrice(product Product) float64 {
    // simulate network delay
    time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

    // simulate price fetching with random prices
    // in reality, you'd parse HTML or call an API
    basePrice := 100.0
    switch product.ID {
    case "laptop-1":
        basePrice = 1200.0
    case "phone-1":
        basePrice = 800.0
    case "tablet-1":
        basePrice = 500.0
    }

    // add some random variation (±10%)
    variation := (rand.Float64() - 0.5) * 0.2
    price := basePrice * (1 + variation)

    return price
}
