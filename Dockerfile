# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build arguments
ARG BUILD_ENV=production
ARG GOOS=linux
ARG GOARCH=amd64

# Build the application
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o /app/tree-api main.go && \
    chmod +x /app/tree-api && \
    ls -la /app/tree-api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/tree-api /app/tree-api

# Copy migrations and other necessary files
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/handlers ./handlers
COPY --from=builder /app/models ./models
COPY --from=builder /app/cache ./cache
COPY --from=builder /app/database ./database

# Verify binary exists and has proper permissions
RUN ls -la /app/tree-api && \
    chmod +x /app/tree-api

# Set environment variables
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/tree-api"] 