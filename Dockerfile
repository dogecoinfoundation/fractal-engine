# syntax=docker/dockerfile:1

# Minimal runtime image that runs the Nix-built fractal-engine binary
# Expectation: Nix build outputs symlink at ./result/bin/fractal-engine

FROM debian:bookworm-slim AS runtime-dist

# Install runtime dependencies: certificates, tzdata, curl for healthchecks
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r app && useradd -r -g app -d /app app
WORKDIR /app

# Copy the prebuilt binary produced by Nix
COPY ./dist/fractal-engine /app/fractal-engine
RUN chmod +x /app/fractal-engine

# Own the workdir
RUN chown -R app:app /app

USER app

# Default environment variables (can be overridden at runtime)
ENV \
  RPC_SERVER_HOST="0.0.0.0" \
  RPC_SERVER_PORT="8891" \
  RPC_API_KEY="" \
  DOGE_NET_NETWORK="tcp" \
  DOGE_NET_ADDRESS="0.0.0.0:8086" \
  DOGE_NET_WEB_ADDRESS="0.0.0.0:8085" \
  EMBED_DOGENET="true" \
  DOGE_SCHEME="http" \
  DOGE_HOST="0.0.0.0" \
  DOGE_PORT="22556" \
  DOGE_USER="test" \
  DOGE_PASSWORD="test" \
  DATABASE_URL="postgres://fractalstore:fractalstore@localhost:5432/fractalstore?sslmode=disable" \
  MIGRATIONS_PATH="db/migrations" \
  PERSIST_FOLLOWER="true" \
  API_RATE_LIMIT_PER_SECOND="10" \
  INVOICE_LIMIT="100" \
  BUY_OFFER_LIMIT="3" \
  SELL_OFFER_LIMIT="3" \
  CORS_ALLOWED_ORIGINS="*"

# Expose default RPC port (can be overridden via RPC_SERVER_PORT env)
EXPOSE 8891

# Optional healthcheck using the configured port
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD sh -c 'curl -fsS "http://127.0.0.1:${RPC_SERVER_PORT:-8891}/health" || exit 1'

# Default command. Override/add flags with `docker run ... fractal-engine --flag value`
ENTRYPOINT ["/app/fractal-engine"]
