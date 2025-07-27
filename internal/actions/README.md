# Actions System

This directory contains all action implementations and the action registry system. Actions are the core building blocks that perform actual operations in Robogo tests.

## Action Categories

### Core Actions
- **`assert`** - Test assertions and validations
- **`log`** - Logging and output messages
- **`variable`** - Variable manipulation and setting

### HTTP Actions
- **`http`** - HTTP requests (GET, POST, PUT, DELETE, etc.)
  - Supports all HTTP methods, headers, authentication
  - JSON and form data handling
  - Response validation and data extraction

### Database Actions
- **`postgres`** - PostgreSQL database operations
  - Query execution, transaction support
  - Connection string flexibility
- **`spanner`** - Google Cloud Spanner operations
  - Distributed database queries
  - Cloud-native SQL support
- **`mongodb`** - MongoDB database operations
  - Document operations: find, insert, update, delete
  - Aggregation pipelines and complex queries
  - BSON document handling with native MongoDB protocol

### Messaging Actions
- **`kafka`** - Apache Kafka producer/consumer
  - Topic management, message publishing
  - Consumer group support
- **`rabbitmq`** - RabbitMQ message operations
  - Queue management, message routing
- **`swift_message`** - SWIFT financial messaging
  - MT103 message generation and parsing

### File Actions
- **`file_read`** - Local file reading operations
  - Text and binary file support
  - Multiple format detection
- **`scp`** - Secure file transfer via SSH/SFTP
  - Upload/download operations
  - Password and key-based authentication

### Data Processing Actions
- **`jq`** - JSON data processing and extraction
  - Complex JSON path queries
  - Data transformation and filtering
- **`xpath`** - XML data processing
  - XPath queries and XML manipulation
- **`json_parse`** - JSON parsing and validation
- **`json_build`** - JSON construction from templates
- **`xml_parse`** - XML parsing operations
- **`xml_build`** - XML document construction
- **`csv_parse`** - CSV file and string parsing
  - Configurable delimiters, headers, row limits
  - File path or string content support
  - JSON-compatible structured output

### String Actions
- **`string_random`** - Random string generation
  - Configurable length and character sets
- **`string_replace`** - String find and replace operations
- **`string_format`** - String formatting and templating
- **`string`** - General string operations

### Utility Actions
- **`uuid`** - UUID generation (v4)
- **`time`** - Time operations and formatting
- **`sleep`** - Delays and timing control
- **`ping`** - Network connectivity testing with ICMP ping
  - Cross-platform support (Windows, macOS, Linux)
  - Configurable packet count and timeout
  - DNS resolution and statistics parsing

### Security & Validation Actions
- **`ssl_cert_check`** - SSL certificate validation and analysis
  - Certificate expiry checking and warnings
  - Chain verification and hostname validation
  - Self-signed certificate handling
  - Cross-platform TLS connection testing

### Encoding Actions
- **`base64_encode`/`base64_decode`** - Base64 encoding operations
- **`url_encode`/`url_decode`** - URL encoding operations
- **`hash`** - Cryptographic hashing (MD5, SHA1, SHA256)

## Action Implementation

### Action Function Signature
```go
type ActionFunc func(args []any, options map[string]any, vars *common.Variables) types.ActionResult
```

### Action Registry
- **Registry Pattern**: Actions registered by name in `action_registry.go`
- **No Global State**: Registry is created per TestRunner instance
- **Built-in Actions**: All standard actions auto-registered
- **Extensible**: New actions can be registered dynamically

### Action Structure
```go
func exampleAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
    // 1. Validate arguments
    if len(args) < 2 {
        return types.MissingArgsError("example", 2, len(args))
    }
    
    // 2. Check for unresolved variables
    if errorResult := validateArgsResolved("example", args); errorResult != nil {
        return *errorResult
    }
    
    // 3. Extract and process arguments
    param1 := fmt.Sprintf("%v", args[0])
    param2 := fmt.Sprintf("%v", args[1])
    
    // 4. Process options
    timeout := "30s"
    if t, ok := options["timeout"].(string); ok {
        timeout = t
    }
    
    // 5. Perform operation
    result, err := performOperation(param1, param2, timeout)
    if err != nil {
        return types.RequestError("operation failed", err.Error())
    }
    
    // 6. Return success result
    return types.ActionResult{
        Status: constants.ActionStatusPassed,
        Data:   result,
    }
}
```

