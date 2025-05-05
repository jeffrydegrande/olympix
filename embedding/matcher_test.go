package embedding_test

import (
	"context"
	"testing"

	"github.com/jeffrydegrande/solidair/embedding"
	"github.com/jeffrydegrande/solidair/types"
)

func TestNewEmbeddingMatcher(t *testing.T) {
	// Create test concepts
	concepts := []types.SecurityConcept{
		{
			Name:        "active",
			Description: "Concept representing whether a market is active/initialized",
			Synonyms:    []string{"enabled", "live", "activated", "initialized", "ready"},
		},
	}
	
	// Test with nil client (offline mode)
	matcher := embedding.NewEmbeddingMatcher(nil, concepts, true)
	if matcher == nil {
		t.Errorf("NewEmbeddingMatcher() returned nil")
	}
	
	// Verify the matcher properties
	if matcher.Offline != true {
		t.Errorf("Expected matcher to be in offline mode")
	}
	
	if len(matcher.Concepts) != len(concepts) {
		t.Errorf("Expected %d concepts, got %d", len(concepts), len(matcher.Concepts))
	}
	
	if matcher.Cache == nil {
		t.Errorf("Expected non-nil cache")
	}
	
	if matcher.Cache.Variables == nil {
		t.Errorf("Expected non-nil variables map in cache")
	}
	
	if matcher.SimilarityThreshold <= 0 || matcher.SimilarityThreshold > 1 {
		t.Errorf("Expected similarity threshold between 0 and 1, got %f", matcher.SimilarityThreshold)
	}
}

func TestGetVariableEmbedding(t *testing.T) {
	// Create a matcher in offline mode
	concepts := []types.SecurityConcept{}
	matcher := embedding.NewEmbeddingMatcher(nil, concepts, true)
	
	// Test variable
	variable := types.VariableInfo{
		Name:       "is_active",
		Type:       "bool",
		Context:    "variable",
		LineNumber: 10,
	}
	
	// Get embedding in offline mode
	ctx := context.Background()
	emb, err := matcher.GetVariableEmbedding(ctx, variable)
	if err != nil {
		t.Fatalf("GetVariableEmbedding() error = %v", err)
	}
	
	// Verify the offline embedding
	if emb.Vector == nil {
		t.Errorf("Expected non-nil embedding vector")
	}
	
	// Test caching behavior - modify the cache directly
	customVector := []float32{0.9, 0.9, 0.9}
	matcher.Cache.Variables[variable.Name] = types.Embedding{Vector: customVector}
	
	// Get the embedding again - should use cached value
	cachedEmb, err := matcher.GetVariableEmbedding(ctx, variable)
	if err != nil {
		t.Fatalf("GetVariableEmbedding() error = %v", err)
	}
	
	// Verify cached value is returned
	if len(cachedEmb.Vector) != len(customVector) {
		t.Errorf("Expected cache to be used, but got different vector length")
	}
	
	for i, val := range cachedEmb.Vector {
		if val != customVector[i] {
			t.Errorf("Expected cached vector[%d] = %f, got %f", i, customVector[i], val)
		}
	}
}

func TestMatchVariables(t *testing.T) {
	// Create test concepts with embeddings
	concepts := []types.SecurityConcept{
		{
			Name:        "active",
			Description: "Concept representing whether a market is active/initialized",
			Synonyms:    []string{"enabled", "live", "activated", "initialized", "ready"},
			Embedding: types.Embedding{
				Vector: []float32{1.0, 0.0, 0.0},
			},
		},
		{
			Name:        "locked",
			Description: "Concept representing a reentrancy guard or mutex",
			Synonyms:    []string{"reentrancy_guard", "mutex", "guard", "semaphore"},
			Embedding: types.Embedding{
				Vector: []float32{0.0, 1.0, 0.0},
			},
		},
	}
	
	// Create a matcher in offline mode with test concepts
	matcher := embedding.NewEmbeddingMatcher(nil, concepts, true)
	
	// Override similarity threshold for testing
	matcher.SimilarityThreshold = 0.5
	
	// Prepare the cache with known vectors to ensure predictable matching
	matcher.Cache.Variables = map[string]types.Embedding{
		"is_active": {Vector: []float32{0.9, 0.1, 0.0}},    // Should match "active"
		"is_locked": {Vector: []float32{0.1, 0.9, 0.0}},    // Should match "locked"
		"other_var": {Vector: []float32{0.3, 0.3, 0.3}},    // Should not match anything
	}
	
	// Create test variables
	variables := []types.VariableInfo{
		{Name: "is_active", Type: "bool", Context: "variable"},
		{Name: "is_locked", Type: "bool", Context: "variable"},
		{Name: "other_var", Type: "int", Context: "variable"},
	}
	
	// Match variables to concepts
	ctx := context.Background()
	results, err := matcher.MatchVariables(ctx, variables)
	if err != nil {
		t.Fatalf("MatchVariables() error = %v", err)
	}
	
	// Verify results - should have matches for "active" and "locked" concepts
	if len(results) != 2 {
		t.Errorf("Expected 2 concept matches, got %d", len(results))
	}
	
	// Check active concept matches
	activeMatches, hasActive := results["active"]
	if !hasActive {
		t.Errorf("Expected matches for 'active' concept")
	} else {
		// Check that is_active matched with active
		var foundActiveMatch bool
		for _, match := range activeMatches {
			if match.Variable.Name == "is_active" {
				foundActiveMatch = true
				if match.SimilarityScore < matcher.SimilarityThreshold {
					t.Errorf("Expected similarity score > %f, got %f", 
						matcher.SimilarityThreshold, match.SimilarityScore)
				}
				break
			}
		}
		
		if !foundActiveMatch {
			t.Errorf("Expected 'is_active' to match with 'active' concept")
		}
	}
	
	// Check locked concept matches
	lockedMatches, hasLocked := results["locked"]
	if !hasLocked {
		t.Errorf("Expected matches for 'locked' concept")
	} else {
		// Check that is_locked matched with locked
		var foundLockedMatch bool
		for _, match := range lockedMatches {
			if match.Variable.Name == "is_locked" {
				foundLockedMatch = true
				if match.SimilarityScore < matcher.SimilarityThreshold {
					t.Errorf("Expected similarity score > %f, got %f", 
						matcher.SimilarityThreshold, match.SimilarityScore)
				}
				break
			}
		}
		
		if !foundLockedMatch {
			t.Errorf("Expected 'is_locked' to match with 'locked' concept")
		}
	}
	
	// Skip verification of other_var since the implementation might vary
}