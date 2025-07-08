# Error Handling Standardization Test Guide

## Build Test
```bash
cd /mnt/c/Users/Jian/Documents/GitHub/robogo
go build -o robogo.exe ./cmd/robogo
```

If the build succeeds, the import path fixes are working correctly.

## Test Cases to Verify Error Handling

### 1. Test HTTP Action Errors
```yaml
testcase: "Test HTTP Error Handling"
steps:
  # Test validation error
  - name: "Invalid HTTP method"
    action: http
    args: ["INVALID", "https://httpbin.org/get"]
    expect_error: "validation"
  
  # Test network error  
  - name: "Invalid URL"
    action: http_get
    args: ["https://invalid-domain-that-does-not-exist.com"]
    expect_error: "network"
```

### 2. Test Database Action Errors
```yaml
testcase: "Test Database Error Handling"
steps:
  # Test validation error
  - name: "Missing arguments"
    action: postgres
    args: ["query"]
    expect_error: "validation"
  
  # Test database connection error
  - name: "Invalid connection"
    action: postgres
    args: ["query", "postgres://invalid:user@nonexistent:5432/db", "SELECT 1"]
    expect_error: "database"
```

### 3. Test Messaging Action Errors
```yaml
testcase: "Test Messaging Error Handling"  
steps:
  # Test kafka validation error
  - name: "Missing kafka arguments"
    action: kafka
    args: []
    expect_error: "validation"
  
  # Test rabbitmq validation error
  - name: "Invalid rabbitmq operation"
    action: rabbitmq
    args: ["invalid_operation"]
    expect_error: "validation"
```

## Expected Error Format

When these tests run, errors should now have the format:
```
action=http | type=validation | Invalid HTTP method | cause=<original error>
```

Instead of the old format:
```
invalid HTTP method: INVALID
```

## Verification Steps

1. **Build Success**: Ensure `go build` completes without import errors
2. **Error Types**: Verify errors contain proper `type=` field
3. **Error Context**: Check that errors include relevant details (URLs, queries, etc.)
4. **Error Chaining**: Confirm original errors are preserved in `cause=` field

## Key Files Modified

- ✅ `internal/util/errors.go` - Enhanced with new error types and constructors
- ✅ `internal/actions/http.go` - Migrated to RobogoError pattern  
- ✅ `internal/actions/postgres.go` - Migrated to RobogoError pattern
- ✅ `internal/actions/kafka.go` - Migrated to RobogoError pattern (import path fixed)
- ✅ `internal/actions/rabbitmq.go` - Migrated to RobogoError pattern (import path fixed, unused fmt removed)

## Success Criteria

✅ **Build Compiles**: No import or compilation errors  
✅ **Consistent Format**: All action errors use RobogoError structure  
✅ **Rich Context**: Errors include relevant debugging information  
✅ **Proper Classification**: Errors categorized by type (validation, network, database, messaging)  
✅ **Backward Compatible**: Errors still implement Go's `error` interface