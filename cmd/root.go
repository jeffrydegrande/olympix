package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "solidair",
	Short: "Solidair analyzes Cairo smart contracts for security vulnerabilities",
	Long: `Solidair is a static analysis tool for Cairo smart contracts designed to detect 
patterns that led to the zkLend hack and other potential vulnerabilities. It uses 
Tree-sitter to parse Cairo code and identify security issues through pattern matching.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("query-dir", "q", "queries", "Directory containing query definitions")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "OpenAI API key (can also be set via OPENAI_API_KEY env var)")
	rootCmd.PersistentFlags().BoolP("offline", "o", false, "Run in offline mode without API calls")
}