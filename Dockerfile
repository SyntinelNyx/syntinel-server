# Build stage
FROM golang:1.22.5-alpine AS builder
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache openssl

# Copy the entire application code and build the binary
COPY . .

RUN chmod +x ./setup.sh
RUN ./setup.sh

RUN go build -o syntinel-server ./cmd/syntinel-server

# Release stage
FROM alpine:latest AS prod

# Set /app as the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/data/ /app/defaults/
COPY --from=builder /app/syntinel-server /app/syntinel-server
COPY --from=builder /app/internal/database/postgresql/schema.sql /app/postgresql/schema.sql

# Expose the port the application will listen on
EXPOSE 8080

# Start the application in the /app directory
CMD ["sh", "-c", "./docker-start.sh"]
