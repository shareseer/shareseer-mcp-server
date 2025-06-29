# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with version info and optimizations
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o shareseer-mcp ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1001 -S shareseer && \
    adduser -u 1001 -S shareseer -G shareseer

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/shareseer-mcp .
COPY --from=builder /app/configs ./configs

# Change ownership to non-root user
RUN chown -R shareseer:shareseer /app

# Switch to non-root user
USER shareseer

# Expose port (updated to match config)
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Command to run
CMD ["./shareseer-mcp"]