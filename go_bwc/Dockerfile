# Multi-stage Dockerfile for Forensic BWC System

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies (if go.mod exists)
RUN if [ -f go.mod ]; then go mod download; fi

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o bwc-system \
    forensic_bwc_system.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 bwc && \
    adduser -D -u 1000 -G bwc bwc

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bwc-system .

# Create storage directory with proper permissions
RUN mkdir -p /app/bwc_storage && \
    chown -R bwc:bwc /app

# Switch to non-root user
USER bwc

# Expose port (if you add API later)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD [ -f /app/bwc-system ] || exit 1

# Run the application
CMD ["./bwc-system"]
