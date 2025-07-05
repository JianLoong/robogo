# Robogo VS Code Extension - Features Guide

## 🎯 Overview

The Robogo VS Code extension provides comprehensive support for the Robogo test automation framework, making it easy to write, debug, and execute test cases with rich IDE features.

## ✨ Core Features

### 1. Syntax Highlighting

**Supported File Types:**
- `.robogo` - Native Robogo test files
- `.yaml` - YAML files with Robogo support
- `.yml` - YAML files with Robogo support

**Highlighted Elements:**
- **Keywords**: `testcase`, `description`, `steps`, `action`, `args`, `result`, `verbose`, `variables`, `secrets`
- **Actions**: `log`, `sleep`, `assert`, `http`, `http_get`, `http_post`, `postgres`, `variable`, `control`
- **Variables**: `${variable_name}` syntax
- **Comments**: `# comment` syntax
- **Strings**: Quoted strings with variable interpolation
- **Numbers**: Integers and floats
- **Booleans**: `true`, `false`

### 2. Intelligent Autocompletion

**Context-Aware Suggestions:**

#### Action Names
- Type `action:` and press `Ctrl+Space`
- Shows all available actions with descriptions
- Fetched from CLI for accuracy

#### Variable Names
- Type `${` to see variable suggestions
- Extracts from `result:` fields, `vars:` section, and `secrets:` section
- Real-time variable discovery

#### HTTP Methods and Headers
- In HTTP action context, suggests methods: `GET`, `POST`, `PUT`, `DELETE`, etc.
- Common headers: `Content-Type`, `Accept`, `Authorization`, etc.
- Certificate options: `cert`, `key`, `ca`

#### Time Formats
- In `get_time` context, suggests formats: `iso`, `datetime`, `date`, `time`, `unix`, etc.

#### PostgreSQL Operations
- In `postgres` context, suggests: `connect`, `query`, `execute`, `close`

#### Variable Operations
- In `variable` context, suggests: `set_variable`, `get_variable`, `list_variables`

#### Verbosity Levels
- Type `verbose:` for options: `true`, `false`, `"basic"`, `"detailed"`, `"debug"`

### 3. Hover Documentation

**Rich Documentation on Hover:**
- **Action Details**: Parameter descriptions, types, requirements
- **Usage Examples**: Complete YAML examples
- **Return Values**: What each action returns
- **Related Actions**: Similar actions for the same functionality
- **Implementation Notes**: Important considerations and best practices

**Example Documentation:**
```markdown
## http

Perform HTTP requests with support for all HTTP methods, custom headers, and SSL certificates.

### Parameters
- **`method`** (string) - **Required** - HTTP method (GET, POST, PUT, DELETE, etc.)
- **`url`** (string) - **Required** - Target URL
- **`body`** (string) - Optional - Request body (for POST/PUT/PATCH)
- **`headers`** (object) - Optional - HTTP headers and SSL options

### Returns
HTTPResponse object with status_code, headers, body, and duration

### Example
```yaml
- action: http
  args: ["GET", "https://api.example.com", {"Authorization": "Bearer token"}]
  result: response
```
```

### 4. Code Snippets

**Available Snippets:**

#### Test Structure
- `robogo-test` - Complete test structure with variables and secrets
- `robogo-test-verbose` - Test structure with global verbosity

#### HTTP Operations
- `robogo-http` - Generic HTTP request
- `robogo-http-get` - HTTP GET request
- `robogo-http-post` - HTTP POST request
- `robogo-http-verbose` - HTTP request with verbose output

#### Database Operations
- `robogo-postgres-query` - PostgreSQL query
- `robogo-postgres-execute` - PostgreSQL execute statement
- `robogo-postgres-connect` - PostgreSQL connection
- `robogo-postgres-verbose` - Database query with detailed verbose output

#### Variable Management
- `robogo-set-variable` - Set variable dynamically
- `robogo-variable-set` - Variable set operation
- `robogo-variable-get` - Variable get operation
- `robogo-variable-verbose` - Variable operation with debug verbose output

#### Control Flow
- `robogo-if` - If/else control flow
- `robogo-for` - For loop control flow
- `robogo-while` - While loop control flow

#### Secrets
- `robogo-file-secret` - File-based secret (single value)
- `robogo-inline-secret` - Inline secret with masking

#### Utility Actions
- `robogo-assert` - Assert two values are equal
- `robogo-get-time` - Get current timestamp
- `robogo-sleep` - Sleep for specified duration
- `robogo-log` - Log a message

### 5. Test Execution

**Run Tests Directly from VS Code:**
- **Right-click** in a `.robogo` file → "Robogo: Run Test"
- **Command Palette** → "Robogo: Run Test"
- **Multiple Output Formats**: Console, JSON, Markdown

**Execution Features:**
- Progress indicators
- Error handling and reporting
- Secret masking in output
- Verbose output support
- Variable substitution

