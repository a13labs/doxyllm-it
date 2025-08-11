package ast

import (
	"testing"
)

func TestEntityTypes(t *testing.T) {
	// Test comment entity
	commentEntity := &Entity{
		Type:         EntityComment,
		OriginalText: "// Line comment",
	}
	if commentEntity.Type != EntityComment {
		t.Errorf("Expected EntityComment, got %v", commentEntity.Type)
	}

	// Test function entity
	functionEntity := &Entity{
		Type: EntityFunction,
		Name: "testFunc",
	}
	if functionEntity.Type != EntityFunction {
		t.Errorf("Expected EntityFunction, got %v", functionEntity.Type)
	}
}

func TestEntityPaths(t *testing.T) {
	// Create a simple hierarchy: namespace::class::method
	root := &Entity{
		Type: EntityUnknown,
		Name: "",
	}

	namespace := &Entity{
		Type: EntityNamespace,
		Name: "MyNamespace",
	}
	root.AddChild(namespace)

	class := &Entity{
		Type: EntityClass,
		Name: "MyClass",
	}
	namespace.AddChild(class)

	method := &Entity{
		Type: EntityMethod,
		Name: "myMethod",
	}
	class.AddChild(method)

	// Test paths
	if root.GetFullPath() != "" {
		t.Errorf("Root should have empty path, got %s", root.GetFullPath())
	}

	if namespace.GetFullPath() != "MyNamespace" {
		t.Errorf("Namespace path should be 'MyNamespace', got %s", namespace.GetFullPath())
	}

	if class.GetFullPath() != "MyNamespace::MyClass" {
		t.Errorf("Class path should be 'MyNamespace::MyClass', got %s", class.GetFullPath())
	}

	if method.GetFullPath() != "MyNamespace::MyClass::myMethod" {
		t.Errorf("Method path should be 'MyNamespace::MyClass::myMethod', got %s", method.GetFullPath())
	}

	// Test scope
	if !root.IsGlobal() {
		t.Errorf("Root should be global")
	}
	if !namespace.IsGlobal() {
		t.Errorf("Namespace should be global since its parent is the unnamed root")
	}
	if method.GetScope() != "MyNamespace::MyClass" {
		t.Errorf("Method scope should be 'MyNamespace::MyClass', got %s", method.GetScope())
	}
}

func TestScopeTree(t *testing.T) {
	content := "// Test file content"
	tree := NewScopeTree("test.hpp", content)

	if tree.Filename != "test.hpp" {
		t.Errorf("Expected filename 'test.hpp', got %s", tree.Filename)
	}

	if tree.Content != content {
		t.Errorf("Expected content '%s', got %s", content, tree.Content)
	}

	if tree.Root == nil {
		t.Errorf("Root should not be nil")
	}

	if tree.Root.Type != EntityRoot {
		t.Errorf("Root type should be EntityRoot, got %v", tree.Root.Type)
	}

	// Test finding root entity
	root := tree.FindEntity("")
	if root != tree.Root {
		t.Errorf("Empty path should return root")
	}

	root = tree.FindEntity("::")
	if root != tree.Root {
		t.Errorf("'::' path should return root")
	}
}

func TestAccessLevel(t *testing.T) {
	tests := []struct {
		level    AccessLevel
		expected string
	}{
		{AccessPublic, "public"},
		{AccessProtected, "protected"},
		{AccessPrivate, "private"},
		{AccessUnknown, "unknown"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("AccessLevel %v should be %s, got %s", test.level, test.expected, test.level.String())
		}
	}
}

func TestEntityType(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityClass, "class"},
		{EntityFunction, "function"},
		{EntityNamespace, "namespace"},
		{EntityUnknown, "unknown"},
	}

	for _, test := range tests {
		if test.entityType.String() != test.expected {
			t.Errorf("EntityType %v should be %s, got %s", test.entityType, test.expected, test.entityType.String())
		}
	}
}
