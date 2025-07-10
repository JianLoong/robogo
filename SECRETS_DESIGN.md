# SECRETS Namespace Design Proposal

## Overview

Implement a dedicated `SECRETS.var` syntax for accessing secrets to improve security clarity, enable better tooling, and provide namespace separation.

## Current vs Proposed Syntax

### Current Approach
```yaml
variables:
  secrets:
    db_password:
      file: "db_secret.txt"
      mask_output: true
    api_key:
      value: "secret-key-123"
  vars:
    db_host: "localhost"
    db_user: "postgres"

steps:
  - name: "Database connection"
    action: postgres
    args: ["connect", "postgresql://${db_user}:${db_password}@${db_host}"]
    # Problem: Can't distinguish secret from regular variable
```

### Proposed Approach
```yaml
variables:
  secrets:
    db_password:
      file: "db_secret.txt"
      mask_output: true
    api_key:
      value: "secret-key-123"
  vars:
    db_host: "localhost"
    db_user: "postgres"

steps:
  - name: "Database connection"
    action: postgres
    args: ["connect", "postgresql://${db_user}:${SECRETS.db_password}@${db_host}"]
    # Clear: SECRETS.db_password is obviously a secret
```

## Implementation Plan

### Phase 1: Dual Support (Backward Compatible)
```go
// Enhanced variable substitution logic
func (vm *VariableManager) substituteString(s string) string {
    // Handle SECRETS.variable pattern
    secretPattern := regexp.MustCompile(`\$\{SECRETS\.([^}]+)\}`)
    s = secretPattern.ReplaceAllStringFunc(s, func(match string) string {
        secretName := secretPattern.FindStringSubmatch(match)[1]
        if value, exists := vm.secrets[secretName]; exists {
            return fmt.Sprintf("%v", value)
        }
        return match // Leave unresolved if secret doesn't exist
    })
    
    // Handle regular ${variable} pattern (existing logic)
    return vm.substituteRegularVariables(s)
}
```

### Phase 2: Enhanced Security Features
```yaml
# Advanced secret configuration
variables:
  secrets:
    database_credentials:
      file: "db_creds.json"
      mask_output: true
      allowed_actions: ["postgres", "mysql"]  # Restrict where secret can be used
      ttl: "1h"                               # Secret expiration
    
    api_token:
      value: "${ENV.API_TOKEN}"               # Environment variable injection
      mask_output: true
      audit_log: true                         # Log secret access (without value)

steps:
  - name: "Database query"
    action: postgres
    args: ["query", "${SECRETS.database_credentials}", "SELECT * FROM users"]
    # Framework validates secret is allowed for 'postgres' action
```

### Phase 3: Security Enhancements
```go
// Secret access validation
type SecretAccessControl struct {
    AllowedActions []string
    AuditLog      bool
    TTL           time.Duration
    AccessCount   int
}

func (vm *VariableManager) validateSecretAccess(secretName, action string) error {
    if control, exists := vm.secretControls[secretName]; exists {
        if len(control.AllowedActions) > 0 && !contains(control.AllowedActions, action) {
            return fmt.Errorf("secret '%s' not allowed for action '%s'", secretName, action)
        }
        if control.AuditLog {
            vm.logSecretAccess(secretName, action)
        }
    }
    return nil
}
```

## Security Benefits

### 1. **Explicit Secret Identification**
```bash
# Easy security auditing
grep "SECRETS\." tests/**/*.robogo | wc -l
# Count: 23 secret usages across test suite

grep "SECRETS\..*password" tests/**/*.robogo
# Find: All password secret references
```

### 2. **IDE Integration**
```json
// VS Code extension enhancement
{
  "patterns": [
    {
      "match": "\\$\\{SECRETS\\.[^}]+\\}",
      "name": "variable.secret.robogo",
      "settings": {
        "foreground": "#ff6b6b",
        "fontStyle": "bold"
      }
    }
  ]
}
```

### 3. **Security Scanning Integration**
```bash
# Security scanner rule
robogo-security-scan --check-secrets tests/
# Output: Found 5 secret references, 2 unmasked outputs detected
```

## Migration Strategy

### Option 1: Gradual Migration (Recommended)
```yaml
# Support both syntaxes during transition
variables:
  secrets:
    db_password:
      file: "secret.txt"
      deprecate_legacy_access: true  # Warn when accessed as ${db_password}

steps:
  - name: "Legacy access (with warning)"
    action: log
    args: ["Password: ${db_password}"]  # Prints warning
    
  - name: "New syntax (recommended)"
    action: log
    args: ["Password: ${SECRETS.db_password}"]  # No warning
```

### Option 2: Configuration-Driven
```yaml
# Global configuration option
robogo_config:
  secrets:
    require_namespace: true   # Force SECRETS.var syntax
    legacy_support: false    # Disable ${secret_name} access
```

## Advanced Features

### 1. **Secret Inheritance**
```yaml
# Base test configuration
variables:
  secrets:
    common_api_key:
      file: "api_key.txt"
      
# Child test inherits and extends
extends: "base_test.robogo"
variables:
  secrets:
    specific_token:
      value: "test-token"
      
steps:
  - name: "Use inherited secret"
    action: http
    args: ["GET", "https://api.example.com", {"Authorization": "Bearer ${SECRETS.common_api_key}"}]
```

### 2. **Secret Scoping**
```yaml
variables:
  secrets:
    global_secret:
      file: "global.txt"
      scope: "global"      # Available to all steps
      
    step_secret:
      file: "step.txt"
      scope: "step:database_ops"  # Only available to specific step
```

### 3. **Secret Rotation Support**
```yaml
variables:
  secrets:
    rotating_key:
      source: "vault://secrets/api-key"
      refresh_interval: "30m"
      fallback: "backup_key.txt"
```

## Implementation Code Example

```go
// Enhanced VariableManager with SECRETS support
type VariableManager struct {
    variables      map[string]interface{}
    secrets        map[string]interface{}
    secretControls map[string]*SecretAccessControl
    mutex          sync.RWMutex
}

func (vm *VariableManager) substituteString(s string) string {
    // Handle SECRETS namespace
    secretPattern := regexp.MustCompile(`\$\{SECRETS\.([^}]+)\}`)
    s = secretPattern.ReplaceAllStringFunc(s, func(match string) string {
        secretName := secretPattern.FindStringSubmatch(match)[1]
        return vm.resolveSecret(secretName)
    })
    
    // Handle regular variables
    varPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
    s = varPattern.ReplaceAllStringFunc(s, func(match string) string {
        varName := varPattern.FindStringSubmatch(match)[1]
        return vm.resolveVariable(varName)
    })
    
    return s
}

func (vm *VariableManager) resolveSecret(name string) string {
    vm.mutex.RLock()
    defer vm.mutex.RUnlock()
    
    if value, exists := vm.secrets[name]; exists {
        // Log access for audit
        vm.auditSecretAccess(name)
        return fmt.Sprintf("%v", value)
    }
    
    return fmt.Sprintf("${SECRETS.%s}", name) // Leave unresolved
}
```

## Conclusion

The `SECRETS.var` syntax provides:
- ✅ **Better Security**: Clear identification of sensitive data usage
- ✅ **Improved Tooling**: Easier auditing and security scanning
- ✅ **Namespace Safety**: Prevents conflicts between secrets and variables
- ✅ **Compliance Ready**: Better for security audits and regulations
- ✅ **Future Extensibility**: Foundation for advanced secret management features

This enhancement would position Robogo as a security-conscious testing framework suitable for enterprise environments while maintaining its developer-friendly approach.