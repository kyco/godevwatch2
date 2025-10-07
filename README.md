# godevwatch

A development proxy tool written in Go.

## Installation

```bash
go install github.com/kyco/godevwatch@latest
```

Or build from source:

```bash
git clone https://github.com/kyco/godevwatch.git
cd godevwatch
go build -o godevwatch
```

## Usage

### Initialize configuration

```bash
godevwatch init
```

This creates a `godevwatch.yaml` file in the current directory with default settings.

### Start the proxy server

```bash
godevwatch
```

This starts a proxy server on port 3000 (or the port specified in `godevwatch.yaml`).

### Configuration

The `godevwatch.yaml` file supports the following options:

```yaml
# Port for the proxy server (default: 3000)
port: 3000
```

### Flags

- `--help`, `-h`: Show help information
- `--version`, `-v`: Show version information

## Example

```bash
# Initialize config
godevwatch init

# Start the proxy (uses port from godevwatch.yaml)
godevwatch

# Or customize the port in godevwatch.yaml first
# port: 8080
godevwatch
```

## License

MIT
