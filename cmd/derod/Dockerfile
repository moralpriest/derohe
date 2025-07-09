# Stage 1: Build the application
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS builder

# Install only essential build tools
RUN apk add --no-cache binutils upx

# Build arguments
ARG P2P_PORT=18089
ARG RPC_PORT=9999
ARG GETWORK_PORT=10100
ARG VERSION=1.0.0

WORKDIR /app

# Copy go.mod and go.sum to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies and clean cache
RUN go mod download && \
    go mod tidy && \
    rm -rf /go/pkg/mod/cache

# Copy the rest of the application source code
COPY . .

# Build derod binary with aggressive optimization (but safe)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -gcflags="-N -l" \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o derod ./cmd/derod && \
    strip --strip-all derod && \
    upx --best --lzma derod && \
    rm -rf /go/pkg /go/src

# Stage 2: Create the final minimal image from scratch
FROM scratch

# Re-declare build arguments for the second stage
ARG P2P_PORT=18089
ARG RPC_PORT=9999
ARG GETWORK_PORT=10100
ARG VERSION=1.0.0

# Add proper labels to final image
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.description="Dero daemon (minimal)"
LABEL org.opencontainers.image.source="https://github.com/civilware/derohe"
LABEL maintainer="moralpriest@tutamail.com"

# Copy only the binary (no SSL certificates for minimal size)
COPY --from=builder /app/derod /derod

# Create /data directory path
VOLUME /data

# Expose ports
EXPOSE ${P2P_PORT}/tcp
EXPOSE ${RPC_PORT}/tcp
EXPOSE ${GETWORK_PORT}/tcp

# Health check - test if network node binary is responsive
# Note: For a proper network health check, consider using external monitoring
# that can test the RPC endpoint (port 9999) from outside the container
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD /derod --version || exit 1

# Default command with arguments
ENTRYPOINT ["/derod"]
CMD ["--rpc-bind=0.0.0.0:9999", "--data-dir=/data"] 