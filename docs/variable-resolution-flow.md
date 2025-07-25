# Variable Resolution Flow

This diagram shows how Robogo's variable system resolves `${variable}` and `${ENV:VARIABLE}` patterns in test arguments and strings. The variable resolution system is fundamental to dynamic test execution and secure credential management.

## Variable Resolution Process

### **Main Resolution Flow**

```mermaid
flowchart TD
    A[String with Variables] --> B[Scan for Pattern]
    B --> C{Pattern Found?}
    C -->|No| D[Return Original]
    C -->|Yes| E[Parse Pattern]
    
    E --> F{Pattern Type?}
    F -->|variable| G[Variable Lookup]
    F -->|ENV:VAR| H[Environment Lookup]
    
    G --> I{Exists?}
    I -->|Yes| J[Get Value]
    I -->|No| K[Mark Unresolved]
    
    H --> L{Exists?}
    L -->|Yes| M[Get Environment Value]
    L -->|No| N[Mark Unresolved]
    
    J --> O{Value Type?}
    O -->|Simple| P[Use Value]
    O -->|Complex| Q[Warn: Use jq]
    
    P --> R[Replace in String]
    M --> R
    K --> R
    N --> R
    Q --> R
    
    R --> S{More Patterns?}
    S -->|Yes| E
    S -->|No| T[Return Final String]
    D --> T
```

### **Variable Type Handling**

```mermaid
flowchart LR
    A[Variable Value] --> B{Type?}
    
    B -->|String| C[Direct Use]
    B -->|Number| D[Convert to String]
    B -->|Boolean| E[Convert to String]
    B -->|Object| F[Complex Data Warning]
    B -->|Array| G[Complex Data Warning]
    
    C --> H[✅ Success]
    D --> H
    E --> H
    
    F --> I[Use jq action to extract]
    G --> I
    I --> J[Extract specific field/item]
    J --> H
```

## Variable Syntax Patterns

### **Stored Variable Access**
```yaml
# Simple variable substitution
variables:
  vars:
    api_url: "https://api.example.com"
    user_id: "12345"

steps:
  - name: "Get user details"
    action: http
    args: ["GET", "${api_url}/users/${user_id}"]
    # Resolves to: ["GET", "https://api.example.com/users/12345"]
```

### **Environment Variable Access**
```yaml
# Secure credential access
variables:
  vars:
    db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:5432/testdb"

steps:
  - name: "Connect to database"
    action: postgres
    args: ["query", "${db_url}", "SELECT version()"]
    # Resolves using actual environment variables
```

### **Mixed Variable Types**
```yaml
# Combining stored and environment variables
variables:
  vars:
    base_url: "https://${ENV:API_HOST}"
    auth_header: "Bearer ${ENV:API_TOKEN}"

steps:
  - name: "Authenticated request"
    action: http
    args: ["GET", "${base_url}/protected"]
    options:
      headers:
        Authorization: "${auth_header}"
```

## Resolution Process Details

### **1. Pattern Scanning**
The system scans input strings for variable patterns:

```go
// Pattern detection regex
var variablePattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Environment variable pattern
var envPattern = regexp.MustCompile(`\$\{ENV:([^}]+)\}`)
```

**Detected Patterns:**
- `${variable_name}` - Stored variable lookup
- `${ENV:VARIABLE_NAME}` - Environment variable lookup
- `${invalid syntax}` - Invalid pattern (marked for warning)

### **2. Variable Storage Lookup**
For `${variable}` patterns:

```go
func (v *Variables) Get(key string) any {
    return v.data[key]  // O(1) map lookup
}
```

**Value Types Handled:**
- **String**: Direct substitution
- **Number**: Converted to string representation
- **Boolean**: Converted to "true"/"false"
- **Object/Array**: Warning + unresolved marker

### **3. Environment Variable Lookup**
For `${ENV:VARIABLE}` patterns:

```go
value := os.Getenv(variableName)
if value == "" {
    // Mark as unresolved
    return fmt.Sprintf("__UNRESOLVED_ENV:%s__", variableName)
}
```

**Environment Sources:**
1. **Explicitly Set**: `export API_TOKEN=xyz`
2. **`.env` File**: Loaded automatically from working directory
3. **Custom `.env`**: Via `--env custom.env` flag

### **4. Complex Data Warning System**
When variables contain objects or arrays:

```yaml
variables:
  vars:
    user_data: {"name": "Alice", "age": 30}

steps:
  # ❌ This triggers a warning
  - name: "Process user"
    action: log
    args: ["User: ${user_data}"]
    # Output: "User: __UNRESOLVED_user_data__"
    # Warning: "Use jq to extract specific fields from complex data"

  # ✅ Correct approach
  - name: "Extract user name"
    action: jq
    args: ["${user_data}", ".name"]
    result: user_name
  
  - name: "Process user name"
    action: log
    args: ["User: ${user_name}"]
    # Output: "User: Alice"
```

## Unresolved Variable Handling

