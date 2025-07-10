# Stage 1: Build the simulator
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS builder

# Install necessary build tools and strip binaries
RUN apk add --no-cache binutils upx

# Build arguments
ARG RPC_PORT=20000
ARG HTTP_PORT=8080
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

# Build the simulator binary with maximum optimization
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -gcflags="-N -l" \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o simulator ./cmd/simulator && \
    strip --strip-all simulator && \
    upx --best --lzma simulator && \
    rm -rf /go/pkg /go/src

# Create user and group files for non-root execution
RUN addgroup -g 1000 dero && \
    adduser -D -s /bin/sh -u 1000 -G dero dero

# Stage 2: Create the final minimal image from scratch
FROM scratch

# Re-declare build arguments for the second stage
ARG RPC_PORT=20000
ARG HTTP_PORT=8080
ARG VERSION=1.0.0

# Add proper labels to final image
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.description="Dero simulator node"
LABEL org.opencontainers.image.source="https://github.com/civilware/derohe"
LABEL maintainer="moralpriest@tutamail.com"

# Copy user/group files from builder stage
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy only the necessary files
COPY --from=builder /app/simulator /simulator

# Create /data directory path
VOLUME /data

# Expose ports
EXPOSE ${RPC_PORT}/tcp
EXPOSE ${HTTP_PORT}/tcp

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD /simulator --version || exit 1

# Switch to non-root user
USER dero

# Default command with arguments
ENTRYPOINT ["/simulator"]
CMD ["--rpc-bind=0.0.0.0:20000", "--http-address=0.0.0.0:8080", "--data-dir=/data"] 