package constants

// Error template names for safe formatting - only used templates
const (
	// Control flow error templates (actually used)
	TemplateInvalidRangeFormat   = "control_flow.invalid_range_format"
	TemplateInvalidStartValue    = "control_flow.invalid_start_value"
	TemplateInvalidEndValue      = "control_flow.invalid_end_value"
	TemplateInvalidCountFormat   = "control_flow.invalid_count_format"
	TemplateIfConditionFailed    = "control_flow.if_condition_failed"
	TemplateWhileConditionFailed = "control_flow.while_condition_failed"
	TemplateUnknownAction        = "control_flow.unknown_action"
)

// Template content mapping - only used templates
var ErrorTemplates = map[string]string{
	// Control flow templates (actually used)
	TemplateInvalidRangeFormat:   "invalid range format: %s",
	TemplateInvalidStartValue:    "invalid start value: %s",
	TemplateInvalidEndValue:      "invalid end value: %s",
	TemplateInvalidCountFormat:   "invalid count format: %s",
	TemplateIfConditionFailed:    "if condition evaluation failed: %v",
	TemplateWhileConditionFailed: "while condition evaluation failed: %v",
	TemplateUnknownAction:        "unknown action: %s",
}
