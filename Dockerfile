# Stage 1: Builder stage - builds Go binary with production optimizations
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    git

# Set up Go workspace
WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build binary with optimizations
# -ldflags: strip debug info, set version from git tags or default to v0.0.0
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o bin/proxy \
      -ldflags="-s -w -X main.Version=$(git describe --tags 2>/dev/null || echo 'v0.0.0')" \
      ./cmd/proxy

# Stage 2: Production stage - minimal distroless image for deployment
FROM gcr.io/distroless/static-debian11

# Copy binary from builder stage
WORKDIR /app
COPY --from=builder /build/bin/proxy ./proxy

# Add Dockerfile and compose files for development (not deployed to production)
COPY Dockerfile docker-compose.yml ./

# Create non-root user for security
RUN addgroup --system --gid 1000 goproxy && \
    adduser --system --uid 1000 --ingroup goproxy proxyuser

# Set permissions
RUN chmod +x ./proxy

# Switch to non-root user
USER proxyuser

# Expose port
EXPOSE 9999

# Health check using the binary itself (no curl dependency)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/proxy --health-check || exit 1

# Run the proxy server
CMD ["./proxy"]
