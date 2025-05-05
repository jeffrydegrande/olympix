package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/jeffrydegrande/solidair/pkg/concepts"
	"github.com/jeffrydegrande/solidair/pkg/embedding"
	"github.com/jeffrydegrande/solidair/pkg/variables"
	"github.com/spf13/cobra"
)

var matchCmd = &cobra.Command{
	Use:   "match [file]",
	Short: "Match variables in a Cairo file to security concepts",
	Long: `Analyze a Cairo file to extract variables and match them against security concepts.
This helps identify variables that may be related to security-sensitive operations.`,
	Args: cobra.ExactArgs(1),
	Run:  matchMain,
}

func init() {
	rootCmd.AddCommand(matchCmd)
}

func matchMain(cmd *cobra.Command, args []string) {
	filename := args[0]
	apiKey, _ := cmd.Flags().GetString("api-key")
	offline, _ := cmd.Flags().GetBool("offline")

	// Load security concepts
	securityConcepts, err := concepts.LoadSecurityConcepts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading security concepts: %v\n", err)
		os.Exit(1)
	}

	// Read the source code
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Parse the source code
	tree, err := cairo.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Extract variables
	vars, err := variables.ExtractVariables(data, tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting variables: %v\n", err)
		os.Exit(1)
	}

	// Set up embedding matcher
	var matcher *embedding.EmbeddingMatcher
	if offline {
		// Offline mode
		matcher = embedding.NewEmbeddingMatcher(nil, securityConcepts, true)
	} else {
		// Online mode with OpenAI API
		if apiKey == "" {
			apiKey = embedding.GetAPIKey()
			if apiKey == "" {
				fmt.Fprintln(os.Stderr, "OpenAI API key not provided. Use --api-key flag or set OPENAI_API_KEY environment variable, or use --offline mode")
				os.Exit(1)
			}
		}
		openAIClient := embedding.NewOpenAIClient(apiKey)
		matcher = embedding.NewEmbeddingMatcher(openAIClient, securityConcepts, false)
	}

	// Match variables to concepts
	ctx := context.Background()
	matchesByConceptMap, err := matcher.MatchVariables(ctx, vars.Variables)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error matching variables: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Variable matches for %s:\n\n", filename)

	if len(matchesByConceptMap) == 0 {
		fmt.Println("No matches found.")
		return
	}

	for concept, matches := range matchesByConceptMap {
		fmt.Printf("Concept: %s\n", concept)
		fmt.Println(strings.Repeat("-", 40))

		for _, match := range matches {
			fmt.Printf("  Variable: %s (line %d)\n", match.Variable.Name, match.Variable.LineNumber)
			fmt.Printf("  Similarity: %.4f\n\n", match.SimilarityScore)
		}

		fmt.Println()
	}
}