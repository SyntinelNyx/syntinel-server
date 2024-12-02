# Build stage
FROM golang:1.22.5-alpine AS builder
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application code and build the binary
COPY . .
RUN go build -o syntinel-server ./cmd/syntinel-server

# Release stage
FROM alpine:latest AS prod

# Set /app as the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/syntinel-server ./syntinel-server
COPY --from=builder /app/config.example.yaml ./config.yaml
COPY --from=builder /app/internal/database/postgresql/schema.sql ./postgresql/schema.sql

# Expose the port the application will listen on
EXPOSE 8080

# Start the application in the /app directory
CMD ["./syntinel-server", "-e", "production", "-p", "8080"]APP_ENV=production
APP_PORT=8080
TLS_CERT_PATH=/path/to/server.crt
TLS_KEY_PATH=/path/to/server.key

REDIS_URL=localhost:6379
DATABASE_URL=postgres://username:password@host:port/database_name

CSRF_SECRET=super_secure_secret
ECDSA_PUBLIC_KEY_PATH=/path/to/ecdsa_public.pem
ECDSA_PRIVATE_KEY_PATH=/path/to/ecdsa_private.pem