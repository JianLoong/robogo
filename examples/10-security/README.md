# Security Examples

Environment variables, data masking, and secure operations examples.

## Examples

### 17-env-var-test.yaml - Environment Variables
**Complexity:** Intermediate  
**Prerequisites:** Environment variable setup  
**Description:** Demonstrates secure credential management using environment variables.

**What you'll learn:**
- `${ENV:VARIABLE}` syntax for environment variables
- Secure credential handling
- Environment variable validation
- Production-ready configuration patterns

**Setup:**
```bash
export TEST_ENV_VAR="test_value"
export API_TOKEN="your_secret_token"
```

**Run it:**
```bash
./robogo run examples/10-security/17-env-var-test.yaml
```

### 18-test-env-missing.yaml - Missing Environment Variables
**Complexity:** Intermediate  
**Prerequisites:** None (demonstrates missing variables)  
**Description:** Handles missing environment variables gracefully with proper error messages.

**What you'll learn:**
- Environment variable validation
- Graceful error handling for missing variables
- Error message clarity
- Defensive programming patterns

**Run it:**
```bash
./robogo run examples/10-security/18-test-env-missing.yaml
```

### 19-no-log-security.yaml - No-Log Mode
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Complete logging suppression for sensitive operations.

**What you'll learn:**
- `no_log: true` for sensitive operations
- Complete output suppression
- Security-first testing patterns
- Sensitive operation handling

**Run it:**
```bash
./robogo run examples/10-security/19-no-log-security.yaml
```

### 20-step-level-masking.yaml - Step-Level Data Masking
**Complexity:** Advanced  
**Prerequisites:** None  
**Description:** Custom field masking with fine-grained controls.

**What you'll learn:**
- `sensitive_fields` configuration
- Custom data masking patterns
- Field-specific security controls
- Granular security configuration

**Run it:**
```bash
./robogo run examples/10-security/20-step-level-masking.yaml
```

## Key Security Concepts

### Environment Variables
```yaml
variables:
  vars:
    # Secure credential access
    api_token: "${ENV:API_TOKEN}"
    db_password: "${ENV:DB_PASSWORD}"
    
    # With fallback values (use carefully)
    debug_mode: "${ENV:DEBUG_MODE:false}"

steps:
  - name: "Secure API call"
    action: http
    args: ["GET", "https://api.example.com/data"]
    options:
      headers:
        Authorization: "Bearer ${api_token}"
```

### No-Log Mode
```yaml
steps:
  - name: "Sensitive operation"
    action: http
    args: ["POST", "/auth/login"]
    options:
      json:
        username: "${ENV:USERNAME}"
        password: "${ENV:PASSWORD}"
    no_log: true  # Suppresses all output for this step
    result: auth_response
```

### Sensitive Field Masking
```yaml
steps:
  - name: "API call with sensitive data"
    action: http
    args: ["POST", "/api/user"]
    options:
      json:
        name: "John Doe"
        email: "john@example.com"
        password: "${ENV:USER_PASSWORD}"
        api_key: "${ENV:API_KEY}"
    sensitive_fields: ["password", "api_key"]  # These fields will be masked in logs
    result: user_response
```

### Combined Security Features
```yaml
steps:
  - name: "Highly sensitive operation"
    action: http
    args: ["POST", "/payment/process"]
    options:
      json:
        card_number: "${ENV:CARD_NUMBER}"
        cvv: "${ENV:CVV}"
        amount: "100.00"
    sensitive_fields: ["card_number", "cvv"]
    no_log: true  # Complete suppression for maximum security
    result: payment_response
```

## Security Best Practices

### 1. Environment Variable Management
```bash
# Use .env files for development
cp .env.example .env
# Edit .env with your values (never commit this file)

# For production, use proper secret management
export API_TOKEN="$(vault kv get -field=token secret/api)"
```

### 2. Automatic Masking
Robogo automatically masks these field names:
- `password`, `passwd`, `pwd`
- `token`, `auth_token`, `access_token`
- `key`, `api_key`, `secret_key`
- `secret`, `client_secret`
- `credential`, `credentials`

### 3. Custom Masking
```yaml
# Add custom sensitive fields
sensitive_fields: ["ssn", "credit_card", "personal_id"]
```

