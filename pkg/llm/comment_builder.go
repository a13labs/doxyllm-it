package llm

import (
	"fmt"
	"regexp"
	"strings"
)

// CommentBuilder handles the construction of structured Doxygen comments
type CommentBuilder struct{}

// NewCommentBuilder creates a new comment builder instance
func NewCommentBuilder() *CommentBuilder {
	return &CommentBuilder{}
}

// BuildStructuredComment creates a properly structured Doxygen comment
func (cb *CommentBuilder) BuildStructuredComment(response *CommentResponse, entityName, entityType string, groupInfo *GroupInfo, context string) string {
	var comment strings.Builder
	comment.WriteString("/**\n")

	description := response.Description

	// Add brief description (first sentence or line)
	lines := strings.Split(description, "\n")
	brief := ""
	if len(lines) > 0 {
		brief = strings.TrimSpace(lines[0])
		// If first line is too short, combine with second line
		if len(brief) < 50 && len(lines) > 1 {
			secondLine := strings.TrimSpace(lines[1])
			if secondLine != "" {
				brief += " " + secondLine
			}
		}
	}

	if brief != "" {
		comment.WriteString(fmt.Sprintf(" * @brief %s\n", brief))
	}

	// Add detailed description if there's more content
	detailed := ""
	if len(lines) > 1 {
		detailedLines := lines[1:]
		if brief != "" && len(lines) > 2 && strings.Contains(brief, strings.TrimSpace(lines[1])) {
			detailedLines = lines[2:] // Skip second line if it was included in brief
		}

		var detailedParts []string
		for _, line := range detailedLines {
			line = strings.TrimSpace(line)
			if line != "" {
				detailedParts = append(detailedParts, line)
			}
		}
		detailed = strings.Join(detailedParts, " ")
	}

	if detailed != "" {
		comment.WriteString(" *\n")
		// Wrap detailed description
		cb.writeWrappedText(&comment, detailed, 80)
	}

	// Add function-specific tags only for actual functions/methods
	if cb.isFunctionType(entityType) {
		params := cb.extractParametersFromContext(context)
		for _, param := range params {
			comment.WriteString(fmt.Sprintf(" * @param %s \n", param))
		}

		if cb.hasReturnValue(context, entityType) {
			comment.WriteString(" * @return \n")
		}
	}

	comment.WriteString(" */")
	return comment.String()
}

// writeWrappedText writes text with proper line wrapping
func (cb *CommentBuilder) writeWrappedText(comment *strings.Builder, text string, maxWidth int) {
	words := strings.Fields(text)
	currentLine := " * "

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			comment.WriteString(currentLine + "\n")
			currentLine = " * " + word
		} else {
			if currentLine != " * " {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if currentLine != " * " {
		comment.WriteString(currentLine + "\n")
	}
}

// isFunctionType determines if an entity type represents a function/method
func (cb *CommentBuilder) isFunctionType(entityType string) bool {
	functionTypes := []string{"function", "method", "constructor", "destructor"}
	lowerType := strings.ToLower(entityType)

	for _, ft := range functionTypes {
		if strings.Contains(lowerType, ft) {
			return true
		}
	}
	return false
}

// extractParametersFromContext extracts parameter names from function context
func (cb *CommentBuilder) extractParametersFromContext(context string) []string {
	var params []string

	// Simple regex to find parameters in function signatures
	funcRegex := regexp.MustCompile(`\([^)]*\)`)
	matches := funcRegex.FindAllString(context, -1)

	for _, match := range matches {
		// Remove parentheses
		paramStr := strings.Trim(match, "()")
		if paramStr == "" || paramStr == "void" {
			continue
		}

		// Split by comma and extract parameter names
		parts := strings.Split(paramStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)

			// Handle default values - split by = and take the left side
			if strings.Contains(part, "=") {
				part = strings.Split(part, "=")[0]
				part = strings.TrimSpace(part)
			}

			// Extract the parameter name (last word)
			words := strings.Fields(part)
			if len(words) > 0 {
				paramName := words[len(words)-1]
				// Remove reference/pointer markers
				paramName = strings.TrimPrefix(paramName, "&")
				paramName = strings.TrimPrefix(paramName, "*")
				paramName = strings.TrimSpace(paramName)

				if paramName != "" && cb.isValidIdentifier(paramName) {
					params = append(params, paramName)
				}
			}
		}
	}

	return params
}

// hasReturnValue checks if a function has a return value
func (cb *CommentBuilder) hasReturnValue(context, entityType string) bool {
	// Constructors and destructors don't have return values
	lowerType := strings.ToLower(entityType)
	if strings.Contains(lowerType, "constructor") || strings.Contains(lowerType, "destructor") {
		return false
	}

	// Check if function returns void explicitly
	if strings.Contains(context, "void ") {
		// Look for void at the beginning of function declarations
		lines := strings.Split(context, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "void ") {
				return false
			}
		}
	}

	return true
}

// isValidIdentifier checks if a string is a valid C++ identifier
func (cb *CommentBuilder) isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest must be letters, digits, or underscores
	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}
