# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

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
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o tree-api main.go && \
    chmod +x tree-api && \
    ls -la tree-api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata postgresql-client

# Copy the binary from builder
COPY --from=builder /build/tree-api /app/tree-api

# Copy migrations
COPY --from=builder /build/migrations /app/migrations

# Add wait-for-it script
COPY --from=builder /build/scripts/wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh

# Verify binary exists and has proper permissions
RUN ls -la /app/tree-api && \
    chmod +x /app/tree-api && \
    ls -la /app/migrations

# Set environment variables
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Run the application with wait-for-it script
CMD ["/bin/sh", "-c", "/app/wait-for-it.sh postgres:5432 -- /app/tree-api"] 