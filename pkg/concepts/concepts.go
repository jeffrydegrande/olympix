package concepts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeffrydegrande/solidair/pkg/types"
	"github.com/pelletier/go-toml/v2"
)

// DefaultSecurityConcepts returns the default security concepts without embeddings
func DefaultSecurityConcepts() []types.SecurityConcept {
	return []types.SecurityConcept{
		{
			Name:        "active",
			Description: "Concept representing whether a market is active/initialized",
			Synonyms:    []string{"enabled", "live", "activated", "initialized", "ready"},
		},
		{
			Name:        "locked",
			Description: "Concept representing a reentrancy guard or mutex",
			Synonyms:    []string{"reentrancy_guard", "mutex", "guard", "semaphore"},
		},
		{
			Name:        "grace_period",
			Description: "Concept representing a waiting period or timelock",
			Synonyms:    []string{"timelock", "delay", "cooldown", "waiting_period"},
		},
		{
			Name:        "admin",
			Description: "Concept representing an administrative role or owner",
			Synonyms:    []string{"owner", "administrator", "authority", "governor"},
		},
		{
			Name:        "min_deposit",
			Description: "Concept representing a minimum deposit or liquidity threshold",
			Synonyms:    []string{"minimum_deposit", "min_liquidity", "threshold", "minimum_amount"},
		},
		{
			Name:        "bounds_check",
			Description: "Concept representing bounds checking or validations",
			Synonyms:    []string{"validation", "assert", "check", "limit", "cap"},
		},
		{
			Name:        "accumulator",
			Description: "Concept representing an accumulator or counter",
			Synonyms:    []string{"counter", "index", "tally", "tracker"},
		},
		{
			Name:        "donation_cap",
			Description: "Concept representing a cap on donations or contributions",
			Synonyms:    []string{"cap", "limit", "maximum", "ceiling"},
		},
	}
}

// SaveConceptsFile saves concept definitions without embeddings
func SaveConceptsFile(concepts []types.SecurityConcept, outputDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Make a copy without embeddings
	conceptsCopy := make([]types.SecurityConcept, len(concepts))
	for i, concept := range concepts {
		conceptsCopy[i] = types.SecurityConcept{
			Name:        concept.Name,
			Description: concept.Description,
			Synonyms:    concept.Synonyms,
		}
	}

	// Save to concepts file
	conceptsFile := filepath.Join(outputDir, "concepts.toml")
	file, err := os.Create(conceptsFile)
	if err != nil {
		return fmt.Errorf("error creating concepts file: %w", err)
	}
	defer file.Close()

	config := struct {
		Concepts []types.SecurityConcept `toml:"concepts"`
	}{
		Concepts: conceptsCopy,
	}

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding concepts TOML: %w", err)
	}

	return nil
}