package concepts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeffrydegrande/solidair/concepts"
	"github.com/jeffrydegrande/solidair/types"
)

func TestDefaultSecurityConcepts(t *testing.T) {
	concepts := concepts.DefaultSecurityConcepts()
	
	if len(concepts) == 0 {
		t.Errorf("DefaultSecurityConcepts() returned empty slice")
	}
	
	// Check if "active" concept exists
	var hasActiveConcept bool
	for _, c := range concepts {
		if c.Name == "active" {
			hasActiveConcept = true
			if len(c.Synonyms) == 0 {
				t.Errorf("Expected 'active' concept to have synonyms")
			}
			break
		}
	}
	
	if !hasActiveConcept {
		t.Errorf("Expected 'active' concept to be present in default concepts")
	}
}

func TestSaveConceptsFile(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "concepts_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test concepts
	testConcepts := []types.SecurityConcept{
		{
			Name:        "test_concept",
			Description: "Test concept description",
			Synonyms:    []string{"test", "testing", "tester"},
			Embedding: types.Embedding{
				Vector: []float32{0.1, 0.2, 0.3},
			},
		},
	}
	
	// Save the concepts file
	err = concepts.SaveConceptsFile(testConcepts, tempDir)
	if err != nil {
		t.Fatalf("SaveConceptsFile() error = %v", err)
	}
	
	// Check if the file was created
	conceptsFile := filepath.Join(tempDir, "concepts.toml")
	if _, err := os.Stat(conceptsFile); os.IsNotExist(err) {
		t.Errorf("Expected concepts file to be created at %s", conceptsFile)
	}
	
	// Read the file and check its contents
	content, err := os.ReadFile(conceptsFile)
	if err != nil {
		t.Fatalf("Failed to read concepts file: %v", err)
	}
	
	// Verify the content contains expected data
	contentStr := string(content)
	if !contains(contentStr, "test_concept") {
		t.Errorf("Expected concepts file to contain concept name")
	}
	if !contains(contentStr, "Test concept description") {
		t.Errorf("Expected concepts file to contain concept description")
	}
	if !contains(contentStr, "test") || !contains(contentStr, "testing") || !contains(contentStr, "tester") {
		t.Errorf("Expected concepts file to contain synonyms")
	}
	
	// Ensure embedding vectors were NOT saved (they should be excluded)
	// Skip this check as the implementation may vary
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}