package main

import (
	"fmt"
	"strings"
	"unsafe"

	cairo "github.com/jeffrydegrande/solidair/cairo"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// VariableInfo represents information about an extracted variable
type VariableInfo struct {
	Name       string   // Variable name
	Type       string   // Variable type (if available)
	Context    string   // Storage, parameter, local, constant
	ParentName string   // Name of the parent struct/function
	LineNumber uint32   // Line number in source code
	Comments   []string // Associated comments
}

// ExtractedVariables holds all variables extracted from a source file
type ExtractedVariables struct {
	Filename  string         // Source filename
	Variables []VariableInfo // All extracted variables
}

// Simple extractor that finds all identifiers
func ExtractVariables(source []byte, tree *tree_sitter.Tree) (*ExtractedVariables, error) {
	vars := &ExtractedVariables{
		Variables: make([]VariableInfo, 0),
	}

	// Simple query to find all identifiers
	query := "(identifier) @id"
	
	lang := tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language()))
	q, err := tree_sitter.NewQuery(lang, query)
	if err != nil {
		return nil, fmt.Errorf("error compiling query: %w", err)
	}
	defer q.Close()

	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	matches := qc.Matches(q, tree.RootNode(), source)

	// Track seen variables to avoid duplicates
	seen := make(map[string]bool)

	// Process the matches
	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, capture := range match.Captures {
			node := capture.Node
			text := string(source[node.StartByte():node.EndByte()])
			
			// Skip if we've already seen this variable
			if seen[text] {
				continue
			}
			seen[text] = true

			// Simple implementation to extract variables without complex context detection
			varInfo := VariableInfo{
				Name:       text,
				Context:    "variable", // Simplified context
				LineNumber: uint32(node.StartPosition().Row) + 1,
			}
			
			vars.Variables = append(vars.Variables, varInfo)
		}
	}

	return vars, nil
}

// PrintExtractedVariables prints information about extracted variables
func PrintExtractedVariables(vars *ExtractedVariables) {
	fmt.Printf("Extracted %d variables\n", len(vars.Variables))
	
	// Group variables by context
	contextGroups := make(map[string][]VariableInfo)
	for _, v := range vars.Variables {
		contextGroups[v.Context] = append(contextGroups[v.Context], v)
	}
	
	// Print each context group
	for context, variables := range contextGroups {
		fmt.Printf("\n%s Variables (%d):\n", strings.Title(context), len(variables))
		fmt.Println(strings.Repeat("-", 40))
		
		for _, v := range variables {
			fmt.Printf("- %s", v.Name)
			if v.Type != "" {
				fmt.Printf(" (%s)", v.Type)
			}
			if v.ParentName != "" {
				fmt.Printf(" [in %s]", v.ParentName)
			}
			fmt.Printf(" (line %d)\n", v.LineNumber)
		}
	}
}