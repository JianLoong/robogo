# Robogo VS Code Extension - Quick Reference

## 🚀 Quick Start

1. **Install**: `.\install-extension.ps1`
2. **Open**: Any `.robogo` file
3. **Test**: Right-click → "Robogo: Run Test"

## ⌨️ Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Autocomplete | `Ctrl+Space` |
| Command Palette | `Ctrl+Shift+P` |
| Run Test | Right-click → "Robogo: Run Test" |
| List Actions | `Ctrl+Shift+P` → "Robogo: List Actions" |

## 📝 Code Snippets

| Snippet | Description |
|---------|-------------|
| `robogo-test` | Complete test structure |
| `robogo-http-get` | HTTP GET request |
| `robogo-postgres-query` | PostgreSQL query |
| `robogo-variable-set` | Set variable |
| `robogo-if` | If/else control flow |
| `robogo-http-verbose` | HTTP with verbose output |

## 🔍 Autocompletion Triggers

| Context | Trigger | Suggestions |
|---------|---------|-------------|
| Action names | `action:` | All available actions |
| Variable names | `${` | Variables from document |
| HTTP methods | `args:` in HTTP action | GET, POST, PUT, DELETE |
| Time formats | `args:` in get_time | iso, datetime, date, time |
| Verbosity levels | `verbose:` | true, false, "basic", "detailed", "debug" |

## 📚 Hover Documentation

**Hover over any action name to see:**
- Parameter descriptions
- Usage examples
- Return values
- Related actions
- Implementation notes

## ⚙️ Configuration

```json
{
  "robogo.executablePath": "robogo",
  "robogo.outputFormat": "console",
  "robogo.showDetailedDocumentation": true
}
```

## 🔧 Verbosity Levels

| Level | Output |
|-------|--------|
| `false` | No verbose output |
| `true` | Action + duration |
| `"basic"` | Action + duration |
| `"detailed"` | Args + duration + output |
| `"debug"` | Everything + verbosity level |

## 📋 Common Patterns

### Basic Test Structure
```yaml
testcase: "Test Name"
description: "Test Description"
verbose: "basic"

variables:
  vars:
    base_url: "https://api.example.com"
  secrets:
    api_key:
      file: "api-key.txt"
      mask_output: true

steps:
  - name: "Step name"
    action: http_get
    args: ["${base_url}/health"]
    result: response
```

### HTTP Request with Verbosity
```yaml
- name: "API call"
  action: http
  args: ["POST", "https://api.example.com", '{"key": "value"}']
  verbose: "detailed"
  result: api_response
```

### Database Operation
```yaml
- name: "Database query"
  action: postgres
  args: ["query", "${db_conn}", "SELECT * FROM users"]
  verbose: "debug"
  result: users
```

### Variable Management
```yaml
- name: "Set variable"
  action: variable
  args: ["set_variable", "my_var", "my_value"]
  result: set_result
```

### Control Flow
```yaml
- name: "Conditional check"
  if:
    condition: "${response.status_code} == 200"
    then:
      - action: log
        args: ["Success"]
    else:
      - action: log
        args: ["Failed"]
```

## 🐛 Troubleshooting

### Extension Not Working
- Restart VS Code
- Check file extension (`.robogo`)
- Verify installation path

### Autocompletion Not Working
- Check Robogo CLI: `robogo list`
- Ensure valid YAML syntax
- Check file type

### Test Execution Fails
- Check Robogo: `robogo --version`
- Verify file syntax
- Check permissions

## 📖 More Information

- **Full Documentation**: `README.md`
- **Features Guide**: `FEATURES.md`
- **Installation Guide**: `INSTALL.md`
- **Demo File**: `demo.robogo` 