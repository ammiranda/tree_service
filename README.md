# Tree Service API

A RESTful API service for managing hierarchical tree structures. The service provides endpoints for creating, reading, updating, and deleting nodes in a tree, with support for pagination and caching.

## Features

- CRUD operations for tree nodes
- Pagination support
- Caching with Redis
- PostgreSQL database backend
- Comprehensive test coverage

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Redis (for caching)
- Docker (optional, for containerized deployment)
- AWS CLI (for AWS deployment)

## Local Development

### Option 1: Using Docker Compose (Recommended)

The easiest way to run the application locally is using Docker Compose, which will set up all required services (PostgreSQL, Redis, and LocalStack) automatically.

1. Start all services:
```bash
make dev-start
```

2. Stop all services:
```bash
make dev-stop
```

3. View logs:
```bash
docker-compose logs -f
```

### Option 2: Manual Setup

### 1. Set up the Database

```bash
# Create the database
createdb tree_service

# Create the test database
createdb tree_service_test
```

### 2. Set up Environment Variables

Create a `.env` file in the root directory:

```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tree_service
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Application configuration
APP_ENV=development
PORT=8080
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Migrations

```bash
go run cmd/migrate/main.go
```

### 5. Start the Server

```bash
go run main.go
```

The server will start on `http://localhost:8080`.

## Running Tests

### 1. Set up Test Environment

Create a `.env.test` file:

```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tree_service_test
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Application configuration
APP_ENV=test
PORT=8081
```

### 2. Run Tests

```bash
# Run all tests
go test ./...

# Run specific test packages
go test ./tests/...
go test ./handlers/...
go test ./repository/...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run a specific test
go test -v ./tests -run TestCache
```

### 3. Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## API Endpoints

### Get Tree
```http
GET /api/tree?page=1&pageSize=10
```
Retrieves the tree structure with pagination support.

Query Parameters:
- `page` (optional): Page number (default: 1)
- `pageSize` (optional): Items per page (default: 10, max: 100)

Response:
```json
{
  "data": [
    {
      "id": 1,
      "label": "root",
      "parentId": null,
      "children": [
        {
          "id": 2,
          "label": "child",
          "parentId": 1,
          "children": []
        }
      ]
    }
  ],
  "pagination": {
    "total": 2,
    "page": 1,
    "pageSize": 10,
    "hasNext": false,
    "hasPrev": false
  }
}
```

### Create Node
```http
POST /api/tree
```
Creates a new node in the tree.

Request Body:
```json
{
  "label": "new node",
  "parentId": 1  // optional
}
```

Response:
```json
{
  "id": 3,
  "label": "new node",
  "parentId": 1
}
```

### Update Node
```http
PUT /api/tree/:id
```
Updates an existing node.

Request Body:
```json
{
  "label": "updated node",
  "parentId": 2  // optional
}
```

Response:
```json
{
  "id": 3,
  "label": "updated node",
  "parentId": 2
}
```

### Delete Node
```http
DELETE /api/tree/:id
```
Deletes a node and all its children.

Response: 204 No Content

## Project Structure
```
.
├── cache/           # Cache implementations
├── cmd/            # Command-line tools
├── config/         # Configuration management
├── handlers/       # HTTP request handlers
├── migrations/     # Database migrations
├── models/         # Data models
├── repository/     # Database repositories
└── tests/          # Test suites
```

## Development

### Adding New Features

1. Create a new branch for your feature
2. Write tests first (TDD approach)
3. Implement the feature
4. Run all tests to ensure nothing is broken
5. Submit a pull request

### Code Style

- Follow Go's standard formatting: `go fmt ./...`
- Run linter: `golangci-lint run`
- Write tests for new features
- Update documentation as needed

### Debugging

- Use `go run main.go -debug` for debug logging
- Check logs in `logs/` directory
- Use `go test -v` for verbose test output

## Development Commands

The project includes a Makefile with several useful commands:

### Build Commands
```bash
make build          # Build the application
make build-lambda   # Build the Lambda function
```

### Testing Commands
```bash
make test          # Run tests
make coverage      # Run tests with coverage
make lint          # Run linter
make fmt           # Format code
```

### Docker Commands
```bash
make docker-build  # Build Docker image
make docker-run    # Run Docker containers
```

### Development Environment
```bash
make dev-start     # Start development environment
make dev-stop      # Stop development environment
```

### LocalStack Commands
```bash
make localstack-init    # Initialize LocalStack resources
make localstack-restore # Restore LocalStack state
make localstack-logs    # Show LocalStack logs
```

### Cleanup
```bash
make clean         # Clean up build artifacts
```

For a complete list of available commands:
```bash
make help
```

## Docker Compose Services

The `docker-compose.yml` file sets up the following services:

- **app**: The main application service
  - Port: 8080
  - Environment variables configured for development
  - Depends on PostgreSQL, Redis, and LocalStack

- **postgres**: PostgreSQL database
  - Port: 5432
  - Database: tree_db
  - User: postgres
  - Password: postgres
  - Persistent volume for data storage

- **redis**: Redis cache
  - Port: 6380 (mapped to container port 6379)
  - Persistent volume for data storage

- **localstack**: AWS service emulator
  - Port: 4566
  - Emulates AWS services (RDS, ElastiCache, Secrets Manager, Lambda, API Gateway)
  - Persistent volume for state storage

All services are connected through a bridge network named `app-network`.

## Deployment

### Docker
Build and run with Docker:
```bash
docker build -t tree-service .
docker run -p 8080:8080 tree-service
```

### AWS Lambda
The service can be deployed as an AWS Lambda function. See `cmd/lambda/main.go` for details.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Security

### Sensitive Files

The following types of files should never be committed to the repository:

- SSL/TLS certificates and keys (`.pem`, `.key`, `.crt`)
- SSH keys
- Environment files (`.env`, `.env.*`)
- AWS credentials (`aws.json`, `aws-credentials.json`)
- Any other files containing secrets or credentials

These files are already excluded in `.gitignore`, but it's important to be aware of this policy.

### Local Development State

The `volume/` directory contains local development state and should never be committed to the repository. This includes:
- LocalStack state and cache
- Temporary files
- Local development certificates
- Machine-specific configurations

This directory is automatically created by Docker Compose and LocalStack, and is excluded from version control.

### Handling Sensitive Files

1. For local development:
   - Create a `.env` file based on `.env.example`
   - Store certificates and keys in a secure location outside the repository
   - Use AWS CLI profiles for AWS credentials

2. For production:
   - Use a secrets management service (AWS Secrets Manager, HashiCorp Vault)
   - Use environment variables in your deployment platform
   - Use IAM roles and policies for AWS services

3. For CI/CD:
   - Use encrypted secrets in your CI/CD platform
   - Never log or expose sensitive values
   - Use temporary credentials when possible 