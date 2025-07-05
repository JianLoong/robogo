# Database Setup for Parallel Testing

This document provides the PostgreSQL setup commands needed for parallel database testing in Robogo.

## Database Configuration

The test files are configured to use:
- **Database**: `robogo_testdb`
- **User**: `robogo_testuser`
- **Host**: `192.168.0.174:5432`
- **Password**: Stored in `secret.txt`

## PostgreSQL Setup Commands

### 1. Create Test Database

Connect to PostgreSQL as a superuser (e.g., `postgres`) and run:

```sql
-- Create the test database
CREATE DATABASE robogo_testdb;

-- Verify the database was created
\l robogo_testdb
```

### 2. Create Test User

```sql
-- Create the test user
CREATE USER robogo_testuser WITH PASSWORD 'your_secure_password';

-- Verify the user was created
\du robogo_testuser
```

### 3. Grant Permissions

```sql
-- Grant all privileges on the test database to the test user
GRANT ALL PRIVILEGES ON DATABASE robogo_testdb TO robogo_testuser;

-- Connect to the test database
\c robogo_testdb

-- Grant schema permissions (important for parallel testing)
GRANT ALL ON SCHEMA public TO robogo_testuser;
GRANT CREATE ON SCHEMA public TO robogo_testuser;

-- Grant table creation permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO robogo_testuser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO robogo_testuser;

-- Verify permissions
\dp
```

### 4. Test User Permissions

```sql
-- Test that the user can connect and create tables
\c robogo_testdb robogo_testuser

-- Test table creation
CREATE TABLE test_permissions (id SERIAL PRIMARY KEY, name VARCHAR(100));

-- Test insert
INSERT INTO test_permissions (name) VALUES ('test');

-- Test select
SELECT * FROM test_permissions;

-- Test update
UPDATE test_permissions SET name = 'updated' WHERE id = 1;

-- Test delete
DELETE FROM test_permissions WHERE id = 1;

-- Clean up
DROP TABLE test_permissions;
```

### 5. Update Secret File

Update the `secret.txt` file with the password you set for `robogo_testuser`:

```bash
# Replace 'your_secure_password' with the actual password
echo "your_secure_password" > secret.txt
```

## Parallel Testing Features

The updated test files include:

### 1. Unique Table Names
- Tables are created with timestamps to prevent conflicts
- Each test run gets isolated table names
- Example: `test_users_1703123456789`

### 2. Proper Cleanup
- Tables are dropped after each test
- Connections are properly closed
- No leftover data between test runs

### 3. Connection Management
- Dynamic connection string building
- Secret-based password management
- Proper connection pooling

## Test Files Updated

The following test files have been updated for parallel testing:

1. **`tests/test-postgres.robogo`**
   - Uses `robogo_testdb` and `robogo_testuser`
   - Creates unique table names with timestamps
   - Proper cleanup and isolation

2. **`tests/test-tdm.robogo`**
   - Updated database configuration
   - Unique table prefixes for parallel runs
   - Enhanced setup and teardown

3. **`tests/test-parallel-db.robogo`**
   - Dedicated test database configuration
   - Parallel batch operations support
   - Table isolation for concurrent runs

4. **`tests/test-db-setup.robogo`** (new)
   - Verifies database configuration
   - Tests basic CRUD operations
   - Validates permissions

## Running Tests

### 1. Verify Setup
```bash
./robogo.exe run tests/test-db-setup.robogo
```

### 2. Run Individual Tests
```bash
./robogo.exe run tests/test-postgres.robogo
./robogo.exe run tests/test-tdm.robogo
```

### 3. Run Parallel Tests
```bash
./robogo.exe run tests/test-parallel-db.robogo
```

## Troubleshooting

### Permission Denied Errors
If you get "permission denied for schema public" errors:

```sql
-- Connect as superuser to robogo_testdb
\c robogo_testdb postgres

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO robogo_testuser;
GRANT CREATE ON SCHEMA public TO robogo_testuser;
```

### Connection Errors
- Verify the database exists: `\l robogo_testdb`
- Verify the user exists: `\du robogo_testuser`
- Check password in `secret.txt`
- Ensure PostgreSQL is running and accessible

### Table Creation Errors
- Verify user has CREATE permission on schema
- Check if tables already exist (they should be unique now)
- Ensure database connection is working

## Security Notes

- The test user has limited permissions only on the test database
- Passwords are stored in `secret.txt` and masked in output
- Test tables are isolated and cleaned up after each run
- No production data is accessed during testing

## Performance Considerations

- Each test creates unique tables to prevent conflicts
- Tables are dropped after tests to free up space
- Connection pooling is used for efficiency
- Parallel operations are limited to prevent database overload 