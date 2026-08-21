# syntax=docker/dockerfile:1

# Build stage: compile derod against the vendored dependency tree.
FROM golang:1.26 AS builder
WORKDIR /src

# Copy module metadata first so dependency layers cache across builds.
COPY go.mod vendor/ vendor/
COPY . .

# Static, stripped build - matches the reproducibility flags used in CI
# (-trimpath -ldflags=-buildid=) so release binaries and the image are
# built from the same recipe.
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-buildid=" \
    -o /out/derod ./cmd/derod

# Runtime stage: minimal Debian with CA certs and a non-root user.
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --home-dir /data dero

COPY --from=builder /out/derod /usr/local/bin/derod

VOLUME ["/data"]

# P2P/GETWORK and JSON-RPC ports (see config/config.go defaults).
EXPOSE 10100 10102

USER dero
ENTRYPOINT ["derod"]
CMD ["--data-dir=/data"]
