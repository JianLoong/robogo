# Robogo VS Code Extension Installer
# This script installs the Robogo extension locally in VS Code

Write-Host "Installing Robogo VS Code Extension..." -ForegroundColor Green

# Get VS Code extensions directory
$vscodeExtensionsDir = "$env:USERPROFILE\.vscode\extensions"

# Create extensions directory if it doesn't exist
if (!(Test-Path $vscodeExtensionsDir)) {
    New-Item -ItemType Directory -Path $vscodeExtensionsDir -Force
    Write-Host "Created VS Code extensions directory: $vscodeExtensionsDir" -ForegroundColor Yellow
}

# Define source and destination paths
$sourceDir = ".vscode/extensions/robogo"
$destDir = "$vscodeExtensionsDir\robogo.robogo-0.1.0"

# Check if source directory exists
if (!(Test-Path $sourceDir)) {
    Write-Host "Error: Source directory not found: $sourceDir" -ForegroundColor Red
    Write-Host "Make sure you are running this script from the project root directory." -ForegroundColor Red
    exit 1
}

# Remove existing installation if it exists
if (Test-Path $destDir) {
    Write-Host "Removing existing installation..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $destDir
}

# Copy extension files
Write-Host "Copying extension files..." -ForegroundColor Yellow
Copy-Item -Recurse -Force $sourceDir $destDir

# Verify installation
if (Test-Path $destDir) {
    Write-Host "Robogo extension installed successfully!" -ForegroundColor Green
    Write-Host "Location: $destDir" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Restart VS Code" -ForegroundColor White
    Write-Host "2. Open a .robogo file to test autocompletion" -ForegroundColor White
    Write-Host "3. Try hovering over keywords for documentation" -ForegroundColor White
    Write-Host ""
    Write-Host "To uninstall, delete: $destDir" -ForegroundColor Gray
} else {
    Write-Host "Installation failed!" -ForegroundColor Red
    exit 1
} 