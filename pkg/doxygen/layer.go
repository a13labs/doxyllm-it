package doxygen

import (
	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/source"
	"fmt"
	"os"
)

// DocLayer represents a layer of documentation entities.
type DocLayer struct {
	entitiesCache []*Entity
	sourceFile    *source.SourceFile
	isModified    bool
}

func NewLayerFromContent(filename string, content string) (*DocLayer, error) {
	src, err := source.NewSourceFromContent(filename, content)
	if err != nil {
		return nil, err
	}

	layer := &DocLayer{
		entitiesCache: make([]*Entity, 0),
		sourceFile:    src,
		isModified:    false,
	}

	instructions := src.GetInstructions()
	for _, instr := range instructions {
		var doxygenComment *DataEntry
		var comment *ast.Entity
		var isNew bool

		if instr.Type == ast.EntityName {
			comment = instr.LineComment
		} else {
			comment = src.GetTree().GetPrecedingEntity(instr)
		}

		// Try to find a preceding comment since this is not a name entity
		if dc, err := newDataEntry(comment); err == nil {
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

			doxygenComment = &DataEntry{
				Raw:        comment,
				Params:     make(map[string]string),
				TParams:    make(map[string]string),
				CustomTags: make(map[string]string),
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

func NewLayer(filename string) (*DocLayer, error) {

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return NewLayerFromContent(filename, string(content))
}

func (l *DocLayer) GetEntities(undocumented bool) []*Entity {
	if undocumented {
		var undocumentedEntities []*Entity
		for _, entity := range l.entitiesCache {
			if entity.isNew {
				undocumentedEntities = append(undocumentedEntities, entity)
			}
		}
		return undocumentedEntities
	}
	return l.entitiesCache
}

func (l *DocLayer) GetUndocumentedEntities() []*Entity {
	return l.GetEntities(true)
}

func (l *DocLayer) IsModified() bool {
	return l.isModified
}

func (l *DocLayer) SaveToString(clanged bool) (string, error) {
	f := formatter.New()

	reconstructed := f.ReconstructCode(l.sourceFile.GetTree())
	if clanged {
		return f.FormatWithClang(reconstructed)
	}

	return reconstructed, nil
}

func (l *DocLayer) Save(clanged bool) error {
	content, err := l.SaveToString(clanged)
	if err != nil {
		return err
	}

	return os.WriteFile(l.sourceFile.GetFilename(), []byte(content), 0644)
}
