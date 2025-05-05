package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

	cairo "github.com/jeffrydegrande/solidair/cairo"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// QueryResult represents a finding from running a query against the code
type QueryResult struct {
	QueryName   string
	QueryFile   string
	Description string
	LineNumber  uint32
	Code        string
}

// ReadQueryFiles reads all .scm files from the specified query directory
func ReadQueryFiles(queryDir string) (map[string]string, error) {
	queries := make(map[string]string)

	err := filepath.WalkDir(queryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process .scm files
		if !d.IsDir() && strings.HasSuffix(path, ".scm") {
			queryContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading query file %s: %w", path, err)
			}

			// Use relative path from queryDir as the key
			relPath, err := filepath.Rel(queryDir, path)
			if err != nil {
				return fmt.Errorf("error getting relative path for %s: %w", path, err)
			}

			queries[relPath] = string(queryContent)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking query directory: %w", err)
	}

	return queries, nil
}

// ExtractQueryMetadata parses the query file to extract metadata from comments
func ExtractQueryMetadata(queryContent string) (string, string) {
	description := ""
	name := ""

	// Extract description from comments
	descRegex := regexp.MustCompile(`(?m)^;\s*Description:\s*(.+)$`)
	if matches := descRegex.FindStringSubmatch(queryContent); len(matches) > 1 {
		description = matches[1]
	}

	// Extract name from comments
	nameRegex := regexp.MustCompile(`(?m)^;\s*Name:\s*(.+)$`)
	if matches := nameRegex.FindStringSubmatch(queryContent); len(matches) > 1 {
		name = matches[1]
	}

	return name, description
}

// RunQueries executes all loaded queries against the source code
func RunQueries(source []byte, tree *tree_sitter.Tree, queries map[string]string) ([]QueryResult, error) {
	var results []QueryResult
	root := tree.RootNode()
	lang := tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language()))

	for queryFile, queryContent := range queries {
		// Extract metadata from the query file
		queryName, description := ExtractQueryMetadata(queryContent)
		if queryName == "" {
			// If no name is specified in the file, use the filename without extension
			queryName = strings.TrimSuffix(filepath.Base(queryFile), filepath.Ext(queryFile))
		}

		// Extract the actual query pattern (remove comments and metadata)
		queryPattern := extractQueryPattern(queryContent)

		// Compile the query
		query, err := tree_sitter.NewQuery(lang, queryPattern)
		if err != nil {
			fmt.Printf("Error compiling query %s: %v\n", queryFile, err)
			continue
		}
		defer query.Close()

		// Execute the query
		qc := tree_sitter.NewQueryCursor()
		defer qc.Close()
		matches := qc.Matches(query, root, source)

		// Process the matches
		for match := matches.Next(); match != nil; match = matches.Next() {
			for _, capture := range match.Captures {
				node := capture.Node
				text := string(source[node.StartByte():node.EndByte()])

				// Get the line number for better reporting
				startPosition := node.StartPosition()

				results = append(results, QueryResult{
					QueryName:   queryName,
					QueryFile:   queryFile,
					Description: description,
					LineNumber:  uint32(startPosition.Row) + 1, // +1 because editors use 1-based line numbers
					Code:        text,
				})

				// We only need one capture per match to report the issue
				break
			}
		}
	}

	return results, nil
}

// extractQueryPattern removes comments and extracts just the query pattern
func extractQueryPattern(queryContent string) string {
	lines := strings.Split(queryContent, "\n")
	var queryLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") {
			continue
		}
		queryLines = append(queryLines, line)
	}

	return strings.Join(queryLines, "\n")
}