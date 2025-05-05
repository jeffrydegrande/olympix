package embedding

import (
	"context"
	"sort"

	"github.com/jeffrydegrande/solidair/types"
)

// EmbeddingMatcher is a system for matching variables to security concepts
type EmbeddingMatcher struct {
	OpenAI              *OpenAIClient
	Concepts            []types.SecurityConcept
	Cache               *EmbeddingCache
	SimilarityThreshold float32
	Offline             bool
}

// NewEmbeddingMatcher creates a new matcher with the provided OpenAI client and concepts
func NewEmbeddingMatcher(client *OpenAIClient, concepts []types.SecurityConcept, offline bool) *EmbeddingMatcher {
	return &EmbeddingMatcher{
		OpenAI:              client,
		Concepts:            concepts,
		Cache:               &EmbeddingCache{Variables: make(map[string]types.Embedding)},
		SimilarityThreshold: 0.7, // Default threshold
		Offline:             offline,
	}
}

// GetVariableEmbedding gets the embedding for a variable, using cache if available
func (m *EmbeddingMatcher) GetVariableEmbedding(ctx context.Context, variable types.VariableInfo) (types.Embedding, error) {
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
		return types.Embedding{}, err
	}

	// Cache the embedding
	m.Cache.Variables[variable.Name] = embedding

	return embedding, nil
}

// getOfflineEmbedding creates a simple embedding for offline mode
// This is a placeholder - in a real implementation, we'd use a more
// sophisticated method for generating offline embeddings
func (m *EmbeddingMatcher) getOfflineEmbedding(name string) types.Embedding {
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

	return types.Embedding{Vector: vector}
}

// MatchVariable finds the best matching security concept for a variable
func (m *EmbeddingMatcher) MatchVariable(ctx context.Context, variable types.VariableInfo) ([]types.ConceptMatch, error) {
	// Get embedding for the variable
	varEmbedding, err := m.GetVariableEmbedding(ctx, variable)
	if err != nil {
		return nil, err
	}

	var matches []types.ConceptMatch

	// Compare with each concept
	for _, concept := range m.Concepts {
		similarity := CosineSimilarity(varEmbedding, concept.Embedding)

		// If similarity is above threshold, add to matches
		if similarity >= m.SimilarityThreshold {
			matches = append(matches, types.ConceptMatch{
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
func (m *EmbeddingMatcher) MatchVariables(ctx context.Context, vars []types.VariableInfo) (map[string][]types.ConceptMatch, error) {
	result := make(map[string][]types.ConceptMatch)

	for _, variable := range vars {
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

// calculateStringSimilarity computes a simple similarity score for offline mode
func calculateStringSimilarity(varName, conceptName string, synonyms []string) float32 {
	// Check for exact match
	if varName == conceptName {
		return 1.0
	}

	// Check if variable name contains concept name
	if Contains(varName, conceptName) {
		return 0.9
	}

	// Check synonyms
	for _, synonym := range synonyms {
		if Contains(varName, synonym) {
			return 0.8
		}
	}

	// Simple n-gram similarity for fallback
	return CalculateNGramSimilarity(varName, conceptName)
}

