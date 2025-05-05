package templates_test

import (
	"reflect"
	"testing"

	"github.com/jeffrydegrande/solidair/templates"
	"github.com/jeffrydegrande/solidair/types"
)

func TestParseQueryTemplate(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		source      string
		want        *templates.QueryTemplate
		wantErr     bool
	}{
		{
			name: "Simple template with no parameters",
			queryString: `; Name: Test Query
; Description: A test query with no parameters
; Concepts: active, locked
(function_item
  name: (identifier) @func_name)`,
			source: "test.scm",
			want: &templates.QueryTemplate{
				Name:        "Test Query",
				Description: "A test query with no parameters",
				Concepts:    []string{"active", "locked"},
				Source:      "test.scm",
				Original:    `; Name: Test Query
; Description: A test query with no parameters
; Concepts: active, locked
(function_item
  name: (identifier) @func_name)`,
				Parameters: map[string]struct{}{},
			},
			wantErr: false,
		},
		{
			name: "Template with parameters",
			queryString: `; Name: Parameterized Query
; Description: A query with parameters
; Concepts: active
(function_item
  name: (identifier) @func_name
  (#match? @func_name "${active}")
  body: (block) @func_body)`,
			source: "param.scm",
			want: &templates.QueryTemplate{
				Name:        "Parameterized Query",
				Description: "A query with parameters",
				Concepts:    []string{"active"},
				Source:      "param.scm",
				Original: `; Name: Parameterized Query
; Description: A query with parameters
; Concepts: active
(function_item
  name: (identifier) @func_name
  (#match? @func_name "${active}")
  body: (block) @func_body)`,
				Parameters: map[string]struct{}{
					"active": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Template with multiple parameters",
			queryString: `; Name: Multi-Param Query
; Description: A query with multiple parameters
; Concepts: active, locked
(function_item
  name: (identifier) @func_name
  (#match? @func_name "${active}")
  body: (block
    (if_statement
      condition: (binary_expression
        left: (identifier) @var_name
        (#eq? @var_name "${locked}")))))`,
			source: "multi_param.scm",
			want: &templates.QueryTemplate{
				Name:        "Multi-Param Query",
				Description: "A query with multiple parameters",
				Concepts:    []string{"active", "locked"},
				Source:      "multi_param.scm",
				Original: `; Name: Multi-Param Query
; Description: A query with multiple parameters
; Concepts: active, locked
(function_item
  name: (identifier) @func_name
  (#match? @func_name "${active}")
  body: (block
    (if_statement
      condition: (binary_expression
        left: (identifier) @var_name
        (#eq? @var_name "${locked}")))))`,
				Parameters: map[string]struct{}{
					"active": {},
					"locked": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Template with incomplete metadata",
			queryString: `; Name: Incomplete Query
(function_item
  name: (identifier) @func_name)`,
			source: "incomplete.scm",
			want: &templates.QueryTemplate{
				Name:        "Incomplete Query",
				Description: "",
				Concepts:    nil,
				Source:      "incomplete.scm",
				Original: `; Name: Incomplete Query
(function_item
  name: (identifier) @func_name)`,
				Parameters: map[string]struct{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := templates.ParseQueryTemplate(tt.queryString, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQueryTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil {
				return
			}
			
			// Compare fields
			if got.Name != tt.want.Name {
				t.Errorf("ParseQueryTemplate().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("ParseQueryTemplate().Description = %v, want %v", got.Description, tt.want.Description)
			}
			if got.Source != tt.want.Source {
				t.Errorf("ParseQueryTemplate().Source = %v, want %v", got.Source, tt.want.Source)
			}
			if got.Original != tt.want.Original {
				t.Errorf("ParseQueryTemplate().Original = %v, want %v", got.Original, tt.want.Original)
			}
			
			// Compare concepts (order matters)
			if !reflect.DeepEqual(got.Concepts, tt.want.Concepts) {
				t.Errorf("ParseQueryTemplate().Concepts = %v, want %v", got.Concepts, tt.want.Concepts)
			}
			
			// Compare parameters (as sets)
			if len(got.Parameters) != len(tt.want.Parameters) {
				t.Errorf("ParseQueryTemplate().Parameters = %v, want %v", got.Parameters, tt.want.Parameters)
			}
			for param := range tt.want.Parameters {
				if _, exists := got.Parameters[param]; !exists {
					t.Errorf("ParseQueryTemplate().Parameters missing %v", param)
				}
			}
		})
	}
}

func TestSubstituteParameters(t *testing.T) {
	tests := []struct {
		name          string
		template      *templates.QueryTemplate
		conceptMatches map[string][]types.ConceptMatch
		want          *templates.ParameterizedQuery
		wantErr       bool
	}{
		{
			name: "Simple substitution",
			template: &templates.QueryTemplate{
				Name:        "Test Query",
				Description: "A test query",
				Concepts:    []string{"active"},
				Source:      "test.scm",
				Original:    `(function (#match? @name "${active}"))`,
				Parameters: map[string]struct{}{
					"active": {},
				},
			},
			conceptMatches: map[string][]types.ConceptMatch{
				"active": {
					{
						Variable: types.VariableInfo{
							Name: "is_active",
						},
						Concept:         "active",
						SimilarityScore: 0.9,
					},
				},
			},
			want: &templates.ParameterizedQuery{
				ProcessedQuery: `(function (#match? @name "is_active"))`,
				Parameters: map[string]string{
					"active": "is_active",
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple substitutions",
			template: &templates.QueryTemplate{
				Name:        "Test Query",
				Description: "A test query",
				Concepts:    []string{"active", "locked"},
				Source:      "test.scm",
				Original:    `(function (#match? @name "${active}") (#eq? @var "${locked}"))`,
				Parameters: map[string]struct{}{
					"active": {},
					"locked": {},
				},
			},
			conceptMatches: map[string][]types.ConceptMatch{
				"active": {
					{
						Variable: types.VariableInfo{
							Name: "is_active",
						},
						Concept:         "active",
						SimilarityScore: 0.9,
					},
				},
				"locked": {
					{
						Variable: types.VariableInfo{
							Name: "reentrancy_guard",
						},
						Concept:         "locked",
						SimilarityScore: 0.8,
					},
				},
			},
			want: &templates.ParameterizedQuery{
				ProcessedQuery: `(function (#match? @name "is_active") (#eq? @var "reentrancy_guard"))`,
				Parameters: map[string]string{
					"active": "is_active",
					"locked": "reentrancy_guard",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing concept match",
			template: &templates.QueryTemplate{
				Name:        "Test Query",
				Description: "A test query",
				Concepts:    []string{"active", "locked"},
				Source:      "test.scm",
				Original:    `(function (#match? @name "${active}") (#eq? @var "${locked}"))`,
				Parameters: map[string]struct{}{
					"active": {},
					"locked": {},
				},
			},
			conceptMatches: map[string][]types.ConceptMatch{
				"active": {
					{
						Variable: types.VariableInfo{
							Name: "is_active",
						},
						Concept:         "active",
						SimilarityScore: 0.9,
					},
				},
				// Missing "locked" concept
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := templates.SubstituteParameters(tt.template, tt.conceptMatches)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubstituteParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			// Set the template field (can't compare directly)
			got.Template = nil
			
			// Check the processed query
			if got.ProcessedQuery != tt.want.ProcessedQuery {
				t.Errorf("SubstituteParameters().ProcessedQuery = %v, want %v", got.ProcessedQuery, tt.want.ProcessedQuery)
			}
			
			// Check parameters
			if !reflect.DeepEqual(got.Parameters, tt.want.Parameters) {
				t.Errorf("SubstituteParameters().Parameters = %v, want %v", got.Parameters, tt.want.Parameters)
			}
		})
	}
}

func TestProcessTemplatedQueries(t *testing.T) {
	tests := []struct {
		name          string
		templates     map[string]*templates.QueryTemplate
		conceptMatches map[string][]types.ConceptMatch
		wantCount     int // Number of processed queries expected
	}{
		{
			name: "Process multiple templates",
			templates: map[string]*templates.QueryTemplate{
				"template1": {
					Name:     "Template 1",
					Concepts: []string{"active"},
					Original: `(function (#match? @name "${active}"))`,
					Parameters: map[string]struct{}{
						"active": {},
					},
				},
				"template2": {
					Name:     "Template 2",
					Concepts: []string{"locked"},
					Original: `(function (#eq? @var "${locked}"))`,
					Parameters: map[string]struct{}{
						"locked": {},
					},
				},
				"template3": {
					Name:     "Template 3",
					Concepts: []string{"missing"},
					Original: `(function (#eq? @var "${missing}"))`,
					Parameters: map[string]struct{}{
						"missing": {},
					},
				},
			},
			conceptMatches: map[string][]types.ConceptMatch{
				"active": {
					{
						Variable: types.VariableInfo{
							Name: "is_active",
						},
						Concept:         "active",
						SimilarityScore: 0.9,
					},
				},
				"locked": {
					{
						Variable: types.VariableInfo{
							Name: "reentrancy_guard",
						},
						Concept:         "locked",
						SimilarityScore: 0.8,
					},
				},
				// Missing "missing" concept
			},
			wantCount: 2, // Should process 2 out of 3 templates
		},
		{
			name: "Template with no concepts",
			templates: map[string]*templates.QueryTemplate{
				"template1": {
					Name:     "Template 1",
					Concepts: []string{}, // No concepts
					Original: `(function)`,
					Parameters: map[string]struct{}{},
				},
			},
			conceptMatches: map[string][]types.ConceptMatch{},
			wantCount:     0, // Should skip template with no concepts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templates.ProcessTemplatedQueries(tt.templates, tt.conceptMatches)
			
			if len(got) != tt.wantCount {
				t.Errorf("ProcessTemplatedQueries() returned %v queries, want %v", len(got), tt.wantCount)
			}
			
			// Verify each returned query has a valid processed query
			for _, query := range got {
				if query.ProcessedQuery == "" {
					t.Errorf("ProcessTemplatedQueries() returned query with empty processed query")
				}
				
				// Check that all parameters were substituted
				for paramName, paramValue := range query.Parameters {
					if paramValue == "" {
						t.Errorf("ProcessTemplatedQueries() returned query with empty parameter value for %s", paramName)
					}
				}
			}
		})
	}
}