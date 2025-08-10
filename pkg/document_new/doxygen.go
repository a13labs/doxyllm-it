// Package document_new provides Doxygen-specific functionality for document manipulation
// This file handles Doxygen comment semantics separately from C++ syntax
package document_new

import (
	"regexp"
	"strings"
)

// DoxygenComment represents a Doxygen comment with structured information
// This replaces the old ast.DoxygenComment and focuses purely on documentation semantics
type DoxygenComment struct {
	Brief      string            // Brief description
	Detailed   string            // Detailed description
	Params     map[string]string // Parameter descriptions (@param)
	TParams    map[string]string // Template parameter descriptions (@tparam)
	Returns    string            // Return description (@return)
	Groups     []string          // Groups this entity belongs to (@ingroup)
	Since      string            // Since version (@since)
	Deprecated string            // Deprecation message (@deprecated)
	CustomTags map[string]string // Custom Doxygen tags
	Raw        string            // Raw comment text for reconstruction
}

// NewDoxygenComment creates a new empty Doxygen comment
func NewDoxygenComment() *DoxygenComment {
	return &DoxygenComment{
		Params:     make(map[string]string),
		TParams:    make(map[string]string),
		CustomTags: make(map[string]string),
		Groups:     make([]string, 0),
	}
}

// ParseDoxygenComment parses a raw comment string into structured Doxygen information
func ParseDoxygenComment(rawComment string) *DoxygenComment {
	comment := NewDoxygenComment()
	comment.Raw = rawComment

	// Remove comment delimiters and clean up
	content := cleanCommentContent(rawComment)
	if content == "" {
		return comment
	}

	// Only parse as Doxygen if it has Doxygen markers
	isDoxygen := strings.Contains(rawComment, "/**") ||
		strings.Contains(rawComment, "///") ||
		strings.Contains(content, "@brief") ||
		strings.Contains(content, "@param") ||
		strings.Contains(content, "@return") ||
		strings.Contains(content, "@tparam") ||
		strings.Contains(content, "@details") ||
		strings.Contains(content, "@since") ||
		strings.Contains(content, "@deprecated") ||
		strings.Contains(content, "@ingroup")

	// If it doesn't have Doxygen markers, return empty comment with only raw content
	if !isDoxygen {
		return comment
	}

	// Parse different Doxygen tags
	comment.Brief = extractBrief(content)
	comment.Detailed = extractDetailed(content)
	comment.Params = extractParams(content)
	comment.TParams = extractTParams(content)
	comment.Returns = extractReturns(content)
	comment.Groups = extractGroups(content)
	comment.Since = extractSince(content)
	comment.Deprecated = extractDeprecated(content)
	comment.CustomTags = extractCustomTags(content)

	return comment
}

// cleanCommentContent removes comment delimiters and normalizes whitespace
func cleanCommentContent(raw string) string {
	// Remove /** */ and /// style delimiters
	content := raw

	// Handle /** */ style comments
	content = regexp.MustCompile(`(?s)/\*\*?(.*?)\*/`).ReplaceAllString(content, "$1")

	// Handle /// style comments
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		// Remove /// or // prefix
		line = regexp.MustCompile(`^\s*///?`).ReplaceAllString(line, "")
		// Remove leading asterisks and spaces
		line = regexp.MustCompile(`^\s*\*\s?`).ReplaceAllString(line, "")
		cleanLines = append(cleanLines, line)
	}

	return strings.TrimSpace(strings.Join(cleanLines, "\n"))
}

// extractBrief extracts the brief description (first sentence or @brief)
func extractBrief(content string) string {
	// Look for explicit @brief tag
	briefRegex := regexp.MustCompile(`(?s)@brief\s+(.+?)(\s*@|$)`)
	if matches := briefRegex.FindStringSubmatch(content); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		// Remove extra whitespace and newlines
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}

	// Extract first sentence as brief if no explicit @brief
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "@") {
			// First non-empty, non-tag line is brief
			return strings.TrimSpace(line)
		}
	}

	return ""
}

// extractDetailed extracts the detailed description
func extractDetailed(content string) string {
	// Look for explicit @details tag
	detailsRegex := regexp.MustCompile(`(?s)@details\s+(.+?)(\s*@|$)`)
	if matches := detailsRegex.FindStringSubmatch(content); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		// Remove extra whitespace and newlines
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}

	// Extract everything after brief that's not tagged content
	lines := strings.Split(content, "\n")
	var detailedLines []string
	inDetailed := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if inDetailed {
				detailedLines = append(detailedLines, "")
			}
			continue
		}

		if strings.HasPrefix(line, "@") {
			break // Stop at first tag
		}

		if inDetailed {
			detailedLines = append(detailedLines, line)
		} else if line != "" {
			// Skip first line (brief), start collecting from second paragraph
			inDetailed = true
		}
	}

	return strings.TrimSpace(strings.Join(detailedLines, "\n"))
}

// extractParams extracts @param tags
func extractParams(content string) map[string]string {
	params := make(map[string]string)

	// Split content into lines and process them sequentially
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for @param at the beginning of a line
		if strings.HasPrefix(line, "@param ") {
			// Extract param name and start of description
			parts := strings.SplitN(line[7:], " ", 2) // Skip "@param "
			if len(parts) < 2 {
				continue
			}

			paramName := strings.TrimSpace(parts[0])
			paramDesc := strings.TrimSpace(parts[1])

			// Collect continuation lines until we hit another tag or end
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				if strings.HasPrefix(nextLine, "@") {
					break // Hit another tag
				}
				paramDesc += " " + nextLine
				i = j // Skip these lines in the main loop
			}

			// Clean up whitespace
			paramDesc = regexp.MustCompile(`\s+`).ReplaceAllString(paramDesc, " ")
			params[paramName] = strings.TrimSpace(paramDesc)
		}
	}

	return params
}

