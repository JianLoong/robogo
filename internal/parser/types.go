package parser

import (
	"time"
)

// TestCase represents a single test case
type TestCase struct {
	Name        string            `yaml:"testcase"`
	Description string            `yaml:"description,omitempty"`
	Templates   map[string]string `yaml:"templates,omitempty"` // Inline template definitions
	Steps       []Step            `yaml:"steps"`
	Timeout     time.Duration     `yaml:"timeout,omitempty"`
	Variables   Variables         `yaml:"variables,omitempty"`
	Verbose     interface{}       `yaml:"verbose,omitempty"` // Global verbosity setting

	// TDM Fields
	DataManagement *DataManagement `yaml:"data_management,omitempty"`
	Environments   []Environment   `yaml:"environments,omitempty"`

	// Parallelism Fields
	Parallel *ParallelConfig `yaml:"parallel,omitempty"`

	// Skip field to allow skipping test cases
	Skip interface{} `yaml:"skip,omitempty"`
}

// DataManagement represents test data management configuration
type DataManagement struct {
	DataSets    []DataSet    `yaml:"data_sets,omitempty"`
	Setup       []Step       `yaml:"setup,omitempty"`
	Teardown    []Step       `yaml:"teardown,omitempty"`
	Validation  []Validation `yaml:"validation,omitempty"`
	Isolation   bool         `yaml:"isolation,omitempty"`   // Enable data isolation
	Cleanup     bool         `yaml:"cleanup,omitempty"`     // Enable automatic cleanup
	Environment string       `yaml:"environment,omitempty"` // Target environment
}

// DataSet represents a collection of test data
type DataSet struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Data        map[string]interface{} `yaml:"data"`
	Schema      map[string]string      `yaml:"schema,omitempty"`      // Data type validation
	Required    []string               `yaml:"required,omitempty"`    // Required fields
	Unique      []string               `yaml:"unique,omitempty"`      // Unique constraint fields
	Relations   []Relation             `yaml:"relations,omitempty"`   // Data relationships
	Version     string                 `yaml:"version,omitempty"`     // Data version
	Environment string                 `yaml:"environment,omitempty"` // Environment-specific data
}

// Relation represents data relationships between fields or datasets
type Relation struct {
	Type      string `yaml:"type"`      // foreign_key, dependency, etc.
	Field     string `yaml:"field"`     // Field name
	Reference string `yaml:"reference"` // Referenced field/dataset
	Required  bool   `yaml:"required"`  // Is relationship required
}

// Validation represents data validation rules
type Validation struct {
	Name              string      `yaml:"name"`
	Type              string      `yaml:"type"`     // format, range, length, custom
	Field             string      `yaml:"field"`    // Field to validate
	Rule              interface{} `yaml:"rule"`     // Validation rule
	Message           string      `yaml:"message"`  // Error message
	Severity          string      `yaml:"severity"` // error, warning, info
	ContinueOnFailure bool        `yaml:"continue_on_failure,omitempty"`
}

// Environment represents environment-specific configuration
type Environment struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Variables   map[string]interface{} `yaml:"variables,omitempty"`
	DataSets    []string               `yaml:"data_sets,omitempty"` // Data sets to load
	Overrides   map[string]interface{} `yaml:"overrides,omitempty"` // Variable overrides
	Secrets     map[string]Secret      `yaml:"secrets,omitempty"`
}

// Variables represents test case variables
type Variables struct {
	Regular map[string]interface{} `yaml:"vars,omitempty"`
	Secrets map[string]Secret      `yaml:"secrets,omitempty"`
}

// Secret represents a secret variable
// Supports inline value or file-based secret
// If both value and file are set, value takes precedence
type Secret struct {
	Value      string `yaml:"value,omitempty"`
	File       string `yaml:"file,omitempty"`
	MaskOutput bool   `yaml:"mask_output,omitempty"`
}

// RetryConfig represents retry configuration for a step
type RetryConfig struct {
	Attempts   int           `yaml:"attempts,omitempty"`   // Number of retry attempts
	Delay      time.Duration `yaml:"delay,omitempty"`      // Delay between retries
	Backoff    string        `yaml:"backoff,omitempty"`    // backoff strategy: fixed, linear, exponential
	Conditions []string      `yaml:"conditions,omitempty"` // When to retry: 5xx, timeout, connection_error, etc.
	MaxDelay   time.Duration `yaml:"max_delay,omitempty"`  // Maximum delay cap
	Jitter     bool          `yaml:"jitter,omitempty"`     // Add randomness to delay
}

// ExpectErrorConfig represents error expectation configuration
type ExpectErrorConfig struct {
	Type    string `yaml:"type,omitempty"`    // any, contains, matches, exact, starts_with, ends_with, not_contains, not_matches
	Message string `yaml:"message,omitempty"` // Expected error message or pattern
}

