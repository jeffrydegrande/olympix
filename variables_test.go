package main

import (
	"os"
	"testing"
	"unsafe"

	cairo "github.com/jeffrydegrande/solidair/cairo"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func TestExtractVariables(t *testing.T) {
	// Read the test file
	data, err := os.ReadFile("examples/good.cairo")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Parse the source code
	parser := tree_sitter.NewParser()
	defer parser.Close()

	err = parser.SetLanguage(tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language())))
	if err != nil {
		t.Fatalf("Failed to set language: %v", err)
	}
	tree := parser.Parse(data, nil)
	defer tree.Close()

	// Extract variables
	vars, err := ExtractVariables(data, tree)
	if err != nil {
		t.Fatalf("Failed to extract variables: %v", err)
	}

	// Check if any variables were extracted
	if len(vars.Variables) == 0 {
		t.Error("No variables were extracted")
	}

	// Define some expected variables
	expectedVars := map[string]bool{
		"SCALE":              true,
		"MIN_ACCUMULATOR":    true,
		"MAX_ACCUMULATOR":    true,
		"reserve_balance":    true,
		"total_supply":       true,
		"lending_accumulator": true,
		"deposit":            true,
		"amount":             true,
		"self":               true,
	}

	// Count how many expected variables we found
	foundCount := 0
	for _, v := range vars.Variables {
		if expectedVars[v.Name] {
			foundCount++
		}
	}

	// Make sure we found at least some of the expected variables
	if foundCount < 3 {
		t.Errorf("Expected to find more variables, only found %d of the expected ones", foundCount)
	}
}

// We're not using this function anymore since we've simplified the implementation