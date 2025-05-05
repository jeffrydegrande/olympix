package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	outputDir string
)

// embeddingEntry stores an embedding with its concept name for easier mapping
type embeddingEntry struct {
	ConceptName string    `toml:"concept_name"`
	Embedding   Embedding `toml:"embedding"`
}

var generateEmbeddingsCmd = &cobra.Command{
	Use:   "generate-embeddings",
	Short: "Generate embeddings for security concepts",
	Long: `Generate embeddings for security concepts using the OpenAI API.
These embeddings will be stored in separate files:
- concepts.toml: Contains concept metadata (names, descriptions, synonyms)
- embeddings.toml: Contains all embeddings in a single file

This approach keeps the files manageable while still allowing for efficient loading.`,
	Run: generateEmbeddingsMain,
}

func init() {
	generateEmbeddingsCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "embeddings", "Output directory for embedding files")
	rootCmd.AddCommand(generateEmbeddingsCmd)
}

// saveConceptsFile saves concept definitions without embeddings
func saveConceptsFile(concepts []SecurityConcept, outputDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Make a copy without embeddings
	conceptsCopy := make([]SecurityConcept, len(concepts))
	for i, concept := range concepts {
		conceptsCopy[i] = SecurityConcept{
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
		Concepts []SecurityConcept `toml:"concepts"`
	}{
		Concepts: conceptsCopy,
	}

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding concepts TOML: %w", err)
	}

	return nil
}

// saveEmbeddingsFile saves all embeddings to a single file
func saveEmbeddingsFile(embeddings []embeddingEntry, outputDir string) error {
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
		Embeddings []embeddingEntry `toml:"embeddings"`
	}{
		Embeddings: embeddings,
	}

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding embeddings TOML: %w", err)
	}

	return nil
}

func generateEmbeddingsMain(cmd *cobra.Command, args []string) {
	// Get API key from flag or environment variable
	apiKey, _ := cmd.Flags().GetString("api-key")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "OpenAI API key not provided. Use --api-key flag or set OPENAI_API_KEY environment variable")
			os.Exit(1)
		}
	}

	// Get the default security concepts
	concepts := DefaultSecurityConcepts()

	// Initialize OpenAI client
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Create the concept metadata file first (without embeddings)
	err := saveConceptsFile(concepts, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving concepts file: %v\n", err)
		os.Exit(1)
	}

	// Collect embeddings for all concepts
	var embeddingEntries []embeddingEntry
	for _, concept := range concepts {
		fmt.Printf("Generating embedding for concept '%s'...\n", concept.Name)

		// Get embedding
		resp, err := client.CreateEmbeddings(
			ctx,
			openai.EmbeddingRequest{
				Input: []string{concept.Name},
				Model: openai.AdaEmbeddingV2,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating embedding for %s: %v\n", concept.Name, err)
			os.Exit(1)
		}

		if len(resp.Data) == 0 {
			fmt.Fprintf(os.Stderr, "No embedding data returned for %s\n", concept.Name)
			os.Exit(1)
		}

		// Convert to float32
		vector := make([]float32, len(resp.Data[0].Embedding))
		for j, v := range resp.Data[0].Embedding {
			vector[j] = float32(v)
		}

		// Add to embedding entries
		embeddingEntries = append(embeddingEntries, embeddingEntry{
			ConceptName: concept.Name,
			Embedding:   Embedding{Vector: vector},
		})
	}

	// Save all embeddings to a single file
	err = saveEmbeddingsFile(embeddingEntries, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving embeddings file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated embeddings and saved to %s\n", outputDir)
	fmt.Printf("- Concepts: %s\n", filepath.Join(outputDir, "concepts.toml"))
	fmt.Printf("- Embeddings: %s\n", filepath.Join(outputDir, "embeddings.toml"))
}
