package doxygen

import (
	"doxyllm-it/pkg/ast"
	"fmt"
	"strings"
)

// DataEntry represents a single entry in the documentation.
type DataEntry struct {
	Raw        *ast.Entity       // Original comment text
	Brief      string            // Brief description
	Detailed   string            // Detailed description
	Params     map[string]string // Parameter documentation (@param)
	TParams    map[string]string // Template parameter documentation (@tparam)
	Returns    string            // Return value documentation
	Throws     []string          // Exception documentation
	Since      string            // Since version
	Deprecated string            // Deprecation notice
	See        []string          // See also references
	Author     string            // Author information
	Version    string            // Version information
	// Group-related tags
	Defgroup   string   // @defgroup tag (for group definitions)
	Ingroup    []string // @ingroup tags (group memberships)
	Addtogroup string   // @addtogroup tag
	// Structural tags
	File       string            // @file tag
	Namespace  string            // @namespace tag
	Class      string            // @class tag
	CustomTags map[string]string // Custom doxygen tags
}

// newDataEntry parses a doxygen comment block
func newDataEntry(comment *ast.Entity) (*DataEntry, error) {
	if comment == nil || comment.Type != ast.EntityComment {
		return nil, fmt.Errorf("invalid comment type")
	}

	doc := &DataEntry{
		Raw:        comment,
		Params:     make(map[string]string),
		TParams:    make(map[string]string),
		CustomTags: make(map[string]string),
	}

	err := doc.Parse(comment.Signature)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (doc *DataEntry) Parse(raw string) error {

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
				doc.SetTag(currentTag, strings.Join(currentContent, " "))
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
			numTags++
			if currentTag == "" {
				// For trailing comments, the text should be treated as brief description
				// without creating verbose @brief tags
				if isTrailingComment {
					if doc.Brief == "" {
						doc.Brief = line
					} else {
						doc.Brief += " " + line
					}
				} else {
					// This is part of the main description for regular comments
					if doc.Brief == "" {
						doc.Brief = line
					} else {
						if doc.Detailed == "" {
							doc.Detailed = line
						} else {
							doc.Detailed += " " + line
						}
					}
				}
			} else {
				currentContent = append(currentContent, line)
			}
		}
	}
	// If no tags were found, treat the comment as a regular description
	if numTags == 0 {
		return fmt.Errorf("no doxygen tags found in comment")
	}

	// Save last tag
	if currentTag != "" {
		doc.SetTag(currentTag, strings.Join(currentContent, " "))
	}

	return nil
}

// setDoxygenTag sets a doxygen tag value
func (doc *DataEntry) SetTag(tag, content string) {
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
	case "tparam":
		parts := strings.SplitN(content, " ", 2)
		if len(parts) == 2 {
			doc.TParams[parts[0]] = parts[1]
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

// GetTag retrieves a doxygen tag value
func (doc *DataEntry) GetTag(tag string) string {
	switch tag {
	case "brief":
		return doc.Brief
	case "details", "detailed":
		return doc.Detailed
	case "param":
		return doc.Params[tag]
	case "tparam":
		return doc.TParams[tag]
	case "return", "returns":
		return doc.Returns
	case "throw", "throws", "exception":
		return strings.Join(doc.Throws, ", ")
	case "since":
		return doc.Since
	case "deprecated":
		return doc.Deprecated
	case "see":
		return strings.Join(doc.See, ", ")
	case "author":
		return doc.Author
	case "version":
		return doc.Version
	// Group-related tags
	case "defgroup":
		return doc.Defgroup
	case "ingroup":
		return strings.Join(doc.Ingroup, ", ")
	case "addtogroup":
		return doc.Addtogroup
	// Structural tags
	case "file":
		return doc.File
	case "namespace":
		return doc.Namespace
	case "class":
		return doc.Class
	default:
		return doc.CustomTags[tag]
	}
}

// AsBlockComment converts the DoxygenComment to a block comment representation
func (doc *DataEntry) AsBlockComment() string {
	var lines []string
	lines = append(lines, "/*")
	if doc.Brief != "" {
		lines = append(lines, fmt.Sprintf(" * %s", doc.Brief))
	}
	if doc.Detailed != "" {
		lines = append(lines, fmt.Sprintf(" * %s", doc.Detailed))
	}
	for param, desc := range doc.Params {
		lines = append(lines, fmt.Sprintf(" * @param %s %s", param, desc))
	}
	for tparam, desc := range doc.TParams {
		lines = append(lines, fmt.Sprintf(" * @tparam %s %s", tparam, desc))
	}
	if doc.Returns != "" {
		lines = append(lines, fmt.Sprintf(" * @return %s", doc.Returns))
	}
	for _, throw := range doc.Throws {
		lines = append(lines, fmt.Sprintf(" * @throw %s", throw))
	}
	if doc.Since != "" {
		lines = append(lines, fmt.Sprintf(" * @since %s", doc.Since))
	}
	if doc.Deprecated != "" {
		lines = append(lines, fmt.Sprintf(" * @deprecated %s", doc.Deprecated))
	}
	for _, see := range doc.See {
		lines = append(lines, fmt.Sprintf(" * @see %s", see))
	}
	if doc.Author != "" {
		lines = append(lines, fmt.Sprintf(" * @author %s", doc.Author))
	}
	if doc.Version != "" {
		lines = append(lines, fmt.Sprintf(" * @version %s", doc.Version))
	}
	if doc.Defgroup != "" {
		lines = append(lines, fmt.Sprintf(" * @defgroup %s", doc.Defgroup))
	}
	for _, ingroup := range doc.Ingroup {
		lines = append(lines, fmt.Sprintf(" * @ingroup %s", ingroup))
	}
	if doc.Addtogroup != "" {
		lines = append(lines, fmt.Sprintf(" * @addtogroup %s", doc.Addtogroup))
	}
	if doc.File != "" {
		lines = append(lines, fmt.Sprintf(" * @file %s", doc.File))
	}
	if doc.Namespace != "" {
		lines = append(lines, fmt.Sprintf(" * @namespace %s", doc.Namespace))
	}
	if doc.Class != "" {
		lines = append(lines, fmt.Sprintf(" * @class %s", doc.Class))
	}
	for tag, content := range doc.CustomTags {
		lines = append(lines, fmt.Sprintf(" * @%s %s", tag, content))
	}
	lines = append(lines, " */")
	return strings.Join(lines, "\n")
}

// AsLineComment returns DoxygenComment as a line comment representation, used for variables (ie: /**< The time point when the timer started. */)
func (doc *DataEntry) AsLineComment() string {
	var commentBuilder strings.Builder
	commentBuilder.WriteString("/»»< ")
	if doc.Brief != "" {
		commentBuilder.WriteString(doc.Brief)
	}
	commentBuilder.WriteString(" */")
	return commentBuilder.String()
}
