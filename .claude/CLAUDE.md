# godevwatch - Development Workflow Monitor for Go

**godevwatch** is a development server and live reload tool for Go applications that provides automatic rebuilds, live browser reloads, and an intelligent development proxy server.

## 🎯 Purpose

This tool solves the common development workflow problem where you need to:
1. Make code changes to your Go application
2. Manually rebuild the application
3. Restart the server
4. Refresh your browser to see changes

Instead, godevwatch automates this entire process, providing a smooth development experience similar to modern frontend frameworks.

## 🚀 Key Features

### Intelligent File Watching
- **Pattern-based watching**: Configure which files/directories to monitor using glob patterns including `**` recursive matching
- **Ignore patterns**: Exclude files like tests, vendor directories, and temporary files
- **Debounced rebuilds**: Prevents rapid successive builds when multiple files change simultaneously

### Smart Build Management
- **Conditional builds**: Only rebuild when relevant files change
- **Build status tracking**: Persistent build state management with detailed status information
- **Concurrent build handling**: Automatically aborts previous builds when new changes are detected
- **Flexible build commands**: Support for any shell command, not just `go build`

### Development Proxy Server
- **Transparent proxying**: Acts as a reverse proxy to your Go backend
- **Graceful degradation**: Shows a "server down" page when backend is unavailable
- **Health monitoring**: Continuously monitors backend availability with configurable checks
- **Auto-reload integration**: Automatically triggers browser refresh when backend comes online

### Browser Auto-Reload
- **Server-Sent Events (SSE)**: Real-time communication with browser clients
- **Multiple client support**: Handle multiple browser tabs/windows simultaneously
- **Connection management**: Robust handling of client connections and disconnections
- **Manual reload triggers**: API endpoints for programmatic reload control

### Process Management
- **Backend lifecycle**: Automatic start/stop/restart of your Go application
- **Graceful shutdown**: Proper cleanup of processes and temporary files
- **Signal handling**: Responds to interrupt signals for clean termination
- **Resource cleanup**: Removes build artifacts and status files on exit

## 📁 Project Architecture

### Core Components

```
godevwatch/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Main command with proxy server startup
│   └── init.go            # Configuration file initialization
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go      # YAML config parsing and defaults
│   ├── watcher/           # File system monitoring
│   │   └── watcher.go     # fsnotify-based file watching with pattern matching
│   ├── build/             # Build execution and tracking
│   │   ├── builder.go     # Build rule execution
│   │   └── tracker.go     # Build status persistence
│   ├── proxy/             # HTTP proxy server
│   │   └── proxy.go       # Reverse proxy with auto-reload integration
│   ├── health/            # Backend health monitoring
│   │   └── monitor.go     # TCP/HTTP health checks and status management
│   ├── process/           # Application process management
│   │   └── runner.go      # Backend process lifecycle
│   ├── ports/             # Port availability checking
│   │   └── checker.go     # Network port utility functions
│   └── logger/            # Logging utilities
│       └── logger.go      # Structured logging with debug mode support
├── example/               # Example backend application
│   └── backend.go         # Simple HTTP server for testing
└── main.go               # Application entrypoint
```

## 🔧 Configuration System

godevwatch uses a YAML configuration file (`godevwatch.yaml`) that provides flexible control over the development workflow:

### Configuration Options

```yaml
# Port for the development proxy server
proxy_port: 3000

# Port of your backend Go server
backend_port: 8080

# Directory where build status files are stored
build_status_dir: tmp/.build-status

# Build rules define conditional build steps based on file changes
build_rules:
  - name: "go-build"
    watch:
      - "**/*.go"              # Watch all Go files recursively
    ignore:
      - "**/*_test.go"         # Ignore test files
      - "vendor/**"            # Ignore vendor directory
      - "node_modules/**"      # Ignore Node.js dependencies
    command: "go build -o ./tmp/main ."

# Command to run your application after successful build
run_cmd: "./tmp/main"
```

### Build Rules System

The build rules system is highly flexible and supports:

- **Multiple rules**: Define different build steps for different file types
- **Pattern matching**: Use glob patterns with `**` for recursive matching
- **Conditional execution**: Rules only execute when matching files change
- **Sequential execution**: Rules run in the order defined
- **Custom commands**: Any shell command can be used, not just Go builds

## 🛠 Installation Methods

### Option 1: Go Install (Recommended)
```bash
go install github.com/kyco/godevwatch@latest
```

### Option 2: Build from Source
```bash
git clone https://github.com/kyco/godevwatch.git
cd godevwatch
go build -o godevwatch
# Optionally move to PATH
sudo mv godevwatch /usr/local/bin/
```

### Option 3: Download Binary
Download the latest release from GitHub releases page for your platform.

## 📖 Usage Guide

### Quick Start

1. **Navigate to your Go project directory**
   ```bash
   cd your-go-project
   ```

2. **Initialize configuration**
   ```bash
   godevwatch init
   ```
   This creates a `godevwatch.yaml` file with sensible defaults.

3. **Start the development server**
   ```bash
   godevwatch
   ```

4. **Open your browser to the proxy**
   ```
   http://localhost:3000
   ```
   (or whatever port you configured in `godevwatch.yaml`)

### Advanced Usage

#### Debug Mode
Enable verbose logging to see detailed information about file watching, build processes, and server operations:
```bash
godevwatch --debug
```

#### Custom Configuration
Modify `godevwatch.yaml` to match your project structure:

