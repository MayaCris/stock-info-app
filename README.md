# Stock Info APP - Backend

A high-performance, scalable backend service for stock market data and financial analysis built with Go and Clean Architecture principles.

## 🏗️ Architecture

This backend follows **Clean Architecture** with clear separation of concerns:

```
cmd/
├── api/           # Application entry point and server setup
├── graphql/       # GraphQL implementation (future)
├── test/          # Integration test utilities
└── worker/        # Background job processing (future)

internal/
├── application/   # Application layer (use cases, services, DTOs)
├── domain/        # Domain layer (entities, repositories, business logic)
├── infrastructure/# Infrastructure layer (database, external APIs, config)
└── presentation/  # Presentation layer (REST handlers, middleware)

pkg/               # Shared utilities and constants
scripts/           # Database scripts and operational utilities
test/              # Test files organized by type
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24.4+
- CockroachDB (cloud or local)
- Redis (optional, for caching)
- Valid API keys for Finnhub and Alpha Vantage

### Environment Setup

**Clone and setup:**
   ```bash
   git clone https://github.com/MayaCris/stock-info-app
   cd stock-info-app/backend
   ```


### Running the Application

```bash
# Validate configuration
go run cmd/api/main.go -config-check

# Test setup without starting server
go run cmd/api/main.go -dry-run

# Start the server
go run cmd/api/main.go

# Build for production
go build -o bin/stock-api cmd/api/main.go
./bin/stock-api
```

### Available Commands

```bash
# Development
go run cmd/api/main.go           # Start server
go run cmd/api/main.go -help     # Show help
go run cmd/api/main.go -version  # Show version info
go run cmd/api/main.go -config-check  # Validate config
go run cmd/api/main.go -dry-run  # Test setup

# Testing
go test ./...                    # Run all tests
go test ./test/unit/...          # Unit tests only
go test ./test/integration/...   # Integration tests only
go test -v -cover ./...          # Tests with coverage
```

## 📊 Key Features

### Data Providers Integration
- **Primary (Finnhub):** Real-time market data, company profiles, financial metrics
- **Secondary (Alpha Vantage):** Historical data, technical indicators, advanced analytics
- **Fallback Strategy:** Automatic provider switching on API failures

### Comprehensive Logging System
- **Application Logger:** General application events and errors
- **Server Logger:** HTTP request/response logging with performance metrics
- **Population Logger:** Database population and data integrity operations
- **Structured Logging:** JSON/text formats with contextual fields

### Advanced Middleware Stack
- **Security:** CORS, rate limiting, request ID generation
- **Monitoring:** Performance metrics, slow request tracking
- **Error Handling:** Standardized error responses with correlation IDs
- **Recovery:** Advanced panic recovery with detailed error context

### Database Features
- **CockroachDB Integration:** Distributed SQL database with ACID compliance
- **Connection Pooling:** Optimized connection management
- **Data Integrity:** Automated validation and repair utilities
- **Migration Support:** Schema versioning and rollback capabilities

## 🔌 API Endpoints

### Core Endpoints
```
GET  /                           # API information and health
GET  /health                     # Detailed health status
GET  /swagger/                   # API documentation (debug mode)
```

### Market Data API (v1)
```
GET  /api/v1/stocks/{symbol}              # Stock information
GET  /api/v1/companies/{symbol}           # Company details
GET  /api/v1/market-data/quote/{symbol}   # Real-time market data
GET  /api/v1/analysis/companies/{id}      # Financial analysis
```

### Alpha Vantage Integration
```
GET  /api/v1/alpha-vantage/historical/{symbol}    # Historical data
GET  /api/v1/alpha-vantage/indicators/{symbol}    # Technical indicators
```

### Response Format
```json
{
  "success": true,
  "data": { ... },
  "request_id": "uuid-string",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 🗄️ Database Schema

### Core Entities
- **companies:** Company profiles and basic information
- **financial_metrics:** Financial ratios and performance metrics
- **historical_data:** Time series market data
- **market_data:** Real-time market information
- **stock_ratings:** Analyst ratings and recommendations
- **technical_indicators:** Technical analysis data


## 🛠️ Configuration

### Application Configuration Structure
```go
type Config struct {
    App           AppConfig           // Application metadata
    Server        ServerConfig        // HTTP server settings
    Database      DatabaseConfig      // Database connection
    Cache         CacheConfig         // Redis configuration
    External      ExternalConfig      // API providers
    Security      SecurityConfig      // Security settings
    Logging       LoggingConfig       // Logging configuration
}
```

## 🧪 Testing

### Test Organization
```
test/
├── fixtures/      # Test data and fixtures
├── integration/   # Integration tests
└── unit/          # Unit tests
```

### Running Tests
```bash
# All tests
go test ./...

# Specific test suites
go test ./test/unit/...
go test ./test/integration/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

**Note:** Some integration tests require database connectivity and may fail in isolated environments.

## 🔒 Security Features

- **Input Validation:** Using go-playground/validator for request validation
- **SQL Injection Prevention:** Parameterized queries with GORM
- **API Key Management:** Environment-based configuration
- **Rate Limiting:** Configurable request limits per client
- **CORS Protection:** Flexible cross-origin resource sharing
- **Request Tracing:** Unique request IDs for audit trails

## 📈 Performance

### Optimization Features
- **Connection Pooling:** Database and HTTP client connection management
- **Concurrent Processing:** Goroutine-based request handling
- **Memory Efficiency:** Streaming data processing for large datasets
- **Caching Strategy:** Multi-level caching with TTL management
- **Background Jobs:** Asynchronous processing for heavy operations

### Monitoring
- **Health Checks:** Comprehensive system health monitoring
- **Performance Metrics:** Request duration and throughput tracking
- **Resource Monitoring:** Database connection and memory usage
- **Error Tracking:** Structured error logging with context

## 🚀 Deployment

### Build Production Binary
```bash
go build -o bin/stock-api cmd/api/main.go
```


### Environment-Specific Configuration
- **Development:** Debug logging, Swagger UI, profiling endpoints
- **Staging:** Production-like settings with enhanced logging
- **Production:** Optimized performance, security hardening, minimal logging

## 🔧 Development

### Code Quality
- **Clean Architecture:** Strict layer separation and dependency inversion
- **SOLID Principles:** Single responsibility, open/closed, dependency inversion
- **Error Handling:** Comprehensive error wrapping and context preservation
- **Code Documentation:** Extensive inline documentation and Swagger annotations

### Contributing Guidelines
1. Follow Clean Architecture principles
2. Maintain high test coverage
3. Use structured logging throughout
4. Validate all external inputs
5. Handle errors gracefully with context

## 📚 Documentation

- **API Documentation:** Available at `/swagger/` in debug mode
- **Architecture Decisions:** Recorded in domain and application layers


## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Support

For issues, questions, or contributions, please refer to the project's issue tracker and documentation.