### 6. Action Management

**List and Explore Actions:**
- **Command Palette** → "Robogo: List Actions"
- Shows all available actions with descriptions
- Fetched from CLI for accuracy
- JSON output option for programmatic access

### 7. Verbosity Support

**Multiple Verbosity Levels:**

| Level | Description | Output |
|-------|-------------|--------|
| `false` | No verbose output | Normal operation |
| `true` | Basic verbose | Action + duration |
| `"basic"` | Basic verbose | Action + duration |
| `"detailed"` | Detailed verbose | Args + duration + output |
| `"debug"` | Debug verbose | Everything + verbosity level |

**Usage:**
```yaml
# Global verbosity
testcase: "Test"
verbose: "detailed"

# Step-level verbosity
- action: http_get
  args: ["https://api.example.com"]
  verbose: "debug"
```

## 🔧 Configuration

### Extension Settings

```json
{
  "robogo.executablePath": "robogo",
  "robogo.outputFormat": "console",
  "robogo.showDetailedDocumentation": true
}
```

### Settings Description

- **`robogo.executablePath`**: Path to the Robogo executable (default: "robogo")
- **`robogo.outputFormat`**: Default output format for test execution (`console`, `json`, `markdown`)
- **`robogo.showDetailedDocumentation`**: Enable/disable detailed parameter documentation in hover tooltips

## 🚀 Getting Started

### 1. Installation
```powershell
# From project root
.\install-extension.ps1
```

### 2. Open a Test File
- Open any `.robogo` file
- See syntax highlighting in action

### 3. Try Autocompletion
- Type `action:` and press `Ctrl+Space`
- See action suggestions with descriptions

### 4. Test Hover Documentation
- Hover over any action name
- See detailed documentation and examples

### 5. Use Code Snippets
- Type `robogo-test` and press `Tab`
- Get a complete test structure

### 6. Run a Test
- Right-click in a `.robogo` file
- Select "Robogo: Run Test"
- See results in output panel

## 📚 Best Practices

### 1. Use Descriptive Step Names
```yaml
- name: "Check API health endpoint"  # Good
  action: http_get
  args: ["https://api.example.com/health"]

- action: http_get  # Less clear
  args: ["https://api.example.com/health"]
```

### 2. Leverage Verbosity for Debugging
```yaml
# Use detailed verbosity for complex operations
- name: "Database query"
  action: postgres
  args: ["query", "${db_conn}", "SELECT * FROM users WHERE active = true"]
  verbose: "detailed"
  result: active_users
```

### 3. Use Variables for Reusability
```yaml
variables:
  vars:
    base_url: "https://api.example.com"
    timeout: "30"

steps:
  - action: http_get
    args: ["${base_url}/health"]
    result: health_check
```

### 4. Secure Secret Management
```yaml
variables:
  secrets:
    api_key:
      file: "api-key.txt"
      mask_output: true

steps:
  - action: http_get
    args: ["https://api.example.com/data"]
    # API key automatically masked in output
```

### 5. Use Control Flow for Complex Logic
```yaml
- name: "Conditional check"
  if:
    condition: "${response.status_code} == 200"
    then:
      - action: log
        args: ["Request successful"]
    else:
      - action: log
        args: ["Request failed"]
```

## 🐛 Troubleshooting

### Extension Not Working
1. **Check installation**: Verify extension is in `%USERPROFILE%\.vscode\extensions\robogo.robogo-0.3.0`
2. **Restart VS Code**: Close and reopen VS Code
3. **Check file type**: Ensure you're editing a `.robogo` file

### Autocompletion Not Working
1. **Check Robogo CLI**: Ensure `robogo` is in your PATH
2. **Test CLI**: Run `robogo list` in terminal
3. **Check file syntax**: Ensure valid YAML syntax

### Test Execution Fails
1. **Check Robogo installation**: Run `robogo --version`
2. **Check file syntax**: Ensure valid `.robogo` file
3. **Check permissions**: Ensure executable permissions

### Verbosity Not Working
1. **Check syntax**: Ensure `verbose:` is properly indented
2. **Check values**: Use valid verbosity levels
3. **Check inheritance**: Step-level overrides global setting

## 📖 Additional Resources

- **Demo File**: `.vscode/extensions/robogo/demo.robogo` - Comprehensive feature demonstration
- **Verbosity Test**: `tests/test-verbosity.robogo` - Verbosity feature examples
- **Installation Guide**: `INSTALL.md` - Detailed installation instructions
- **Main README**: `README.md` - Extension overview and quick start

## 🤝 Contributing

To improve the extension:

1. **Edit source**: Modify `.vscode/extensions/robogo/src/extension.ts`
2. **Compile**: Run `npm run compile` in extension directory
3. **Test**: Restart VS Code to see changes
4. **Package**: Run `npm run build` to create new VSIX file

## 📄 License

[Add your license information here] 