FROM alpine
ADD ca-certificates.crt /etc/ssl/certs/
ADD static /static
ADD config.yaml
ADD pdfsigner /
COPY passwd /etc/passwd
WORKDIR /
USER user
CMD ["./pdfsinger", "serve", "--config", "./config.yaml"]
