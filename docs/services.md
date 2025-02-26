# (Multiple) Services Mode

The services mode allows running multiple PDFSigner operations concurrently, combining watch folders and API endpoints in a single process.

## Usage

Basic command: `pdfsigner services`

Required configuration:

```sh
--config path/to/config/file
```

### Examples

Run all configured services:

```sh
pdfsigner services
```

Run specific services:

```sh
pdfsigner services service1 service2
```

## Configuration

Services are defined in the configuration file. 

Each service can be:
- Watch folder monitor
- Web API endpoint
- Combination of both

[See configuration documentation](configuration.md) for detailed setup instructions.


