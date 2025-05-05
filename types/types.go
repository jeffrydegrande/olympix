package types

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

// EmbeddingEntry stores an embedding with its concept name for easier mapping
type EmbeddingEntry struct {
	ConceptName string    `toml:"concept_name"`
	Embedding   Embedding `toml:"embedding"`
}

// VariableInfo represents information about an extracted variable
type VariableInfo struct {
	Name       string   // Variable name
	Type       string   // Variable type (if available)
	Context    string   // Storage, parameter, local, constant
	ParentName string   // Name of the parent struct/function
	LineNumber uint32   // Line number in source code
	Comments   []string // Associated comments
}

// ConceptMatch represents a match between a variable and a security concept
type ConceptMatch struct {
	Variable        VariableInfo // The matched variable
	Concept         string       // The security concept (e.g., "active", "initialized")
	SimilarityScore float32      // 0.0-1.0 score of the match quality
}