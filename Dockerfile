# Build stage: compile from source so `docker build .` works from a plain
# checkout. The previous single-stage file COPYed a prebuilt `gibidify` that
# .dockerignore excludes, so every build failed. Release images are built by
# GoReleaser from Dockerfile.release.
FROM golang:1.27.3-alpine3.24@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gibidify . \
	&& ./scripts/third-party-notices.sh /out/THIRD_PARTY_NOTICES

# Use a minimal base image
FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# Add unprivileged user with a fixed UID (Alpine uses BusyBox adduser), so the
# runtime user resolves the same way on every host (hadolint DL3066)
RUN adduser -D -u 10001 -s /bin/sh gibidify

COPY --from=builder --chmod=0755 /out/gibidify /usr/local/bin/gibidify
# The dependencies' licenses require their notices in binary distributions.
COPY --from=builder /src/LICENSE /out/THIRD_PARTY_NOTICES /usr/share/doc/gibidify/

# CLI image — no service to health-check
HEALTHCHECK NONE

USER 10001

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gibidify"]