### 4. Conditional Security
```yaml
- name: "Debug information"
  if: "${ENV:DEBUG_MODE} == 'true'"
  action: log
  args: ["Debug: ${response_data}"]
  sensitive_fields: ["response_data"]  # Mask even debug info
```

### 5. Secure Assertions
```yaml
# Don't expose sensitive data in assertions
- name: "Verify authentication"
  action: assert
  args: ["${auth_response.status}", "==", "200"]
  # Instead of: args: ["${auth_response.token}", "!=", ""]
```

## Environment File Template

Create a `.env` file with these patterns:

```bash
# API Credentials
API_BASE_URL=https://api.example.com
API_TOKEN=your_secret_token_here
API_KEY=your_api_key_here

# Database Credentials
DB_HOST=localhost
DB_PORT=5432
DB_USER=username
DB_PASSWORD=secure_password
DB_NAME=database_name

# Authentication
USERNAME=test_user
PASSWORD=secure_password

# Feature Flags
DEBUG_MODE=false
ENABLE_LOGGING=true

# External Services
KAFKA_BROKERS=localhost:9092
REDIS_URL=redis://localhost:6379
```

## Secret Management Philosophy

**Robogo's Approach: External Secret Management**

Robogo intentionally **does not implement built-in secret management** and follows the principle that **secret management is an infrastructure responsibility**, not an automation tool responsibility.

### Why External Secret Management?

- **🔐 Security Best Practice**: Use proven, audited systems (Vault, AWS Secrets Manager, etc.)
- **🏗️ Infrastructure Separation**: Secret management belongs in deployment/infrastructure layer
- **🔌 Standards Compliance**: Works with enterprise security policies and audit requirements
- **🚫 No Reinvention**: Avoid building custom security infrastructure
- **📈 Scalability**: Integrates with existing organizational secret management

### Integration Patterns

**Development Environment:**
```bash
# Use .env files (never commit)
echo "API_TOKEN=dev-token-123" >> .env
./robogo run workflow.yaml
```

**Production Environments:**
```bash
# HashiCorp Vault
export API_TOKEN="$(vault kv get -field=token secret/myapp/api)"

# AWS Secrets Manager
export DB_PASSWORD="$(aws secretsmanager get-secret-value --secret-id prod/db --query SecretString --output text | jq -r .password)"

# Azure Key Vault
export CERT_PASSWORD="$(az keyvault secret show --vault-name MyVault --name cert-password --query value -o tsv)"

# Kubernetes Secrets (mounted as files)
export DATABASE_URL="$(cat /var/secrets/database-url)"
```

**CI/CD Pipeline Examples:**
```yaml
# GitHub Actions
- name: Run Robogo Workflow
  env:
    API_TOKEN: ${{ secrets.API_TOKEN }}
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  run: ./robogo run production-workflow.yaml

# GitLab CI
script:
  - export API_TOKEN="$VAULT_API_TOKEN"
  - ./robogo run deployment-workflow.yaml

# Jenkins with Vault
script:
  - export DATABASE_URL=$(vault kv get -field=url secret/database)
  - ./robogo run integration-tests.yaml
```

### What Robogo Provides

**✅ Secure Consumption:**
- Environment variable injection (`${ENV:SECRET}`)
- Automatic sensitive field masking
- No-log mode for sensitive operations
- Custom masking configuration

**❌ What Robogo Doesn't Do:**
- Secret storage or retrieval
- Credential rotation
- Secret encryption/decryption
- Integration with specific secret stores

### Security Checklist

**Development:**
- ✅ Use environment variables for all secrets (`${ENV:SECRET}`)
- ✅ Never hardcode credentials in YAML files
- ✅ Use `.env` files for development (add to `.gitignore`)
- ✅ Use `no_log: true` for sensitive operations
- ✅ Configure `sensitive_fields` for custom data
- ✅ Validate environment variables exist before use

**Production:**
- ✅ Integrate with proper secret management system (Vault, AWS, Azure, etc.)
- ✅ Use pipeline/infrastructure to inject secrets as environment variables
- ✅ Implement secret rotation at infrastructure level
- ✅ Test both success and failure scenarios
- ✅ Document required environment variables
- ✅ Follow organizational security policies
- ✅ Regular security audits of secret access patterns