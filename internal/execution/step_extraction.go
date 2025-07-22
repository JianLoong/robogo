package execution

import (
	"fmt"
	"regexp"

	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// applyExtraction applies the specified extraction to the data
func (s *BasicExecutionStrategy) applyExtraction(data any, config *types.ExtractConfig) (any, error) {
	if data == nil {
		return nil, types.NewNilDataError()
	}

	switch config.Type {
	case "jq":
		return s.applyJQExtraction(data, config.Path)
	case "xpath":
		return s.applyXPathExtraction(data, config.Path)
	case "regex":
		return s.applyRegexExtraction(data, config.Path, config.Group)
	default:
		return nil, types.NewUnsupportedExtractionTypeError(config.Type)
	}
}

// applyJQExtraction applies JQ extraction to data
func (s *BasicExecutionStrategy) applyJQExtraction(data any, path string) (any, error) {
	jqAction, exists := s.actionRegistry.Get("jq")
	if !exists {
		return nil, types.NewExtractionError("jq action not available")
	}
	
	result := jqAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyXPathExtraction applies XPath extraction to data  
func (s *BasicExecutionStrategy) applyXPathExtraction(data any, path string) (any, error) {
	xpathAction, exists := s.actionRegistry.Get("xpath")
	if !exists {
		return nil, types.NewExtractionError("xpath action not available")
	}
	
	result := xpathAction([]any{data, path}, map[string]any{}, s.variables)
	if result.Status != constants.ActionStatusPassed {
		return nil, types.NewExtractionError(result.GetMessage())
	}
	
	return result.Data, nil
}

// applyRegexExtraction applies regex extraction to data
func (s *BasicExecutionStrategy) applyRegexExtraction(data any, pattern string, group int) (any, error) {
	// Convert data to string
	var text string
	switch v := data.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		text = fmt.Sprintf("%v", v)
	}
	
	// Apply regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, types.NewInvalidRegexPatternError(pattern, err.Error())
	}
	
	matches := re.FindStringSubmatch(text)
	if matches == nil {
		return nil, types.NewNoRegexMatchError(pattern)
	}
	
	// Default to group 1, or use specified group
	if group == 0 {
		group = 1
	}
	
	if group >= len(matches) {
		return nil, types.NewInvalidCaptureGroupError(group, len(matches)-1)
	}
	
	return matches[group], nil
}