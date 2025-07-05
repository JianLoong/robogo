# Robogo VS Code Extension

This extension provides enhanced support for Robogo test automation framework files in VS Code.

## Features

### 🎯 Autocompletion
- **Keyword Autocompletion**: Get suggestions for available keywords as you type
- **Structure Autocompletion**: Auto-complete common YAML/Robogo structure elements
- **Argument Suggestions**: See available arguments for each keyword

### 📖 Hover Documentation
- **Keyword Documentation**: Hover over keywords to see detailed documentation
- **Argument Information**: View expected arguments and their types
- **Usage Examples**: See practical examples of how to use each keyword

### 🎨 Syntax Highlighting
- **Robogo-specific syntax**: Enhanced highlighting for `.robogo` files
- **YAML compatibility**: Works with both `.yaml` and `.robogo` files
- **Keyword highlighting**: Special highlighting for Robogo keywords

## File Support

The extension supports:
- `.robogo` files (native Robogo format)
- `.yaml` files (YAML format)
- `.yml` files (YAML format)

## Configuration

### Executable Path
You can configure the path to the Robogo executable in VS Code settings:

```json
{
  "robogo.executablePath": "/path/to/robogo"
}
```

## Usage

1. **Install the extension** in VS Code
2. **Open a `.robogo` file** - the extension will automatically activate
3. **Start typing** - you'll get autocompletion suggestions
4. **Hover over keywords** to see documentation
5. **Use the command palette** to access Robogo commands

## Commands

The extension provides several commands accessible via the command palette:

- `Robogo: List Keywords` - Show all available keywords
- `Robogo: Get Completions` - Get keyword completions for a prefix

## Development

To build the extension:

```bash
cd .vscode/extensions/robogo
npm install
npm run compile
```

## Requirements

- VS Code 1.60.0 or higher
- Robogo CLI installed and accessible in PATH

## Contributing

Contributions are welcome! Please see the main Robogo repository for contribution guidelines. 