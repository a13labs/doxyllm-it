package ast

import (
	"testing"
)

func TestEntityTypes(t *testing.T) {
	// Test comment entity
	commentEntity := &Entity{
		Type: EntityComment,
		Body: []string{"// Line comment"},
	}
	if commentEntity.Type != EntityComment {
		t.Errorf("Expected EntityComment, got %v", commentEntity.Type)
	}

	// Test function entity
	functionEntity := &Entity{
		Type: EntityCallable,
		Name: "testFunc",
	}
	if functionEntity.Type != EntityCallable {
		t.Errorf("Expected EntityFunction, got %v", functionEntity.Type)
	}
}

func TestEntityPaths(t *testing.T) {
	// Create a simple hierarchy: namespace::class::method
	root := &Entity{
		Type: EntityRoot,
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

	function := &Entity{
		Type: EntityCallable,
		Name: "myFunction",
	}
	class.AddChild(function)

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

	if function.GetFullPath() != "MyNamespace::MyClass::myFunction" {
		t.Errorf("Function path should be 'MyNamespace::MyClass::myFunction', got %s", function.GetFullPath())
	}

	if function.GetScope() != "MyNamespace::MyClass" {
		t.Errorf("Function scope should be 'MyNamespace::MyClass', got %s", function.GetScope())
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
		{EntityCallable, "function"},
		{EntityNamespace, "namespace"},
	}

	for _, test := range tests {
		if test.entityType.String() != test.expected {
			t.Errorf("EntityType %v should be %s, got %s", test.entityType, test.expected, test.entityType.String())
		}
	}
}