## Security Features

### Sensitive Data Masking
- **Automatic Detection**: Passwords, tokens, API keys automatically masked
- **Custom Fields**: `sensitive_fields` option for custom masking
- **Security Context**: Actions receive security information via options
  - `__no_log`: Suppress all logging for this action
  - `sensitive_fields`: Array of field names to mask

### Environment Variables
- Actions can access environment variables via `${ENV:VAR}` syntax
- Secure credential management without hardcoding secrets
- Integration with .env file loading

## Error Handling

### Error Types
- **MissingArgsError**: Insufficient arguments provided
- **InvalidArgError**: Invalid argument format or value
- **RequestError**: Operation failed (network, database, etc.)
- **ValidationError**: Data validation failed

### Error Structure
```go
// Technical errors (system issues)
types.NewErrorBuilder(types.ErrorCategoryNetwork, "CONNECTION_FAILED").
    WithTemplate("Failed to connect to %s: %s").
    WithContext("host", hostname).
    WithContext("error", err.Error()).
    Build(hostname, err.Error())

// Logical failures (test failures)
message := fmt.Sprintf("Expected %v but got %v", expected, actual)
return types.ActionResult{
    Status: constants.ActionStatusFailed,
    FailureInfo: &types.FailureInfo{
        Message: message,
        Expected: expected,
        Actual: actual,
    },
}
```

## Variable Integration

### Variable Substitution
- All action arguments go through variable substitution before execution
- Supports `${variable}` and `${ENV:VARIABLE}` syntax
- Unresolved variables cause action failures with helpful error messages

### Variable Access
```go
// Get variable value
value := vars.Get("my_variable")

// Set variable (usually done by execution strategies)
vars.Set("result_variable", actionResult)

// Check if variable exists
if vars.Has("optional_variable") {
    // Use variable
}
```

## Testing Actions

### Example Test Structure
```yaml
testcase: "HTTP API Test"
steps:
  - name: "Make API request"
    action: http
    args: ["GET", "https://api.example.com/users"]
    options:
      headers:
        Authorization: "Bearer ${ENV:API_TOKEN}"
      timeout: "30s"
    result: api_response
    
  - name: "Validate response"
    action: assert
    args: ["${api_response.status}", "==", "200"]
```

### Action Testing Guidelines
1. **Positive Cases**: Test successful operations
2. **Error Cases**: Test various failure scenarios
3. **Edge Cases**: Test boundary conditions and invalid inputs
4. **Security Cases**: Test sensitive data masking
5. **Integration Cases**: Test with real external services when possible

## Adding New Actions

1. **Create Action Function**: Follow the standard signature and structure
2. **Register Action**: Add to `registerBuiltinActions()` in `action_registry.go`
3. **Add Dependencies**: Update `go.mod` if external libraries needed
4. **Write Tests**: Create example YAML test cases
5. **Document**: Add to this README and main documentation

## Performance Considerations

- **Connection Management**: Actions open/close connections per operation
- **No Connection Pooling**: Keeps architecture simple and prevents hanging processes
- **Timeouts**: All network operations should have configurable timeouts
- **Resource Cleanup**: Always use `defer` for resource cleanup

## File Structure

```
actions/
├── action_registry.go    # Action registration and management
├── assert.go            # Assertion actions
├── encoding.go          # Encoding/decoding actions
├── file.go              # File operation actions
├── http.go              # HTTP request actions
├── jq.go                # JSON processing actions
├── json.go              # JSON manipulation actions
├── kafka.go             # Kafka messaging actions
├── log.go               # Logging actions
├── postgres.go          # PostgreSQL database actions
├── rabbitmq.go          # RabbitMQ messaging actions
├── scp.go               # SCP file transfer actions
├── sleep.go             # Sleep/timing actions
├── spanner.go           # Google Spanner actions
├── string_random.go     # Random string generation
├── string_utils.go      # String manipulation actions
├── swift.go             # SWIFT messaging actions
├── time.go              # Time operations
├── uuid.go              # UUID generation
├── variable.go          # Variable manipulation
├── xml_build.go         # XML construction
├── xml_parse.go         # XML parsing
└── xpath.go             # XPath processing
```