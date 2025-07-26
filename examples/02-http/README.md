# HTTP Examples

HTTP requests, REST APIs, and TLS handling examples.

## Examples

### 01-http-get.yaml - Basic HTTP GET
**Complexity:** Beginner  
**Prerequisites:** None (uses HTTPBin service)  
**Description:** Simple HTTP GET request with response validation and data extraction.

**What you'll learn:**
- Basic HTTP GET requests with the `http` action
- Response data extraction using `jq` action
- Assertion testing with the `assert` action
- Variable substitution in URLs

**Run it:**
```bash
./robogo run examples/02-http/01-http-get.yaml
```

### 02-http-post.yaml - HTTP POST with JSON
**Complexity:** Beginner  
**Prerequisites:** None (uses HTTPBin service)  
**Description:** HTTP POST request with JSON payload and response validation.

**What you'll learn:**
- HTTP POST requests with JSON data
- Request body construction
- Response status code validation
- JSON response parsing

**Run it:**
```bash
./robogo run examples/02-http/02-http-post.yaml
```

### 02-http-post-with-json-build.yaml - Dynamic JSON Construction
**Complexity:** Intermediate  
**Prerequisites:** None (uses HTTPBin service)  
**Description:** HTTP POST using the `json_build` action to construct request payloads dynamically.

**What you'll learn:**
- Dynamic JSON construction with `json_build`
- Template-based JSON creation
- Variable interpolation in JSON templates
- Complex request payload handling

**Run it:**
```bash
./robogo run examples/02-http/02-http-post-with-json-build.yaml
```

### 36-http-skip-tls.yaml - TLS Verification Disabled
**Complexity:** Intermediate  
**Prerequisites:** None  
**Description:** HTTP requests with TLS verification disabled for testing against self-signed certificates.

**What you'll learn:**
- TLS verification control
- Self-signed certificate handling
- HTTP options configuration
- Security considerations for testing

**Run it:**
```bash
./robogo run examples/02-http/36-http-skip-tls.yaml
```

### 37-http-tls-validation.yaml - Strict TLS Validation
**Complexity:** Intermediate  
**Prerequisites:** None  
**Description:** HTTP requests with strict TLS validation enabled.

**What you'll learn:**
- Strict TLS certificate validation
- Certificate chain verification
- Secure HTTP communication
- Production-ready TLS handling

**Run it:**
```bash
./robogo run examples/02-http/37-http-tls-validation.yaml
```

## Key Concepts

### HTTP Action Options
```yaml
- name: "HTTP request with options"
  action: http
  args: ["POST", "https://api.example.com/data"]
  options:
    headers:
      Authorization: "Bearer ${token}"
      Content-Type: "application/json"
    json:
      key: "value"
    timeout: "30s"
    skip_tls_verify: false
```

### Response Data Extraction
```yaml
# Extract status code
- name: "Get status code"
  action: jq
  args: ["${response}", ".status_code"]
  result: status_code

# Extract response body
- name: "Get response body"
  action: jq
  args: ["${response}", ".body"]
  result: response_body
```

### Common Patterns
- Always use `jq` to extract data from HTTP responses
- Store responses in variables for later processing
- Use assertions to validate expected outcomes
- Handle both success and error scenarios