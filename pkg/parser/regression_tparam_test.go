package parser

import (
	"testing"
)

// TestRegressionTParamVsParam tests the specific issue where @tparam was incorrectly
// parsed as @param, leading to wrong documentation format
func TestRegressionTParamVsParam(t *testing.T) {
	// This is the exact scenario described in the issue
	originalComment := `/**
 * @brief Alias for std::vector<T>.
 * @tparam T The type of elements in the list.
 */`

	// Parse the comment
	result := ParseDoxygenComment(originalComment)
	if result == nil {
		t.Fatal("ParseDoxygenComment returned nil")
	}

	// Verify the regression is fixed: @tparam should NOT be in Params
	if len(result.Params) != 0 {
		t.Errorf("REGRESSION: Found %d params, should be 0. Params: %v", len(result.Params), result.Params)
	}

	// Verify @tparam is correctly in TParams
	if len(result.TParams) != 1 {
		t.Errorf("Expected 1 tparam, got %d. TParams: %v", len(result.TParams), result.TParams)
	}

	// Verify the specific tparam value
	if value, exists := result.TParams["T"]; !exists {
		t.Error("TParam 'T' not found")
	} else if value != "The type of elements in the list." {
		t.Errorf("TParams['T'] = %q, want %q", value, "The type of elements in the list.")
	}

	// Verify brief is correct
	if result.Brief != "Alias for std::vector<T>." {
		t.Errorf("Brief = %q, want %q", result.Brief, "Alias for std::vector<T>.")
	}
}

// TestFormatterPreservesCorrectTParamFormat tests that the formatter 
// outputs the correct @tparam tags, not @param
// NOTE: This test is covered by the formatter package tests
func TestFormatterPreservesCorrectTParamFormat(t *testing.T) {
	// This functionality is tested in pkg/formatter/formatter_tparam_test.go
	// The main regression test above ensures the parser correctly separates
	// @param and @tparam, which is the core fix for the reported issue.
	t.Skip("This functionality is comprehensively tested in pkg/formatter/formatter_tparam_test.go")
}
