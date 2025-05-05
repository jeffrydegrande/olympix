# Embeddings Design Document

## Overview

Embeddings will be used to match variable names with semantic concepts. Instead of relying on exact string matching, embeddings allow us to find variables with similar meanings to our target concepts. This is particularly useful for finding security-related variables regardless of their exact naming.

## Requirements

1. Generate embeddings for variable names extracted from Cairo contracts
2. Establish reference embeddings for security-related concepts
3. Implement similarity matching between extracted variables and reference concepts
4. Provide a reasonable performance with minimal external dependencies

## Embedding Options

We'll consider several approaches for generating embeddings:

### 1. Local Embedding Models

**Pros:**
- No external API calls or dependencies
- Faster execution
- Works offline

**Cons:**
- Potentially lower quality embeddings
- May require larger binary size
- Limited context understanding

**Options:**
- Sentence Transformers (Go bindings for BERT-like models)
- Word2Vec or GloVe pre-trained embeddings
- FastText with pre-trained vectors

### 2. Simple Word Similarity Approaches

**Pros:**
- Fast and lightweight
- No dependencies
- Easy to implement

**Cons:**
- Less semantic understanding
- Limited to basic similarity detection

**Options:**
- Word tokenization and overlap scoring
- Edit distance with semantic components
- n-gram similarity with weighting

### 3. External API Integration

**Pros:**
- High-quality embeddings with deep semantic understanding
- Maintained and updated by providers
- More accurate for specialized domains

**Cons:**
- Requires API keys and internet connection
- Network latency
- Usage costs

**Options:**
- OpenAI Text Embeddings API
- Hugging Face Inference API
- Custom API endpoint

## Recommended Approach

For simplicity, performance, and to avoid external dependencies, we recommend a simple but effective local approach:

**Hybrid Word-Semantic Similarity with Caching**

This approach combines multiple techniques:
1. Word segmentation (camelCase, snake_case, etc.)
2. Stemming or lemmatization of words
3. Synonym matching via a built-in dictionary
4. Weighted n-gram similarity for unknown terms
5. In-memory caching of computed similarities

## Implementation Design

### 1. Data Structures

```go
// Embedding represents a semantic embedding for a variable or concept
type Embedding struct {
    Original string    // Original name
    Segments []string  // Segmented words
    Stems    []string  // Stemmed segments
    Type     string    // Variable type, if available
    Context  string    // Variable context (storage, parameter, etc.)
}

// ConceptMatch represents a match between a variable and a security concept
type ConceptMatch struct {
    Variable       VariableInfo  // The matched variable
    Concept        string        // The security concept (e.g., "active", "initialized")
    SimilarityScore float64      // 0.0-1.0 score of the match quality
    MatchReason    string        // Human-readable explanation of why they matched
}

// EmbeddingCache provides caching for computed embeddings
type EmbeddingCache struct {
    Variables map[string]Embedding       // Cache of variable embeddings
    Concepts  map[string]Embedding       // Cache of concept embeddings
    Scores    map[string]map[string]float64  // Cache of similarity scores
}
```

### 2. Word Segmentation

```go
// SegmentVariableName breaks a variable name into word segments
func SegmentVariableName(name string) []string {
    // Handle camelCase
    // Handle snake_case
    // Handle other patterns
    // Return array of words
}
```

### 3. Stemming Functions

```go
// StemWord reduces a word to its root form
func StemWord(word string) string {
    // Implement basic stemming algorithm
    // or use Porter stemming
    // Return stemmed word
}
```

### 4. Similarity Calculation

```go
// CalculateSimilarity computes similarity between a variable and concept
func CalculateSimilarity(varEmbed, conceptEmbed Embedding) (float64, string) {
    // Check for direct matches and synonyms
    // Calculate weighted overlap of stemmed segments
    // Consider context/type information
    // Return similarity score and reason
}
```

### 5. Reference Concepts

We'll establish a set of reference security concepts with their expected semantic meanings:

