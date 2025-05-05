package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/jeffrydegrande/solidair/pkg/concepts"
	"github.com/jeffrydegrande/solidair/pkg/embedding"
	"github.com/jeffrydegrande/solidair/pkg/templates"
	"github.com/jeffrydegrande/solidair/pkg/variables"
	"github.com/spf13/cobra"
)

var analyzeSmartCmd = &cobra.Command{
	Use:   "analyze-smart [file]",
	Short: "Analyze a Cairo file with semantic variable matching",
	Long: `Analyze a Cairo file using both static pattern matching and semantic variable matching.
This command extracts variables, matches them to security concepts using embeddings,
and then uses the matched variables to parameterize the query templates.`,
	Args: cobra.ExactArgs(1),
	Run:  analyzeSmartMain,
}

func init() {
	rootCmd.AddCommand(analyzeSmartCmd)
}

func analyzeSmartMain(cmd *cobra.Command, args []string) {
	filename := args[0]
	queryDir, _ := cmd.Flags().GetString("query-dir")
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
	defer tree.Close()

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
	conceptMatches, err := matcher.MatchVariables(ctx, vars.Variables)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error matching variables: %v\n", err)
		os.Exit(1)
	}

	// Read regular and templated queries
	queries, err := ReadQueryFiles(queryDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading query files: %v\n", err)
		os.Exit(1)
	}

	if len(queries) == 0 {
		fmt.Printf("No query files found in %s\n", queryDir)
		os.Exit(0)
	}

	fmt.Printf("Loaded %d queries\n", len(queries))

	// Parse templates
	queryTemplates := make(map[string]*templates.QueryTemplate)
	for source, content := range queries {
		template, err := templates.ParseQueryTemplate(content, source)
		if err != nil {
			fmt.Printf("Warning: Error parsing query template %s: %v\n", source, err)
			continue
		}
		queryTemplates[source] = template
	}

	// Process templated queries
	parameterizedQueries := templates.ProcessTemplatedQueries(queryTemplates, conceptMatches)

	// Run standard queries (non-templated)
	standardResults, err := RunQueries(data, tree, queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running standard queries: %v\n", err)
		os.Exit(1)
	}

	// Run parameterized queries
	var templatedResults []QueryResult
	for _, paramQuery := range parameterizedQueries {
		// Create a map with a single entry for this query
		queryMap := map[string]string{
			paramQuery.Template.Source: paramQuery.ProcessedQuery,
		}

		// Run the parameterized query
		results, err := RunQueries(data, tree, queryMap)
		if err != nil {
			fmt.Printf("Warning: Error running parameterized query %s: %v\n",
				paramQuery.Template.Name, err)
			continue
		}

		// Add parameter information to results
		for i := range results {
			results[i].QueryName = paramQuery.Template.Name + " (parametrized)"
		}

		templatedResults = append(templatedResults, results...)
	}

	// Combine results
	allResults := append(standardResults, templatedResults...)

	// Print the results
	if len(allResults) == 0 {
		fmt.Println("No vulnerabilities found.")
	} else {
		fmt.Printf("Found %d potential vulnerabilities:\n\n", len(allResults))

		for i, result := range allResults {
			fmt.Printf("Vulnerability #%d: %s\n", i+1, result.QueryName)
			fmt.Printf("  Source: %s\n", result.QueryFile)
			if result.Description != "" {
				fmt.Printf("  Description: %s\n", result.Description)
			}
			fmt.Printf("  Line: %d\n", result.LineNumber)
			fmt.Printf("  Code: %s\n\n", result.Code)
		}
	}
}