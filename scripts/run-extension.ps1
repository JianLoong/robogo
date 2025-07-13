# Robogo VS Code Extension - Quick Run
# Just run this script to build and launch the extension

Write-Host "🚀 Robogo VS Code Extension - Quick Launch" -ForegroundColor Green

$extensionDir = ".vscode/extensions/robogo"

# Step 1: Build the extension
Write-Host "Building extension..." -ForegroundColor Yellow
Push-Location $extensionDir
npm run compile
$buildSuccess = $LASTEXITCODE -eq 0
Pop-Location

if ($buildSuccess) {
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host "🎯 Launching extension in new VS Code window..." -ForegroundColor Cyan
    
    # Launch VS Code with the extension in development mode
    code --new-window --extensionDevelopmentPath="$PWD\.vscode\extensions\robogo"
    
    Write-Host "✨ Extension launched! The new VS Code window has your extension loaded." -ForegroundColor Green
}
else {
    Write-Host "❌ Build failed! Check the errors above." -ForegroundColor Red
    Write-Host "💡 Try running: cd .vscode/extensions/robogo && npm install" -ForegroundColor Yellow
} 