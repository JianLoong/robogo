# Templates Package

The `templates` package provides error message template management for consistent error formatting throughout the Robogo framework.

## Overview

This package serves as a bridge between the error handling system and the template constants, providing a centralized way to access and manage error message templates for consistent user-facing error messages.

## Architecture

### Simple Template System
- **No Complex Templating**: Uses simple string templates with placeholder substitution
- **Constant-Based**: All templates defined as constants in the `constants` package
- **Bridge Pattern**: Acts as an interface between error handling and template storage

### Components

```
internal/templates/
└── init.go                # Template access and initialization functions
```

## Core Functions

### `InitializeErrorTemplates()`
Returns the complete map of error templates from the constants package.

```go
func InitializeErrorTemplates() map[string]string {
    return constants.ErrorTemplates
}
```

**Usage**: Initialize the template system at startup or when needed.

### `GetTemplateConstant(templateName string)`
Retrieves a specific template string by its constant name.

```go
func GetTemplateConstant(templateName string) string {
    templates := constants.ErrorTemplates
    if template, exists := templates[templateName]; exists {
        return template
    }
    return ""
}
```

**Parameters**:
- `templateName`: The constant name of the template (e.g., `constants.TemplateUnknownAction`)

**Returns**: The template string, or empty string if not found

## Template Usage Pattern

### In Error Handling
```go
// Example from basic_strategy.go
errorResult := types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_ACTION").
    WithTemplate(templates.GetTemplateConstant(constants.TemplateUnknownAction)).
    WithContext("action", step.Action).
    WithContext("step", step.Name).
    Build(step.Action)
```

### Template Definition (in constants package)
```go
// From constants/error_templates.go
const TemplateUnknownAction = "TEMPLATE_UNKNOWN_ACTION"

var ErrorTemplates = map[string]string{
    TemplateUnknownAction: "Unknown action '%s' in step '%s'",
    // ... other templates
}
```

## Template Categories

The templates support various error categories:

### **Validation Templates**
- **Unknown Action**: Action not found in registry
- **Invalid Arguments**: Malformed or missing arguments
- **Configuration Errors**: Invalid step or test configuration

### **Execution Templates**
- **Action Failures**: Action-specific error messages
- **Variable Resolution**: Unresolved variable errors
- **Security Violations**: Security-related error messages

### **System Templates**
- **Network Errors**: Connection and timeout messages
- **File System**: File access and permission errors
- **Resource Limits**: Memory, timeout, and capacity errors

## Design Principles

### **Centralized Management**
- All templates stored in constants package
- Single source of truth for error messages
- Consistent formatting across the framework

### **Type Safety**
- Template names as typed constants
- Compile-time checking of template references
- No magic strings in error handling code

### **Separation of Concerns**
- Templates package handles template access
- Constants package stores template definitions
- Error handling focuses on context and formatting

### **Simple Implementation**
- No complex template engines
- Basic string substitution patterns
- Minimal overhead and dependencies

## Usage Examples

### **Basic Template Retrieval**
```go
import "github.com/JianLoong/robogo/internal/templates"
import "github.com/JianLoong/robogo/internal/constants"

// Get a specific template
template := templates.GetTemplateConstant(constants.TemplateUnknownAction)
// Returns: "Unknown action '%s' in step '%s'"

// Use with error builder
errorResult := types.NewErrorBuilder(category, code).
    WithTemplate(template).
    WithContext("action", actionName).
    WithContext("step", stepName).
    Build(actionName)
```

### **Template Initialization**
```go
// Initialize all templates (typically at startup)
allTemplates := templates.InitializeErrorTemplates()

// Access templates directly
template := allTemplates[constants.TemplateUnknownAction]
```

### **Error Message Generation**
```go
// Template: "Failed to connect to %s after %d attempts"
errorMsg := types.NewErrorBuilder(types.ErrorCategoryNetwork, "CONNECTION_FAILED").
    WithTemplate(templates.GetTemplateConstant(constants.TemplateConnectionFailed)).
    WithContext("host", hostname).
    WithContext("attempts", retryCount).
    Build(fmt.Sprintf("connection to %s failed", hostname))
```

## Integration with Error System

### **Error Builder Integration**
The templates package integrates seamlessly with the error builder pattern:

```go
// Step 1: Get template
template := templates.GetTemplateConstant(constants.TemplateSpecific)

// Step 2: Build error with template
error := types.NewErrorBuilder(category, code).
    WithTemplate(template).          // Set template message
    WithContext("key", value).       // Add context data
    Build(primaryMessage)            // Build final error
```

### **Variable Substitution**
Templates support simple placeholder substitution:

```go
// Template: "Variable '${variable}' not found in step '${step}'"
// Context: {"variable": "api_url", "step": "Make API call"}
// Result: "Variable 'api_url' not found in step 'Make API call'"
```

## Best Practices

### **Template Design**
1. **Clear Messages**: Use descriptive, user-friendly language
2. **Context Aware**: Include relevant context in template placeholders
3. **Consistent Tone**: Maintain consistent voice across all templates
4. **Actionable**: Provide hints for resolution when possible

### **Template Usage**
1. **Use Constants**: Always use typed constants for template names
2. **Provide Context**: Include relevant context data for substitution
3. **Fallback Handling**: Handle missing templates gracefully
4. **Performance**: Cache templates when used frequently

### **Error Integration**
1. **Structured Errors**: Use templates with structured error types
2. **Context First**: Add context before building final error
3. **Meaningful Codes**: Use clear error codes with templates
4. **User Experience**: Prioritize clarity for end users

## Performance Considerations

### **Memory Usage**
- **Static Storage**: Templates stored as constants (minimal memory)
- **No Dynamic Allocation**: Template strings are compile-time constants
- **Shared Access**: Single template map shared across all usage

### **Access Performance**
- **O(1) Lookup**: Direct map access for template retrieval
- **No Processing**: Templates returned as-is, no parsing overhead
- **Minimal Function Calls**: Simple bridge functions with low overhead

## Error Handling

### **Missing Templates**
```go
template := templates.GetTemplateConstant("NONEXISTENT_TEMPLATE")
// Returns: "" (empty string)

// Graceful handling
if template == "" {
    template = "An error occurred: %s" // fallback template
}
```

### **Template Validation**
The package provides simple validation through existence checking:

```go
templates := templates.InitializeErrorTemplates()
if _, exists := templates[constants.TemplateSpecific]; !exists {
    // Handle missing template
}
```

## Future Considerations

### **Potential Enhancements**
1. **Template Validation**: Compile-time template format validation
2. **Internationalization**: Multi-language template support
3. **Dynamic Templates**: Runtime template modification for customization
4. **Template Metrics**: Usage tracking for template optimization

### **Current Limitations**
1. **Static Only**: No runtime template modification
2. **Simple Substitution**: No complex template logic
3. **English Only**: Single language support
4. **No Validation**: No format checking for placeholders

The templates package provides a simple, efficient foundation for consistent error messaging while maintaining the framework's KISS principles and avoiding over-engineering common in template systems.

---

**Note**: This is different from the root-level `templates/` directory which contains SWIFT financial message templates for the `swift_message` action. This `internal/templates` package is specifically for error message template management.