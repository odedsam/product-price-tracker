package main

import (
	"database/sql"
	"time"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    database := &Database{db: db}
    if err := database.createTables(); err != nil {
        return nil, err
    }

    return database, nil
}

func (d *Database) createTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS products (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            url TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS price_entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            product_id TEXT NOT NULL,
            price REAL NOT NULL,
            timestamp DATETIME NOT NULL,
            FOREIGN KEY (product_id) REFERENCES products (id)
        )`,
        `CREATE INDEX IF NOT EXISTS idx_price_entries_product_id ON price_entries (product_id)`,
        `CREATE INDEX IF NOT EXISTS idx_price_entries_timestamp ON price_entries (timestamp)`,
    }

    for _, query := range queries {
        if _, err := d.db.Exec(query); err != nil {
            return err
        }
    }

    return nil
}

func (d *Database) InsertProduct(product Product) error {
    query := `INSERT OR REPLACE INTO products (id, name, url) VALUES (?, ?, ?)`
    _, err := d.db.Exec(query, product.ID, product.Name, product.URL)
    return err
}

func (d *Database) GetAllProducts() ([]Product, error) {
    query := `SELECT id, name, url FROM products ORDER BY name`
    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var product Product
        if err := rows.Scan(&product.ID, &product.Name, &product.URL); err != nil {
            return nil, err
        }
        products = append(products, product)
    }

    return products, nil
}

func (d *Database) GetProductsWithLatestPrices() ([]ProductWithLatestPrice, error) {
    query := `
        SELECT
            p.id, p.name, p.url,
            pe.price, pe.timestamp
        FROM products p
        LEFT JOIN (
            SELECT DISTINCT product_id,
                   FIRST_VALUE(price) OVER (PARTITION BY product_id ORDER BY timestamp DESC) as price,
                   FIRST_VALUE(timestamp) OVER (PARTITION BY product_id ORDER BY timestamp DESC) as timestamp,
                   ROW_NUMBER() OVER (PARTITION BY product_id ORDER BY timestamp DESC) as rn
            FROM price_entries
        ) pe ON p.id = pe.product_id AND pe.rn = 1
        ORDER BY p.name`

    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []ProductWithLatestPrice
    for rows.Next() {
        var product ProductWithLatestPrice
        var price sql.NullFloat64
        var timestamp sql.NullTime

        if err := rows.Scan(&product.ID, &product.Name, &product.URL, &price, &timestamp); err != nil {
            return nil, err
        }

        if price.Valid {
            product.LatestPrice = &price.Float64
        }
        if timestamp.Valid {
            product.LastUpdated = &timestamp.Time
        }

        products = append(products, product)
    }

    return products, nil
}

func (d *Database) InsertPriceEntry(productID string, price float64, timestamp time.Time) error {
    query := `INSERT INTO price_entries (product_id, price, timestamp) VALUES (?, ?, ?)`
    _, err := d.db.Exec(query, productID, price, timestamp)
    return err
}

func (d *Database) GetPriceHistory(productID string, limit int) ([]PriceEntry, error) {
    query := `
        SELECT id, product_id, price, timestamp
        FROM price_entries
        WHERE product_id = ?
        ORDER BY timestamp DESC
        LIMIT ?`

    rows, err := d.db.Query(query, productID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []PriceEntry
    for rows.Next() {
        var entry PriceEntry
        if err := rows.Scan(&entry.ID, &entry.ProductID, &entry.Price, &entry.Timestamp); err != nil {
            return nil, err
        }
        entries = append(entries, entry)
    }

    return entries, nil
}

func (d *Database) ProductExists(productID string) (bool, error) {
    query := `SELECT COUNT(*) FROM products WHERE id = ?`
    var count int
    err := d.db.QueryRow(query, productID).Scan(&count)
    return count > 0, err
}

func (d *Database) Close() error {
    return d.db.Close()
}
