# Implementation Plan

- [x] 1. Create core error handling infrastructure





  - Implement ErrorCategory enumeration and ErrorInfo struct
  - Create ErrorBuilder system with safe formatting
  - Add SafeFormatter with template validation
  - _Requirements: 1.1, 2.1, 2.3, 6.1_

- [x] 2. Implement enhanced ActionResult structure










  - Remove Error and Reason string fields and replace with ErrorInfo struct
  - Create NewErrorResult replacement with structured errors
  - Update CLI code to read error/reason messages from ErrorInfo.Message
  - Handle skip reasons through ErrorInfo instead of separate Reason field
  - Update ActionResult creation throughout codebase
  - _Requirements: 1.1, 5.1, 5.3_

- [x] 3. Create centralized error factory system










  - Implement ErrorFactory with predefined templates
  - Add template registration and validation system
  - Create factory methods for different error categories
  - _Requirements: 2.1, 2.4, 6.1, 6.2_

- [ ] 4. Fix format string security issues
  - Replace variable format strings with safe template system
  - Update assert.go to use ErrorBuilder instead of NewErrorResult
  - Validate all error message creation points for security
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 5. Enhance assertion system with detailed context



  - Create AssertionContext struct with type information
  - Implement enhanced assertion error reporting
  - Add comparison type detection (numeric vs string)
  - Update compareNumeric function to provide detailed context
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 6. Improve variable resolution error handling
  - Create VariableContext for tracking resolution failures
  - Implement SubstituteWithContext method with detailed error info
  - Add variable access path tracking for nested failures
  - Update variable substitution to provide resolution traces
  - _Requirements: 4.1, 4.2, 4.3_

- [ ] 7. Add step context enrichment
  - Create StepContext and LoopContext structs
  - Update control flow executor to track context information
  - Include loop iteration details in error messages
  - Add step number and name to all error contexts
  - _Requirements: 3.4, 5.3_

- [ ] 8. Update HTTP action error handling
  - Replace HTTP action error messages with structured errors
  - Add network-specific error categories and context
  - Include request details in error information
  - _Requirements: 1.1, 5.1_

- [ ] 9. Update database action error handling
  - Replace database action error messages with structured errors
  - Add database-specific error categories
  - Include connection details in error context
  - _Requirements: 1.1, 5.1_

- [ ] 10. Update all remaining actions to use new error system
  - Migrate log, variable, uuid, time actions to ErrorBuilder
  - Update kafka and rabbitmq actions with structured errors
  - Ensure consistent error formatting across all actions
  - _Requirements: 6.1, 6.4_

- [ ] 11. Implement error aggregation system
  - Create ErrorCollector for aggregating multiple errors
  - Update TestRunner to collect and categorize errors
  - Add error summary generation for test results
  - _Requirements: 5.4_

- [ ] 12. Add debugging and logging enhancements
  - Implement optional variable change logging
  - Add debug mode for detailed error context
  - Create structured logging for error information
  - _Requirements: 4.4, 6.3_

- [ ] 13. Set up testing infrastructure and create unit tests
  - Add testing dependencies to go.mod (testify or similar)
  - Create test files for ErrorBuilder, SafeFormatter, and error creation functions
  - Write unit tests for ErrorBuilder with all categories and templates
  - Write unit tests for SafeFormatter with various input types and security scenarios
  - Write unit tests for AssertionContext creation and formatting
  - Write unit tests for VariableContext tracking and error generation
  - _Requirements: All requirements validation_

- [ ] 14. Create integration tests for error handling
  - Write integration tests for end-to-end error flow through execution pipeline
  - Test error context preservation in loops and conditions
  - Test error aggregation across multiple failing steps
  - Validate structured error output format in real test scenarios
  - _Requirements: All requirements validation_

- [ ] 15. Add manual testing and validation
  - Create test YAML files that trigger various error conditions
  - Manually verify error message improvements and format string security
  - Test error output in CLI to ensure proper display of ErrorInfo.Message
  - Validate that all error scenarios produce helpful, secure error messages
  - _Requirements: 2.1, 2.2, 2.3_