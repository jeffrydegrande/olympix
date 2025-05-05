package concepts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeffrydegrande/solidair/concepts"
	"github.com/jeffrydegrande/solidair/types"
	"github.com/pelletier/go-toml/v2"
)

func TestLoadSecurityConcepts(t *testing.T) {
	// Setup: Create a temporary directory with test concepts files
	tempDir, err := os.MkdirTemp("", "embeddings_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save the current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create embeddings directory in the temp dir
	embeddingsDir := filepath.Join(tempDir, "embeddings")
	if err := os.MkdirAll(embeddingsDir, 0755); err != nil {
		t.Fatalf("Failed to create embeddings dir: %v", err)
	}

	// Create test concept data
	testConcepts := []types.SecurityConcept{
		{
			Name:        "test_concept",
			Description: "Test concept description",
			Synonyms:    []string{"test", "testing", "tester"},
		},
	}

	// Create test embeddings
	testEmbeddings := []types.EmbeddingEntry{
		{
			ConceptName: "test_concept",
			Embedding: types.Embedding{
				Vector: []float32{0.1, 0.2, 0.3},
			},
		},
	}

	// Write the concepts file
	conceptsFile := filepath.Join(embeddingsDir, "concepts.toml")
	conceptsConfig := struct {
		Concepts []types.SecurityConcept `toml:"concepts"`
	}{
		Concepts: testConcepts,
	}
	conceptsData, err := os.Create(conceptsFile)
	if err != nil {
		t.Fatalf("Failed to create concepts file: %v", err)
	}
	if err := toml.NewEncoder(conceptsData).Encode(conceptsConfig); err != nil {
		conceptsData.Close()
		t.Fatalf("Failed to write concepts data: %v", err)
	}
	conceptsData.Close()

	// Write the embeddings file
	embeddingsFile := filepath.Join(embeddingsDir, "embeddings.toml")
	embeddingsConfig := struct {
		Embeddings []types.EmbeddingEntry `toml:"embeddings"`
	}{
		Embeddings: testEmbeddings,
	}
	embeddingsData, err := os.Create(embeddingsFile)
	if err != nil {
		t.Fatalf("Failed to create embeddings file: %v", err)
	}
	if err := toml.NewEncoder(embeddingsData).Encode(embeddingsConfig); err != nil {
		embeddingsData.Close()
		t.Fatalf("Failed to write embeddings data: %v", err)
	}
	embeddingsData.Close()

	// Create a symlink to make the embeddings dir visible from current directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original directory

	// Test loading concepts with fixtures in place
	loadedConcepts, err := concepts.LoadSecurityConcepts()
	if err != nil {
		t.Fatalf("LoadSecurityConcepts() error = %v", err)
	}

	// Verify the loaded concepts
	if len(loadedConcepts) == 0 {
		t.Errorf("Expected to load concepts, got empty slice")
	}
	
	// Check for our test concept
	var foundConcept bool
	for _, c := range loadedConcepts {
		if c.Name == "test_concept" {
			foundConcept = true
			
			// Check that the embedding was loaded
			if len(c.Embedding.Vector) != 3 {
				t.Errorf("Expected embedding vector to be loaded, got %v", c.Embedding.Vector)
			}
			
			// Check the vector values
			if c.Embedding.Vector[0] != 0.1 || c.Embedding.Vector[1] != 0.2 || c.Embedding.Vector[2] != 0.3 {
				t.Errorf("Expected specific embedding values, got %v", c.Embedding.Vector)
			}
			
			break
		}
	}
	
	if !foundConcept {
		t.Errorf("Expected to find test_concept in loaded concepts")
	}
}

// TestLoadSecurityConceptsFallback tests that the default concepts are loaded when no files are found
func TestLoadSecurityConceptsFallback(t *testing.T) {
	// Setup a clean directory with no concept files
	tempDir, err := os.MkdirTemp("", "embeddings_test_fallback")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save the current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to the empty directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original directory

	// Test loading concepts when no files are present (should fall back to defaults)
	loadedConcepts, err := concepts.LoadSecurityConcepts()
	if err != nil {
		t.Fatalf("LoadSecurityConcepts() error = %v", err)
	}

	// Verify default concepts are returned
	if len(loadedConcepts) == 0 {
		t.Errorf("Expected default concepts, got empty slice")
	}
	
	// Check for "active" concept which is part of defaults
	var hasActiveConcept bool
	for _, c := range loadedConcepts {
		if c.Name == "active" {
			hasActiveConcept = true
			break
		}
	}
	
	if !hasActiveConcept {
		t.Errorf("Expected to find 'active' concept in default concepts")
	}
}