# Development Container Setup

This directory contains the configuration for a development container that provides a consistent development environment for the Gobot project.

## Quick Start

### Option 1: Using VS Code Dev Containers (Recommended)

1. Install the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) in VS Code
2. Open the project in VS Code
3. When prompted, click "Reopen in Container" or use `Ctrl+Shift+P` and select "Dev Containers: Reopen in Container"
4. Wait for the container to build and start

### Option 2: Using Docker Compose

1. Make sure you have Docker and Docker Compose installed
2. Run the following command from the project root:
   ```bash
   docker-compose up -d gobot-dev
   ```
3. Connect to the container:
   ```bash
   docker-compose exec gobot-dev bash
   ```

## What's Included

### Development Tools
- **Go 1.21** - Latest stable version
- **Git** - Version control
- **VS Code Extensions** - Go, YAML, Docker, GitLens, and testing extensions
- **Go Tools** - goimports, golangci-lint, gocov, and more

### Development Features
- **Hot Reload** - Code changes are immediately reflected
- **Volume Mounts** - Your local code is mounted into the container
- **Go Module Caching** - Faster dependency downloads
- **Test Database** - PostgreSQL for future database features

### Port Forwarding
- **8080** - Web UI (future)
- **3000** - API (future)
- **5432** - PostgreSQL database

## Development Workflow

1. **Start Development**:
   ```bash
   # Inside the container
   go mod download
   go run main.go
   ```

2. **Run Tests**:
   ```bash
   go test ./...
   go test -race ./...
   go test -cover ./...
   ```

3. **Code Quality**:
   ```bash
   go vet ./...
   golangci-lint run
   goimports -w .
   ```

4. **Build**:
   ```bash
   go build -o robot-go main.go
   ```

## Environment Variables

The container is configured with the following Go environment variables:
- `GO111MODULE=on` - Enable Go modules
- `GOPRIVATE=*` - Allow private repositories
- `CGO_ENABLED=1` - Enable CGO for potential native dependencies

## Troubleshooting

### Container Won't Start
- Make sure Docker is running
- Check if ports 8080, 3000, or 5432 are already in use
- Try rebuilding: `docker-compose build --no-cache`

### Go Tools Not Found
- The tools are installed in the post-create command
- If they're missing, run: `go install golang.org/x/tools/cmd/goimports@latest`

### Permission Issues
- The container runs as a non-root user (vscode)
- Use `sudo` if you need elevated permissions

## Contributing

When contributing to this project:
1. Use the dev container for consistent development
2. Follow the Go coding standards
3. Write tests for new features
4. Run the full test suite before submitting PRs

## Future Enhancements

- Add more development tools as needed
- Configure additional VS Code extensions
- Add support for different Go versions
- Include additional databases or services 