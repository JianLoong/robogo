# Robogo VS Code Extension

Enhanced support for the Robogo test automation framework with syntax highlighting, autocompletion, and comprehensive documentation.

## Features

### 🎯 **Enhanced Documentation**
- **Hover Documentation**: Hover over any action name to see detailed documentation including:
  - Parameter descriptions with types and requirements
  - Return value information
  - Usage examples
  - Related actions
  - Implementation notes

- **Signature Help**: Get parameter information when typing action calls
- **Contextual Documentation**: Documentation adapts based on the action context

### 🔧 **Autocompletion**
- Action name suggestions with descriptions
- Parameter autocompletion for specific actions (HTTP methods, time formats, etc.)
- Variable name suggestions
- YAML structure completion

### 🎨 **Syntax Highlighting**
- Custom syntax highlighting for `.robogo` files
- Dark theme support
- Proper YAML syntax highlighting

### ⚡ **Test Execution**
- Run tests directly from VS Code
- Multiple output formats (console, JSON, markdown)
- Progress indicators and error handling

### 📋 **Action Management**
- List all available actions
- View action details and examples
- Search and filter actions

## New Features in v0.3.0

- **Dynamic Variable Management**: Use the `variable` action to set, get, or list variables during test execution.
- **File-based and Inline Secrets**: Load secrets from files (single value per file) or define inline secrets, with output masking.
- **PostgreSQL Support**: Use the `postgres` action for database queries, execution, and connection management.
- **Verbosity Support**: Enable detailed output with multiple verbosity levels (`basic`, `detailed`, `debug`) for better debugging and monitoring.
- **Retry Mechanism**: Comprehensive retry support with configurable backoff strategies, conditions, and jitter.
- **Expanded Snippets**: Quickly scaffold tests, secrets, control flow, HTTP, database actions, verbose operations, and retry configurations with new code snippets.
- **Action List Always Up-to-Date**: The extension fetches actions from the CLI (`robogo list`) for accurate autocompletion and documentation.

## Snippets

- `robogo-set-variable`: Set a variable dynamically
- `robogo-file-secret`: Define a file-based secret (single value)
- `robogo-inline-secret`: Define an inline secret
- `robogo-postgres-query`: PostgreSQL query
- `robogo-postgres-secret`: PostgreSQL query using secret variable
- `robogo-control-if`: Control flow (if)
- `robogo-control-for`: Control flow (for loop)
- `robogo-http`, `robogo-http-get`, `robogo-http-post`: HTTP actions
- `robogo-assert`, `robogo-get-time`, `robogo-sleep`, `robogo-log`: Utility actions
- `robogo-test`: Complete test structure
- `robogo-http-verbose`: HTTP request with verbose output
- `robogo-postgres-verbose`: Database query with detailed verbose output
- `robogo-variable-verbose`: Variable operation with debug verbose output
- `robogo-test-verbose`: Complete test with global verbose setting
- `robogo-retry-basic`: Basic retry configuration with fixed delay
- `robogo-retry-exponential`: Retry with exponential backoff strategy
- `robogo-retry-linear`: Retry with linear backoff and jitter
- `robogo-retry-rate-limit`: Retry configuration for rate limiting scenarios
- `robogo-http-retry`: HTTP request with comprehensive retry configuration

## Action List/Autocomplete

The extension fetches actions from the CLI (`robogo completions` or `robogo list`) so it always stays up to date.

## Usage

### Getting Documentation

1. **Hover Documentation**: Simply hover over any action name in your `.robogo` file to see comprehensive documentation:
   ```yaml
   - action: http  # ← Hover here for detailed docs
     args: ["GET", "https://api.example.com"]
   ```

2. **Signature Help**: Type an action name and press `Ctrl+Shift+Space` (or `Cmd+Shift+Space` on Mac) to see parameter information.

3. **Autocompletion**: Start typing an action name and press `Ctrl+Space` to see suggestions with descriptions.

### Running Tests

1. **Command Palette**: Press `Ctrl+Shift+P` and run "Robogo: Run Test"
2. **Context Menu**: Right-click in a `.robogo` file and select "Robogo: Run Test"
3. **Output**: Results appear in the output panel or as a new document (for JSON format)

### Listing Actions

1. **Command Palette**: Press `Ctrl+Shift+P` and run "Robogo: List Actions"
2. **Output**: All available actions with descriptions appear in the output panel

## Retry Mechanism

Robogo provides comprehensive retry functionality to handle transient failures and improve test reliability. **The retry block can be used with any action, not just HTTP.**

### Retry for Any Action

