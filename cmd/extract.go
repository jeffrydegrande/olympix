package cmd

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/jeffrydegrande/solidair/cairo"
	"github.com/spf13/cobra"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var extractCmd = &cobra.Command{
	Use:   "extract [file]",
	Short: "Extract variable names from a Cairo file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

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

		vars.Filename = filename
		PrintExtractedVariables(vars)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
}