FROM alpine:3

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install certificates
RUN apk add --no-cache ca-certificates

# Create application directories
RUN mkdir -p /usr/local/bin \
    /etc/pdfsigner \
    /var/lib/pdfsigner \ 
    /var/lib/pdfsigner/input \
    /var/lib/pdfsigner/output

# Copy application files
COPY config.example.yaml /etc/pdfsigner/config.yaml
COPY pdfsigner.lic /etc/pdfsigner/pdfsigner.lic
COPY pdfsigner /usr/local/bin/pdfsigner

# Set permissions and ownership
RUN chown -R appuser:appgroup /etc/pdfsigner /var/lib/pdfsigner
RUN chmod 755 /usr/local/bin/pdfsigner
    
# Define volume for configuration
VOLUME ["/etc/pdfsigner", "/var/lib/pdfsigner/input", "/var/lib/pdfsigner/output"]

WORKDIR /var/lib/pdfsigner

USER appuser

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["pdfsigner", "serve", "--config", "/etc/pdfsigner/config.yaml"]
