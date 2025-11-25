# Multi-stage build for all services
# Usage: docker build --build-arg SERVICE=api-gateway -t onichange/api-gateway:latest .

ARG SERVICE=api-gateway
ARG GO_VERSION=1.24
ARG BUILD_REF=unknown
ARG BUILD_DATE=unknown

# Stage 1: Builder
FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    make

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG SERVICE
ARG BUILD_REF
ARG BUILD_DATE

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s \
        -X 'main.Version=${BUILD_REF}' \
        -X 'main.BuildDate=${BUILD_DATE}' \
        -X 'main.Service=${SERVICE}' \
        -extldflags '-static'" \
    -trimpath \
    -o /app/${SERVICE} \
    ./cmd/${SERVICE}

# Stage 2: Runtime (Distroless for security)
FROM gcr.io/distroless/static:nonroot

# Metadata
LABEL org.opencontainers.image.title="OniChange ${SERVICE}"
LABEL org.opencontainers.image.description="High-Performance & Secure Omni-Channel POS System - ${SERVICE}"
LABEL org.opencontainers.image.vendor="OniChange"
LABEL org.opencontainers.image.version="${BUILD_REF}"

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
ARG SERVICE
COPY --from=builder /app/${SERVICE} /app

# Use non-root user (UID 1000)
USER nonroot:nonroot

# Expose port (will be overridden in docker-compose)
EXPOSE 8080

# Health check is configured in docker-compose.yml
# (distroless image doesn't have shell, so HTTP healthcheck is used)

# Run the application
ENTRYPOINT ["/app"]

