package actions

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// xmlParseAction parses an XML string into structured data
func xmlParseAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("xml_parse", 1, len(args))
	}

	// Get the XML string to parse
	xmlStr, ok := args[0].(string)
	if !ok {
		return types.InvalidArgError("xml_parse", "XML string", "first argument must be a string")
	}

	// Parse the XML into a map-like structure
	result, err := parseXMLToMap(xmlStr)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "XML_PARSE_ERROR").
			WithTemplate("Failed to parse XML: %s").
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   result,
	}
}

// parseXMLToMap parses XML into a map structure
func parseXMLToMap(xmlStr string) (map[string]interface{}, error) {
	// First, try to determine if this is a simple XML structure
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	
	// Token-based parsing for flexibility
	var result map[string]interface{}
	stack := make([]map[string]interface{}, 0)
	var current map[string]interface{}
	
	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("XML parsing error: %w", err)
		}
		
		switch elem := token.(type) {
		case xml.StartElement:
			// Create new map for this element
			newMap := make(map[string]interface{})
			
			// Add attributes if present
			if len(elem.Attr) > 0 {
				attrs := make(map[string]string)
				for _, attr := range elem.Attr {
					attrs[attr.Name.Local] = attr.Value
				}
				newMap["@attributes"] = attrs
			}
			
			// If this is the root element, make it our result
			if result == nil {
				result = make(map[string]interface{})
				result[elem.Name.Local] = newMap
				current = newMap
			} else {
				// Add to current parent
				if existing, exists := current[elem.Name.Local]; exists {
					// Convert to array if multiple elements with same name
					switch v := existing.(type) {
					case []interface{}:
						current[elem.Name.Local] = append(v, newMap)
					default:
						current[elem.Name.Local] = []interface{}{v, newMap}
					}
				} else {
					current[elem.Name.Local] = newMap
				}
				
				// Push current onto stack and make new map current
				stack = append(stack, current)
				current = newMap
			}
			
		case xml.EndElement:
			// Pop from stack
			if len(stack) > 0 {
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			
		case xml.CharData:
			// Add text content if not just whitespace
			content := strings.TrimSpace(string(elem))
			if content != "" {
				// If current map is empty (no child elements), just set the text
				if len(current) == 0 || (len(current) == 1 && current["@attributes"] != nil) {
					current["#text"] = content
				} else {
					// Mix of text and elements - store as #text
					if existing := current["#text"]; existing != nil {
						current["#text"] = fmt.Sprintf("%v%s", existing, content)
					} else {
						current["#text"] = content
					}
				}
			}
			
		case xml.Comment:
			// Store comments if needed (optional)
			comment := strings.TrimSpace(string(elem))
			if comment != "" {
				if comments, exists := current["#comments"]; exists {
					if commentArr, ok := comments.([]string); ok {
						current["#comments"] = append(commentArr, comment)
					} else {
						current["#comments"] = []string{fmt.Sprintf("%v", comments), comment}
					}
				} else {
					current["#comments"] = comment
				}
			}
		}
	}
	
	if result == nil {
		return nil, fmt.Errorf("no valid XML content found")
	}
	
	return result, nil
}