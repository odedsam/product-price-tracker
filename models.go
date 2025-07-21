package main

import (
	"time"
)

// Product represents a product to track
type Product struct {
    ID   string `json:"id" db:"id"`
    Name string `json:"name" db:"name"`
    URL  string `json:"url" db:"url"`
}

// PriceEntry represents a price data point
type PriceEntry struct {
    ID        int       `json:"id" db:"id"`
    ProductID string    `json:"product_id" db:"product_id"`
    Price     float64   `json:"price" db:"price"`
    Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ProductWithLatestPrice combines product info with its latest price
type ProductWithLatestPrice struct {
    Product
    LatestPrice *float64   `json:"latest_price,omitempty"`
    LastUpdated *time.Time `json:"last_updated,omitempty"`
}
