# PDFSigner

A comprehensive PDF digital signature solution that supports multiple signing methods, automated workflows, and API integration.

[![Build & Test](https://github.com/digitorus/pdfsigner/workflows/Build%20&%20Test/badge.svg)](https://github.com/digitorus/pdfsigner/actions/workflows/go.yml)
[![GolangCI-Lint](https://github.com/digitorus/pdfsigner/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/digitorus/pdfsigner/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/digitorus/pdfsigner/workflows/CodeQL/badge.svg)](https://github.com/digitorus/pdfsigner/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/digitorus/pdfsigner)](https://goreportcard.com/report/github.com/digitorus/pdfsigner)
[![Coverage Status](https://codecov.io/gh/digitorus/pdfsigner/graph/badge.svg?token=SylidcS2uJ)](https://codecov.io/gh/digitorus/pdfsigner)
[![Go Reference](https://pkg.go.dev/badge/github.com/digitorus/pdfsigner.svg)](https://pkg.go.dev/github.com/digitorus/pdfsigner)

## Overview

PDFSigner is a robust application written in Go that provides:

- Digital signature creation and verification for PDF documents
- Multiple signing methods (PEM, PKCS#11) 
- Automated folder watching for batch processing
- RESTful API for remote operations
- Multiple concurrent signing services
- Configurable signature workflows

## Quick Start

1. [Install the license](docs/license.md)
2. Choose your preferred signing method:
   - [Command line tool](docs/command-line-signer.md)
   - [Watch folder automation](docs/watch-and-sign.md)
   - [Web API integration](docs/web-api.md)

## Documentation

Full documentation covering all features is available in the [docs](./docs/) directory.

## Development Status

This project is under active development. While core functionality is stable, APIs and features may evolve. We welcome bug reports, contributions and suggestions.

## License

Dual licensed under [GNU](LICENSE.md) and Commercial licenses.
For commercial licensing, please contact us at https://digitorus.com.
