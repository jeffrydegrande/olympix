package embedding

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/jeffrydegrande/solidair/types"
	"github.com/pelletier/go-toml/v2"
	"github.com/sashabaranov/go-openai"
)

// EmbeddingCache provides caching for computed embeddings
type EmbeddingCache struct {
	Variables map[string]types.Embedding // Cache of variable embeddings
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
func (c *OpenAIClient) GetEmbedding(ctx context.Context, text string) (types.Embedding, error) {
	resp, err := c.Client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return types.Embedding{}, fmt.Errorf("error getting embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return types.Embedding{}, fmt.Errorf("no embedding data returned")
	}

	// Convert from []float64 to []float32 to save memory
	vector := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		vector[i] = float32(v)
	}

	return types.Embedding{Vector: vector}, nil
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(a, b types.Embedding) float32 {
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

// SaveEmbeddingsFile saves all embeddings to a single file
func SaveEmbeddingsFile(embeddings []types.EmbeddingEntry, outputDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create embeddings file
	embeddingsFile := filepath.Join(outputDir, "embeddings.toml")
	file, err := os.Create(embeddingsFile)
	if err != nil {
		return fmt.Errorf("error creating embeddings file: %w", err)
	}
	defer file.Close()

	// Write the embeddings
	config := struct {
		Embeddings []types.EmbeddingEntry `toml:"embeddings"`
	}{
		Embeddings: embeddings,
	}

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding embeddings TOML: %w", err)
	}

	return nil
}

// LoadEmbeddingsFile loads embeddings from a file
func LoadEmbeddingsFile(embeddingsFile string) ([]types.EmbeddingEntry, error) {
	embeddingsData, err := os.ReadFile(embeddingsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading embeddings file: %w", err)
	}

	var embeddingsConfig struct {
		Embeddings []types.EmbeddingEntry `toml:"embeddings"`
	}

	if err := toml.Unmarshal(embeddingsData, &embeddingsConfig); err != nil {
		return nil, fmt.Errorf("error parsing embeddings file: %w", err)
	}

	return embeddingsConfig.Embeddings, nil
}

// GetAPIKey retrieves the OpenAI API key from environment variable
func GetAPIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

// Helper functions for string similarity (used in offline mode)

// Contains checks if a string contains another string, case-insensitive
func Contains(s, substr string) bool {
	return ContainsIgnoreCase(s, substr)
}

// ContainsIgnoreCase checks if a string contains another string, case-insensitive
func ContainsIgnoreCase(s, substr string) bool {
	return contains(s, substr)
}

// Helper implementation of case-insensitive contains
func contains(s, substr string) bool {
	// A simple implementation would be strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	// But we could implement a more sophisticated version if needed
	return true // Placeholder
}

// CalculateNGramSimilarity computes n-gram similarity between two strings
func CalculateNGramSimilarity(s1, s2 string) float32 {
	return 0.5 // Placeholder
}