### **Unresolved Markers**
When variables cannot be resolved:

```yaml
# If 'missing_var' is not defined
- action: log
  args: ["Value: ${missing_var}"]
  # Becomes: "Value: __UNRESOLVED_missing_var__"

# If 'MISSING_ENV' environment variable is not set
- action: log
  args: ["API: ${ENV:MISSING_ENV}"]
  # Becomes: "API: __UNRESOLVED_ENV:MISSING_ENV__"
```

### **Critical Action Validation**
Some actions validate that all variables are resolved:

```go
func validateArgsResolved(actionName string, args []any) *types.ActionResult {
    for i, arg := range args {
        if argStr, ok := arg.(string); ok {
            if strings.Contains(argStr, "__UNRESOLVED_") {
                return &types.ActionResult{
                    Status: constants.ActionStatusError,
                    ErrorInfo: &types.ErrorInfo{
                        Category: types.ErrorCategoryVariable,
                        Code:     "UNRESOLVED_VARIABLE",
                        Message:  fmt.Sprintf("Unresolved variable in %s argument", actionName),
                    },
                }
            }
        }
    }
    return nil
}
```

**Critical Actions** (require all variables resolved):
- `assert` - Must have concrete values for comparison
- `http` - URLs and data must be fully resolved
- `postgres` - Connection strings must be complete

## Resolution Examples

### **Simple Substitution**
```yaml
variables:
  vars:
    name: "Alice"
    age: 25

# Input: "User ${name} is ${age} years old"
# Output: "User Alice is 25 years old"
```

### **Environment Integration**
```bash
# .env file
API_HOST=api.production.com
API_TOKEN=prod_token_abc123
DB_PASSWORD=secure_password_xyz789
```

```yaml
variables:
  vars:
    api_url: "https://${ENV:API_HOST}/v1"
    db_url: "postgres://user:${ENV:DB_PASSWORD}@localhost/db"

# api_url resolves to: "https://api.production.com/v1"
# db_url resolves to: "postgres://user:secure_password_xyz789@localhost/db"
```

### **Nested Resolution**
```yaml
variables:
  vars:
    protocol: "https"
    host: "${ENV:API_HOST}"
    base_url: "${protocol}://${host}"

# Resolution order:
# 1. ${ENV:API_HOST} → "api.example.com"
# 2. ${host} → "api.example.com"  
# 3. ${protocol} → "https"
# 4. Final: "https://api.example.com"
```

### **Complex Data Extraction**
```yaml
steps:
  - name: "Get user data"
    action: http
    args: ["GET", "/users/123"]
    result: user_response

  # ❌ Wrong - complex object substitution
  - name: "Log user"
    action: log
    args: ["User: ${user_response}"]
    # Result: "User: __UNRESOLVED_user_response__"

  # ✅ Right - extract specific fields with jq
  - name: "Extract user name"
    action: jq
    args: ["${user_response}", ".body.data.name"]
    result: user_name
    
  - name: "Log user name"
    action: log
    args: ["User name: ${user_name}"]
    # Result: "User name: Alice"
```

## Performance Characteristics

### **Resolution Performance**
- **Pattern Scanning**: O(n) where n is string length
- **Variable Lookup**: O(1) map access for stored variables
- **Environment Lookup**: O(1) system call for environment variables
- **String Substitution**: O(n) string replacement

### **Memory Usage**
- **Variable Storage**: Minimal overhead, native Go map
- **Pattern Caching**: Compiled regex patterns cached
- **String Operations**: Efficient built-in string replacement

### **Optimization Strategies**
- **Lazy Resolution**: Variables resolved only when accessed
- **Single Pass**: Multiple patterns resolved in one scan
- **Early Exit**: No scanning if no `${` patterns found

## Error Handling & Debugging

### **Common Resolution Issues**

**1. Typos in Variable Names**
```yaml
variables:
  vars:
    api_url: "https://api.example.com"

# Typo in variable name
- action: http
  args: ["GET", "${api_ur}/users"]  # Missing 'l'
  # Result: "__UNRESOLVED_api_ur__"
```

**2. Missing Environment Variables**
```bash
# Missing in environment
./robogo run test.yaml
# Results in: "__UNRESOLVED_ENV:MISSING_VAR__"
```

**3. Complex Data Access**
```yaml
# Wrong approach
- action: log
  args: ["Data: ${complex_object}"]
  # Gets: "__UNRESOLVED_complex_object__"

# Right approach  
- action: jq
  args: ["${complex_object}", ".field"]
  result: extracted_field
```

### **Debugging Variable Resolution**
```yaml
# Debug all variables
- name: "Debug variables"
  action: variable
  args: ["debug"]
  # Shows all stored variables and their values

# Check environment loading
- name: "Check environment"
  action: log
  args: ["DB Host: ${ENV:DB_HOST}"]
  # Shows resolved environment variable or unresolved marker
```

This variable resolution system provides the foundation for dynamic, secure test execution while maintaining simplicity and clear error reporting for debugging complex variable scenarios.