You can add a `retry` block to any step, regardless of the action type. This is useful for handling transient errors in assertions, database operations, variable management, and more.

#### Example: Retry for Assertion
```yaml
- name: "Assert with retry"
  action: assert
  args: [1, 2, "Should eventually match"]
  retry:
    attempts: 5
    delay: 1s
    conditions: ["all"]
```

#### Example: Retry for Database Query
```yaml
- name: "Retry DB query"
  action: postgres
  args: ["query", "postgres://user:pass@localhost/db", "SELECT 1"]
  retry:
    attempts: 3
    delay: 2s
    backoff: "linear"
    conditions: ["connection_error", "5xx"]
```

#### Example: Retry for Variable Set
```yaml
- name: "Set variable with retry"
  action: variable
  args: ["set_variable", "foo", "bar"]
  retry:
    attempts: 2
    delay: 1s
    conditions: ["all"]
```

> **Note:** The retry logic is applied at the step level. Any action can be retried if it returns an error and the retry conditions match.

### Retry Configuration

```yaml
- name: "HTTP request with retry"
  action: http
  args:
    - "GET"
    - "https://api.example.com"
  retry:
    attempts: 3              # Number of retry attempts
    delay: 1s               # Base delay between retries
    backoff: "exponential"  # Backoff strategy: fixed, linear, exponential
    max_delay: 10s          # Maximum delay cap
    jitter: true            # Add randomness to delay
    conditions: ["5xx", "timeout"]  # When to retry
```

### Backoff Strategies

| Strategy | Description | Example Delays |
|----------|-------------|----------------|
| `fixed` | Same delay for all retries | 1s, 1s, 1s |
| `linear` | Linear increase in delay | 1s, 2s, 3s |
| `exponential` | Exponential increase in delay | 1s, 2s, 4s, 8s |

### Retry Conditions

| Condition | Description | Triggers |
|-----------|-------------|----------|
| `5xx` | Server errors | 500, 502, 503, 504 |
| `4xx` | Client errors (rate limiting) | 429 |
| `timeout` | Request timeouts | Timeout errors |
| `connection_error` | Network issues | Connection refused, unreachable |
| `rate_limit` | Rate limiting | 429 status codes |
| `all` | Retry on any error | All error conditions |

### Usage Examples

#### Basic Retry
```yaml
- action: http_get
  args: ["https://api.example.com"]
  retry:
    attempts: 3
    delay: 1s
    conditions: ["5xx"]
```

#### Exponential Backoff
```yaml
- action: http_post
  args: ["https://api.example.com", "{\"data\": \"value\"}"]
  retry:
    attempts: 4
    delay: 1s
    backoff: "exponential"
    max_delay: 10s
    conditions: ["5xx", "timeout"]
```

#### Rate Limit Handling
```yaml
- action: http_get
  args: ["https://api.example.com"]
  retry:
    attempts: 2
    delay: 3s
    backoff: "fixed"
    conditions: ["rate_limit", "4xx"]
```

#### Connection Error Recovery
```yaml
- action: http_get
  args: ["https://api.example.com"]
  retry:
    attempts: 3
    delay: 2s
    backoff: "linear"
    jitter: true
    conditions: ["connection_error"]
```

### Retry Output

When retries occur, you'll see output like:
```
🔄 Attempt 1/3 (delay: 1s): HTTP 500 error
🔄 Attempt 2/3 (delay: 2s): HTTP 500 error
✅ Success after 3 attempts
```

## Verbosity Support

Robogo supports multiple verbosity levels to help with debugging and monitoring:

### Verbosity Levels

| Level | Description | Output |
|-------|-------------|--------|
| `false` | No verbose output | Normal operation |
| `true` | Basic verbose | Action + duration |
| `"basic"` | Basic verbose | Action + duration |
| `"detailed"` | Detailed verbose | Args + duration + output |
| `"debug"` | Debug verbose | Everything + verbosity level |

### Usage Examples

#### Global Verbosity (Test Case Level)
```yaml
testcase: "Verbose Test"
verbose: "detailed"  # All steps get detailed output
steps:
  - action: log
    args: ["All steps will be verbose"]
```

#### Step-Level Verbosity (Overrides Global)
```yaml
- name: "Debug HTTP request"
  action: http_get
  args: ["https://api.example.com"]
  verbose: "debug"  # Overrides global setting
```

#### Disable Verbosity for Specific Step
```yaml
- name: "Silent operation"
  action: log
  args: ["This step will be silent"]
  verbose: false
```

### Verbose Output Examples

#### Basic Verbosity
```
🔍 log: 1.2ms
📝 Output: This step will show basic verbose output
```

