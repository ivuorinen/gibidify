# Use a minimal base image
FROM alpine:3.24.1

# Add unprivileged user (Alpine uses BusyBox adduser)
RUN adduser -D -s /bin/sh gibidify

# Copy the gibidify binary into the container with executable permissions
COPY --chmod=0755 gibidify /usr/local/bin/gibidify

# CLI image — no service to health-check
HEALTHCHECK NONE

# Use the new user
USER gibidify

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gibidify"]
