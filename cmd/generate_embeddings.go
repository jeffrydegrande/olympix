package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeffrydegrande/solidair/concepts"
	"github.com/jeffrydegrande/solidair/embedding"
	"github.com/jeffrydegrande/solidair/types"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	outputDir string
)

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

func generateEmbeddingsMain(cmd *cobra.Command, args []string) {
	// Get API key from flag or environment variable
	apiKey, _ := cmd.Flags().GetString("api-key")
	if apiKey == "" {
		apiKey = embedding.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "OpenAI API key not provided. Use --api-key flag or set OPENAI_API_KEY environment variable")
			os.Exit(1)
		}
	}

	// Get the default security concepts
	securityConcepts := concepts.DefaultSecurityConcepts()

	// Initialize OpenAI client
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Create the concept metadata file first (without embeddings)
	err := concepts.SaveConceptsFile(securityConcepts, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving concepts file: %v\n", err)
		os.Exit(1)
	}

	// Collect embeddings for all concepts
	var embeddingEntries []types.EmbeddingEntry
	for _, concept := range securityConcepts {
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
		embeddingEntries = append(embeddingEntries, types.EmbeddingEntry{
			ConceptName: concept.Name,
			Embedding:   types.Embedding{Vector: vector},
		})
	}

	// Save all embeddings to a single file
	err = embedding.SaveEmbeddingsFile(embeddingEntries, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving embeddings file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated embeddings and saved to %s\n", outputDir)
	fmt.Printf("- Concepts: %s\n", filepath.Join(outputDir, "concepts.toml"))
	fmt.Printf("- Embeddings: %s\n", filepath.Join(outputDir, "embeddings.toml"))
}

