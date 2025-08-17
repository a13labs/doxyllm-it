package doxygen

import (
	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"strings"
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

	// Filter out the comment from the raw string by searching for the comment delimiter
	// since sometimes the LLM may pass also the code after the doxygen comment
	commentStart := strings.Index(raw, "/*")
	commentEnd := strings.Index(raw, "*/")
	if commentStart != -1 && commentEnd != -1 && commentEnd > commentStart {
		raw = raw[commentStart : commentEnd+2]
	}

	err := e.doc.Parse(raw)
	if err != nil {
		return err
	}
	if e.instr.Type == ast.EntityName {
		e.doc.Raw.Signature = e.doc.AsLineComment()
	} else {
		e.doc.Raw.Signature = e.doc.AsBlockComment()
	}
	e.layer.isModified = true
	return nil
}

func (e *Entity) Context(includeParent bool, includeSiblings bool) string {
	f := formatter.New()
	return f.ExtractEntityContext(e.instr)
}
