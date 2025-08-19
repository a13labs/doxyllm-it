package doxygen

import (
	"context"
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/llm"
)

// mockLLMService implements a mock LLM service for testing
type mockLLMService struct {
	generateFunc func(ctx context.Context, req llm.LLMRequest) (*llm.LLMResponse, error)
}

func (m *mockLLMService) Generate(ctx context.Context, req llm.LLMRequest) (*llm.LLMResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &llm.LLMResponse{
		Response: "Mock description",
		Metadata: make(map[string]string),
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
		generateFunc: func(ctx context.Context, req llm.LLMRequest) (*llm.LLMResponse, error) {
			return &llm.LLMResponse{
				Response: "Generated description",
				Metadata: make(map[string]string),
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
		ExcludeTypes: []ast.EntityType{ast.EntityName},
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
