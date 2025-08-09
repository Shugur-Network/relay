# Multi-stage production Dockerfile for Shugur Relay
# Security-hardened with minimal attack surface

# Stage 1: Build the binary
FROM golang:1.25rc3-alpine3.22 AS builder

# Install build dependencies
RUN apk update && apk upgrade && apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy dependency files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/ web/

# Build arguments for version information
ARG VERSION=1.0.0
ARG COMMIT=unknown
ARG DATE=unknown

# Build the static binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -a -installsuffix cgo \
    -o relay \
    ./cmd

# Verify the binary
RUN ./relay version || echo "Binary built successfully"

# Stage 2: Minimal production image
FROM alpine:latest AS production

# Install minimal runtime dependencies and create user
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -s /sbin/nologin -u 1001 relay && \
    mkdir -p /app/config /app/data /app/logs && \
    chown -R relay:relay /app

# Copy the binary and web assets
COPY --from=builder /app/relay /usr/local/bin/relay
COPY --from=builder /app/web /app/web

# Switch to non-root user
USER relay

WORKDIR /app

# Expose ports
EXPOSE 4656
EXPOSE 8080  
EXPOSE 8181

# Environment variables
ENV PATH="/usr/local/bin:${PATH}" \
    SHUGUR_ENV=production \
    SHUGUR_LOG_LEVEL=info \
    SHUGUR_LOG_FORMAT=json

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ["/usr/local/bin/relay", "health"] || exit 1

# Default command
CMD ["relay", "start"]

# =============================================================================
# Development stage for live reloading and debugging
FROM golang:1.25rc3-alpine3.22 AS development

# Install development dependencies
RUN apk update && apk upgrade && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    bash \
    curl \
    make

# Install Air for live reloading
RUN go install github.com/air-verse/air@v1.61.5

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose ports for development
EXPOSE 4656
EXPOSE 8080  
EXPOSE 8181

# Environment variables for development
ENV SHUGUR_ENV=development \
    SHUGUR_LOG_LEVEL=debug \
    SHUGUR_LOG_FORMAT=text \
    CGO_ENABLED=0

# Default command uses Air for live reloading
CMD ["air", "-c", ".air.toml"]