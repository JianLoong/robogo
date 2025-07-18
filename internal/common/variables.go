package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/expr-lang/expr"
)

// Variables - simple variable storage and substitution
type Variables struct {
	data map[string]any
}

func NewVariables() *Variables {
	return &Variables{
		data: make(map[string]any),
	}
}

func (v *Variables) Set(key string, value any) {
	// If value is a JSON string, try to parse it
	if str, ok := value.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			v.data[key] = parsed
			return
		}
	}
	v.data[key] = value
}

func (v *Variables) Get(key string) any {
	if val, exists := v.data[key]; exists {
		return val
	}
	return nil
}

func (v *Variables) Load(vars map[string]any) {
	for k, val := range vars {
		v.data[k] = val
	}
}

// Simple variable substitution using ${variable} syntax, now using expr
func (v *Variables) Substitute(template string) string {
	if template == "" {
		return template
	}

	result := template
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		exprStr := result[start+2 : end]
		if exprStr == "" {
			result = result[:start] + result[end+1:]
			continue
		}

		// Evaluate using expr
		output, err := expr.Eval(exprStr, v.data)
		var replacement string
		if err != nil || output == nil {
			fmt.Printf("[WARN] Unresolved variable or expr error in template: %q, error: %v\n", exprStr, err)
			replacement = "__UNRESOLVED__"
		} else {
			replacement = fmt.Sprintf("%v", output)
		}
		result = result[:start] + replacement + result[end+1:]
	}
	// Warn if unresolved marker remains
	if strings.Contains(result, "__UNRESOLVED__") {
		fmt.Printf("[WARN] Unresolved variable in template: %q\n", result)
	}
	return result
}

// Substitute variables in arguments
func (v *Variables) SubstituteArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			result[i] = v.Substitute(str)
		} else {
			result[i] = arg
		}
	}
	return result
}
