# 1 Builder (needs gcc for CGO/sqlite3)
FROM golang:1.24.5-bookworm AS build
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src

# cache modules early
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# app source
COPY . .

# CGO needed for github.com/mattn/go-sqlite3
ENV CGO_ENABLED=1 GOOS=linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" \
      -o /out/job-watcher ./cmd/job-watcher

# 2 Runtime
FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

# non-root user
RUN useradd -r -u 10001 -g users appuser

WORKDIR /app

# binary
COPY --from=build /out/job-watcher /app/job-watcher

# templates and static files (your code loads these from disk)
COPY static/ /app/static/
COPY internal/tmpl/ /app/internal/tmpl/

# mount point for Fly Volume (SQLite lives here)
VOLUME ["/data"]

ENV PORT=8080
EXPOSE 8080
USER appuser
ENTRYPOINT ["/app/job-watcher"]
