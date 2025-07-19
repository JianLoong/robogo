package constants

// Error template names for safe formatting
const (
	// Control flow error templates
	TemplateInvalidRangeFormat   = "control_flow.invalid_range_format"
	TemplateInvalidStartValue    = "control_flow.invalid_start_value"
	TemplateInvalidEndValue      = "control_flow.invalid_end_value"
	TemplateInvalidCountFormat   = "control_flow.invalid_count_format"
	TemplateIfConditionFailed    = "control_flow.if_condition_failed"
	TemplateWhileConditionFailed = "control_flow.while_condition_failed"
	TemplateUnknownAction        = "control_flow.unknown_action"

	// Database error templates
	TemplatePostgresConnectionFailed = "postgres.connection_failed"
	TemplatePostgresPingFailed       = "postgres.ping_failed"
	TemplatePostgresQueryFailed      = "postgres.query_failed"
	TemplatePostgresColumnsFailed    = "postgres.columns_failed"
	TemplatePostgresScanFailed       = "postgres.scan_failed"
	TemplatePostgresExecuteFailed    = "postgres.execute_failed"
	TemplatePostgresUnknownOperation = "postgres.unknown_operation"

	TemplateSpannerConnectionFailed   = "spanner.connection_failed"
	TemplateSpannerQueryFailed        = "spanner.query_failed"
	TemplateSpannerColumnsFailed      = "spanner.columns_failed"
	TemplateSpannerScanFailed         = "spanner.scan_failed"
	TemplateSpannerRowIterationFailed = "spanner.row_iteration_failed"
	TemplateSpannerDMLFailed          = "spanner.dml_failed"
	TemplateSpannerUnknownOperation   = "spanner.unknown_operation"

	// HTTP error templates
	TemplateHTTPRequestCreateFailed = "http.request_create_failed"
	TemplateHTTPRequestFailed       = "http.request_failed"
	TemplateHTTPResponseReadFailed  = "http.response_read_failed"
	TemplateHTTPTimeoutError        = "http.timeout_error"
	TemplateHTTPConnectionError     = "http.connection_error"
	TemplateHTTPDNSError            = "http.dns_error"
	TemplateHTTPTLSError            = "http.tls_error"

	// Kafka error templates
	TemplateKafkaPublishFailed    = "kafka.publish_failed"
	TemplateKafkaConsumeFailed    = "kafka.consume_failed"
	TemplateKafkaConsumeTimeout   = "kafka.consume_timeout"
	TemplateKafkaInvalidCount     = "kafka.invalid_count"
	TemplateKafkaUnknownOperation = "kafka.unknown_operation"

	// RabbitMQ error templates
	TemplateRabbitMQConnectionFailed = "rabbitmq.connection_failed"
	TemplateRabbitMQConnectionClosed = "rabbitmq.connection_closed"
	TemplateRabbitMQChannelFailed    = "rabbitmq.channel_failed"
	TemplateRabbitMQPublishFailed    = "rabbitmq.publish_failed"
	TemplateRabbitMQUnknownOperation = "rabbitmq.unknown_operation"

	// Assertion error templates
	TemplateAssertFailed              = "assert.failed"
	TemplateAssertComparisonFailed    = "assert.comparison_failed"
	TemplateAssertUnsupportedOperator = "assert.unsupported_operator"
	TemplateAssertMissingArgs         = "assert.missing_args"
	TemplateAssertBooleanFailed       = "assert.boolean_failed"
	TemplateAssertTypeMismatch        = "assert.type_mismatch"
	TemplateAssertDetailedComparison  = "assert.detailed_comparison"

	// Variable error templates
	TemplateVariableMissingArgs      = "variable.missing_args"
	TemplateVariableNotFound         = "variable.not_found"
	TemplateVariableAccessError      = "variable.access_error"
	TemplateVariableExpressionError  = "variable.expression_error"
	TemplateVariableResolutionFailed = "variable.resolution_failed"
	TemplateVariableDetailedFailure  = "variable.detailed_failure"

	// Log error templates
	TemplateLogMissingArgs = "log.missing_args"

	// General validation templates
	TemplateMissingArgs = "validation.missing_args"
)

