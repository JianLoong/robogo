# Common Package

This package provides core utilities and shared functionality used throughout the Robogo framework. It follows the KISS principle with simple, direct implementations of essential features.

## Components

### 📝 **Variables System** (`variables.go`)

Simple variable storage and substitution engine that powers the entire framework.

**Core Functionality:**
- **Variable Storage**: Store and retrieve test variables with `${variable}` syntax
- **Environment Variables**: Access system environment variables with `${ENV:VARIABLE}` syntax  
- **String Substitution**: Replace variable placeholders in strings and arguments
- **Unresolved Detection**: Track variables that couldn't be resolved for debugging

**Key Methods:**
```go
// Basic operations
vars := NewVariables()
vars.Set("api_url", "https://api.example.com")
value := vars.Get("api_url")

// Substitution
result := vars.SubstituteString("Request to ${api_url}/users")
// Returns: "Request to https://api.example.com/users"

// Environment variables
vars.SubstituteString("Database: ${ENV:DB_HOST}:${ENV:DB_PORT}")
// Returns: "Database: localhost:5432" (from environment)
```

**Variable Syntax:**
- `${variable_name}` - Substitute stored variable
- `${ENV:VARIABLE_NAME}` - Substitute environment variable
- `__UNRESOLVED_variable_name__` - Marker for failed resolution

### 🔒 **Security System** (`security.go`)

Comprehensive data masking and security controls for sensitive information.

**Security Features:**
- **Automatic Masking**: Detects and masks common sensitive patterns (passwords, tokens, keys)
- **Custom Field Masking**: User-specified fields to mask in step properties
- **No-Log Mode**: Complete logging suppression for sensitive operations
- **JSON-Aware**: Intelligent masking within JSON structures

**Built-in Patterns:**
```go
// Automatically masked patterns
"password=secret123"     → "password=***"
"token=abc123def"        → "token=***"
"Authorization: Bearer"  → "Authorization: Bearer ***"
"api_key=xyz789"         → "api_key=***"
```

**Usage in Tests:**
```yaml
# Complete suppression
- name: "Login with credentials"
  action: http
  args: ["POST", "/login", '{"password": "secret"}']
  no_log: true

# Custom field masking
- name: "Process user data"
  action: http
  args: ["POST", "/users", "${user_data}"]
  sensitive_fields: ["ssn", "credit_card"]
```

### 🌐 **Environment Loading** (`dotenv.go`)

Simple `.env` file loading for secure credential management.

**Features:**
- **Automatic Loading**: Loads `.env` file from working directory
- **Custom Files**: Support for custom `.env` file paths via `--env` flag
- **Variable Precedence**: Explicitly set environment variables override `.env` values
- **Error Handling**: Graceful handling of missing or malformed `.env` files

**Environment File Format:**
```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=robogo_user
DB_PASSWORD=secure_password123

# API credentials
API_BASE_URL=https://api.production.com
API_TOKEN=prod_token_xyz789
```

**Usage:**
```bash
# Automatic .env loading
./robogo run test.yaml

# Custom environment file
./robogo --env production.env run test.yaml
```

## Design Principles

### 🎯 **KISS Architecture**
- **No Complex Templating**: Simple `${variable}` substitution instead of Jinja2-style templates
- **Direct Implementation**: No interfaces or abstractions, just concrete functionality
- **Single Responsibility**: Each file handles one clear concern

### 🔐 **Security by Design**
- **Fail-Safe Defaults**: Security controls enabled by default
- **Granular Control**: Step-level security without action-level complexity
- **Clear Patterns**: Obvious syntax for security-sensitive operations

### 🧩 **Framework Integration**
- **Variable System**: Used by all actions for argument substitution
- **Security System**: Applied automatically during execution
- **Environment System**: Supports secure credential management

## Usage Examples

### Variable Substitution in Tests
```yaml
variables:
  vars:
    base_url: "https://api.example.com"
    user_id: "12345"
    
steps:
  - name: "Get user details"
    action: http
    args: ["GET", "${base_url}/users/${user_id}"]
    result: user_data
    
  - name: "Extract user name"
    action: jq
    args: ["${user_data}", ".name"]
    result: user_name
```

### Environment Variable Access
```yaml
variables:
  vars:
    # Secure database connection
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/testdb"
    
steps:
  - name: "Test database connection"
    action: postgres
    args: ["query", "${db_url}", "SELECT version()"]
```

### Security-Aware Testing
```yaml
steps:
  - name: "Authenticate user (sensitive)"
    action: http
    args: ["POST", "/auth", '{"username": "testuser", "password": "secret123"}']
    no_log: true  # Complete logging suppression
    result: auth_response
    
  - name: "Process payment (custom masking)"
    action: http
    args: ["POST", "/payments", "${payment_data}"]
    sensitive_fields: ["credit_card", "cvv", "account_number"]
```

## Error Handling

### Variable Resolution
- **Unresolved Variables**: Marked with `__UNRESOLVED_variable_name__` for debugging
- **Environment Missing**: Clear warnings when `${ENV:VAR}` variables are not set
- **Helpful Suggestions**: Guidance on using `jq` for complex data extraction

### Security Validation
- **Pattern Detection**: Automatic detection of sensitive patterns in logs and output
- **Custom Validation**: User-defined sensitive fields validated and masked
- **Fail-Safe**: Defaults to masking when in doubt

## Performance Considerations

### Variable Substitution
- **Simple Pattern Matching**: Uses `strings.Replace()` for efficient substitution
- **Lazy Evaluation**: Variables resolved only when needed
- **Memory Efficient**: Minimal overhead for variable storage

### Security Masking
- **Pre-computed Patterns**: Sensitive patterns compiled once, used repeatedly
- **JSON-Aware**: Efficient JSON field masking without full parsing
- **Context-Sensitive**: Different masking strategies for different data types

## Integration Points

### With Execution System
- Variables substituted before action execution
- Security masking applied to step results and logs
- Environment variables loaded during framework initialization

### With Actions
- All actions receive pre-substituted arguments via `Variables.SubstituteArgs()`
- Security settings applied automatically without action awareness
- Consistent variable access pattern across all actions

### With CLI
- Environment file loading integrated with CLI argument parsing
- Variable debugging available through CLI flags
- Security controls accessible via step properties

## Contributing

### Adding New Security Patterns
1. **Update Pattern Detection**: Add patterns to `security.go` 
2. **Test Coverage**: Ensure new patterns are properly masked
3. **Documentation**: Update examples showing pattern detection

### Extending Variable System  
1. **Maintain Simplicity**: Avoid complex templating features
2. **Preserve Performance**: Keep substitution fast and memory-efficient
3. **Clear Syntax**: New variable syntax should be obvious and consistent

### Environment Integration
1. **Follow Standards**: Use standard `.env` file format
2. **Error Messages**: Provide clear guidance for missing variables
3. **Security First**: Never log or expose environment variable values

This common package forms the foundation of Robogo's simplicity and security, providing essential functionality without complexity or over-engineering.