# ----------------------
# 1. Build stage
# ----------------------
FROM golang:1.25-alpine AS builder

# Install necessary dependencies including build tools for C dependencies
RUN apk add --no-cache git ca-certificates gcc libc-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY . .

# Build binaries
RUN go build -o sso ./cmd/sso

# ----------------------
# 2. Run stage
# ----------------------
FROM alpine:latest

# Download necessary CA certificates (if the application makes HTTPS requests)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/sso /usr/local/bin/sso
# Copy config
COPY config /app/config

# Set the working directory
WORKDIR /app

# Open port (if the application listens)
EXPOSE 8080

# Start command
CMD ["sso"]