// Step represents a single test step
// 'name' is now mandatory for every step and must be unique within the test case
type Step struct {
	Name    string                 `yaml:"name,omitempty"`
	Action  string                 `yaml:"action"`
	Args    []interface{}          `yaml:"args"`
	Options map[string]interface{} `yaml:"options,omitempty"`
	Result  string                 `yaml:"result,omitempty"`
	Verbose interface{}            `yaml:"verbose,omitempty"` // true/false or "basic"/"detailed"/"debug"
	Retry   *RetryConfig           `yaml:"retry,omitempty"`   // Retry configuration

	// Control flow fields
	If    *ConditionalBlock `yaml:"if,omitempty"`    // If statement
	For   *LoopBlock        `yaml:"for,omitempty"`   // For loop
	While *LoopBlock        `yaml:"while,omitempty"` // While loop

	ContinueOnFailure bool `yaml:"continue_on_failure,omitempty"` // Continue on failure

	// Error expectation - can be string (simple) or object (detailed)
	ExpectError interface{} `yaml:"expect_error,omitempty"` // string or ExpectErrorConfig

	// Skip field to allow skipping steps
	Skip interface{} `yaml:"skip,omitempty"`
}

// ConditionalBlock represents an if/else block
type ConditionalBlock struct {
	Condition string `yaml:"condition"`      // Condition to evaluate
	Then      []Step `yaml:"then"`           // Steps to execute if true
	Else      []Step `yaml:"else,omitempty"` // Steps to execute if false
}

// LoopBlock represents a for or while loop
type LoopBlock struct {
	Condition     string `yaml:"condition"`                // For: range/array, While: condition
	Steps         []Step `yaml:"steps"`                    // Steps to execute in loop
	MaxIterations int    `yaml:"max_iterations,omitempty"` // Prevent infinite loops
}

// TestResult represents the result of running a test case
type TestResult struct {
	TestCase       *TestCase
	Status         string
	Duration       time.Duration
	TotalSteps     int
	PassedSteps    int
	FailedSteps    int
	SkippedSteps   int
	StepResults    []StepResult
	ErrorMessage   string
	CapturedOutput string // New field to store captured console output

	// TDM Results
	DataResults *DataResults `yaml:"data_results,omitempty"`
}

// DataResults represents TDM execution results
type DataResults struct {
	SetupStatus    string                 `yaml:"setup_status"`
	TeardownStatus string                 `yaml:"teardown_status"`
	Validations    []ValidationResult     `yaml:"validations"`
	DataSets       map[string]DataSetInfo `yaml:"data_sets"`
	CleanupStatus  string                 `yaml:"cleanup_status"`
}

// ValidationResult represents the result of a data validation
type ValidationResult struct {
	Name     string `yaml:"name"`
	Status   string `yaml:"status"`
	Message  string `yaml:"message"`
	Severity string `yaml:"severity"`
}

// DataSetInfo represents information about a loaded data set
type DataSetInfo struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Records  int    `yaml:"records"`
	Status   string `yaml:"status"`
	LoadTime string `yaml:"load_time"`
}

// StepResult represents the result of a single step
type StepResult struct {
	Step      Step
	Status    string
	Duration  time.Duration
	Output    string
	Error     string
	Timestamp time.Time

	// Enhanced display fields
	DisplayName   string                 `json:"display_name,omitempty"`   // Human-readable step name
	Category      string                 `json:"category,omitempty"`       // setup, main, teardown, validation
	VerboseOutput string                 `json:"verbose_output,omitempty"` // Detailed output for verbose mode
	Warnings      []string               `json:"warnings,omitempty"`       // Non-fatal warnings
	Metadata      map[string]interface{} `json:"metadata,omitempty"`       // Additional context
}

// ParallelConfig represents parallelism configuration
type ParallelConfig struct {
	Enabled        bool `yaml:"enabled,omitempty"`             // Enable parallel execution
	MaxConcurrency int  `yaml:"max_concurrency,omitempty"`     // Maximum concurrent operations
	TestCases      bool `yaml:"test_cases,omitempty"`          // Enable parallel test case execution
	Steps          bool `yaml:"steps,omitempty"`               // Enable parallel step execution
	HTTPRequests   bool `yaml:"http_requests,omitempty"`       // Enable parallel HTTP requests
	DatabaseOps    bool `yaml:"database_operations,omitempty"` // Enable parallel database operations
	DataValidation bool `yaml:"data_validation,omitempty"`     // Enable parallel data validation
	FileOperations bool `yaml:"file_operations,omitempty"`     // Enable parallel file operations
}

// LoadTestingConfig represents load testing configuration
type LoadTestingConfig struct {
	Enabled    bool `yaml:"enabled,omitempty"`     // Enable load testing
	MaxWorkers int  `yaml:"max_workers,omitempty"` // Maximum worker goroutines
	RateLimit  int  `yaml:"rate_limit,omitempty"`  // Requests per second limit
}
