# syntax=docker/dockerfile:1

# ------------------------------------------------------------------------------
# Builder stage: build the fractal-engine binary with CGO enabled
# ------------------------------------------------------------------------------
FROM golang:bookworm AS builder

# Enable CGO (required for github.com/mattn/go-sqlite3)
ENV CGO_ENABLED=1 \
    GO111MODULE=on

WORKDIR /src

# Install build deps for CGO builds (gcc, libc headers, etc.)
RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Cache module downloads first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the application binary
# Target: cmd/fractal-engine/fractal_engine.go
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags "-s -w" -o /out/fractal-engine ./cmd/fractal-engine

# ------------------------------------------------------------------------------
# Runtime stage: minimal image to run the binary
# ------------------------------------------------------------------------------
FROM debian:bookworm-slim AS runtime

# Install runtime dependencies: certificates, tzdata, curl for healthchecks
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r app && useradd -r -g app -d /app app
WORKDIR /app

# Copy the built binary
COPY --from=builder /out/fractal-engine /usr/local/bin/fractal-engine

# Own the workdir
RUN chown -R app:app /app

USER app

# Expose default RPC port (configurable via env/flags)
EXPOSE 8891

# Optional healthcheck: assumes /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8891/health || exit 1

# Default command. You can override/add flags with `docker run ... fractal-engine --flag value`
ENTRYPOINT ["/usr/local/bin/fractal-engine"]
