# setup-spanner-emulator.ps1
# Run this script to create the default project, instance, and database in the Spanner emulator

$env:SPANNER_EMULATOR_HOST = "localhost:9010"
$project = "test-project"
$instance = "test-instance"
$database = "test-database"

# Set gcloud to use the emulator
Write-Host "[Spanner Emulator Setup] Configuring gcloud for emulator..."
gcloud config configurations create emulator --no-user-output-enabled 2>$null | Out-Null
gcloud config set project $project --no-user-output-enabled | Out-Null

# Create instance if it doesn't exist
Write-Host "[Spanner Emulator Setup] Checking for instance '$instance'..."
$instanceExists = gcloud spanner instances list | Select-String $instance
if (-not $instanceExists) {
    Write-Host "[Spanner Emulator Setup] Creating instance '$instance'..."
    gcloud spanner instances create $instance --config=emulator-config --description="Test Instance" --nodes=1
}
else {
    Write-Host "[Spanner Emulator Setup] Instance '$instance' already exists."
}

# Drop and recreate database with schema
Write-Host "[Spanner Emulator Setup] Dropping existing database '$database' if it exists..."
gcloud spanner databases delete $database --instance=$instance --quiet 2>$null | Out-Null

Write-Host "[Spanner Emulator Setup] Creating database '$database' with schema..."
gcloud spanner databases create $database --instance=$instance --ddl="CREATE TABLE integration_test (id STRING(MAX) NOT NULL, name STRING(MAX)) PRIMARY KEY (id)" --ddl="CREATE TABLE payments (transaction_id STRING(64) NOT NULL, sender_bic STRING(16), sender_account STRING(34), beneficiary_account STRING(34), amount FLOAT64) PRIMARY KEY (transaction_id)"

Write-Host "[Spanner Emulator Setup] Database '$database' created successfully with table 'integration_test' and 'payments'." 