package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
	"github.com/sashabaranov/go-openai"
)

// Use os.ReadFile instead of embed for flexibility
// We'll read from the absolute path

// Embedding represents a vector embedding for a variable or concept
type Embedding struct {
	Vector []float32 `toml:"vector"` // The embedding vector
}

// SecurityConcept represents a security-related concept with its embedding
type SecurityConcept struct {
	Name        string    `toml:"name"`        // Concept name (e.g., "active", "initialized")
	Description string    `toml:"description"` // Description of what this concept represents
	Synonyms    []string  `toml:"synonyms"`    // Synonyms for this concept
	Embedding   Embedding `toml:"embedding"`   // Pre-computed embedding for this concept
}

// ConceptMatch represents a match between a variable and a security concept
type ConceptMatch struct {
	Variable        VariableInfo // The matched variable
	Concept         string       // The security concept (e.g., "active", "initialized")
	SimilarityScore float32      // 0.0-1.0 score of the match quality
}

// EmbeddingCache provides caching for computed embeddings
type EmbeddingCache struct {
	Variables map[string]Embedding // Cache of variable embeddings
}

// OpenAIClient represents a client for the OpenAI API
type OpenAIClient struct {
	Client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client with the provided API key
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		Client: openai.NewClient(apiKey),
	}
}

// GetEmbedding calculates an embedding for the given text using OpenAI's API
func (c *OpenAIClient) GetEmbedding(ctx context.Context, text string) (Embedding, error) {
	resp, err := c.Client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return Embedding{}, fmt.Errorf("error getting embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return Embedding{}, fmt.Errorf("no embedding data returned")
	}

	// Convert from []float64 to []float32 to save memory
	vector := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		vector[i] = float32(v)
	}

	return Embedding{Vector: vector}, nil
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(a, b Embedding) float32 {
	// Early check for empty vectors
	if len(a.Vector) == 0 || len(b.Vector) == 0 {
		return 0
	}

	// Vectors should be the same length
	if len(a.Vector) != len(b.Vector) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a.Vector); i++ {
		dotProduct += a.Vector[i] * b.Vector[i]
		normA += a.Vector[i] * a.Vector[i]
		normB += b.Vector[i] * b.Vector[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// EmbeddingMatcher is a system for matching variables to security concepts
type EmbeddingMatcher struct {
	OpenAI              *OpenAIClient
	Concepts            []SecurityConcept
	Cache               *EmbeddingCache
	SimilarityThreshold float32
	Offline             bool
}

// NewEmbeddingMatcher creates a new matcher with the provided OpenAI client and concepts
func NewEmbeddingMatcher(client *OpenAIClient, concepts []SecurityConcept, offline bool) *EmbeddingMatcher {
	return &EmbeddingMatcher{
		OpenAI:              client,
		Concepts:            concepts,
		Cache:               &EmbeddingCache{Variables: make(map[string]Embedding)},
		SimilarityThreshold: 0.7, // Default threshold
		Offline:             offline,
	}
}

// GetVariableEmbedding gets the embedding for a variable, using cache if available
func (m *EmbeddingMatcher) GetVariableEmbedding(ctx context.Context, variable VariableInfo) (Embedding, error) {
	// Check if we have a cached embedding
	if embedding, ok := m.Cache.Variables[variable.Name]; ok {
		return embedding, nil
	}

	// If we're in offline mode, use a simple fallback method
	if m.Offline {
		return m.getOfflineEmbedding(variable.Name), nil
	}

	// Get embedding from OpenAI
	embedding, err := m.OpenAI.GetEmbedding(ctx, variable.Name)
	if err != nil {
		return Embedding{}, err
	}

	// Cache the embedding
	m.Cache.Variables[variable.Name] = embedding

	return embedding, nil
}

// getOfflineEmbedding creates a simple embedding for offline mode
// This is a placeholder - in a real implementation, we'd use a more
// sophisticated method for generating offline embeddings
func (m *EmbeddingMatcher) getOfflineEmbedding(name string) Embedding {
	// Create a simple embedding based on string characteristics
	// This is just a placeholder that creates a vector with a few dimensions
	vector := make([]float32, 3)

	// Fill with some values based on the string
	for i := 0; i < len(vector); i++ {
		if i < len(name) {
			vector[i] = float32(name[i%len(name)]) / 255.0
		} else {
			vector[i] = 0
		}
	}

	return Embedding{Vector: vector}
}

// MatchVariable finds the best matching security concept for a variable
func (m *EmbeddingMatcher) MatchVariable(ctx context.Context, variable VariableInfo) ([]ConceptMatch, error) {
	// Get embedding for the variable
	varEmbedding, err := m.GetVariableEmbedding(ctx, variable)
	if err != nil {
		return nil, err
	}

	var matches []ConceptMatch

	// Compare with each concept
	for _, concept := range m.Concepts {
		similarity := CosineSimilarity(varEmbedding, concept.Embedding)

		// If similarity is above threshold, add to matches
		if similarity >= m.SimilarityThreshold {
			matches = append(matches, ConceptMatch{
				Variable:        variable,
				Concept:         concept.Name,
				SimilarityScore: similarity,
			})
		}
	}

	// Sort matches by similarity score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].SimilarityScore > matches[j].SimilarityScore
	})

	return matches, nil
}

