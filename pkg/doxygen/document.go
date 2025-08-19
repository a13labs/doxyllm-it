package doxygen

import (
	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/source"
	"fmt"
	"os"
)

// Document represents a layer of documentation entities.
type Document struct {
	entitiesCache []*Entity
	sourceFile    *source.SourceFile
	isModified    bool
}

func NewFromContent(filename string, content string) (*Document, error) {
	src, err := source.NewSourceFromContent(filename, content)
	if err != nil {
		return nil, err
	}

	layer := &Document{
		entitiesCache: make([]*Entity, 0),
		sourceFile:    src,
		isModified:    false,
	}

	instructions := src.GetInstructions()
	for _, instr := range instructions {
		var doxygenComment *DocumentationEntry
		var comment *ast.Entity
		var isNew bool

		if instr.Type == ast.EntityName {
			comment = instr.LineComment
		} else {
			comment = src.GetTree().GetPrecedingEntity(instr)
		}

		// Try to find a preceding comment since this is not a name entity
		if dc, err := createDocumentationEntry(comment); err == nil {
			doxygenComment = dc
		} else {
			isNew = true
			comment = &ast.Entity{
				Type:      ast.EntityComment,
				Signature: "/** @brief No description provided. */",
			}
			if instr.Type == ast.EntityName {
				instr.LineComment = comment
			} else {
				src.GetTree().InsertBefore(instr, comment)
			}

			doxygenComment = &DocumentationEntry{
				Raw:        comment,
				CustomTags: make(DoxygenTags, 0),
			}
		}

		// Create the Doxygen entity and add it to the cache
		entity := &Entity{
			layer: layer,
			instr: instr,
			doc:   doxygenComment,
			isNew: isNew,
		}
		layer.entitiesCache = append(layer.entitiesCache, entity)
	}

	return layer, nil
}

func NewFromFile(filename string) (*Document, error) {

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return NewFromContent(filename, string(content))
}

func (l *Document) GetEntities(newOnly bool) []*Entity {
	if newOnly {
		var newOnlyEntities []*Entity
		for _, entity := range l.entitiesCache {
			if entity.isNew {
				newOnlyEntities = append(newOnlyEntities, entity)
			}
		}
		return newOnlyEntities
	}
	return l.entitiesCache
}

func (l *Document) GetUndocumentedEntities() []*Entity {
	return l.GetEntities(true)
}

func (l *Document) GetAllEntities() []*Entity {
	return l.GetEntities(false)
}

func (l *Document) IsModified() bool {
	return l.isModified
}

func (l *Document) SaveToString(clanged bool) (string, error) {
	f := formatter.New()

	reconstructed := f.ReconstructCode(l.sourceFile.GetTree(), -1, true)
	if clanged {
		return f.FormatWithClang(reconstructed)
	}

	return reconstructed, nil
}

func (l *Document) Save(clanged bool) error {
	content, err := l.SaveToString(clanged)
	if err != nil {
		return err
	}
	return os.WriteFile(l.sourceFile.GetFilename(), []byte(content), 0644)
}
