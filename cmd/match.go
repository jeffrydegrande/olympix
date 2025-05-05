package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/spf13/cobra"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var matchCmd = &cobra.Command{
	Use:   "match [file]",
	Short: "Match variables in a Cairo file to security concepts",
	Args:  cobra.ExactArgs(1),
	Run:   matchMain,
}

func init() {
	rootCmd.AddCommand(matchCmd)
}

func matchMain(cmd *cobra.Command, args []string) {
	filename := args[0]
	apiKey, _ := cmd.Flags().GetString("api-key")
	offline, _ := cmd.Flags().GetBool("offline")

	// Load security concepts
	concepts, err := LoadSecurityConcepts()
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
	parser := tree_sitter.NewParser()
	defer parser.Close()

	err = parser.SetLanguage(tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting language: %v\n", err)
		os.Exit(1)
	}
	tree := parser.Parse(data, nil)
	defer tree.Close()

	// Extract variables
	vars, err := ExtractVariables(data, tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting variables: %v\n", err)
		os.Exit(1)
	}

	// Set up embedding matcher
	var matcher *EmbeddingMatcher
	if offline {
		// Offline mode
		matcher = NewEmbeddingMatcher(nil, concepts, true)
	} else {
		// Online mode with OpenAI API
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				fmt.Fprintln(os.Stderr, "OpenAI API key not provided. Use --api-key flag or set OPENAI_API_KEY environment variable, or use --offline mode")
				os.Exit(1)
			}
		}
		openAIClient := NewOpenAIClient(apiKey)
		matcher = NewEmbeddingMatcher(openAIClient, concepts, false)
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
