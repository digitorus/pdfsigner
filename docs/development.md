# Development Guide

## License System

The PDFSigner licensing system provides secure license management with these key components:

- **License**: Contains user information, rate limits, and expiration date
- **Public Key**: Used to validate license authenticity
- **HMAC Key**: Used to secure the storage of license limits

### License Generation

For license generation, use the command-line tool:

