# PowerShell script to setup Spanner emulator instance and database using gcloud

Write-Host "Setting up Spanner emulator..."

# Set emulator endpoint
$env:SPANNER_EMULATOR_HOST = "localhost:9010"

# Wait for emulator to be ready
# Write-Host "Waiting for Spanner emulator..."
# do {
#     try {
#         $response = Invoke-WebRequest -Uri "http://localhost:9020/v1/projects" -Method GET -ErrorAction Stop
#         $ready = $true
#     }
#     catch {
#         Write-Host "Waiting for emulator..."
#         Start-Sleep -Seconds 2
#         $ready = $false
#     }
# } while (-not $ready)

# Create project first
Write-Host "Creating project..."
try {
    gcloud projects create test-project --name="Test Project"
    Write-Host "Project created successfully"
}
catch {
    Write-Host "Project may already exist: $($_.Exception.Message)"
}

Write-Host ""

# Create instance
Write-Host "Creating Spanner instance..."
try {
    gcloud spanner instances create test-instance --config=emulator-config --description="Test instance for Robogo" --nodes=1 --project=test-project
    Write-Host "Instance created successfully"
}
catch {
    Write-Host "Instance creation result: $($_.Exception.Message)"
}

Write-Host ""

# Create database  
Write-Host "Creating Spanner database..."
try {
    gcloud spanner databases create test-database --instance=test-instance --project=test-project
    Write-Host "Database created successfully"
}
catch {
    Write-Host "Database creation result: $($_.Exception.Message)"
}

Write-Host ""
Write-Host "Spanner setup completed!"
Write-Host "Instance: test-instance"
Write-Host "Database: test-database"
Write-Host "Connection string: projects/test-project/instances/test-instance/databases/test-database"