```yaml
# Example for a more complex project
proxy_port: 4000
backend_port: 8080

build_rules:
  - name: "generate-code"
    watch:
      - "schemas/**/*.json"
    command: "go generate ./..."

  - name: "build-server"
    watch:
      - "cmd/**/*.go"
      - "internal/**/*.go"
      - "pkg/**/*.go"
    ignore:
      - "**/*_test.go"
      - "**/*.pb.go"  # Ignore generated protobuf files
    command: "go build -o ./bin/server ./cmd/server"

run_cmd: "./bin/server"
```

## 🔄 Workflow Integration

### Development Process

1. **Start godevwatch** in your project root
2. **Open browser** to the proxy URL (e.g., `http://localhost:3000`)
3. **Edit your Go files** - the tool automatically:
   - Detects file changes using efficient filesystem monitoring
   - Triggers appropriate build rules based on file patterns
   - Rebuilds only when necessary (debounced)
   - Restarts your backend application
   - Refreshes your browser when the backend is ready

### Integration with IDEs

godevwatch works seamlessly with any editor or IDE:
- **VS Code**: No special configuration needed
- **GoLand**: Works with any external build tools
- **Vim/Neovim**: Perfect for terminal-based development
- **Sublime Text**: No conflicts with built-in Go tools

### CI/CD Compatibility

The tool is designed for development only and doesn't interfere with:
- Production builds
- CI/CD pipelines
- Docker containers
- Deployment processes

## 🌐 Browser Integration

### Auto-Reload Mechanism

godevwatch uses Server-Sent Events (SSE) to communicate with browsers:

1. **Initial load**: Browser connects to `/__reload` endpoint
2. **File changes**: Watcher detects changes and triggers builds
3. **Build completion**: Backend process restarts
4. **Health check**: Monitor detects backend availability
5. **Reload signal**: Browser receives reload event and refreshes

### Multiple Browser Support

- Multiple browser tabs/windows are supported
- Each connection is managed independently
- Connection cleanup happens automatically
- No browser extensions or plugins required

### Custom Integration

For advanced users, the auto-reload system can be customized:

```javascript
// Custom SSE connection in your HTML
const eventSource = new EventSource('/__reload');
eventSource.onmessage = function(event) {
    if (event.data === 'reload') {
        location.reload();
    }
};
```

## 🔍 API Endpoints

godevwatch exposes several endpoints for monitoring and control:

### Health Check
```
GET /__health
```
Returns HTTP 200 when backend is available, 503 when down.

### Build Status
```
GET /__build-status
```
Returns JSON with current build information:
```json
{
  "current_build": {
    "build_id": "1633024800-abc123",
    "rule_name": "go-build",
    "status": "success",
    "timestamp": 1633024800
  }
}
```

### Auto-Reload Stream
```
GET /__reload
```
Server-Sent Events stream for browser auto-reload.

## 🐛 Troubleshooting

### Common Issues

#### Port Conflicts
```
Error: proxy server failed to start: listen tcp :3000: bind: address already in use
```
**Solution**: Change `proxy_port` in `godevwatch.yaml` or kill the process using the port.

#### Build Failures
```
Error: build failed (go-build): exit status 2
```
**Solution**: Check your Go code for compilation errors. Use `--debug` flag for detailed build output.

#### Backend Won't Start
```
Warning: Failed to start backend: fork/exec ./tmp/main: no such file or directory
```
**Solution**: Ensure your build command creates the executable at the specified path.

### Debug Mode

Use the `--debug` flag to get detailed information:
```bash
godevwatch --debug
```

This shows:
- File system events
- Build command execution
- Process management details
- Health check status
- Client connection information

### Log Analysis

godevwatch uses structured logging with prefixes:
- `[proxy]`: Proxy server and health monitoring
- `[watcher]`: File system monitoring and build triggers
- `[build]`: Build execution and status
- `[process]`: Backend process management

## 🔒 Security Considerations

### Development Only

godevwatch is designed exclusively for development:
- **Not for production**: The proxy adds overhead and debugging features
- **Local network only**: Binds to localhost by default
- **No authentication**: Assumes trusted development environment
- **File system access**: Requires read access to watch directories

### Network Security

- Proxy server only listens on specified port
- No external network connections required
- Build commands run with same permissions as user
- Temporary files created in configurable directory

## 🎛 Command Line Interface

### Main Command
```bash
godevwatch [flags]
```

### Flags
- `--debug`: Enable verbose debug logging
- `--version, -v`: Show version information
- `--help, -h`: Show help information

### Subcommands

#### Initialize Configuration
```bash
godevwatch init
```
Creates `godevwatch.yaml` with default settings. Prompts for confirmation if file exists.

## 🔄 Version History & Compatibility

### Current Version: 0.1.0

**Go Version Compatibility**: Requires Go 1.25.1 or later

**Dependencies**:
- `github.com/fsnotify/fsnotify`: File system notifications
- `github.com/spf13/cobra`: CLI framework
- `github.com/manifoldco/promptui`: Interactive prompts
- `gopkg.in/yaml.v3`: YAML configuration parsing

### Platform Support
- **macOS**: Full support (tested)
- **Linux**: Full support
- **Windows**: Compatible (file watching may have platform-specific behavior)

## 🤝 Contributing & Development

### Project Structure
The project follows standard Go conventions with clear separation of concerns:
- CLI commands in `cmd/`
- Internal packages in `internal/`
- Example applications in `example/`
- Configuration templates embedded in binaries

### Development Setup
1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Build with `go build`
4. Test with the example backend: `go build -o tmp/backend example/backend.go`

### Testing
- Unit tests for core functionality
- Integration tests with example backend
- Manual testing across different Go project structures

## 📄 License

MIT License - see LICENSE file for details.

---

**godevwatch** streamlines Go development by eliminating manual build-refresh cycles, allowing developers to focus on writing code while the tool handles the development workflow automation.