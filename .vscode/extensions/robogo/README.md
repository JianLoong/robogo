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
- **Expanded Snippets**: Quickly scaffold tests, secrets, control flow, HTTP, and database actions with new code snippets.
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