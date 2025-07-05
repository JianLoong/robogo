# Gobot - Project Structure

```
gobot/
├── 📁 .devcontainer/           # Development container configuration
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── README.md
├── 📁 cmd/                     # CLI applications
│   └── gobot/
│       └── main.go            # Main CLI entry point
├── 📁 internal/                # Private application code
│   ├── 📁 parser/             # YAML test case parsing
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── types.go           # Test case data structures
│   ├── 📁 keywords/           # Keyword execution engine
│   │   ├── builtin.go         # Built-in keywords (log, sleep, assert)
│   │   ├── builtin_test.go
│   │   ├── executor.go        # Keyword execution logic
│   │   ├── executor_test.go
│   │   └── types.go           # Keyword interface definitions
│   ├── 📁 runner/             # Test execution orchestration
│   │   ├── runner.go          # Main test runner
│   │   ├── runner_test.go
│   │   ├── parallel.go        # Parallel execution logic
│   │   └── types.go           # Runner data structures
│   ├── 📁 git/                # Git integration
│   │   ├── client.go          # Git operations
│   │   ├── client_test.go
│   │   └── types.go
│   ├── 📁 secrets/            # Secret management
│   │   ├── resolver.go        # Environment variable resolution
│   │   ├── resolver_test.go
│   │   └── types.go
│   ├── 📁 http/               # HTTP client with mTLS support
│   │   ├── client.go          # HTTP client with mTLS
│   │   ├── client_test.go
│   │   └── types.go
│   └── 📁 reporting/          # Test result reporting
│       ├── reporter.go        # Console and file reporting
│       ├── reporter_test.go
│       └── types.go
├── 📁 pkg/                     # Public libraries (if needed)
│   └── (future public APIs)
├── 📁 testcases/              # Example test cases
│   ├── basic.yaml             # Basic test case examples
│   ├── parallel.yaml          # Parallel execution examples
│   ├── http.yaml              # HTTP testing examples
│   └── secrets.yaml           # Secret usage examples
├── 📁 tests/                  # Framework tests
│   ├── 📁 integration/        # Integration tests
│   │   ├── full_pipeline_test.go
│   │   ├── git_integration_test.go
│   │   └── parallel_test.go
│   ├── 📁 e2e/               # End-to-end tests
│   │   ├── cli_test.go
│   │   └── yaml_parsing_test.go
│   └── 📁 fixtures/          # Test data and fixtures
│       ├── test_cases/
│       └── git_repos/
├── 📁 docs/                   # Documentation
│   ├── README.md
│   ├── API.md
│   ├── EXAMPLES.md
│   └── CONTRIBUTING.md
├── 📁 scripts/                # Build and deployment scripts
│   ├── build.sh
│   ├── test.sh
│   └── release.sh
├── 📁 .github/                # GitHub workflows
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── security.yml
├── 📄 .gitignore
├── 📄 .dockerignore
├── 📄 docker-compose.yml
├── 📄 go.mod
├── 📄 go.sum
├── 📄 Makefile               # Build automation
├── 📄 PRD.md                 # Product Requirements Document
├── 📄 PROJECT_STRUCTURE.md   # This file
├── 📄 README.md
└── 📄 LICENSE
```

## Key Design Principles

### 1. **Separation of Concerns**
- **`internal/`** - Private application logic
- **`cmd/`** - CLI entry points
- **`pkg/`** - Future public APIs
- **`testcases/`** - Example usage

### 2. **Modular Architecture**
- **Parser** - YAML parsing and validation
- **Keywords** - Extensible keyword system
- **Runner** - Test orchestration and execution
- **Git** - Repository integration
- **Secrets** - Secure credential management
- **HTTP** - Network testing with mTLS
- **Reporting** - Result output and formatting

### 3. **Testing Strategy**
- **Unit tests** - Alongside each package
- **Integration tests** - Cross-component testing
- **E2E tests** - Full workflow testing
- **Fixtures** - Reusable test data

### 4. **Development Experience**
- **Dev containers** - Consistent environment
- **Docker Compose** - Local development
- **GitHub Actions** - CI/CD automation
- **Makefile** - Common development tasks

## Future Extensions

### Phase 2: Enhanced Features
```
gobot/
├── 📁 internal/
│   ├── 📁 plugins/           # Plugin system
│   ├── 📁 web/              # Web UI
│   ├── 📁 api/              # REST API
│   └── 📁 database/         # Test result storage
├── 📁 pkg/
│   ├── 📁 sdk/              # Go SDK for extensions
│   └── 📁 client/           # HTTP client library
└── 📁 web/                  # Frontend application
```

### Phase 3: Enterprise Features
```
gobot/
├── 📁 internal/
│   ├── 📁 auth/             # Authentication/Authorization
│   ├── 📁 scheduler/        # Test scheduling
│   ├── 📁 notifications/    # Alert system
│   └── 📁 analytics/        # Test analytics
└── 📁 deployments/          # Kubernetes, Docker, etc.
```

## File Naming Conventions

- **Go files**: `snake_case.go` for main files, `snake_case_test.go` for tests
- **YAML files**: `kebab-case.yaml`
- **Directories**: `snake_case/`
- **Constants**: `UPPER_SNAKE_CASE`
- **Variables**: `camelCase`
- **Functions**: `PascalCase` for exported, `camelCase` for private

## Dependencies Strategy

### Core Dependencies (MVP)
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/spf13/cobra` - CLI framework
- `github.com/stretchr/testify` - Testing utilities

### Future Dependencies
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/gin-gonic/gin` - Web framework
- `github.com/golang-jwt/jwt` - JWT authentication
- `gorm.io/gorm` - Database ORM 