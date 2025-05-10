# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Final stage
FROM public.ecr.aws/lambda/provided:al2

# Copy the binary from builder
COPY --from=builder /app/main ${LAMBDA_RUNTIME_DIR}/bootstrap

# Set the CMD to your handler
CMD [ "bootstrap" ] 