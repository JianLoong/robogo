package actions

import (
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// xmlBuildAction creates an XML string from nested YAML arguments
func xmlBuildAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	var xmlData any

	// If we have exactly one argument, use it as the XML data
	if len(args) == 1 {
		xmlData = args[0]
	} else if len(args) == 0 {
		// No args, build from options if provided
		if len(options) > 0 {
			xmlData = options
		} else {
			return types.InvalidArgError("xml_build", "XML data", "at least one argument or options")
		}
	} else {
		// Multiple args - treat as an array (wrap in a root element)
		xmlData = map[string]any{"items": args}
	}

	// Perform variable substitution on the data structure
	substitutedData := substituteVariablesInData(xmlData, vars)

	// Convert the structured data to XML
	xmlString, err := buildXMLFromData(substitutedData, options)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "XML_BUILD_ERROR").
			WithTemplate("Failed to build XML: %s").
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   xmlString,
	}
}

// buildXMLFromData converts structured data to XML string
func buildXMLFromData(data interface{}, options map[string]any) (string, error) {
	var builder strings.Builder
	
	// Add XML declaration if requested
	if addDeclaration, ok := options["declaration"]; ok && addDeclaration == true {
		builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		builder.WriteString("\n")
	}
	
	// Get root element name from options, default to "root"
	rootElement := "root"
	if root, ok := options["root_element"]; ok {
		rootElement = fmt.Sprintf("%v", root)
	}
	
	err := buildXMLElement(&builder, rootElement, data, 0)
	if err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// buildXMLElement recursively builds XML elements
func buildXMLElement(builder *strings.Builder, name string, data interface{}, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	
	switch v := data.(type) {
	case map[string]interface{}:
		// Handle object as XML element with potential attributes and children
		builder.WriteString(fmt.Sprintf("%s<%s", indentStr, name))
		
		var text string
		var children = make(map[string]interface{})
		
		// Separate attributes, text, and children
		for key, value := range v {
			if key == "@attributes" {
				if attrs, ok := value.(map[string]string); ok {
					for attrName, attrValue := range attrs {
						builder.WriteString(fmt.Sprintf(` %s="%s"`, attrName, attrValue))
					}
				}
			} else if key == "text" {
				text = fmt.Sprintf("%v", value)
			} else {
				children[key] = value
			}
		}
		
		if text != "" && len(children) == 0 {
			// Simple text element
			builder.WriteString(fmt.Sprintf(">%s</%s>\n", text, name))
		} else if len(children) == 0 {
			// Empty element
			builder.WriteString("/>\n")
		} else {
			// Element with children
			builder.WriteString(">\n")
			if text != "" {
				builder.WriteString(fmt.Sprintf("%s  %s\n", indentStr, text))
			}
			for childName, childValue := range children {
				err := buildXMLElement(builder, childName, childValue, indent+1)
				if err != nil {
					return err
				}
			}
			builder.WriteString(fmt.Sprintf("%s</%s>\n", indentStr, name))
		}
		
	case []interface{}:
		// Handle array - each element becomes a separate XML element with the same name
		for _, item := range v {
			err := buildXMLElement(builder, name, item, indent)
			if err != nil {
				return err
			}
		}
		
	default:
		// Simple value
		builder.WriteString(fmt.Sprintf("%s<%s>%v</%s>\n", indentStr, name, v, name))
	}
	
	return nil
}

