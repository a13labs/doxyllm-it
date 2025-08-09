package ast

import (
	"testing"
)

func TestDoxygenCommentParamsAndTParams(t *testing.T) {
	tests := []struct {
		name           string
		comment        *DoxygenComment
		expectedParams map[string]string
		expectedTParams map[string]string
	}{
		{
			name: "Empty comment",
			comment: &DoxygenComment{
				Params:  make(map[string]string),
				TParams: make(map[string]string),
			},
			expectedParams:  map[string]string{},
			expectedTParams: map[string]string{},
		},
		{
			name: "Function with params only",
			comment: &DoxygenComment{
				Params: map[string]string{
					"x": "The x coordinate",
					"y": "The y coordinate",
				},
				TParams: make(map[string]string),
			},
			expectedParams: map[string]string{
				"x": "The x coordinate",
				"y": "The y coordinate",
			},
			expectedTParams: map[string]string{},
		},
		{
			name: "Template with tparams only",
			comment: &DoxygenComment{
				Params: make(map[string]string),
				TParams: map[string]string{
					"T": "The type of elements in the container",
					"U": "The type of the allocator",
				},
			},
			expectedParams: map[string]string{},
			expectedTParams: map[string]string{
				"T": "The type of elements in the container",
				"U": "The type of the allocator",
			},
		},
		{
			name: "Template function with both params and tparams",
			comment: &DoxygenComment{
				Params: map[string]string{
					"value": "The value to insert",
					"index": "The index where to insert",
				},
				TParams: map[string]string{
					"T": "The type of elements",
					"Allocator": "The allocator type",
				},
			},
			expectedParams: map[string]string{
				"value": "The value to insert",
				"index": "The index where to insert",
			},
			expectedTParams: map[string]string{
				"T": "The type of elements",
				"Allocator": "The allocator type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify Params
			if len(tt.comment.Params) != len(tt.expectedParams) {
				t.Errorf("Params length mismatch: got %d, want %d", len(tt.comment.Params), len(tt.expectedParams))
			}
			for key, value := range tt.expectedParams {
				if got, exists := tt.comment.Params[key]; !exists || got != value {
					t.Errorf("Params[%q] = %q, want %q", key, got, value)
				}
			}

			// Verify TParams
			if len(tt.comment.TParams) != len(tt.expectedTParams) {
				t.Errorf("TParams length mismatch: got %d, want %d", len(tt.comment.TParams), len(tt.expectedTParams))
			}
			for key, value := range tt.expectedTParams {
				if got, exists := tt.comment.TParams[key]; !exists || got != value {
					t.Errorf("TParams[%q] = %q, want %q", key, got, value)
				}
			}
		})
	}
}

func TestDoxygenCommentInitialization(t *testing.T) {
	comment := &DoxygenComment{
		Params:  make(map[string]string),
		TParams: make(map[string]string),
	}

	// Verify maps are properly initialized
	if comment.Params == nil {
		t.Error("Params map not initialized")
	}
	if comment.TParams == nil {
		t.Error("TParams map not initialized")
	}

	// Test that we can add to both maps without nil pointer errors
	comment.Params["test_param"] = "test description"
	comment.TParams["test_tparam"] = "test template description"

	if comment.Params["test_param"] != "test description" {
		t.Error("Failed to add to Params map")
	}
	if comment.TParams["test_tparam"] != "test template description" {
		t.Error("Failed to add to TParams map")
	}
}
