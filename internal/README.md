# Internal Architecture

This directory contains the core implementation of Robogo's test automation framework, organized following KISS (Keep It Simple and Straightforward) principles.

## Directory Structure

```
internal/
├── actions/           # Action implementations and registry
├── common/           # Shared utilities (variables, security, dotenv)
├── constants/        # Configuration constants
├── execution/        # Execution strategies and core logic
├── templates/        # Template management
├── types/           # Core data structures
├── cli.go           # Direct CLI implementation
├── parser.go        # YAML test file parsing
└── runner.go        # Test execution orchestration
```

## Core Architecture Principles

### 1. KISS Architecture
- **No Dependency Injection**: Direct object construction throughout
- **No Over-abstraction**: Simple, direct implementations
- **Minimal Interfaces**: Only where absolutely necessary
- **Direct Construction**: TestRunner creates dependencies directly

### 2. Strategy Pattern for Execution
- **Priority-based routing**: Higher priority strategies handle more specific cases
- **Clean separation**: Each strategy handles one concern
- **Delegation pattern**: Strategies can route back to the router

### 3. Simple Variable System
- Single `Variables` struct with map storage
- `${variable}` and `${ENV:VARIABLE}` substitution
- No complex templating engines or multiple abstraction layers

## Key Components

### TestRunner (`runner.go`)
- **Purpose**: Orchestrates complete test execution lifecycle
- **Responsibilities**: 
  - Parse YAML test files
  - Manage variables and strategy router
  - Execute setup, main steps, and teardown
  - Generate test results and summaries
- **Architecture**: Direct construction, no dependency injection

### CLI (`cli.go`)
- **Purpose**: Direct command-line interface handling
- **Commands**: `run`, `list`, `version`
- **Design**: No abstractions, handles commands directly

### Parser (`parser.go`)
- **Purpose**: YAML test file parsing and validation
- **Responsibilities**: Convert YAML to internal types
- **Design**: Simple, direct parsing with error handling

## Execution Flow

1. **CLI** receives command and delegates to appropriate handler
2. **TestRunner** parses YAML and creates execution environment
3. **ExecutionStrategyRouter** routes steps to appropriate strategies
4. **Strategies** execute actions and handle control flow
5. **Actions** perform actual operations (HTTP, DB, etc.)
6. **Results** are collected, formatted, and displayed

For detailed execution flow, see: [../docs/execution-flow-diagram.md](../docs/execution-flow-diagram.md)

## Security Features

- **Step-level security**: `no_log` and `sensitive_fields` properties
- **Automatic masking**: Passwords, tokens, API keys automatically detected
- **Environment variables**: Secure credential management via `${ENV:VAR}`
- **Output sanitization**: All logs and errors go through masking layer

## Error Handling

- **Dual system**: ErrorInfo (technical) vs FailureInfo (logical)
- **Structured errors**: Category, code, template, context
- **Consistent patterns**: All execution strategies return ActionResult
- **User-friendly**: Clear error messages with suggestions

## Recent Improvements

- **Architecture simplification**: Removed 6+ abstraction layers
- **File organization**: Split large files into focused modules
- **Error standardization**: Consistent error handling patterns
- **Security enhancements**: Comprehensive data masking system
- **SCP support**: Secure file transfer via SSH/SFTP

## Development Guidelines

1. **Follow KISS principles**: Avoid over-engineering
2. **Direct construction**: No dependency injection
3. **Single responsibility**: One concern per file/function
4. **Error handling**: Always return ActionResult, never panic
5. **Security first**: Mask sensitive data by default
6. **Test coverage**: Every action should have example test cases

## Dependencies

- **Minimal external dependencies**: Only essential libraries
- **Go standard library**: Preferred for common operations  
- **Specific libraries**: Only for specialized functionality (SSH, databases, etc.)

See individual directory READMEs for detailed component documentation.