package parser

import (
	"strings"

	"doxyllm-it/pkg/ast"
)

// ParseDoxygenComment parses a doxygen comment block
func ParseDoxygenComment(comment string) *ast.DoxygenComment {
	if comment == "" {
		return nil
	}

	doc := &ast.DoxygenComment{
		Raw:        comment,
		Params:     make(map[string]string),
		CustomTags: make(map[string]string),
	}

	// Clean up the comment (remove /** */ and leading *)
	lines := strings.Split(comment, "\n")
	var cleanLines []string

	for i, line := range lines {
		clean := strings.TrimSpace(line)

		// Remove comment markers
		if i == 0 && strings.HasPrefix(clean, "/**") {
			clean = strings.TrimPrefix(clean, "/**")
		}
		if i == len(lines)-1 && strings.HasSuffix(clean, "*/") {
			clean = strings.TrimSuffix(clean, "*/")
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

	for _, line := range cleanLines {
		if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "\\") {
			// Save previous tag
			if currentTag != "" {
				setDoxygenTag(doc, currentTag, strings.Join(currentContent, " "))
			}
			// Start new tag
			parts := strings.SplitN(line[1:], " ", 2)
			currentTag = parts[0]
			currentContent = []string{}

			if len(parts) > 1 {
				currentContent = append(currentContent, parts[1])
			}
		} else {
			if currentTag == "" {
				// This is part of the main description
				if doc.Brief == "" {
					doc.Brief = line
				} else {
					if doc.Detailed == "" {
						doc.Detailed = line
					} else {
						doc.Detailed += " " + line
					}
				}
			} else {
				currentContent = append(currentContent, line)
			}
		}
	}

	// Save last tag
	if currentTag != "" {
		setDoxygenTag(doc, currentTag, strings.Join(currentContent, " "))
	}

	return doc
}

// setDoxygenTag sets a doxygen tag value
func setDoxygenTag(doc *ast.DoxygenComment, tag, content string) {
	switch tag {
	case "brief":
		doc.Brief = content
	case "details", "detailed":
		doc.Detailed = content
	case "param":
		parts := strings.SplitN(content, " ", 2)
		if len(parts) == 2 {
			doc.Params[parts[0]] = parts[1]
		}
	case "return", "returns":
		doc.Returns = content
	case "throw", "throws", "exception":
		doc.Throws = append(doc.Throws, content)
	case "since":
		doc.Since = content
	case "deprecated":
		doc.Deprecated = content
	case "see":
		doc.See = append(doc.See, content)
	case "author":
		doc.Author = content
	case "version":
		doc.Version = content
	// Group-related tags
	case "defgroup":
		doc.Defgroup = content
	case "ingroup":
		doc.Ingroup = append(doc.Ingroup, content)
	case "addtogroup":
		doc.Addtogroup = content
	// Structural tags
	case "file":
		doc.File = content
	case "namespace":
		doc.Namespace = content
	case "class":
		doc.Class = content
	default:
		doc.CustomTags[tag] = content
	}
}
