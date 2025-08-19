package doxygen

import (
	"doxyllm-it/pkg/ast"
	"fmt"
	"strings"
)

// DocumentationEntry represents a single entry in the documentation.
type DocumentationEntry struct {
	Raw        *ast.Entity // Original comment text
	CustomTags DoxygenTags
}

// createDocumentationEntry parses a doxygen comment block
func createDocumentationEntry(comment *ast.Entity) (*DocumentationEntry, error) {
	if comment == nil || comment.Type != ast.EntityComment {
		return nil, fmt.Errorf("invalid comment type")
	}

	doc := &DocumentationEntry{
		Raw:        comment,
		CustomTags: make(DoxygenTags, 0),
	}

	err := doc.Parse(comment.Signature)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (doc *DocumentationEntry) Parse(raw string) error {

	// Clean up the comment (remove /** */ and leading *)
	lines := strings.Split(raw, "\n")
	var cleanLines []string
	isTrailingComment := strings.HasPrefix(raw, "/**<") ||
		strings.HasPrefix(raw, "///<") ||
		strings.HasPrefix(raw, "//!<")

	for i, line := range lines {
		clean := strings.TrimSpace(line)

		// Remove comment markers
		if i == 0 && strings.HasPrefix(clean, "/**") {
			clean = strings.TrimPrefix(clean, "/**")
			// For trailing comments, preserve the < character as it's semantically important
			// Don't strip it here - it will be handled by the formatter
		}
		if i == len(lines)-1 && strings.HasSuffix(clean, "*/") {
			clean = strings.TrimSuffix(clean, "*/")
		}
		if after, ok := strings.CutPrefix(clean, "///"); ok {
			clean = after
			// For trailing comments, preserve the < character as it's semantically important
			// Don't strip it here - it will be handled by the formatter
		}
		if after, ok := strings.CutPrefix(clean, "//!"); ok {
			clean = after
			// For trailing comments, preserve the < character as it's semantically important
			// Don't strip it here - it will be handled by the formatter
		}
		clean = strings.TrimPrefix(clean, "*")

		clean = strings.TrimSpace(clean)
		if clean != "" {
			cleanLines = append(cleanLines, clean)
		}
	}

	// Parse doxygen tags
	var currentTag string
	var currentContent []string
	numTags := 0

	for _, line := range cleanLines {
		if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "\\") {
			// Save previous tag
			if currentTag != "" {
				if strings.HasPrefix(currentTag, "param") {
					content := strings.Join(currentContent, " ")
					tokens := strings.SplitN(content, " ", 2)
					doc.CustomTags.SetParam(tokens[0], tokens[1])
				} else if currentTag == "tparam" {
					content := strings.Join(currentContent, " ")
					tokens := strings.SplitN(content, " ", 2)
					doc.CustomTags.SetTParam(tokens[0], tokens[1])
				} else {
					doc.CustomTags.Set(currentTag, strings.Join(currentContent, " "))
				}
			}
			// Start new tag
			parts := strings.SplitN(line[1:], " ", 2)
			currentTag = parts[0]
			currentContent = []string{}

			if len(parts) > 1 {
				currentContent = append(currentContent, parts[1])
			}
			numTags++
		} else {
			if currentTag == "" {
				// For trailing comments, the text should be treated as brief description
				// without creating verbose @brief tags
				if isTrailingComment {
					briefTag := doc.CustomTags.Get("brief")
					if len(briefTag) > 0 {
						doc.CustomTags.Set("brief", briefTag[0].Value+" "+line)
					} else {
						doc.CustomTags.Set("brief", line)
					}
				}
			} else {
				currentContent = append(currentContent, line)
			}
		}
	}

	// Save last tag
	if currentTag != "" {
		if strings.HasPrefix(currentTag, "param") {
			content := strings.Join(currentContent, " ")
			tokens := strings.SplitN(content, " ", 2)
			doc.CustomTags.SetParam(tokens[0], tokens[1])
		} else if currentTag == "tparam" {
			content := strings.Join(currentContent, " ")
			tokens := strings.SplitN(content, " ", 2)
			doc.CustomTags.SetTParam(tokens[0], tokens[1])
		} else {
			doc.CustomTags.Set(currentTag, strings.Join(currentContent, " "))
		}
	}

	return nil
}

// AsBlockComment converts the DoxygenComment to a block comment representation
func (doc *DocumentationEntry) AsBlockComment() string {
	var lines []string
	lines = append(lines, "/**")
	tags := doc.CustomTags.AsStrings()
	for _, tag := range tags {
		lines = append(lines, fmt.Sprintf(" * %s", tag))
	}
	lines = append(lines, " */")
	return strings.Join(lines, "\n")
}

// AsLineComment returns DoxygenComment as a line comment representation, used for variables (ie: /**< The time point when the timer started. */)
func (doc *DocumentationEntry) AsLineComment() string {
	var commentBuilder strings.Builder
	commentBuilder.WriteString("/**< ")
	brief := doc.CustomTags.Get("brief")
	if len(brief) > 0 {
		commentBuilder.WriteString(brief[0].Value)
	}
	commentBuilder.WriteString(" */")
	return commentBuilder.String()
}
