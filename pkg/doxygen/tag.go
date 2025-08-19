package doxygen

import (
	"fmt"
	"strings"
)

/*
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
*/
type DoxygenTag struct {
	Key   string
	Value string
}

type DoxygenTags []DoxygenTag

func (tags *DoxygenTags) Add(name, content string) {
	*tags = append(*tags, DoxygenTag{Key: name, Value: content})
}

func (tags *DoxygenTags) Get(name string) DoxygenTags {
	var contents DoxygenTags
	for _, tag := range *tags {
		if tag.Key == name {
			contents = append(contents, tag)
		}
	}
	return contents
}

func (tags *DoxygenTags) Set(name, content string) {
	for i, tag := range *tags {
		if tag.Key == name {
			(*tags)[i].Value = content
			return
		}
	}
	// If the tag doesn't exist, add it
	tags.Add(name, content)
}

func (tags *DoxygenTags) Remove(name string) {
	for i, tag := range *tags {
		if tag.Key == name {
			*tags = append((*tags)[:i], (*tags)[i+1:]...)
			return
		}
	}
}

func (tags *DoxygenTags) SetParam(name, content string) {

	for _, tag := range *tags {
		if tag.Key == "param" && strings.HasPrefix(tag.Value, name+" ") {
			varName := strings.SplitN(tag.Value, " ", 2)[0]
			if varName == name {
				tag.Value = fmt.Sprintf("%s %s", varName, content)
				return
			}
		}
	}
	tags.Add("param", fmt.Sprintf("%s %s", name, content))
}

func (tags *DoxygenTags) GetParam(name, dir string) (string, bool) {

	for _, tag := range *tags {
		if tag.Key == "param" && strings.HasPrefix(tag.Value, name+" ") {
			tokens := strings.SplitN(tag.Value, " ", 2)
			if tokens[0] == name {
				return tokens[1], true
			}
		}
	}
	return "", false
}

func (tags *DoxygenTags) SetTParam(name, content string) {

	for _, tag := range *tags {
		if tag.Key == "tparam" && strings.HasPrefix(tag.Value, name+" ") {
			varName := strings.SplitN(tag.Value, " ", 2)[0]
			if varName == name {
				tag.Value = fmt.Sprintf("%s %s", varName, content)
				return
			}
		}
	}
	// If the tag doesn't exist, add it
	tags.Add("tparam", fmt.Sprintf("%s %s", name, content))
}

func (tags *DoxygenTags) GetTParam(name string) (string, bool) {
	params := tags.Get("tparam")
	for _, tag := range params {
		tokens := strings.SplitN(tag.Value, " ", 2)
		if tokens[0] == name {
			return tokens[1], true
		}
	}
	return "", false
}

func (tag *DoxygenTag) AsString() string {
	return fmt.Sprintf("@%s %s", tag.Key, tag.Value)
}

func (tags *DoxygenTags) AsStrings() []string {
	var tagStrings []string
	for _, tag := range *tags {
		tagStrings = append(tagStrings, tag.AsString())
	}
	return tagStrings
}
