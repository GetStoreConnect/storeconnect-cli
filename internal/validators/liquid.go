package validators

import (
	"fmt"
	"regexp"
	"strings"
)

// LiquidValidator validates Liquid template syntax
type LiquidValidator struct{}

// NewLiquidValidator creates a new Liquid validator
func NewLiquidValidator() *LiquidValidator {
	return &LiquidValidator{}
}

// Validate validates Liquid template syntax
func (v *LiquidValidator) Validate(content string) []string {
	var errors []string

	// Check for unclosed tags
	tagRegex := regexp.MustCompile(`\{%\s*(\w+).*?%\}`)
	tags := tagRegex.FindAllStringSubmatch(content, -1)

	stack := []string{}
	for _, match := range tags {
		tagName := match[1]

		// Opening tags
		if isOpeningTag(tagName) {
			stack = append(stack, tagName)
		}

		// Closing tags
		if isClosingTag(tagName) {
			expectedOpen := getOpeningTag(tagName)
			if len(stack) == 0 {
				errors = append(errors, fmt.Sprintf("Unexpected closing tag: {%% %s %%}", tagName))
			} else {
				lastOpen := stack[len(stack)-1]
				if lastOpen != expectedOpen {
					errors = append(errors, fmt.Sprintf("Mismatched tags: opened %s, closed with %s", lastOpen, tagName))
				}
				stack = stack[:len(stack)-1]
			}
		}
	}

	// Check for unclosed tags
	for _, tag := range stack {
		errors = append(errors, fmt.Sprintf("Unclosed tag: {%% %s %%}", tag))
	}

	// Check for unmatched braces
	if strings.Count(content, "{{") != strings.Count(content, "}}") {
		errors = append(errors, "Unmatched variable braces {{ }}")
	}

	return errors
}

func isOpeningTag(tag string) bool {
	openingTags := []string{"if", "unless", "for", "case", "capture", "block", "tablerow", "comment"}
	for _, t := range openingTags {
		if t == tag {
			return true
		}
	}
	return false
}

func isClosingTag(tag string) bool {
	return strings.HasPrefix(tag, "end")
}

func getOpeningTag(closingTag string) string {
	return strings.TrimPrefix(closingTag, "end")
}
