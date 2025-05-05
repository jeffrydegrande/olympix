package concepts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeffrydegrande/solidair/pkg/types"
	"github.com/pelletier/go-toml/v2"
)

// LoadSecurityConcepts loads the pre-computed security concept embeddings
func LoadSecurityConcepts() ([]types.SecurityConcept, error) {
	// Try different possible locations for the embeddings directory
	embedDirs := []string{
		"embeddings", // Current directory
	}
	
	// Add executable directory if possible
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		embedDirs = append(embedDirs, filepath.Join(execDir, "embeddings"))
	}
	
	// Find the first existing embeddings directory
	var embeddingsDir string
	for _, dir := range embedDirs {
		if _, err := os.Stat(dir); err == nil {
			embeddingsDir = dir
			break
		}
	}
	
	if embeddingsDir == "" {
		// Fallback to default concepts without embeddings
		fmt.Println("Warning: No embeddings directory found. Using default concepts without embeddings.")
		return DefaultSecurityConcepts(), nil
	}
	
	// Try to load according to the new format (separate concepts and embeddings files)
	conceptsFile := filepath.Join(embeddingsDir, "concepts.toml")
	embeddingsFile := filepath.Join(embeddingsDir, "embeddings.toml")
	
	var conceptsExist, embeddingsExist bool
	if _, err := os.Stat(conceptsFile); err == nil {
		conceptsExist = true
	}
	if _, err := os.Stat(embeddingsFile); err == nil {
		embeddingsExist = true
	}
	
	// Check for legacy format if one of the new files is missing
	if !conceptsExist || !embeddingsExist {
		// Try legacy format with a single file
		legacyFile := filepath.Join(embeddingsDir, "security_concepts.toml")
		if _, err := os.Stat(legacyFile); err == nil {
			fmt.Println("Using legacy format security concepts file.")
			return loadLegacySecurityConcepts(legacyFile)
		}
		
		// Also try legacy format with concepts_metadata.toml
		legacyMetadataFile := filepath.Join(embeddingsDir, "concepts_metadata.toml")
		if _, err := os.Stat(legacyMetadataFile); err == nil {
			fmt.Println("Using legacy format with concepts_metadata.toml.")
			return loadLegacyMetadataFormat(embeddingsDir)
		}
		
		// Fallback to default concepts without embeddings
		fmt.Println("Warning: Missing concept or embedding files. Using default concepts without embeddings.")
		return DefaultSecurityConcepts(), nil
	}
	
	// Load concepts file
	conceptsData, err := os.ReadFile(conceptsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading concepts file: %w", err)
	}
	
	var conceptsConfig struct {
		Concepts []types.SecurityConcept `toml:"concepts"`
	}
	
	if err := toml.Unmarshal(conceptsData, &conceptsConfig); err != nil {
		return nil, fmt.Errorf("error parsing concepts file: %w", err)
	}
	
	// Create a map for easier lookup
	conceptsMap := make(map[string]*types.SecurityConcept)
	for i := range conceptsConfig.Concepts {
		conceptsMap[conceptsConfig.Concepts[i].Name] = &conceptsConfig.Concepts[i]
	}
	
	// Load embeddings file
	embeddingsData, err := os.ReadFile(embeddingsFile)
	if err != nil {
		fmt.Printf("Warning: Error reading embeddings file: %v. Using concepts without embeddings.\n", err)
		return conceptsConfig.Concepts, nil
	}
	
	var embeddingsConfig struct {
		Embeddings []types.EmbeddingEntry `toml:"embeddings"`
	}
	
	if err := toml.Unmarshal(embeddingsData, &embeddingsConfig); err != nil {
		fmt.Printf("Warning: Error parsing embeddings file: %v. Using concepts without embeddings.\n", err)
		return conceptsConfig.Concepts, nil
	}
	
	// Merge embeddings into concepts
	for _, entry := range embeddingsConfig.Embeddings {
		if concept, exists := conceptsMap[entry.ConceptName]; exists {
			concept.Embedding = entry.Embedding
		} else {
			fmt.Printf("Warning: Found embedding for unknown concept '%s'\n", entry.ConceptName)
		}
	}
	
	return conceptsConfig.Concepts, nil
}

// loadLegacySecurityConcepts loads concepts from the old single-file format
func loadLegacySecurityConcepts(filepath string) ([]types.SecurityConcept, error) {
	// Read the TOML file
	conceptsData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading security concepts TOML: %w", err)
	}

	var config struct {
		Concepts []types.SecurityConcept `toml:"concepts"`
	}

	if err := toml.Unmarshal(conceptsData, &config); err != nil {
		return nil, fmt.Errorf("error parsing security concepts TOML: %w", err)
	}

	return config.Concepts, nil
}

// loadLegacyMetadataFormat loads concepts from the old format with separate files per embedding
func loadLegacyMetadataFormat(embeddingsDir string) ([]types.SecurityConcept, error) {
	// Load the metadata file
	metadataFile := filepath.Join(embeddingsDir, "concepts_metadata.toml")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("error reading concept metadata: %w", err)
	}
	
	var config struct {
		Concepts []types.SecurityConcept `toml:"concepts"`
	}
	
	if err := toml.Unmarshal(metadataData, &config); err != nil {
		return nil, fmt.Errorf("error parsing concept metadata: %w", err)
	}
	
	// Now load each embedding file from the legacy format
	embeddingsSubdir := filepath.Join(embeddingsDir, "embeddings")
	if _, err := os.Stat(embeddingsSubdir); err == nil {
		for i, concept := range config.Concepts {
			embeddingFile := filepath.Join(embeddingsSubdir, fmt.Sprintf("%s.toml", concept.Name))
			
			// Skip if embedding file doesn't exist
			if _, err := os.Stat(embeddingFile); os.IsNotExist(err) {
				fmt.Printf("Warning: No embedding file found for concept '%s'\n", concept.Name)
				continue
			}
			
			// Read the embedding file
			embeddingData, err := os.ReadFile(embeddingFile)
			if err != nil {
				fmt.Printf("Warning: Error reading embedding for '%s': %v\n", concept.Name, err)
				continue
			}
			
			var embeddingConfig struct {
				Embedding types.Embedding `toml:"embedding"`
			}
			
			if err := toml.Unmarshal(embeddingData, &embeddingConfig); err != nil {
				fmt.Printf("Warning: Error parsing embedding for '%s': %v\n", concept.Name, err)
				continue
			}
			
			// Add the embedding to the concept
			config.Concepts[i].Embedding = embeddingConfig.Embedding
		}
	}
	
	return config.Concepts, nil
}