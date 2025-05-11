# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with explicit architecture settings
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

# Final stage
FROM amazon/aws-lambda-provided:al2

# Copy the binary from builder
COPY --from=builder /app/main ${LAMBDA_RUNTIME_DIR}/main

# Set the CMD to your handler
CMD [ "main" ]