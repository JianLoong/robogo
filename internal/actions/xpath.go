package actions

import (
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// xpathAction executes XPath queries on XML strings
func xpathAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("xpath", 2, len(args))
	}

	xmlStr := fmt.Sprintf("%v", args[0])
	xpathQuery := fmt.Sprintf("%v", args[1])

	// Parse the XML document
	doc, err := xmlquery.Parse(strings.NewReader(xmlStr))
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "XML_PARSE_ERROR").
			WithTemplate("Failed to parse XML: %s").
			Build(err.Error())
	}

	// Check if we want multiple results or single result
	multiple := false
	if multi, ok := options["multiple"]; ok {
		if m, ok := multi.(bool); ok {
			multiple = m
		}
	}

	if multiple {
		// Find all matching nodes
		nodes := xmlquery.Find(doc, xpathQuery)
		var results []string
		
		for _, node := range nodes {
			if node.Type == xmlquery.AttributeNode {
				results = append(results, node.InnerText())
			} else if node.Type == xmlquery.TextNode {
				results = append(results, node.Data)
			} else {
				results = append(results, node.InnerText())
			}
		}
		
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   results,
		}
	} else {
		// Find first matching node
		node := xmlquery.FindOne(doc, xpathQuery)
		if node == nil {
			return types.ActionResult{
				Status: constants.ActionStatusPassed,
				Data:   nil,
			}
		}
		
		var result string
		if node.Type == xmlquery.AttributeNode {
			result = node.InnerText()
		} else if node.Type == xmlquery.TextNode {
			result = node.Data
		} else {
			result = node.InnerText()
		}
		
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   result,
		}
	}
}