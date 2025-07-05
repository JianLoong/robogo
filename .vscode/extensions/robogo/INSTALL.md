# Robogo VS Code Extension Installation Guide

## Quick Installation

### Option 1: Using the Install Script (Recommended)

1. **Navigate to the project root**:
   ```powershell
   cd C:\Users\Jian\Documents\GitHub\robogo
   ```

2. **Run the installation script**:
   ```powershell
   .\install-extension.ps1
   ```

3. **Restart VS Code** to activate the extension

### Option 2: Manual Installation

1. **Copy the extension files**:
   ```powershell
   # Get VS Code extensions directory
   $vscodeExtensionsDir = "$env:USERPROFILE\.vscode\extensions"
   
   # Create extensions directory if it doesn't exist
   if (!(Test-Path $vscodeExtensionsDir)) {
       New-Item -ItemType Directory -Path $vscodeExtensionsDir -Force
   }
   
   # Copy extension files
   Copy-Item -Recurse -Force ".vscode\extensions\robogo" "$vscodeExtensionsDir\robogo.robogo-0.3.0"
   ```

2. **Restart VS Code**

### Option 3: Install from VSIX File

1. **Build the VSIX package**:
   ```powershell
   cd .vscode\extensions\robogo
   npm run build
   ```

2. **Install the VSIX file**:
   - Open VS Code
   - Press `Ctrl+Shift+P` to open Command Palette
   - Type "Extensions: Install from VSIX..."
   - Select the `robogo-0.3.0.vsix` file

## Verification

After installation, you can verify the extension is working:

1. **Open a `.robogo` file** - you should see syntax highlighting
2. **Try autocompletion** - type `action:` and press `Ctrl+Space`
3. **Test hover documentation** - hover over any action name
4. **Run a test** - right-click in a `.robogo` file and select "Robogo: Run Test"

## Features to Test

### 1. Syntax Highlighting
- Open the `demo.robogo` file
- Verify keywords, actions, and variables are highlighted

### 2. Autocompletion
- Type `action:` and press `Ctrl+Space`
- Type `args:` in an HTTP action and see method suggestions
- Type `${` to see variable suggestions
- Type `verbose:` to see verbosity level suggestions

### 3. Hover Documentation
- Hover over any action name (e.g., `http`, `postgres`, `variable`)
- See detailed documentation with examples

### 4. Code Snippets
- Type `robogo-test` and press `Tab` for complete test structure
- Type `robogo-http-get` for HTTP GET request
- Type `robogo-variable-set` for variable management
- Type `robogo-http-verbose` for HTTP with verbose output

### 5. Test Execution
- Right-click in a `.robogo` file
- Select "Robogo: Run Test"
- See results in the output panel

### 6. Action Management
- Press `Ctrl+Shift+P`
- Type "Robogo: List Actions"
- See all available actions with descriptions

### 7. Verbosity Support
- Open `tests/test-verbosity.robogo`
- Run the test to see different verbosity levels in action
- Try adding `verbose: "detailed"` to any step

## Configuration

The extension can be configured in VS Code settings:

```json
{
  "robogo.executablePath": "robogo",
  "robogo.outputFormat": "console",
  "robogo.showDetailedDocumentation": true
}
```

## Troubleshooting

### Extension Not Working
1. **Check installation**: Verify the extension is in `%USERPROFILE%\.vscode\extensions\robogo.robogo-0.3.0`
2. **Restart VS Code**: Close and reopen VS Code
3. **Check output**: View → Output → Robogo for any error messages

### Autocompletion Not Working
1. **Check Robogo executable**: Ensure `robogo` is in your PATH
2. **Test CLI**: Run `robogo list` in terminal to verify CLI works
3. **Check file type**: Ensure you're editing a `.robogo` file

### Test Execution Fails
1. **Check Robogo installation**: Run `robogo --version`
2. **Check file syntax**: Ensure the `.robogo` file is valid YAML
3. **Check permissions**: Ensure you can execute the Robogo binary

## Uninstallation

To uninstall the extension:

```powershell
# Remove the extension directory
Remove-Item -Recurse -Force "$env:USERPROFILE\.vscode\extensions\robogo.robogo-0.3.0"
```

## Support

If you encounter issues:

1. **Check the demo file**: `demo.robogo` shows all features working
2. **Review the README**: `.vscode\extensions\robogo\README.md` has detailed documentation
3. **Test with simple file**: Create a minimal `.robogo` file to isolate issues

## Development

To modify the extension:

1. **Edit source files**: Modify `.vscode\extensions\robogo\src\extension.ts`
2. **Compile**: Run `npm run compile` in the extension directory
3. **Test**: Restart VS Code to see changes
4. **Package**: Run `npm run build` to create new VSIX file 