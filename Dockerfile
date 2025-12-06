# Build stage - builds the binary for the target architecture
FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS builder

# Build arguments automatically set by buildx
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary for the target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
	go build -ldflags="-s -w" -o gibidify .

# Runtime stage - minimal image with the binary
FROM alpine:3.23.0

# Install ca-certificates for HTTPS and create non-root user
# hadolint ignore=DL3018
# kics-scan ignore-line
RUN apk add --no-cache ca-certificates && \
	adduser -D -s /bin/sh gibidify

# Copy the binary from builder
COPY --from=builder /build/gibidify /usr/local/bin/gibidify

# Use non-root user
USER gibidify

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gibidify"]
