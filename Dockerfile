# Build stage
FROM golang:1.25.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc g++ musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both binaries
RUN go build -o /app/bin/process_raid ./cmd/process_raid
RUN go build -o /app/bin/update_from_schaledb ./cmd/update_from_schaledb

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/process_raid /app/bin/process_raid
COPY --from=builder /app/bin/update_from_schaledb /app/bin/update_from_schaledb

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Set timezone to Asia/Seoul (optional, adjust as needed)
ENV TZ=Asia/Seoul

ENTRYPOINT ["/app/docker-entrypoint.sh"]
