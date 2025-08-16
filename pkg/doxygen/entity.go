package doxygen

import (
	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
)

// Entity represents a single entity with its associated documentation.
type Entity struct {
	layer *DocLayer   // The layer this entity belongs to
	instr *ast.Entity // The entity this comment is associated with
	doc   *DataEntry
	isNew bool
}

func (e *Entity) GetInstruction() *ast.Entity {
	return e.instr
}

func (e *Entity) IsNew() bool {
	return e.isNew
}

func (e *Entity) ApplyRaw(raw string) error {
	err := e.doc.Parse(raw)
	if err != nil {
		return err
	}
	e.layer.isModified = true
	return nil
}

func (e *Entity) Context(includeParent bool, includeSiblings bool) string {
	f := formatter.New()
	return f.ExtractEntityContext(e.instr, includeParent, includeSiblings)
}
