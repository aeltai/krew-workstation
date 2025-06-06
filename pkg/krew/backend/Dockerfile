FROM golang:1.21-alpine AS builder

# Install required packages for building
RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Start fresh from a smaller image
FROM alpine:latest

# Install required packages for runtime
RUN apk add --no-cache curl netcat-openbsd git

# Install kubectl
RUN curl -LO https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Copy the binary from builder
COPY --from=builder /app/main /app/main

WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD nc -z localhost 3000 || exit 1

# Run the application
CMD ["/app/main"] 