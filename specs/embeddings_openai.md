# Embeddings Design for Variable Matching using OpenAI API

## Overview

We'll use semantic embeddings to match variable names with security-related concepts, regardless of the exact naming convention used. The implementation will use the OpenAI text embeddings API through the official Go client library. To maintain portability and avoid database dependencies, we'll store reference embeddings as TOML files that will be embedded in the compiled binary.

## Components

1. **OpenAI Embeddings Integration**: Generate embeddings for variable names using OpenAI's API
2. **Reference Embeddings**: Pre-computed embeddings for common security concepts, stored as TOML
3. **Embedding Cache**: In-memory cache to avoid redundant API calls
4. **Distance Calculation**: Cosine similarity function to match embeddings
5. **Embedding Matcher**: System to find the best matches between variables and security concepts

## Implementation Details

### 1. Data Structures

```go
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
    Variable        VariableInfo    // The matched variable
    Concept         string          // The security concept (e.g., "active", "initialized")
    SimilarityScore float32         // 0.0-1.0 score of the match quality
}

// EmbeddingCache provides caching for computed embeddings
type EmbeddingCache struct {
    Variables map[string]Embedding // Cache of variable embeddings
}
```

### 2. OpenAI API Integration

```go
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
```

### 3. Reference Embeddings in TOML

We'll create TOML files with pre-computed embeddings for security concepts. The format will be:

```toml
# security_concepts.toml
[[concepts]]
name = "active"
description = "Concept representing whether a market is active/initialized"
synonyms = ["enabled", "live", "activated", "initialized", "ready"]
embedding.vector = [0.1, 0.2, 0.3, ...] # Pre-computed embedding

[[concepts]]
name = "locked"
description = "Concept representing a reentrancy guard or mutex"
synonyms = ["reentrancy_guard", "mutex", "guard", "semaphore"]
embedding.vector = [0.4, 0.5, 0.6, ...] # Pre-computed embedding

# More concepts...
```

These files will be embedded in the binary using Go's embed package:

```go
//go:embed embeddings/security_concepts.toml
var securityConceptsTOML string

// LoadSecurityConcepts loads the pre-computed security concept embeddings
func LoadSecurityConcepts() ([]SecurityConcept, error) {
    var config struct {
        Concepts []SecurityConcept `toml:"concepts"`
    }
    
    if err := toml.Unmarshal([]byte(securityConceptsTOML), &config); err != nil {
        return nil, fmt.Errorf("error parsing security concepts TOML: %w", err)
    }
    
    return config.Concepts, nil
}
```

### 4. Distance Calculation

```go
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
```

### 5. Embedding Matcher

```go
// EmbeddingMatcher is a system for matching variables to security concepts
type EmbeddingMatcher struct {
    OpenAI     *OpenAIClient
    Concepts   []SecurityConcept
    Cache      *EmbeddingCache
    SimilarityThreshold float32
}

// NewEmbeddingMatcher creates a new matcher with the provided OpenAI client and concepts
func NewEmbeddingMatcher(client *OpenAIClient, concepts []SecurityConcept) *EmbeddingMatcher {
    return &EmbeddingMatcher{
        OpenAI:     client,
        Concepts:   concepts,
        Cache:      &EmbeddingCache{Variables: make(map[string]Embedding)},
        SimilarityThreshold: 0.7, // Default threshold
    }
}

// GetVariableEmbedding gets the embedding for a variable, using cache if available
func (m *EmbeddingMatcher) GetVariableEmbedding(ctx context.Context, variable VariableInfo) (Embedding, error) {
    // Check if we have a cached embedding
    if embedding, ok := m.Cache.Variables[variable.Name]; ok {
        return embedding, nil
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
```

### 6. Generating Reference Embeddings

We'll create a utility to pre-compute embeddings for security concepts:

```go
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
```

## API Key Handling

To avoid hardcoding OpenAI API keys, we'll support three methods:

1. Environment variable: `OPENAI_API_KEY`
2. Configuration file: `~/.config/solidair/config.toml`
3. Command-line flag: `--api-key`

## Offline Mode

To support users without an OpenAI API key, we'll provide an offline mode that uses only the pre-computed embeddings:

```go
// OfflineEmbeddingMatcher implements matching using only pre-computed embeddings
type OfflineEmbeddingMatcher struct {
    Concepts   []SecurityConcept
    SimilarityThreshold float32
}

// MatchVariableOffline matches a variable using string similarity instead of embeddings
func (m *OfflineEmbeddingMatcher) MatchVariableOffline(variable VariableInfo) []ConceptMatch {
    var matches []ConceptMatch
    
    for _, concept := range m.Concepts {
        // Simple similarity: check if variable name contains concept name or synonyms
        similarity := calculateStringSimilarity(variable.Name, concept.Name, concept.Synonyms)
        
        if similarity >= m.SimilarityThreshold {
            matches = append(matches, ConceptMatch{
                Variable:        variable,
                Concept:         concept.Name,
                SimilarityScore: similarity,
            })
        }
    }
    
    // Sort matches by similarity
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].SimilarityScore > matches[j].SimilarityScore
    })
    
    return matches
}
```

## Security Concept Definitions

We'll include these key security concepts in our reference embeddings:

1. **active**: Represents whether a market is active/enabled/initialized
2. **locked**: Represents a reentrancy guard or mutex
3. **grace_period**: Represents a waiting period or timelock
4. **admin**: Represents an administrative role or owner
5. **min_deposit**: Represents a minimum deposit or liquidity threshold
6. **bounds_check**: Represents bounds checking or validations
7. **accumulator**: Represents an accumulator or counter
8. **donation_cap**: Represents a cap on donations or contributions

## Integration with Main Program Flow

The embedding system will be integrated into the main program as follows:

1. Load security concepts with pre-computed embeddings
2. Extract variables from the Cairo source code
3. Match variables to security concepts
4. Use matched variables in templated queries

## Performance Considerations

- **Caching**: Cache embeddings to avoid duplicate API calls
- **Rate Limiting**: Respect OpenAI API rate limits
- **Parallelization**: Process multiple variables concurrently
- **Pre-computed Embeddings**: Use pre-computed embeddings for common concepts