# Tree API

A simple API for managing a tree structure of nodes.

## Features

- RESTful API endpoints
- PostgreSQL persistence layer
- In-memory and DynamoDB caching options
- AWS Lambda deployment
- Comprehensive test coverage

## Requirements

- Go 1.21 or later
- PostgreSQL
- AWS SDK for Go v2
- AWS Lambda
- AWS Secrets Manager (for RDS credentials)

## Setup

1. Set up your PostgreSQL database
2. Configure AWS credentials
3. Set up AWS Secrets Manager with RDS credentials
4. Install dependencies:
   ```bash
   go mod download
   ```

## Running Tests

```bash
go test ./...
```

## Deployment

The application is designed to be deployed as an AWS Lambda function. The Lambda function will handle API Gateway events and interact with RDS for persistence.

## API Endpoints

- `GET /api/tree` - Get the entire tree structure
- `POST /api/tree` - Create a new node

## Architecture

The application follows a clean architecture pattern with:
- Repository pattern for data access
- Interface-based design for flexibility
- Dependency injection for testability
- Caching layer for performance 