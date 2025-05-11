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