# Variables
BINARY_NAME=tree-api
LAMBDA_ZIP=lambda.zip
GO=go
DOCKER=docker
DOCKER_COMPOSE=docker-compose
TERRAFORM=terraform

# Go related variables
GOFILES?=$$(find . -name '*.go' -not -path "./vendor/*")

# Colors for terminal output
GREEN=\033[0;32m
NC=\033[0m # No Color

.PHONY: all build clean test coverage lint fmt docker-build docker-run terraform-init terraform-plan terraform-apply terraform-destroy dev-start dev-stop localstack-init localstack-restore localstack-logs

all: clean build

# Go commands
build:
	@echo "$(GREEN)Building application...$(NC)"
	$(GO) build -o $(BINARY_NAME) ./cmd/api

build-lambda:
	@echo "$(GREEN)Building Lambda function...$(NC)"
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME) ./cmd/lambda
	zip $(LAMBDA_ZIP) $(BINARY_NAME)
	rm $(BINARY_NAME)

test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v ./tests/...

coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -coverprofile=coverage.out ./tests/...
	$(GO) tool cover -html=coverage.out

lint:
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

# Docker commands
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	$(DOCKER_COMPOSE) build

docker-run:
	@echo "$(GREEN)Running Docker containers...$(NC)"
	$(DOCKER_COMPOSE) up

# Terraform commands
terraform-init:
	@echo "$(GREEN)Initializing Terraform...$(NC)"
	cd terraform && $(TERRAFORM) init

terraform-plan:
	@echo "$(GREEN)Planning Terraform changes...$(NC)"
	cd terraform && $(TERRAFORM) plan

terraform-apply:
	@echo "$(GREEN)Applying Terraform changes...$(NC)"
	cd terraform && $(TERRAFORM) apply

terraform-destroy:
	@echo "$(GREEN)Destroying Terraform resources...$(NC)"
	cd terraform && $(TERRAFORM) destroy

# Development commands
dev-start:
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(DOCKER_COMPOSE) up -d

dev-stop:
	@echo "$(GREEN)Stopping development environment...$(NC)"
	$(DOCKER_COMPOSE) down

# LocalStack commands
localstack-init:
	@echo "$(GREEN)Initializing LocalStack...$(NC)"
	chmod +x scripts/init-localstack.sh
	./scripts/init-localstack.sh

localstack-restore:
	@echo "$(GREEN)Restoring LocalStack state...$(NC)"
	chmod +x scripts/restore-localstack.sh
	./scripts/restore-localstack.sh

localstack-logs:
	@echo "$(GREEN)Showing LocalStack logs...$(NC)"
	$(DOCKER_COMPOSE) logs -f localstack

# Cleanup
clean:
	@echo "$(GREEN)Cleaning up...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f $(LAMBDA_ZIP)
	rm -f coverage.out
	rm -f test/test.db

# Help command
help:
	@echo "$(GREEN)Available commands:$(NC)"
	@echo "  make build          - Build the application"
	@echo "  make build-lambda   - Build the Lambda function"
	@echo "  make test          - Run tests"
	@echo "  make coverage      - Run tests with coverage"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker containers"
	@echo "  make terraform-init    - Initialize Terraform"
	@echo "  make terraform-plan    - Plan Terraform changes"
	@echo "  make terraform-apply   - Apply Terraform changes"
	@echo "  make terraform-destroy - Destroy Terraform resources"
	@echo "  make dev-start     - Start development environment"
	@echo "  make dev-stop      - Stop development environment"
	@echo "  make localstack-init   - Initialize LocalStack resources"
	@echo "  make localstack-restore - Restore LocalStack state"
	@echo "  make localstack-logs   - Show LocalStack logs"
	@echo "  make clean         - Clean up build artifacts" 