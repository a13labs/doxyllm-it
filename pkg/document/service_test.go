package document

import (
	"context"
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/llm"
)

// mockLLMService implements a mock LLM service for testing
type mockLLMService struct {
	generateFunc func(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error)
}

func (m *mockLLMService) GenerateDocumentation(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &llm.DocumentationResult{
		Comment:     "/** @brief Mock comment */",
		Description: "Mock description",
		Metadata:    make(map[string]string),
	}, nil
}

func (m *mockLLMService) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockLLMService) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{Name: "mock-model"}
}

func TestNewDocumentationService(t *testing.T) {
	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.llmService != mockLLM {
		t.Fatal("Expected LLM service to be set correctly")
	}
}

func TestProcessUndocumentedEntities(t *testing.T) {
	// Create a test document with undocumented entities
	testContent := `
namespace TestNamespace {
    class TestClass {
    public:
        void undocumentedMethod();
        
        /** @brief Documented method */
        void documentedMethod();
    };
}
`

	doc, err := NewFromContent("test.hpp", testContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create mock LLM service
	mockLLM := &mockLLMService{
		generateFunc: func(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error) {
			return &llm.DocumentationResult{
				Comment:     "/** @brief Generated comment for " + req.EntityName + " */",
				Description: "Generated description",
				Metadata:    make(map[string]string),
			}, nil
		},
	}

	service := NewDocumentationService(mockLLM)

	// Test processing with default options
	ctx := context.Background()
	opts := ProcessingOptions{
		DryRun: false,
	}

	result, err := service.ProcessUndocumentedEntities(ctx, doc, opts)
	if err != nil {
		t.Fatalf("Failed to process entities: %v", err)
	}

	if result.EntitiesProcessed == 0 {
		t.Error("Expected some entities to be processed")
	}

	if result.EntitiesUpdated != result.EntitiesProcessed {
		t.Errorf("Expected all processed entities to be updated, got %d updated out of %d processed",
			result.EntitiesUpdated, result.EntitiesProcessed)
	}

	if len(result.UpdatedEntities) == 0 {
		t.Error("Expected some entities to be updated")
	}
}

func TestProcessUndocumentedEntities_DryRun(t *testing.T) {
	testContent := `
class TestClass {
public:
    void undocumentedMethod();
};
`

	doc, err := NewFromContent("test.hpp", testContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	ctx := context.Background()
	opts := ProcessingOptions{
		DryRun: true,
	}

	result, err := service.ProcessUndocumentedEntities(ctx, doc, opts)
	if err != nil {
		t.Fatalf("Failed to process entities: %v", err)
	}

	// In dry run, entities should be "processed" but not actually updated
	if result.EntitiesUpdated != 0 {
		t.Errorf("Expected no entities to be updated in dry run, got %d", result.EntitiesUpdated)
	}

	if len(result.UpdatedEntities) == 0 {
		t.Error("Expected entities to be listed in dry run")
	}
}

func TestProcessUndocumentedEntities_MaxEntities(t *testing.T) {
	testContent := `
class TestClass {
public:
    void method1();
    void method2();
    void method3();
};
`

	doc, err := NewFromContent("test.hpp", testContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	ctx := context.Background()
	opts := ProcessingOptions{
		MaxEntities: 2,
		DryRun:      true,
	}

	result, err := service.ProcessUndocumentedEntities(ctx, doc, opts)
	if err != nil {
		t.Fatalf("Failed to process entities: %v", err)
	}

	if result.EntitiesProcessed > 2 {
		t.Errorf("Expected at most 2 entities to be processed, got %d", result.EntitiesProcessed)
	}
}

func TestProcessUndocumentedEntities_ExcludeTypes(t *testing.T) {
	testContent := `
namespace TestNamespace {
    class TestClass {
    public:
        void testMethod();
        int testVariable;
    };
}
`

	doc, err := NewFromContent("test.hpp", testContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	ctx := context.Background()
	opts := ProcessingOptions{
		ExcludeTypes: []ast.EntityType{ast.EntityVariable, ast.EntityField},
		DryRun:       true,
	}

	result, err := service.ProcessUndocumentedEntities(ctx, doc, opts)
	if err != nil {
		t.Fatalf("Failed to process entities: %v", err)
	}

	// Verify that variables were excluded
	for _, entityPath := range result.UpdatedEntities {
		if strings.Contains(entityPath, "testVariable") {
			t.Error("Expected variables to be excluded from processing")
		}
	}
}

func TestProcessEntitiesNeedingGroupUpdate(t *testing.T) {
	testContent := `
/** @brief Documented class without group */
class TestClass {
public:
    /** @brief Documented method without group */
    void testMethod();
};
`

	doc, err := NewFromContent("test.hpp", testContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	group := &GroupConfig{
		Name:  "testgroup",
		Title: "Test Group",
	}

	ctx := context.Background()
	result, err := service.ProcessEntitiesNeedingGroupUpdate(ctx, doc, group)
	if err != nil {
		t.Fatalf("Failed to process group updates: %v", err)
	}

	if result.EntitiesProcessed == 0 {
		t.Error("Expected some entities to need group updates")
	}
}

func TestShouldSkipEntity(t *testing.T) {
	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	tests := []struct {
		name     string
		entity   *ast.Entity
		expected bool
	}{
		{
			name: "single letter entity",
			entity: &ast.Entity{
				Name: "T",
				Type: ast.EntityClass,
			},
			expected: true,
		},
		{
			name: "std namespace",
			entity: &ast.Entity{
				Name: "std",
				Type: ast.EntityNamespace,
			},
			expected: true,
		},
		{
			name: "normal class",
			entity: &ast.Entity{
				Name: "MyClass",
				Type: ast.EntityClass,
			},
			expected: false,
		},
		{
			name: "local variable",
			entity: &ast.Entity{
				Name: "temp",
				Type: ast.EntityVariable,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ShouldSkipEntity(tt.entity)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for entity %s", tt.expected, result, tt.entity.Name)
			}
		})
	}
}

func TestGetEntityTypeDescription(t *testing.T) {
	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	tests := []struct {
		entityType ast.EntityType
		expected   string
	}{
		{ast.EntityNamespace, "namespace"},
		{ast.EntityClass, "class"},
		{ast.EntityFunction, "function"},
		{ast.EntityMethod, "method"},
		{ast.EntityVariable, "variable"},
	}

	for _, tt := range tests {
		entity := &ast.Entity{Type: tt.entityType}
		result := service.getEntityTypeDescription(entity)
		if result != tt.expected {
			t.Errorf("Expected %s, got %s for type %v", tt.expected, result, tt.entityType)
		}
	}
}

func TestParseGeneratedComment(t *testing.T) {
	mockLLM := &mockLLMService{}
	service := NewDocumentationService(mockLLM)

	commentText := `/**
 * @brief This is a test comment
 * 
 * More detailed description here.
 */`

	comment := service.parseGeneratedComment(commentText)

	if comment.Raw != commentText {
		t.Error("Expected raw comment to be preserved")
	}

	if comment.Brief != "This is a test comment" {
		t.Errorf("Expected brief to be extracted, got: %s", comment.Brief)
	}
}
