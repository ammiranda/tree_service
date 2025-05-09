# Tree Service API

A RESTful API service for managing hierarchical tree structures. The service provides endpoints for creating, reading, updating, and deleting nodes in a tree, with support for pagination and caching.

## Features

- CRUD operations for tree nodes
- Pagination support
- Caching with DynamoDB (optional)
- PostgreSQL database backend
- Comprehensive test coverage

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Docker (optional, for containerized deployment)
- AWS CLI (optional, for DynamoDB cache)

## Local Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd tree-service
```

2. Set up environment variables:
```bash
# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=tree_service
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_SSL_MODE=disable

# AWS configuration (if using DynamoDB cache)
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=your_region
```

3. Create the database:
```bash
createdb tree_service
```

4. Install dependencies:
```bash
go mod download
```

5. Run migrations:
```bash
go run cmd/migrate/main.go
```

6. Start the server:
```bash
go run main.go
```

The server will start on `http://localhost:8080`.

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

## Running Tests

1. Set up test environment:
```bash
export APP_ENV=test
export DB_NAME=tree_service_test
```

2. Run all tests:
```bash
go test ./...
```

3. Run specific test packages:
```bash
go test ./tests/...
go test ./handlers/...
```

4. Run with coverage:
```bash
go test -cover ./...
```

## Development

### Project Structure
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

### Adding New Features

1. Create a new branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and run tests:
```bash
go test ./...
```

3. Submit a pull request

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