#### Detailed Verbosity
```
🔍 Verbose http_get Operation:
   Args: [https://httpbin.org/get]
   Duration: 245ms
   Output: 200
```

#### Debug Verbosity
```
🐛 Debug variable Operation:
   Args: [set_variable test_var debug_value]
   Duration: 0.5ms
   Verbosity Level: debug
   Output: Variable set successfully
```

## Configuration

The extension can be configured through VS Code settings:

```json
{
  "robogo.executablePath": "robogo",
  "robogo.outputFormat": "console",
  "robogo.showDetailedDocumentation": true
}
```

### Settings

- `robogo.executablePath`: Path to the Robogo executable (default: "robogo")
- `robogo.outputFormat`: Default output format for test execution (console, json, markdown)
- `robogo.showDetailedDocumentation`: Enable/disable detailed parameter documentation in hover tooltips

## Documentation Features

### Action Documentation Includes:

1. **Parameter Details**:
   - Parameter names and types
   - Required vs optional parameters
   - Default values
   - Detailed descriptions

2. **Return Values**:
   - What the action returns
   - Return type information
   - Error conditions

3. **Usage Examples**:
   - Complete YAML examples
   - Common use cases
   - Best practices

4. **Related Actions**:
   - Similar actions for the same functionality
   - Alternative approaches
   - Action groupings

5. **Implementation Notes**:
   - Important considerations
   - Performance notes
   - Security considerations

### Example Documentation

When you hover over the `http` action, you'll see:

```markdown
## http

Perform HTTP requests with support for all HTTP methods, custom headers, and SSL certificates.

### Parameters

- **`method`** (string) - **Required**
  HTTP method (GET, POST, PUT, DELETE, etc.)

- **`url`** (string) - **Required**
  Target URL

- **`body`** (string) - Optional
  Request body (for POST/PUT/PATCH)

- **`headers`** (object) - Optional
  HTTP headers and SSL options

### Returns

HTTPResponse object with status_code, headers, body, and duration

### Notes

SSL options: 'cert' (client certificate), 'key' (private key), 'ca' (CA certificate). All can be file paths or PEM content.

### Example

```yaml
- action: http
  args: ["GET", "https://secure.example.com/api", {"cert": "client.crt", "key": "client.key", "ca": "ca.crt", "Authorization": "Bearer ..."}]
  result: response
```

### Related Actions

- `http_get` - Perform HTTP GET request
- `http_post` - Perform HTTP POST request
```

## Supported File Types

- `.robogo` - Robogo test files
- `.yaml` - YAML files (with Robogo support)
- `.yml` - YAML files (with Robogo support)

## Commands

| Command | Description | Shortcut |
|---------|-------------|----------|
| `robogo.runTest` | Run the current test file | Context menu |
| `robogo.listActions` | List all available actions | Command palette |

## Requirements

- VS Code 1.60.0 or higher
- Robogo executable installed and accessible in PATH

## Installation

1. Clone or download this extension
2. Run the installation script: `./install-extension.ps1`
3. Restart VS Code
4. Open a `.robogo` file to test the features

## Troubleshooting

### Extension Not Working
1. Ensure Robogo is installed and accessible: `robogo --version`
2. Check the executable path in settings
3. Restart VS Code after installation

### Documentation Not Showing
1. Verify the file has a `.robogo` extension
2. Check that `robogo.showDetailedDocumentation` is enabled
3. Try hovering over action names (not just anywhere in the file)

### Autocompletion Issues
1. Make sure you're typing in the correct context (after `action:`)
2. Check that the Robogo executable can run `robogo list --output json`
3. Verify the extension is activated for the file type

## Contributing

To contribute to this extension:

1. Fork the repository
2. Make your changes in the `src/` directory
3. Run `npm run compile` to build
4. Test your changes
5. Submit a pull request

## License

This extension is part of the Robogo project and follows the same license terms. 

## Example Test Case

```yaml
testcase: "Variable and Secret Example"
description: "Demonstrate variable and secret usage"

variables:
  vars:
    db_host: "localhost"
    db_port: "5432"
    db_name: "postgres"
    db_user: "postgres"
  secrets:
    db_password:
      file: "secret.txt"
      mask_output: true

steps:
  - name: "Set a dynamic variable"
    action: variable
    args: ["set_variable", "dynamic_var", "dynamic_value"]
    result: set_result
  - name: "Log dynamic variable"
    action: log
    args: ["Dynamic variable set result: ${set_result}"]
  - name: "Show masked secret"
    action: log
    args: ["Database password: ${db_password}"]
``` 