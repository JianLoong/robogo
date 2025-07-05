# Installation Guide

This guide will help you install and set up Gobot on your system.

## Prerequisites

Before installing Gobot, ensure you have the following:

- **Go 1.21 or later** - [Download from golang.org](https://golang.org/dl/)
- **Git** - For version control and test case management
- **Docker** (optional) - For containerized development

## Installation Methods

### Method 1: Go Install (Recommended)

The easiest way to install Gobot is using Go's `install` command:

```bash
go install github.com/your-org/gobot/cmd/gobot@latest
```

This will:
- Download the latest version of Gobot
- Compile it for your platform
- Install it to your `$GOPATH/bin` directory

### Method 2: From Source

If you want to build from source or contribute to development:

```bash
# Clone the repository
git clone https://github.com/your-org/gobot.git
cd gobot

# Build the binary
go build -o gobot cmd/gobot/main.go

# Install to your system
sudo mv gobot /usr/local/bin/
```

### Method 3: Docker (Development)

For development or if you prefer containerized execution:

```bash
# Pull the latest image
docker pull your-org/gobot:latest

# Run Gobot in a container
docker run --rm -v $(pwd):/workspace your-org/gobot:latest run testcases/example.yaml
```

## Verification

After installation, verify that Gobot is working correctly:

```bash
# Check version
gobot version

# Expected output:
# Gobot v1.0.0
# Build: 2024-01-15T10:30:00Z
# Go: go1.21.5
```

## Configuration

### Environment Variables

Gobot uses environment variables for configuration. Set these in your shell profile:

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
export GOBOT_CONFIG_PATH="$HOME/.gobot/config.yaml"
export GOBOT_LOG_LEVEL="info"
export GOBOT_CACHE_DIR="$HOME/.gobot/cache"
```

### Configuration File

Create a configuration file at `~/.gobot/config.yaml`:

```yaml
# Gobot Configuration
log_level: info
cache_dir: ~/.gobot/cache
timeout: 300s

# Git settings
git:
  clone_timeout: 60s
  fetch_timeout: 30s

# HTTP settings
http:
  timeout: 30s
  retry_attempts: 3
  retry_delay: 1s

# Security settings
security:
  allow_insecure: false
  verify_ssl: true
```

## Platform-Specific Instructions

### Linux

#### Ubuntu/Debian
```bash
# Install Go if not already installed
sudo apt update
sudo apt install golang-go

# Install Gobot
go install github.com/your-org/gobot/cmd/gobot@latest

# Add to PATH if needed
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

#### CentOS/RHEL/Fedora
```bash
# Install Go if not already installed
sudo dnf install golang

# Install Gobot
go install github.com/your-org/gobot/cmd/gobot@latest
```

### macOS

#### Using Homebrew
```bash
# Install Go if not already installed
brew install go

# Install Gobot
go install github.com/your-org/gobot/cmd/gobot@latest
```

#### Manual Installation
```bash
# Download and install Go from golang.org
# Then install Gobot
go install github.com/your-org/gobot/cmd/gobot@latest
```

### Windows

#### Using Chocolatey
```powershell
# Install Go if not already installed
choco install golang

# Install Gobot
go install github.com/your-org/gobot/cmd/gobot@latest
```

#### Manual Installation
1. Download Go from [golang.org/dl](https://golang.org/dl/)
2. Install Go following the official instructions
3. Open PowerShell and run:
   ```powershell
   go install github.com/your-org/gobot/cmd/gobot@latest
   ```
4. Add `%GOPATH%\bin` to your PATH environment variable

## Development Setup

For contributors and developers:

### Using Dev Containers

1. Install VS Code and the Dev Containers extension
2. Clone the repository
3. Open in VS Code and select "Reopen in Container"
4. The development environment will be automatically set up

### Local Development

```bash
# Clone the repository
git clone https://github.com/your-org/gobot.git
cd gobot

# Install dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
go test ./...

# Build
go build -o gobot cmd/gobot/main.go
```

## Troubleshooting

### Common Issues

#### "command not found: gobot"
- Ensure Go is installed and in your PATH
- Check that `$GOPATH/bin` is in your PATH
- Verify the installation with `go list -f '{{.Dir}}' github.com/your-org/gobot/cmd/gobot`

#### Permission Denied
- On Linux/macOS: `chmod +x $(go env GOPATH)/bin/gobot`
- On Windows: Run PowerShell as Administrator

#### Network Issues
- Check your internet connection
- Verify proxy settings if behind a corporate firewall
- Try using a different DNS server

#### Go Version Issues
- Ensure you have Go 1.21 or later: `go version`
- Update Go if needed: [golang.org/dl](https://golang.org/dl/)

### Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Search [GitHub Issues](https://github.com/your-org/gobot/issues)
3. Ask in [GitHub Discussions](https://github.com/your-org/gobot/discussions)
4. Report a new issue with detailed information

## Next Steps

After successful installation:

1. Follow the [Quick Start Guide](quickstart.md) to run your first test
2. Read the [CLI Reference](cli-reference.md) to learn all available commands
3. Explore [Test Case Writing](test-cases.md) for best practices

## Uninstallation

To remove Gobot:

```bash
# Remove the binary
rm $(go env GOPATH)/bin/gobot

# Remove configuration (optional)
rm -rf ~/.gobot

# Remove cache (optional)
rm -rf ~/.gobot/cache
``` 