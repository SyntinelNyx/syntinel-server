# Build stage
FROM golang:1.22.5-alpine AS builder
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application code and build the binary
COPY . .
RUN go build -o syntinel-server ./cmd/syntinel

# Release stage
FROM alpine:latest AS prod

# Copy the built binary from the builder stage
COPY --from=builder /app/syntinel-server /

# Expose the port the application will listen on
EXPOSE 8080

CMD ["/syntinel-server", "-e", "production", "-p", "8080"]
