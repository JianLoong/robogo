package common

import (
	"fmt"
	"strings"
)

// Variables - simple variable storage and substitution
type Variables struct {
	data map[string]interface{}
}

func NewVariables() *Variables {
	return &Variables{
		data: make(map[string]interface{}),
	}
}

func (v *Variables) Set(key string, value interface{}) {
	v.data[key] = value
}

func (v *Variables) Get(key string) interface{} {
	if val, exists := v.data[key]; exists {
		return val
	}
	return nil
}

func (v *Variables) Load(vars map[string]interface{}) {
	for k, val := range vars {
		v.data[k] = val
	}
}

// Simple variable substitution using ${variable} syntax
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
		
		varName := result[start+2 : end]
		if varName == "" {
			result = result[:start] + result[end+1:]
			continue
		}
		
		// Handle dot notation (simple version)
		value := v.resolveDotNotation(varName)
		replacement := fmt.Sprintf("%v", value)
		
		result = result[:start] + replacement + result[end+1:]
	}
	
	return result
}

// Simple dot notation resolver
func (v *Variables) resolveDotNotation(varName string) interface{} {
	if !strings.Contains(varName, ".") {
		// Simple variable
		if val, exists := v.data[varName]; exists {
			return val
		}
		return fmt.Sprintf("${%s}", varName) // Return unresolved
	}
	
	parts := strings.Split(varName, ".")
	baseVar := parts[0]
	
	current, exists := v.data[baseVar]
	if !exists {
		return fmt.Sprintf("${%s}", varName) // Return unresolved
	}
	
	// Navigate through properties
	for _, part := range parts[1:] {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
				continue
			}
		}
		// Property not found
		return fmt.Sprintf("${%s}", varName)
	}
	
	return current
}

// Substitute variables in arguments
func (v *Variables) SubstituteArgs(args []interface{}) []interface{} {
	result := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			result[i] = v.Substitute(str)
		} else {
			result[i] = arg
		}
	}
	return result
}