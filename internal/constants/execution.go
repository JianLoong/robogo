package constants

// ActionStatus represents the lifecycle state of an action.
type ActionStatus string

const (
	ActionStatusPassed  ActionStatus = "PASS"
	ActionStatusFailed  ActionStatus = "FAIL"
	ActionStatusError   ActionStatus = "ERROR"
	ActionStatusSkipped ActionStatus = "SKIPPED"
)

// Comparison operators
const (
	OperatorEqual              = "=="
	OperatorNotEqual           = "!="
	OperatorGreaterThan        = ">"
	OperatorLessThan           = "<"
	OperatorGreaterThanOrEqual = ">="
	OperatorLessThanOrEqual    = "<="
	OperatorContains           = "contains"
	OperatorStartsWith         = "starts_with"
	OperatorEndsWith           = "ends_with"
)

// HTTP operations supported
const (
	HTTPGet    = "GET"
	HTTPPost   = "POST"
	HTTPPut    = "PUT"
	HTTPPatch  = "PATCH"
	HTTPDelete = "DELETE"
	HTTPHead   = "HEAD"
)

// Database operation constants
const (
	OperationQuery   = "query"
	OperationSelect  = "select"
	OperationExecute = "execute"
	OperationInsert  = "insert"
	OperationUpdate  = "update"
	OperationDelete  = "delete"
)

// Messaging operation constants
const (
	OperationPublish    = "publish"
	OperationConsume    = "consume"
	OperationListTopics = "list_topics"
)

// Variable operation constants
const (
	VariableOperationSet    = "set"
	VariableOperationGet    = "get"
	VariableOperationList   = "list"
	VariableOperationDelete = "delete"
	VariableOperationDebug  = "debug"
)