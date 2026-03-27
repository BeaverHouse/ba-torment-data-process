# Build stage
FROM golang:1.25.5-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build single binary
RUN go build -o /app/bin/batorment .

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/batorment /app/bin/batorment

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Set timezone to Asia/Seoul (optional, adjust as needed)
ENV TZ=Asia/Seoul

ENTRYPOINT ["/app/docker-entrypoint.sh"]
