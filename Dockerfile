# Use a minimal base image
FROM alpine:latest

# Copy the gibidify binary into the container
COPY gibidify /usr/local/bin/gibidify

# Ensure the binary is executable
RUN chmod +x /usr/local/bin/gibidify

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gibidify"]