// Template content mapping - these will be registered with SafeFormatter
var ErrorTemplates = map[string]string{
	// Control flow templates
	TemplateInvalidRangeFormat:   "invalid range format: %s",
	TemplateInvalidStartValue:    "invalid start value: %s",
	TemplateInvalidEndValue:      "invalid end value: %s",
	TemplateInvalidCountFormat:   "invalid count format: %s",
	TemplateIfConditionFailed:    "if condition evaluation failed: %v",
	TemplateWhileConditionFailed: "while condition evaluation failed: %v",
	TemplateUnknownAction:        "unknown action: %s",

	// Database templates
	TemplatePostgresConnectionFailed: "failed to open postgres connection: %v",
	TemplatePostgresPingFailed:       "failed to ping postgres database: %v",
	TemplatePostgresQueryFailed:      "failed to execute query: %v",
	TemplatePostgresColumnsFailed:    "failed to get columns: %v",
	TemplatePostgresScanFailed:       "failed to scan row: %v",
	TemplatePostgresExecuteFailed:    "failed to execute statement: %v",
	TemplatePostgresUnknownOperation: "unknown postgres operation: %s",

	TemplateSpannerConnectionFailed:   "failed to open spanner database: %v",
	TemplateSpannerQueryFailed:        "query failed: %v",
	TemplateSpannerColumnsFailed:      "failed to get columns: %v",
	TemplateSpannerScanFailed:         "failed to scan row: %v",
	TemplateSpannerRowIterationFailed: "row iteration error: %v",
	TemplateSpannerDMLFailed:          "DML failed: %v",
	TemplateSpannerUnknownOperation:   "unsupported spanner operation: %s",

	// HTTP templates
	TemplateHTTPRequestCreateFailed: "failed to create HTTP request: %v",
	TemplateHTTPRequestFailed:       "HTTP request failed: %v",
	TemplateHTTPResponseReadFailed:  "failed to read HTTP response body: %v",
	TemplateHTTPTimeoutError:        "HTTP request timed out after %s: %v",
	TemplateHTTPConnectionError:     "HTTP connection failed to %s: %v",
	TemplateHTTPDNSError:            "DNS resolution failed for %s: %v",
	TemplateHTTPTLSError:            "TLS/SSL error for %s: %v",

	// Kafka templates
	TemplateKafkaPublishFailed:    "failed to publish message: %v",
	TemplateKafkaConsumeFailed:    "failed to consume message: %v",
	TemplateKafkaConsumeTimeout:   "failed to consume message: timeout waiting for messages (topic may not exist or be empty)",
	TemplateKafkaInvalidCount:     "count must be >= 0, got %d",
	TemplateKafkaUnknownOperation: "unknown kafka operation: %s",

	// RabbitMQ templates
	TemplateRabbitMQConnectionFailed: "failed to connect to RabbitMQ: %v",
	TemplateRabbitMQConnectionClosed: "RabbitMQ connection closed immediately",
	TemplateRabbitMQChannelFailed:    "failed to open channel: %v",
	TemplateRabbitMQPublishFailed:    "failed to publish message: %v",
	TemplateRabbitMQUnknownOperation: "unknown rabbitmq operation: %s",

	// Assertion templates
	TemplateAssertFailed:              "assertion failed: %v",
	TemplateAssertComparisonFailed:    "assertion failed: %v %v %v",
	TemplateAssertUnsupportedOperator: "unsupported operator: %v",
	TemplateAssertMissingArgs:         "assert action requires at least 1 argument",
	TemplateAssertBooleanFailed:       "boolean assertion failed: expected true, got %v (%s)",
	TemplateAssertTypeMismatch:        "type mismatch in assertion: %s vs %s - %s",
	TemplateAssertDetailedComparison:  "detailed assertion failed: %s",

	// Variable templates
	TemplateVariableMissingArgs:      "variable action requires at least 2 arguments",
	TemplateVariableNotFound:         "variable '%s' not found - available variables: %s",
	TemplateVariableAccessError:      "failed to access '%s' in variable path '%s': %s",
	TemplateVariableExpressionError:  "expression evaluation failed for '%s': %s",
	TemplateVariableResolutionFailed: "variable resolution failed: %d unresolved variables",
	TemplateVariableDetailedFailure:  "%s",

	// Log templates
	TemplateLogMissingArgs: "log action requires at least 1 argument",

	// General validation templates
	TemplateMissingArgs: "%s action requires at least %d arguments%s",

	// Test compatibility templates
	"assertion.failed": "Assertion failed: expected %v %v %v, but got %v",
}
