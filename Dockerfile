# Build stage
FROM golang:1.25.5-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both binaries
RUN go build -o /app/bin/process_raid ./cmd/process_raid
RUN go build -o /app/bin/update_from_schaledb ./cmd/update_from_schaledb
RUN go build -o /app/bin/total_analysis ./cmd/total_analysis

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/process_raid /app/bin/process_raid
COPY --from=builder /app/bin/update_from_schaledb /app/bin/update_from_schaledb
COPY --from=builder /app/bin/total_analysis /app/bin/total_analysis

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Set timezone to Asia/Seoul (optional, adjust as needed)
ENV TZ=Asia/Seoul

ENTRYPOINT ["/app/docker-entrypoint.sh"]
