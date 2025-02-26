FROM alpine
ADD ca-certificates.crt /etc/ssl/certs/
ADD static /static
ADD config.toml
ADD pdfsigner /
COPY passwd /etc/passwd
WORKDIR /
USER user
CMD ["./pdfsinger", "serve", "multiple-signers", "--config", "./config.yaml", "--serve-address", "0.0.0.0", "--serve-port", "3000", "simple"]