```go
// SecurityConcepts maps concept names to their descriptions and synonyms
var SecurityConcepts = map[string]struct {
    Description string
    Synonyms    []string
}{
    "active": {
        Description: "Whether a market or feature is activated/enabled",
        Synonyms:    []string{"enabled", "live", "on", "operative", "running"},
    },
    "initialized": {
        Description: "Whether a contract or component has been properly set up",
        Synonyms:    []string{"setup", "configured", "prepared", "ready"},
    },
    "locked": {
        Description: "State preventing reentrant calls or concurrent access",
        Synonyms:    []string{"mutex", "reentrancy_guard", "semaphore", "exclusive"},
    },
    // Additional security concepts...
}
```

### 6. Main Embedding Process

```go
// GenerateEmbeddings creates embeddings for a set of extracted variables
func GenerateEmbeddings(variables []VariableInfo, cache *EmbeddingCache) []Embedding {
    var embeddings []Embedding
    
    for _, v := range variables {
        // Check cache first
        if embed, found := cache.Variables[v.Name]; found {
            embeddings = append(embeddings, embed)
            continue
        }
        
        // Create new embedding
        embed := Embedding{
            Original: v.Name,
            Segments: SegmentVariableName(v.Name),
            Type:     v.Type,
            Context:  v.Context,
        }
        
        // Stem the segments
        for _, seg := range embed.Segments {
            embed.Stems = append(embed.Stems, StemWord(seg))
        }
        
        embeddings = append(embeddings, embed)
        cache.Variables[v.Name] = embed
    }
    
    return embeddings
}
```

### 7. Concept Matching

```go
// MatchVariablesToConcepts finds variables that match security concepts
func MatchVariablesToConcepts(variables []VariableInfo, 
                             cache *EmbeddingCache) []ConceptMatch {
    var matches []ConceptMatch
    
    // Generate embeddings for variables
    varEmbeddings := GenerateEmbeddings(variables, cache)
    
    // Generate embeddings for concepts (if not cached)
    for concept := range SecurityConcepts {
        if _, found := cache.Concepts[concept]; !found {
            embed := Embedding{
                Original: concept,
                Segments: SegmentVariableName(concept),
            }
            for _, seg := range embed.Segments {
                embed.Stems = append(embed.Stems, StemWord(seg))
            }
            cache.Concepts[concept] = embed
        }
    }
    
    // Compare each variable against each concept
    for i, varEmbed := range varEmbeddings {
        for concept, conceptEmbed := range cache.Concepts {
            // Check cache for score
            cacheKey := varEmbed.Original + ":" + conceptEmbed.Original
            var score float64
            var reason string
            
            if scoreMap, found := cache.Scores[varEmbed.Original]; found {
                if cachedScore, ok := scoreMap[concept]; ok {
                    score = cachedScore
                    reason = "Cached result"
                }
            } else {
                score, reason = CalculateSimilarity(varEmbed, conceptEmbed)
                
                // Update cache
                if _, exists := cache.Scores[varEmbed.Original]; !exists {
                    cache.Scores[varEmbed.Original] = make(map[string]float64)
                }
                cache.Scores[varEmbed.Original][concept] = score
            }
            
            // If score is above threshold, add to matches
            if score >= 0.7 {  // Configurable threshold
                matches = append(matches, ConceptMatch{
                    Variable:        variables[i],
                    Concept:         concept,
                    SimilarityScore: score,
                    MatchReason:     reason,
                })
            }
        }
    }
    
    // Sort matches by similarity score (highest first)
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].SimilarityScore > matches[j].SimilarityScore
    })
    
    return matches
}
```

## Integration with Main Flow

The embedding system will be integrated as follows:

1. Extract variables from the Cairo source code
2. Generate embeddings for the variables
3. Match variables to security concepts
4. Pass the matched variables to the parameterized query system

## Performance Considerations

1. **Caching**: Implement caching for embeddings and similarity scores
2. **Efficient Data Structures**: Use optimized structures for quick lookups
3. **Lazy Computation**: Only compute similarities when needed
4. **Parallelization**: Consider parallel processing for large contracts

## Testing Strategy

1. **Unit Tests**: Test each component (segmentation, stemming, similarity)
2. **Integration Tests**: Verify the entire embedding pipeline
3. **Benchmark Tests**: Measure performance with different-sized inputs
4. **Real-world Tests**: Validate with actual Cairo contracts

## Next Steps

After implementing the embedding system:
1. Integrate with the variable extraction component
2. Connect to the parameterized query system
3. Fine-tune thresholds and scoring based on evaluation results