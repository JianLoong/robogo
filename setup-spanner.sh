#!/bin/bash
# Simple script to setup Spanner emulator instance and database using gcloud

echo "Setting up Spanner emulator..."

# Set emulator endpoint
export SPANNER_EMULATOR_HOST=localhost:9010

# Wait for emulator to be ready
echo "Waiting for Spanner emulator..."
until curl -f http://localhost:9020/v1/projects >/dev/null 2>&1; do
    echo "Waiting for emulator..."
    sleep 2
done

# Create project first
echo "Creating project..."
gcloud projects create test-project --name="Test Project" || echo "Project may already exist"

echo ""

# Create instance
echo "Creating Spanner instance..."
gcloud spanner instances create test-instance \
    --config=emulator-config \
    --description="Test instance for Robogo" \
    --nodes=1 \
    --project=test-project

echo ""

# Create database  
echo "Creating Spanner database..."
gcloud spanner databases create test-database \
    --instance=test-instance \
    --project=test-project

echo ""
echo "Spanner setup completed!"
echo "Instance: test-instance"
echo "Database: test-database"
echo "Connection string: projects/test-project/instances/test-instance/databases/test-database"