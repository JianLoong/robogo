# Secrets Handling Design

## Overview

Implement secure secret management in Robogo's variable system through environment variable integration, automatic secret detection, and comprehensive output masking while maintaining backward compatibility.

## Architecture

### Core Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Secret        │    │    Variable      │    │   Output        │
│   Resolver      │───▶│   Substitution   │───▶│   Masker        │
│                 │    │   Engine         │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ - ENV vars      │    │ - ${secrets.KEY} │    │ - Log masking   │
│ - Secret files  │    │ - ${env.VAR}     │    │ - Error masking │
│ - Auto-detect   │    │ - Regular vars   │    │ - Result masking│
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Components and Interfaces

### 1. Secret Resolver

```go
// internal/common/secrets.go
type SecretResolver struct {
    envVars     map[string]string
    secretFiles map[string]string
    sensitiveKeys map[string]bool
}

type SecretSource interface {
    GetSecret(key string) (string, bool)
    ListSecrets() []string
}

// Environment variable source
type EnvSecretSource struct{}
func (e *EnvSecretSource) GetSecret(key string) (string, bool) {
    return os.LookupEnv(key)
}

// File-based secret source  
type FileSecretSource struct {
    filePath string
    secrets  map[string]string
}
```

### 2. Enhanced Variable System

```go
// internal/common/variables.go (enhanced)
type Variables struct {
    store          *VariableStore
    engine         *SubstitutionEngine
    secretResolver *SecretResolver
    sensitiveVars  map[string]bool
}

// New substitution patterns
// ${secrets.SECRET_NAME} - Secret from environment or config
// ${env.VAR_NAME}        - Explicit environment variable (non-secret)
// ${secret_var}          - Auto-detected sensitive variable
```

### 3. Output Masking System

```go
// internal/common/masker.go
type OutputMasker struct {
    sensitiveValues map[string]bool
    patterns        []*regexp.Regexp
}

func (m *OutputMasker) MaskSensitiveData(text string) string {
    // Replace sensitive values with [MASKED]
    // Apply pattern-based masking for common secrets
}
```

### 4. Secret Detection

```go
// internal/common/detector.go
var SensitiveKeywords = []string{
    "password", "passwd", "pwd",
    "secret", "key", "token", 
    "api_key", "apikey", "auth",
    "credential", "cred",
}

var SensitivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`^[A-Za-z0-9+/]{40,}={0,2}$`), // Base64 tokens
    regexp.MustCompile(`^[a-f0-9]{32,}$`),             // Hex tokens
    regexp.MustCompile(`^sk-[a-zA-Z0-9]{48}$`),        // OpenAI API keys
}
```

## Data Models

### Secret Configuration

```yaml
# .robogo-secrets (optional file)
secrets:
  database_password: "actual_password_here"
  api_token: "secret_token_value"
  
# Auto-detection overrides
sensitivity:
  force_sensitive:
    - "custom_secret_var"
  force_non_sensitive:
    - "public_api_url"
```

### Enhanced Variable Usage

```yaml
# test-with-secrets.yaml
testcase: "API Test with Secrets"

variables:
  vars:
    api_url: "https://api.example.com"
    api_key: "${secrets.API_SECRET_KEY}"    # Secret (GitHub Actions style)
    db_password: "${secrets.DB_PASSWORD}"   # Secret from env or config
    public_data: "not sensitive"
    build_number: "${env.BUILD_NUMBER}"     # Non-secret env var

steps:
  - name: "Authenticate with API"
    action: http
    args: ["POST", "${api_url}/auth"]
    options:
      headers:
        Authorization: "Bearer ${api_key}"   # Will be masked in logs
    result: auth_response
```

## Error Handling

### Secure Error Messages

```go
// Enhanced error handling for secrets
func (eb *ErrorBuilder) WithSensitiveContext(key string, value any) *ErrorBuilder {
    if eb.isSensitive(key) {
        eb.context[key] = "[MASKED]"
    } else {
        eb.context[key] = value
    }
    return eb
}
```

### Missing Secret Errors

```go
func MissingSecretError(secretName, source string) ActionResult {
    return NewErrorBuilder(ErrorCategoryVariable, "MISSING_SECRET").
        WithTemplate("Secret '%s' not found in %s").
        WithSuggestion("Set environment variable: export %s=your_secret_value").
        WithSuggestion("Or add to .robogo-secrets file").
        Build(secretName, source, secretName)
}
```

## Testing Strategy

### Unit Tests
- Secret resolution from different sources
- Variable substitution with secrets
- Output masking functionality
- Auto-detection accuracy

### Integration Tests
- End-to-end secret handling in real tests
- Environment variable integration
- File-based secret loading
- Error scenarios with missing secrets

### Security Tests
- Verify no secrets appear in logs
- Test masking under all conditions
- Validate error message sanitization
- Confirm file permission handling

## Implementation Phases

### Phase 1: Core Secret Resolution
- GitHub Actions-style secret support (`${secrets.KEY}`)
- Environment variable support (`${env.VAR}`)
- Basic output masking
- Secret detection patterns

### Phase 2: Enhanced Features  
- File-based secrets
- Configuration overrides
- Advanced masking patterns

### Phase 3: Security Hardening
- Comprehensive log sanitization
- Error message security review
- Performance optimization

## Security Considerations

### Secret Storage
- Environment variables (recommended)
- Local files with restricted permissions (600)
- Never in version control (.gitignore patterns)

### Memory Management
- Clear sensitive data from memory when possible
- Avoid string concatenation with secrets
- Use secure comparison for secret validation

### Logging Security
- All output goes through masking layer
- Sensitive patterns detected automatically
- Debug mode still respects masking

## Backward Compatibility

- Existing variable syntax unchanged
- New secret features are opt-in
- No performance impact when secrets not used
- All existing tests continue to work