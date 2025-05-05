package embedding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeffrydegrande/solidair/embedding"
	"github.com/jeffrydegrande/solidair/types"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        types.Embedding
		b        types.Embedding
		expected float32
	}{
		{
			name: "Identical vectors",
			a: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.0},
			},
			b: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.0},
			},
			expected: 1.0, // Identical vectors have cosine similarity of 1
		},
		{
			name: "Orthogonal vectors",
			a: types.Embedding{
				Vector: []float32{1.0, 0.0, 0.0},
			},
			b: types.Embedding{
				Vector: []float32{0.0, 1.0, 0.0},
			},
			expected: 0.0, // Orthogonal vectors have cosine similarity of 0
		},
		{
			name: "Opposite vectors",
			a: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.0},
			},
			b: types.Embedding{
				Vector: []float32{-1.0, -2.0, -3.0},
			},
			expected: -1.0, // Opposite vectors have cosine similarity of -1
		},
		{
			name: "Similar vectors",
			a: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.0},
			},
			b: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.5},
			},
			expected: 0.9978, // Approximate value
		},
		{
			name: "Empty vectors",
			a: types.Embedding{
				Vector: []float32{},
			},
			b: types.Embedding{
				Vector: []float32{},
			},
			expected: 0.0, // Empty vectors should return 0
		},
		{
			name: "Different length vectors",
			a: types.Embedding{
				Vector: []float32{1.0, 2.0, 3.0},
			},
			b: types.Embedding{
				Vector: []float32{1.0, 2.0},
			},
			expected: 0.0, // Different length vectors should return 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embedding.CosineSimilarity(tt.a, tt.b)
			
			// For approximate equality
			if tt.name == "Similar vectors" {
				if result < 0.99 || result > 1.0 {
					t.Errorf("CosineSimilarity() = %v, expected approximately %v", result, tt.expected)
				}
			} else {
				// For exact equality (allowing for float precision issues)
				if abs(result-tt.expected) > 0.0001 {
					t.Errorf("CosineSimilarity() = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestSaveLoadEmbeddingsFile(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "embeddings_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test embeddings
	testEmbeddings := []types.EmbeddingEntry{
		{
			ConceptName: "test_concept",
			Embedding: types.Embedding{
				Vector: []float32{0.1, 0.2, 0.3},
			},
		},
		{
			ConceptName: "another_concept",
			Embedding: types.Embedding{
				Vector: []float32{0.4, 0.5, 0.6},
			},
		},
	}
	
	// Save the embeddings file
	err = embedding.SaveEmbeddingsFile(testEmbeddings, tempDir)
	if err != nil {
		t.Fatalf("SaveEmbeddingsFile() error = %v", err)
	}
	
	// Check if the file was created
	embeddingsFile := filepath.Join(tempDir, "embeddings.toml")
	if _, err := os.Stat(embeddingsFile); os.IsNotExist(err) {
		t.Errorf("Expected embeddings file to be created at %s", embeddingsFile)
	}
	
	// Load the embeddings back
	loadedEmbeddings, err := embedding.LoadEmbeddingsFile(embeddingsFile)
	if err != nil {
		t.Fatalf("LoadEmbeddingsFile() error = %v", err)
	}
	
	// Verify the loaded embeddings
	if len(loadedEmbeddings) != len(testEmbeddings) {
		t.Errorf("Expected %d embeddings, got %d", len(testEmbeddings), len(loadedEmbeddings))
	}
	
	// Check each embedding
	for i, entry := range loadedEmbeddings {
		if entry.ConceptName != testEmbeddings[i].ConceptName {
			t.Errorf("Expected concept name %s, got %s", testEmbeddings[i].ConceptName, entry.ConceptName)
		}
		
		// Check vector length
		if len(entry.Embedding.Vector) != len(testEmbeddings[i].Embedding.Vector) {
			t.Errorf("Expected vector length %d, got %d", len(testEmbeddings[i].Embedding.Vector), len(entry.Embedding.Vector))
		}
		
		// Check vector values
		for j, val := range entry.Embedding.Vector {
			if val != testEmbeddings[i].Embedding.Vector[j] {
				t.Errorf("Expected vector[%d] = %f, got %f", j, testEmbeddings[i].Embedding.Vector[j], val)
			}
		}
	}
}

func TestGetAPIKey(t *testing.T) {
	// Save existing API key
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)
	
	// Set a test API key
	testKey := "test_api_key_123"
	os.Setenv("OPENAI_API_KEY", testKey)
	
	// Test getting the API key
	key := embedding.GetAPIKey()
	if key != testKey {
		t.Errorf("GetAPIKey() = %v, expected %v", key, testKey)
	}
}

func TestStringHelpers(t *testing.T) {
	// Test the string helper functions
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected bool
	}{
		{
			name:     "Contains - identical strings",
			s1:       "active",
			s2:       "active",
			expected: true,
		},
		{
			name:     "Contains - substring",
			s1:       "is_active_flag",
			s2:       "active",
			expected: true,
		},
		{
			name:     "ContainsIgnoreCase",
			s1:       "IsActive",
			s2:       "active",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embedding.Contains(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("Contains() = %v, expected %v", result, tt.expected)
			}
			
			// Also test the case-insensitive version
			result = embedding.ContainsIgnoreCase(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("ContainsIgnoreCase() = %v, expected %v", result, tt.expected)
			}
		})
	}
	
	// Test NGram similarity
	sim := embedding.CalculateNGramSimilarity("active", "inactive")
	if sim < 0 || sim > 1 {
		t.Errorf("CalculateNGramSimilarity() = %v, expected value between 0 and 1", sim)
	}
}

// Helper function to calculate absolute value of float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}