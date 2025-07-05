# Installation Guide

This guide will help you install and set up Robogo on your system.

## Prerequisites

Before installing Robogo, ensure you have the following:

- **Go 1.21 or later** - [Download from golang.org](https://golang.org/dl/)
- **Git** - For version control and test case management
- **Docker** (optional) - For containerized development

## Installation Methods

### Method 1: From Source (Recommended)

The easiest way to install Robogo is to build from source:

```bash
# Clone the repository
git clone https://github.com/your-org/robogo.git
cd robogo

# Install dependencies
go mod download

# Build the binary
go build -o robogo cmd/robogo/main.go

# Install to your system (optional)
sudo mv robogo /usr/local/bin/
```

### Method 2: Go Install

You can also install using Go's `install` command:

```bash
go install github.com/your-org/robogo/cmd/robogo@latest
```

This will:
- Download the latest version of Robogo
- Compile it for your platform
- Install it to your `$GOPATH/bin` directory

### Method 3: Docker (Development)

For development or if you prefer containerized execution:

```bash
# Build the Docker image
docker build -t robogo .

# Run Robogo in a container
docker run --rm -v $(pwd):/workspace robogo run testcases/hello-world.yaml
```

## Verification

After installation, verify that Robogo is working correctly:

```bash
# Check version
./robogo --version

# Expected output:
# robogo version 0.1.0 (commit: dev, date: unknown)
```

## Configuration

### Environment Variables

Robogo uses environment variables for configuration. Set these in your shell profile:

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
export ROBOGO_CONFIG_PATH="$HOME/.robogo/config.yaml"
export ROBOGO_LOG_LEVEL="info"
export ROBOGO_CACHE_DIR="$HOME/.robogo/cache"
```

### Configuration File

Create a configuration file at `~/.robogo/config.yaml`:

```yaml
# Robogo Configuration
log_level: info
cache_dir: ~/.robogo/cache
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

# Clone and build Robogo
git clone https://github.com/your-org/robogo.git
cd robogo
go build -o robogo cmd/robogo/main.go
sudo mv robogo /usr/local/bin/
```

#### CentOS/RHEL/Fedora
```bash
# Install Go if not already installed
sudo dnf install golang

# Clone and build Robogo
git clone https://github.com/your-org/robogo.git
cd robogo
go build -o robogo cmd/robogo/main.go
sudo mv robogo /usr/local/bin/
```

### macOS

#### Using Homebrew
```bash
# Install Go if not already installed
brew install go

# Clone and build Robogo
git clone https://github.com/your-org/robogo.git
cd robogo
go build -o robogo cmd/robogo/main.go
sudo mv robogo /usr/local/bin/
```

#### Manual Installation
```bash
# Download and install Go from golang.org
# Then clone and build Robogo
git clone https://github.com/your-org/robogo.git
cd robogo
go build -o robogo cmd/robogo/main.go
sudo mv robogo /usr/local/bin/
```

### Windows

#### Using Chocolatey
```powershell
# Install Go if not already installed
choco install golang

# Clone and build Robogo
git clone https://github.com/your-org/robogo.git
cd robogo
go build -o robogo.exe cmd/robogo/main.go
# Add to PATH or move to a directory in PATH
```

#### Manual Installation
1. Download Go from [golang.org/dl](https://golang.org/dl/)
2. Install Go following the official instructions
3. Open PowerShell and run:
   ```powershell
   git clone https://github.com/your-org/robogo.git
   cd robogo
   go build -o robogo.exe cmd/robogo/main.go
   ```
4. Add the directory containing `robogo.exe` to your PATH environment variable

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
git clone https://github.com/your-org/robogo.git
cd robogo

# Install dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
go test ./...

# Build the binary
go build -o robogo cmd/robogo/main.go
```

## Troubleshooting

### Common Issues

#### "command not found: robogo"
- Ensure the binary is in your PATH
- Try running with `./robogo` from the project directory
- Check that the build was successful

#### "go: module github.com/your-org/robogo: not found"
- This is expected for local development
- The module name is a placeholder and should be updated for your organization

#### Permission Denied
- Use `sudo` when moving the binary to system directories
- Ensure the binary has execute permissions: `chmod +x robogo`

## Next Steps

After installation:

1. **Run your first test**: `./robogo run testcases/hello-world.yaml`
2. **Explore the CLI**: `./robogo --help`
3. **List available actions**: `./robogo list`
4. **Read the [Quick Start Guide](quickstart.md)** for your first test case 