// extractTParams extracts @tparam tags
func extractTParams(content string) map[string]string {
	tparams := make(map[string]string)

	// Split content into lines and process them sequentially
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for @tparam at the beginning of a line
		if strings.HasPrefix(line, "@tparam ") {
			// Extract tparam name and start of description
			parts := strings.SplitN(line[8:], " ", 2) // Skip "@tparam "
			if len(parts) < 2 {
				continue
			}

			tparamName := strings.TrimSpace(parts[0])
			tparamDesc := strings.TrimSpace(parts[1])

			// Collect continuation lines until we hit another tag or end
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				if strings.HasPrefix(nextLine, "@") {
					break // Hit another tag
				}
				tparamDesc += " " + nextLine
				i = j // Skip these lines in the main loop
			}

			// Clean up whitespace
			tparamDesc = regexp.MustCompile(`\s+`).ReplaceAllString(tparamDesc, " ")
			tparams[tparamName] = strings.TrimSpace(tparamDesc)
		}
	}

	return tparams
}

// extractReturns extracts @return or @returns tag
func extractReturns(content string) string {
	returnRegex := regexp.MustCompile(`(?s)@returns?\s+(.+?)(\s*@|$)`)
	if matches := returnRegex.FindStringSubmatch(content); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		// Remove extra whitespace and newlines
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}
	return ""
}

// extractGroups extracts @ingroup tags
func extractGroups(content string) []string {
	var groups []string
	groupRegex := regexp.MustCompile(`@ingroup\s+(\w+)`)
	matches := groupRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			groups = append(groups, strings.TrimSpace(match[1]))
		}
	}

	return groups
}

// extractSince extracts @since tag
func extractSince(content string) string {
	sinceRegex := regexp.MustCompile(`(?s)@since\s+(.+?)(\s*@|$)`)
	if matches := sinceRegex.FindStringSubmatch(content); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}
	return ""
}

// extractDeprecated extracts @deprecated tag
func extractDeprecated(content string) string {
	deprecatedRegex := regexp.MustCompile(`(?s)@deprecated\s+(.+?)(\s*@|$)`)
	if matches := deprecatedRegex.FindStringSubmatch(content); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}
	return ""
}

// extractCustomTags extracts any other @tag entries
func extractCustomTags(content string) map[string]string {
	customTags := make(map[string]string)

	// Known tags to skip
	knownTags := map[string]bool{
		"brief": true, "details": true, "param": true, "tparam": true,
		"return": true, "returns": true, "ingroup": true, "since": true,
		"deprecated": true,
	}

	// Split content into lines and process them sequentially
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for @tag at the beginning of a line
		if strings.HasPrefix(line, "@") {
			// Extract tag name and start of value
			parts := strings.SplitN(line[1:], " ", 2) // Skip "@"
			if len(parts) < 2 {
				continue
			}

			tagName := strings.TrimSpace(parts[0])
			if knownTags[tagName] {
				continue // Skip known tags
			}

			tagValue := strings.TrimSpace(parts[1])

			// Collect continuation lines until we hit another tag or end
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				if strings.HasPrefix(nextLine, "@") {
					break // Hit another tag
				}
				tagValue += " " + nextLine
				i = j // Skip these lines in the main loop
			}

			// Clean up whitespace
			tagValue = regexp.MustCompile(`\s+`).ReplaceAllString(tagValue, " ")
			customTags[tagName] = strings.TrimSpace(tagValue)
		}
	}

	return customTags
}

// IsEmpty returns true if the comment has no meaningful content
func (dc *DoxygenComment) IsEmpty() bool {
	return dc.Brief == "" && dc.Detailed == "" && len(dc.Params) == 0 &&
		len(dc.TParams) == 0 && dc.Returns == "" && len(dc.Groups) == 0 &&
		dc.Since == "" && dc.Deprecated == "" && len(dc.CustomTags) == 0
}

// HasDoxygenContent returns true if this appears to be a Doxygen comment
func (dc *DoxygenComment) HasDoxygenContent() bool {
	// Check for Doxygen-style markers in raw content
	if strings.Contains(dc.Raw, "/**") || strings.Contains(dc.Raw, "///") {
		return true
	}

	// Check for any explicit Doxygen tags in raw content
	if strings.Contains(dc.Raw, "@brief") || strings.Contains(dc.Raw, "@param") ||
		strings.Contains(dc.Raw, "@return") || strings.Contains(dc.Raw, "@tparam") ||
		strings.Contains(dc.Raw, "@details") || strings.Contains(dc.Raw, "@since") ||
		strings.Contains(dc.Raw, "@deprecated") || strings.Contains(dc.Raw, "@ingroup") {
		return true
	}

	// Check for structured content (parsed fields)
	if dc.Brief != "" || dc.Detailed != "" || dc.Returns != "" ||
		dc.Since != "" || dc.Deprecated != "" {
		return true
	}

	// Check for parameters or template parameters
	if len(dc.Params) > 0 || len(dc.TParams) > 0 {
		return true
	}

	// Check for groups or custom tags
	if len(dc.Groups) > 0 || len(dc.CustomTags) > 0 {
		return true
	}

	// If none of the above, it's probably not a Doxygen comment
	return false
}
