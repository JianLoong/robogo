# 🔐 Robogo Secrets Management Examples

This directory demonstrates Robogo's comprehensive secret management capabilities with practical examples.

## 🎯 **What's Demonstrated**

✅ **Clean `SECRETS.var` namespace** (no collision with regular variables)  
✅ **File-based secrets** (recommended for production)  
✅ **Inline secrets** (convenient for development)  
✅ **Masked vs unmasked secrets** (security control)  
✅ **Secrets in HTTP authentication**  
✅ **Debug output with proper masking**  
✅ **Variable substitution with secrets**

## 📂 **Files Overview**

### 🔑 **Secret Files** (`secrets/`)
```
secrets/
├── api-key.txt          # Example API key
├── db-password.txt      # Example database password  
└── jwt-token.txt        # Example JWT token
```

### 🧪 **Example Tests**
```
simple-secrets-demo.robogo     # Clean, focused demo (RECOMMENDED)
secrets-showcase.robogo        # Comprehensive feature demonstration
test-secrets-example.robogo    # Advanced usage patterns
```

## 🚀 **Quick Start**

### 1. **Run the Simple Demo**
```bash
# Basic run
./robogo.exe run examples/simple-secrets-demo.robogo

# With debug output to see secret masking
./robogo.exe run examples/simple-secrets-demo.robogo --debug-vars
```

### 2. **Expected Debug Output**
```
🔍 Variable Resolution Debug (execution):
   ✅ Resolved Variables:
      ${SECRETS.api_key} → [MASKED] (secret from file, masked=true)
      ${SECRETS.session_token} → [MASKED] (secret from inline, masked=true)
      ${SECRETS.api_version} → 1.0.0 (secret from inline, masked=false)
```

## 📝 **Secret Configuration Syntax**

### **File-based Secrets** (Production-Ready)
```yaml
variables:
  secrets:
    api_key:
      file: "path/to/secret.txt"
      mask_output: true  # Will show [MASKED] in debug output
```

### **Inline Secrets** (Development Convenience)
```yaml
variables:
  secrets:
    client_secret:
      value: "secret_value_here"
      mask_output: true  # Will show [MASKED] in debug output
```

### **Unmasked Secrets** (For Non-Sensitive Data)
```yaml
variables:
  secrets:
    api_version:
      value: "v1.2.3"
      mask_output: false  # Will show actual value in debug output
```

## 🔒 **Security Features**

### **1. Automatic Masking in Debug Output**
When using `--debug-vars`, secrets with `mask_output: true` show as:
```
${secret_name} → [MASKED] (secret from file, masked=true)
```

### **2. Source Tracking**
Debug output shows where secrets come from:
- `(secret from file, masked=true)` - File-based secret
- `(secret from inline, masked=false)` - Inline secret

### **3. Flexible Masking Control**
- `mask_output: true` - Secret value hidden in all output
- `mask_output: false` - Secret value visible (for non-sensitive data)

## 🌐 **Usage in HTTP Requests**

```yaml
steps:
  - name: "Authenticated API request"
    action: http
    args: ["GET", "${api_url}/secure-endpoint"]
    options:
      headers:
        Authorization: "Bearer ${SECRETS.api_key}"        # Masked in logs
        X-Session-ID: "${SECRETS.session_token}"         # Masked in logs
        X-API-Version: "${SECRETS.api_version}"          # Visible in logs
```

## 🔄 **Usage in Variable Construction**

```yaml
steps:
  - name: "Build secure database URL"
    action: variable
    args:
      - "set"
      - "db_url"
      - "postgresql://user:${SECRETS.db_password}@host:5432/db"
    # ${SECRETS.db_password} will be masked in debug output
```

## 📊 **Best Practices**

### **✅ DO:**
- Use file-based secrets for production credentials
- Set `mask_output: true` for all sensitive data
- Use descriptive secret names
- Store secret files outside the repository
- Use unmasked secrets only for truly non-sensitive data (versions, public endpoints)

### **❌ DON'T:**
- Commit secret files to version control
- Use inline secrets for production credentials
- Set `mask_output: false` for passwords, tokens, or keys
- Hardcode secrets in test files

## 🧪 **Testing Your Setup**

### **1. Verify Secret Masking**
```bash
./robogo.exe run examples/simple-secrets-demo.robogo --debug-vars
```
Look for `[MASKED]` in the debug output.

### **2. Test Secret Substitution**
Check that your HTTP requests work correctly with secrets:
```bash
./robogo.exe run examples/simple-secrets-demo.robogo
```
Should complete successfully with 200 responses.

## 🔍 **Troubleshooting**

### **Problem: Secrets not loading**
```
Error: Failed to read secret file 'path/to/secret.txt'
```
**Solution:** Ensure the file exists and is readable.

### **Problem: Secrets visible in output**
```
Debug shows: API=actual-secret-value
```
**Solution:** Check that `mask_output: true` is set in your secret configuration.

### **Problem: Variables not substituting**
```
${SECRETS.secret_name} appears literally in output
```
**Solution:** Verify the secret is properly defined in the `variables.secrets` section and you're using the correct `SECRETS.var` syntax.

## 📚 **Advanced Examples**

For more complex scenarios, see:
- `secrets-showcase.robogo` - Comprehensive demonstration
- `test-secrets-example.robogo` - Advanced patterns with database operations

## 🔗 **Related Documentation**

- [Main Robogo Documentation](../CLAUDE.md)
- [Architecture Overview](../docs/ARCHITECTURE.md)
- [Security Best Practices](../SECRETS_DESIGN.md)