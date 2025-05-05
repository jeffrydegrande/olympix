package cmd

import (
	"fmt"
	"os"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/jeffrydegrande/solidair/variables"
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract [file]",
	Short: "Extract variable names from a Cairo file",
	Args:  cobra.ExactArgs(1),
	Run:   extractMain,
}

func init() {
	rootCmd.AddCommand(extractCmd)
}

func extractMain(cmd *cobra.Command, args []string) {
	filename := args[0]

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

	vars.Filename = filename
	variables.PrintExtractedVariables(vars)
}

