package parser_new

import (
	"testing"
)

// TestRegressionTParamVsParam tests the specific issue where @tparam was incorrectly
// parsed as @param, leading to wrong documentation format
// NOTE: This test is now in the Document layer since Doxygen parsing moved there
func TestRegressionTParamVsParam(t *testing.T) {
	t.Skip("This test has been moved to the Document layer - Doxygen parsing is no longer in the Parser layer")
}

// TestFormatterPreservesCorrectTParamFormat tests that the formatter
// outputs the correct @tparam tags, not @param
// NOTE: This test is covered by the formatter package tests
func TestFormatterPreservesCorrectTParamFormat(t *testing.T) {
	t.Skip("This functionality is tested in pkg/formatter/formatter_tparam_test.go - not a parser test")
}
