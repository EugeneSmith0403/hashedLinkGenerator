# Advanced HTTP Go Project

A production-ready URL shortener service built with Go, featuring JWT authentication, click statistics, and event-driven architecture.

## Features

- **URL Shortening**: Create and manage shortened links
- **JWT Authentication**: Secure user authentication and authorization
- **Click Statistics**: Track and analyze link usage with event-driven statistics collection
- **RESTful API**: Clean HTTP API design
- **Middleware Support**: CORS, logging, and authentication middleware
- **PostgreSQL Database**: Robust data persistence with GORM
- **Docker Support**: Easy deployment with Docker Compose
- **Validation**: Request validation using go-playground/validator
- **Event Bus**: Asynchronous event handling for statistics

## Tech Stack

- **Go** 1.25.5
- **PostgreSQL** with GORM
- **JWT** (golang-jwt/jwt)
- **Docker & Docker Compose**
- **Validator** (go-playground/validator)

## Project Structure

```
.
├── cmd/                    # Application entry point
│   └── main.go            # Main application
├── configs/               # Configuration management
├── internal/              # Private application code
│   ├── auth/             # Authentication logic
│   ├── jwt/              # JWT service
│   ├── link/             # Link shortening logic
│   ├── stats/            # Statistics tracking
│   ├── user/             # User management
│   └── models/           # Database models
├── pkg/                   # Public packages
│   ├── db/               # Database connection
│   ├── event/            # Event bus implementation
│   ├── middleware/       # HTTP middlewares
│   ├── request/          # Request handling utilities
│   └── response/         # Response utilities
├── migrations/           # Database migrations
├── docker-compose.yml    # Docker configuration
└── .env.example         # Environment variables template
```

## Getting Started

### Prerequisites

- Go 1.25.5 or higher
- PostgreSQL (or use Docker)
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`:
```env
DSN="host=localhost user=postgres password=my_pass dbname=your_db port=5432 sslmode=disable"
TOKEN="your_secret_token_here"
```

4. Create the test environment file in the `cmd` directory:
```bash
cat > cmd/.env.test << 'EOF'
DSN="host=localhost user=postgres_test password=my_pass dbname=test_link port=5433 sslmode=disable"
TOKEN="test_secret"
EOF
```

### Running with Docker

1. Start the application with Docker Compose:
```bash
docker-compose up -d
```

2. The API will be available at `http://localhost:8081`

### Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL (if not using Docker)

3. Run the application:
```bash
go run cmd/main.go
```

The server will start on port `8081`.

## Testing

Run the test suite:
```bash
./test.sh
```

Or run tests manually:
```bash
go test ./...
```

For integration tests with Docker:
```bash
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## API Endpoints

### Authentication
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login and receive JWT token

### Links
- `POST /links` - Create a shortened link (requires authentication)
- `GET /links` - Get user's links (requires authentication)
- `GET /{shortCode}` - Redirect to original URL

### Statistics
- `GET /stats/{linkId}` - Get link statistics (requires authentication)

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DSN` | PostgreSQL connection string | `host=localhost user=postgres password=pass dbname=db port=5432 sslmode=disable` |
| `TOKEN` | JWT secret key | `your_secret_token_here` |

## Development

### Code Structure

The project follows a clean architecture pattern:

- **cmd/**: Application entry points
- **internal/**: Domain logic (not importable by external projects)
- **pkg/**: Reusable packages (can be imported by external projects)
- **configs/**: Configuration loading and management

### Key Components

- **Event Bus**: Asynchronous event handling for decoupled statistics tracking
- **Middleware Chain**: Composable HTTP middleware for cross-cutting concerns
- **Repository Pattern**: Data access abstraction layer
- **Dependency Injection**: Manual DI for better testability

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.
