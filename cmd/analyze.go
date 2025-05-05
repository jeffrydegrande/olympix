package cmd

import (
	"fmt"
	"os"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze a Cairo file for security vulnerabilities",
	Args:  cobra.ExactArgs(1),
	Run:   analyzeMain,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func analyzeMain(cmd *cobra.Command, args []string) {
	filename := args[0]
	queryDir, _ := cmd.Flags().GetString("query-dir")

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

	// Read all query files
	queries, err := ReadQueryFiles(queryDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading query files: %v\n", err)
		os.Exit(1)
	}

	if len(queries) == 0 {
		fmt.Printf("No query files found in %s\n", queryDir)
		os.Exit(0)
	}

	fmt.Printf("Loaded %d queries from %s\n", len(queries), queryDir)

	// Run all queries against the source code
	results, err := RunQueries(data, tree, queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running queries: %v\n", err)
		os.Exit(1)
	}

	// Print the results
	if len(results) == 0 {
		fmt.Println("No vulnerabilities found.")
	} else {
		fmt.Printf("Found %d potential vulnerabilities:\n\n", len(results))

		for i, result := range results {
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