// MatchVariables matches multiple variables to security concepts
func (m *EmbeddingMatcher) MatchVariables(ctx context.Context, variables []VariableInfo) (map[string][]ConceptMatch, error) {
	result := make(map[string][]ConceptMatch)

	for _, variable := range variables {
		matches, err := m.MatchVariable(ctx, variable)
		if err != nil {
			return nil, err
		}

		// Group matches by concept
		for _, match := range matches {
			result[match.Concept] = append(result[match.Concept], match)
		}
	}

	return result, nil
}

// LoadSecurityConcepts loads the pre-computed security concept embeddings
func LoadSecurityConcepts() ([]SecurityConcept, error) {
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
		Concepts []SecurityConcept `toml:"concepts"`
	}

	if err := toml.Unmarshal(conceptsData, &conceptsConfig); err != nil {
		return nil, fmt.Errorf("error parsing concepts file: %w", err)
	}

	// Create a map for easier lookup
	conceptsMap := make(map[string]*SecurityConcept)
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
		Embeddings []embeddingEntry `toml:"embeddings"`
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
func loadLegacySecurityConcepts(filepath string) ([]SecurityConcept, error) {
	// Read the TOML file
	conceptsData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading security concepts TOML: %w", err)
	}

	var config struct {
		Concepts []SecurityConcept `toml:"concepts"`
	}

	if err := toml.Unmarshal(conceptsData, &config); err != nil {
		return nil, fmt.Errorf("error parsing security concepts TOML: %w", err)
	}

	return config.Concepts, nil
}

// loadLegacyMetadataFormat loads concepts from the old format with separate files per embedding
func loadLegacyMetadataFormat(embeddingsDir string) ([]SecurityConcept, error) {
	// Load the metadata file
	metadataFile := filepath.Join(embeddingsDir, "concepts_metadata.toml")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("error reading concept metadata: %w", err)
	}

	var config struct {
		Concepts []SecurityConcept `toml:"concepts"`
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
				Embedding Embedding `toml:"embedding"`
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

// GenerateSecurityConceptsEmbeddings generates embeddings for security concepts
func GenerateSecurityConceptsEmbeddings(ctx context.Context, client *OpenAIClient, concepts []SecurityConcept) ([]SecurityConcept, error) {
	result := make([]SecurityConcept, len(concepts))

	for i, concept := range concepts {
		// Copy concept data
		result[i] = concept

		// Generate embedding
		embedding, err := client.GetEmbedding(ctx, concept.Name)
		if err != nil {
			return nil, err
		}

		result[i].Embedding = embedding
	}

	return result, nil
}

// SaveSecurityConceptsToTOML saves security concepts with embeddings to a TOML file
func SaveSecurityConceptsToTOML(concepts []SecurityConcept, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	config := struct {
		Concepts []SecurityConcept `toml:"concepts"`
	}{
		Concepts: concepts,
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}

// GetAPIKey retrieves the OpenAI API key from environment variable or config file
func GetAPIKey() string {
	// First try environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		return apiKey
	}

	// Then try config file (implementation would depend on your config system)
	// This is just a placeholder
	return ""
}
