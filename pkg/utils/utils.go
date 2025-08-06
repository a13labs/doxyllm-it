// Package utils provides utility functions for the parser
package utils

import (
	"bufio"
	"regexp"
	"strings"
)

// DoxygenCommentParser handles parsing of doxygen comments from source code
type DoxygenCommentParser struct {
	content string
	lines   []string
}

// NewDoxygenCommentParser creates a new doxygen comment parser
func NewDoxygenCommentParser(content string) *DoxygenCommentParser {
	return &DoxygenCommentParser{
		content: content,
		lines:   strings.Split(content, "\n"),
	}
}

// FindCommentForLine finds the doxygen comment that precedes the given line
func (p *DoxygenCommentParser) FindCommentForLine(lineNum int) string {
	if lineNum <= 1 || lineNum > len(p.lines) {
		return ""
	}

	// Look backwards from the line to find a doxygen comment
	for i := lineNum - 2; i >= 0; i-- {
		line := strings.TrimSpace(p.lines[i])

		// Stop if we hit a non-comment, non-empty line
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "/*") && !strings.HasSuffix(line, "*/") {
			break
		}

		// Check if this is the start of a doxygen comment block
		if strings.HasPrefix(line, "/**") {
			// Found the start, extract the full comment
			return p.extractCommentBlock(i, lineNum-1)
		}

		// Check for single-line doxygen comments
		if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
			return p.extractSingleLineComments(i, lineNum-1)
		}
	}

	return ""
}

// extractCommentBlock extracts a /* */ style comment block
func (p *DoxygenCommentParser) extractCommentBlock(startLine, endLine int) string {
	var comment strings.Builder

	for i := startLine; i <= endLine; i++ {
		if i >= len(p.lines) {
			break
		}

		line := p.lines[i]
		comment.WriteString(line)
		if i < endLine {
			comment.WriteString("\n")
		}

		// Stop if we hit the end of comment
		if strings.Contains(line, "*/") {
			break
		}
	}

	return comment.String()
}

// extractSingleLineComments extracts consecutive single-line comments
func (p *DoxygenCommentParser) extractSingleLineComments(startLine, endLine int) string {
	var comment strings.Builder

	for i := startLine; i <= endLine; i++ {
		if i >= len(p.lines) {
			break
		}

		line := strings.TrimSpace(p.lines[i])

		// Stop if not a single-line comment
		if !strings.HasPrefix(line, "///") && !strings.HasPrefix(line, "//!") {
			break
		}

		comment.WriteString(p.lines[i])
		if i < endLine {
			comment.WriteString("\n")
		}
	}

	return comment.String()
}

// IsDoxygenComment checks if a string is a doxygen comment
func IsDoxygenComment(text string) bool {
	trimmed := strings.TrimSpace(text)
	return strings.HasPrefix(trimmed, "/**") ||
		strings.HasPrefix(trimmed, "///") ||
		strings.HasPrefix(trimmed, "//!") ||
		strings.HasPrefix(trimmed, "/*!")
}

// CleanComment removes comment markers and normalizes whitespace
func CleanComment(comment string) string {
	if comment == "" {
		return ""
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(comment))

	for scanner.Scan() {
		line := scanner.Text()
		cleaned := cleanCommentLine(line)
		if cleaned != "" || result.Len() > 0 {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(cleaned)
		}
	}

	return result.String()
}

// cleanCommentLine cleans a single line of a comment
func cleanCommentLine(line string) string {
	// Remove leading/trailing whitespace
	line = strings.TrimSpace(line)

	// Remove comment markers
	patterns := []string{
		`^/\*\*`, // /**
		`^/\*!`,  // /*!
		`^\*/`,   // */
		`^///`,   // ///
		`^//!`,   // //!
		`^\*\s?`, // * or *
		`\*/$`,   // */
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		line = re.ReplaceAllString(line, "")
	}

	return strings.TrimSpace(line)
}

// ExtractBrief extracts the brief description from a doxygen comment
func ExtractBrief(comment string) string {
	lines := strings.Split(CleanComment(comment), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for @brief or \brief
		if strings.HasPrefix(line, "@brief ") {
			return strings.TrimSpace(line[7:])
		}
		if strings.HasPrefix(line, "\\brief ") {
			return strings.TrimSpace(line[7:])
		}

		// If no explicit @brief, the first non-empty line is the brief
		if line != "" && !strings.HasPrefix(line, "@") && !strings.HasPrefix(line, "\\") {
			return line
		}
	}

	return ""
}

// SplitPath splits a C++ qualified name into parts
func SplitPath(path string) []string {
	if path == "" {
		return []string{}
	}

	// Handle global scope
	if path == "::" {
		return []string{}
	}

	// Remove leading/trailing ::
	path = strings.Trim(path, ":")

	if path == "" {
		return []string{}
	}

	return strings.Split(path, "::")
}

// JoinPath joins path parts into a C++ qualified name
func JoinPath(parts []string) string {
	if len(parts) == 0 {
		return "::"
	}

	return strings.Join(parts, "::")
}

// IsValidCppIdentifier checks if a string is a valid C++ identifier
func IsValidCppIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// Must start with letter or underscore
	if !isLetter(rune(name[0])) && name[0] != '_' {
		return false
	}

	// Rest must be letters, digits, or underscores
	for _, char := range name[1:] {
		if !isLetter(char) && !isDigit(char) && char != '_' {
			return false
		}
	}

	return true
}

// isLetter checks if a rune is a letter
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isDigit checks if a rune is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// RemoveTemplateParams removes template parameters from a type name
func RemoveTemplateParams(typeName string) string {
	depth := 0
	var result strings.Builder

	for _, char := range typeName {
		if char == '<' {
			depth++
		} else if char == '>' {
			depth--
		} else if depth == 0 {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ParseAccessLevel parses C++ access level keywords
func ParseAccessLevel(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "public:") {
		return "public", true
	}
	if strings.HasPrefix(trimmed, "protected:") {
		return "protected", true
	}
	if strings.HasPrefix(trimmed, "private:") {
		return "private", true
	}

	return "", false
}
