# =============================================================================
# Multi-stage Dockerfile for Go Cats API
# Optimized for security, size, and performance
# =============================================================================

# Stage 1: Build Environment
FROM golang:1.23-alpine AS builder

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

# Install build dependencies and security updates
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    && apk upgrade --no-cache

# Create non-root user for building
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod ./
COPY go.sum* ./

# Download dependencies with verification
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -a -installsuffix cgo \
    -o backend .

# Run tests during build (fail fast if tests fail)
RUN go test -v ./...

# Stage 2: Security Scanner (Optional - can be disabled for faster builds)
FROM alpine:latest AS security-scanner

# Install security scanning tools
RUN apk add --no-cache \
    ca-certificates \
    curl

# Copy binary for scanning
COPY --from=builder /build/backend /tmp/backend

# Basic security checks (can be extended)
RUN echo "Security scan placeholder - binary size:" && \
    ls -la /tmp/backend

# Stage 3: Runtime Environment
FROM scratch AS runtime

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /build/backend /backend

# Copy static assets if they exist
COPY --from=builder /build/swagger-ui /swagger-ui
COPY --from=builder /build/openapi.yml /openapi.yml

# Use non-root user
USER appuser

# Note: No health check in scratch container due to lack of tools
# GitHub Actions services will check port availability instead

# Expose port
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/backend"]

# =============================================================================
# Alternative: Distroless Runtime (Uncomment to use instead of scratch)
# =============================================================================

# Stage 3 Alternative: Distroless Runtime (More secure than Alpine, larger than scratch)
FROM gcr.io/distroless/static:nonroot AS distroless-runtime

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/backend /backend

# Copy static assets
COPY --from=builder /build/swagger-ui /swagger-ui
COPY --from=builder /build/openapi.yml /openapi.yml

# Expose port
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/backend"]

# =============================================================================
# Development Stage (For development with hot reload)
# =============================================================================

FROM golang:1.23-alpine AS development

# Install development tools
RUN apk add --no-cache \
    ca-certificates \
    curl \
    git \
    make \
    tzdata

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Create app directory
WORKDIR /app

# Copy dependency files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Default command for development
CMD ["air", "-c", ".air.toml"]