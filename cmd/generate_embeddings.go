package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	outputFile string
)

var generateEmbeddingsCmd = &cobra.Command{
	Use:   "generate-embeddings",
	Short: "Generate embeddings for security concepts",
	Long: `Generate embeddings for security concepts using the OpenAI API.
These embeddings will be stored in a TOML file and used for semantic matching
of variable names against security concepts.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get API key from flag or environment variable
		apiKey, _ := cmd.Flags().GetString("api-key")
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				fmt.Fprintln(os.Stderr, "OpenAI API key not provided. Use --api-key flag or set OPENAI_API_KEY environment variable")
				os.Exit(1)
			}
		}

		// Define security concepts
		concepts := []SecurityConcept{
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

		// Initialize OpenAI client
		client := openai.NewClient(apiKey)
		ctx := context.Background()

		// Generate embeddings for each concept
		for i, concept := range concepts {
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

			// Update concept with embedding
			concepts[i].Embedding = Embedding{Vector: vector}
		}

		// Ensure directory exists
		dir := "embeddings"
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Save to TOML
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		config := struct {
			Concepts []SecurityConcept `toml:"concepts"`
		}{
			Concepts: concepts,
		}

		encoder := toml.NewEncoder(file)
		if err := encoder.Encode(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding TOML: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully generated embeddings and saved to %s\n", outputFile)
	},
}

func init() {
	generateEmbeddingsCmd.Flags().StringVarP(&outputFile, "output", "o", "embeddings/security_concepts.toml", "Output TOML file path")
	rootCmd.AddCommand(generateEmbeddingsCmd)
}