# Stage 1: Build the application
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS builder

# Install necessary build tools and strip binaries
RUN apk add --no-cache binutils upx

# Build arguments with default value
ARG VERSION=1.0.0

WORKDIR /app

# Copy go.mod and go.sum to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies and clean cache
RUN go mod download && \
    go mod tidy && \
    rm -rf /go/pkg/mod/cache

# Copy only the necessary source files for better caching
COPY cmd/ ./cmd/
COPY walletapi/ ./walletapi/
COPY transaction/ ./transaction/
COPY rpc/ ./rpc/
COPY globals/ ./globals/
COPY config/ ./config/
COPY cryptography/ ./cryptography/
COPY glue/ ./glue/
COPY block/ ./block/
COPY errormsg/ ./errormsg/
COPY astrobwt/ ./astrobwt/

# Build the wallet CLI with maximum optimization
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -gcflags="-N -l" \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o dero-wallet-cli ./cmd/dero-wallet-cli && \
    strip --strip-all dero-wallet-cli && \
    upx --best --lzma dero-wallet-cli && \
    rm -rf /go/pkg /go/src

# Create user and group files for non-root execution
RUN addgroup -g 1000 dero && \
    adduser -D -s /bin/sh -u 1000 -G dero dero

# Stage 2: Create the final minimal image
FROM scratch

# Declare ARG in final stage for LABEL
ARG VERSION=1.0.0

# Add proper labels to final image
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.description="Dero wallet CLI"
LABEL org.opencontainers.image.source="https://github.com/civilware/derohe"
LABEL maintainer="moralpriest@tutamail.com"

# Copy user/group files from builder stage
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary from builder
COPY --from=builder /app/dero-wallet-cli /dero-wallet-cli

# Health check (if the application supports it)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD /dero-wallet-cli --version || exit 1

# Switch to non-root user
USER dero

# Default command
ENTRYPOINT ["/dero-wallet-cli"]
CMD ["--help"] 