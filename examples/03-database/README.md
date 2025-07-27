# Database Examples

PostgreSQL, Google Cloud Spanner, and database data extraction examples.

## Prerequisites

Start the database services:
```bash
docker-compose up -d
```

This will start:
- PostgreSQL: localhost:5432 (user: robogo_testuser, pass: robogo_testpass, db: robogo_testdb)
- Spanner Emulator: localhost:9010

## Examples

### 03-postgres-basic.yaml - Basic PostgreSQL Operations
**Complexity:** Beginner  
**Prerequisites:** PostgreSQL service running  
**Description:** Basic PostgreSQL queries and operations.

**What you'll learn:**
- PostgreSQL connection strings
- Basic SQL queries with the `postgres` action
- Database result handling
- Connection management

**Run it:**
```bash
./robogo run examples/03-database/03-postgres-basic.yaml
```

### 03-postgres-secure.yaml - Secure Database Connections
**Complexity:** Intermediate  
**Prerequisites:** PostgreSQL service + .env file  
**Description:** PostgreSQL operations using environment variables for credentials.

**What you'll learn:**
- Environment variable usage for database credentials
- Secure connection string construction
- `.env` file integration
- Production-ready database testing

**Setup:**
```bash
# Ensure .env file exists with database credentials
cp .env.example .env
```

**Run it:**
```bash
./robogo run examples/03-database/03-postgres-secure.yaml
```

### 04-postgres-advanced.yaml - Advanced Database Operations
**Complexity:** Advanced  
**Prerequisites:** PostgreSQL service running  
**Description:** Advanced PostgreSQL operations including transactions and complex queries.

**What you'll learn:**
- Complex SQL queries
- Transaction handling
- Advanced database patterns
- Error handling in database operations

**Run it:**
```bash
./robogo run examples/03-database/04-postgres-advanced.yaml
```

### 06-spanner-basic.yaml - Google Cloud Spanner
**Complexity:** Intermediate  
**Prerequisites:** Spanner emulator running  
**Description:** Basic Google Cloud Spanner operations.

**What you'll learn:**
- Spanner connection configuration
- Spanner SQL syntax
- Distributed database operations
- Cloud-native database testing

**Setup:**
```bash
# After starting docker-compose, set up Spanner
# Linux/Mac:
SPANNER_EMULATOR_HOST=localhost:9010 ./setup-spanner.sh
# Windows:
.\setup-spanner.ps1
```

**Run it:**
```bash
./robogo run examples/03-database/06-spanner-basic.yaml
```

### 07-spanner-advanced.yaml - Advanced Spanner Operations
**Complexity:** Advanced  
**Prerequisites:** Spanner emulator + setup  
**Description:** Advanced Google Cloud Spanner operations and patterns.

**What you'll learn:**
- Complex Spanner queries
- Spanner-specific SQL features
- Advanced distributed database patterns
- Performance considerations

**Run it:**
```bash
./robogo run examples/03-database/07-spanner-advanced.yaml
```

### 29-database-extraction.yaml - Database Result Extraction
**Complexity:** Advanced  
**Prerequisites:** PostgreSQL service running  
**Description:** Advanced patterns for extracting and processing database query results.

**What you'll learn:**
- Complex result extraction patterns
- Data transformation from database results
- Multi-step database workflows
- Result validation and processing

**Run it:**
```bash
./robogo run examples/03-database/29-database-extraction.yaml
```

### 40-mongodb-basic.yaml - Basic MongoDB Operations
**Complexity:** Intermediate  
**Prerequisites:** MongoDB service running  
**Description:** Basic MongoDB document operations including insert, find, update, delete, and count.

**What you'll learn:**
- MongoDB connection and authentication
- Document insertion (single and multiple)
- Finding documents with filters and projections
- Updating documents with operators
- Deleting documents
- Counting documents
- Basic aggregation pipelines

**Setup:**
```bash
# Start MongoDB (using Docker)
docker run -d --name mongodb -p 27017:27017 mongo:latest

# Or if you have MongoDB installed locally
mongod --dbpath /path/to/data
```

**Run it:**
```bash
./robogo run examples/03-database/40-mongodb-basic.yaml
```

### 41-mongodb-advanced.yaml - Advanced MongoDB Operations
**Complexity:** Advanced  
**Prerequisites:** MongoDB service running  
**Description:** Advanced MongoDB operations including complex queries, aggregation pipelines, and error handling.

**What you'll learn:**
- Complex query filters with operators
- Array and nested field queries
- Advanced aggregation pipelines
- Data transformation and analytics
- Upsert operations
- Performance considerations
- Error handling patterns

**Run it:**
```bash
./robogo run examples/03-database/41-mongodb-advanced.yaml
```

## Key Concepts

### PostgreSQL Connection
```yaml
variables:
  vars:
    # Direct connection string
    db_url: "postgres://robogo_testuser:robogo_testpass@localhost:5432/robogo_testdb?sslmode=disable"
    
    # Or using environment variables (recommended)
    secure_db_url: "postgres://${ENV:DB_USER}:${ENV:DB_PASSWORD}@${ENV:DB_HOST}:${ENV:DB_PORT}/${ENV:DB_NAME}?sslmode=disable"

steps:
  - name: "Execute query"
    action: postgres
    args: ["query", "${db_url}", "SELECT version()"]
    result: db_result
```

### Spanner Connection
```yaml
variables:
  vars:
    spanner_db: "projects/test-project/instances/test-instance/databases/test-database"

steps:
  - name: "Spanner query"
    action: spanner
    args: ["query", "${spanner_db}", "SELECT 1 as test_value"]
    result: spanner_result
```

### Result Processing
```yaml
# Extract data from database results
- name: "Extract first row"
  action: jq
  args: ["${db_result}", ".rows[0]"]
  result: first_row

# Extract specific column
- name: "Extract column value"
  action: jq
  args: ["${db_result}", ".rows[0].column_name"]
  result: column_value
```

## Environment Variables

For secure database testing, set these in your `.env` file:

```bash
# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=robogo_testuser
DB_PASSWORD=robogo_testpass
DB_NAME=robogo_testdb

# Spanner (if using real GCP)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
SPANNER_PROJECT_ID=your-project-id
SPANNER_INSTANCE_ID=your-instance-id
SPANNER_DATABASE_ID=your-database-id
```

## Common Patterns

- Use environment variables for all database credentials
- Always handle connection errors gracefully
- Use `jq` to extract specific data from query results
- Test both successful queries and error conditions
- Include cleanup operations in teardown sections