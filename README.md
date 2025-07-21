# Product Price Tracker

A minimal product price tracker built with Go that periodically fetches product prices and stores them in a SQLite database. The application provides REST API endpoints to retrieve tracked products and their price history.

## Features

- **Concurrent Price Tracking**: Uses goroutines and channels for background price fetching
- **SQLite Database**: Persistent storage for products and price history
- **REST API**: HTTP endpoints to list products and retrieve price history
- **Thread-Safe**: Uses sync.RWMutex for safe concurrent access
- **Worker Pool**: Efficient concurrent processing of multiple products
- **Graceful Shutdown**: Clean shutdown handling with context cancellation

## Project Structure

```
price-tracker/
├── go.mod
├── main.go          # Application entry point
├── models.go        # Data structures
├── database.go      # SQLite database operations
├── tracker.go       # Price tracking logic with concurrency
├── api.go          # HTTP server and REST API endpoints
└── README.md       # This file
```

## Prerequisites

- Go 1.21 or later
- SQLite3 (usually pre-installed on most systems)

## Installation & Setup

1. **Clone or create the project:**
   ```bash
   mkdir price-tracker
   cd price-tracker
   ```

2. **Initialize Go module:**
   ```bash
   go mod init price-tracker
   ```

3. **Copy all the source files** (main.go, models.go, database.go, tracker.go, api.go)

4. **Install dependencies:**
   ```bash
   go mod tidy
   ```

## Running the Application

1. **Start the application:**
   ```bash
   go run .
   ```

2. **The application will:**
   - Create a SQLite database (`prices.db`) in the current directory
   - Add sample products (laptop, phone, tablet)
   - Start background price tracking (every 30 seconds)
   - Start HTTP server on port 8080

3. **Access the application:**
   - Web interface: http://localhost:8080
   - API endpoints: http://localhost:8080/api/v1/

## API Endpoints

### 1. List All Products
```
GET /api/v1/products
```
Returns all tracked products with their latest prices.

**Example Response:**
```json
[
  {
    "id": "laptop-1",
    "name": "Gaming Laptop",
    "url": "https://example.com/laptop-1",
    "latest_price": 1184.50,
    "last_updated": "2025-07-21T10:30:00Z"
  }
]
```

### 2. Get Price History
```
GET /api/v1/products/{id}/history?limit=50
```
Returns price history for a specific product.

**Parameters:**
- `limit` (optional): Number of records to return (default: 50)

**Example Response:**
```json
{
  "product_id": "laptop-1",
  "count": 5,
  "history": [
    {
      "id": 1,
      "product_id": "laptop-1",
      "price": 1184.50,
      "timestamp": "2025-07-21T10:30:00Z"
    }
  ]
}
```

### 3. Health Check
```
GET /api/v1/health
```
Returns application health status.

## Architecture & Concurrency

### Concurrency Features

1. **Background Price Tracking**:
   - Uses goroutines to run price tracking in the background
   - Context-based cancellation for clean shutdown

2. **Worker Pool Pattern**:
   - Multiple goroutines process products concurrently
   - Channels coordinate work distribution and result collection

3. **Thread-Safe Data Access**:
   - `sync.RWMutex` protects concurrent access to product map
   - Database operations are naturally thread-safe with SQLite

4. **Graceful Shutdown**:
   - Signal handling for clean application termination
   - HTTP server graceful shutdown with timeout

### Key Components

- **PriceTracker**: Core tracking logic with concurrent workers
- **Database**: SQLite operations with proper indexing
- **APIServer**: HTTP server with middleware and CORS support
- **Models**: Clean data structures for products and price entries

## Configuration

You can modify these settings in `main.go`:

- **Tracking Interval**: Change `30*time.Second` to adjust price checking frequency
- **Server Port**: Modify `:8080` to use a different port
- **Worker Count**: Adjust `numWorkers` in `trackAllProducts()` method
- **Database Path**: Change `prices.db` to use a different database file

## Simulated Price Fetching

The current implementation simulates price fetching with random variations. In a real-world scenario, you would:

1. Make HTTP requests to actual product URLs
2. Parse HTML content or call APIs
3. Extract price information using selectors or JSON parsing
4. Handle rate limiting and error cases

Replace the `fetchPrice()` method in `tracker.go` with actual web scraping or API calls.

## Database Schema

### Products Table
```sql
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Price Entries Table
```sql
CREATE TABLE price_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id TEXT NOT NULL,
    price REAL NOT NULL,
    timestamp DATETIME NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products (id)
);
```

## Example Usage

After starting the application, you can:

1. **View the web interface** at http://localhost:8080
2. **List all products**: `curl http://localhost:8080/api/v1/products`
3. **Get price history**: `curl http://localhost:8080/api/v1/products/laptop-1/history`
4. **Check health**: `curl http://localhost:8080/api/v1/health`

## Building for Production

```bash
# Build binary
go build -o price-tracker .

# Run binary
./price-tracker
```

## Notes

- The application creates sample products on startup for demonstration
- Price tracking starts automatically with 30-second intervals
- All data persists in the SQLite database between runs
- Logs provide visibility into tracking operations and API requests
