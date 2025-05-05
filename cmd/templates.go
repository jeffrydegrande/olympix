package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryTemplate represents a Tree-sitter query with template parameters
type QueryTemplate struct {
	Name        string              // Query name
	Description string              // Query description
	Concepts    []string            // Required concepts
	Source      string              // Source file
	Original    string              // Original query string with templates
	Parameters  map[string]struct{} // Template parameters
}

// ParameterizedQuery represents a query with actual parameter values
type ParameterizedQuery struct {
	Template        *QueryTemplate    // Original template
	Parameters      map[string]string // Parameter values
	ProcessedQuery  string            // Query with parameters substituted
}

// ParseQueryTemplate parses a query string to extract template information
func ParseQueryTemplate(queryContent, source string) (*QueryTemplate, error) {
	template := &QueryTemplate{
		Parameters: make(map[string]struct{}),
		Source:     source,
		Original:   queryContent,
	}
	
	// Extract metadata from comments
	nameRegex := regexp.MustCompile(`(?m)^;\s*Name:\s*(.+)$`)
	if matches := nameRegex.FindStringSubmatch(queryContent); len(matches) > 1 {
		template.Name = matches[1]
	}
	
	descRegex := regexp.MustCompile(`(?m)^;\s*Description:\s*(.+)$`)
	if matches := descRegex.FindStringSubmatch(queryContent); len(matches) > 1 {
		template.Description = matches[1]
	}
	
	// Extract concepts
	conceptsRegex := regexp.MustCompile(`(?m)^;\s*Concepts:\s*(.+)$`)
	if matches := conceptsRegex.FindStringSubmatch(queryContent); len(matches) > 1 {
		conceptsStr := matches[1]
		concepts := strings.Split(conceptsStr, ",")
		for i, c := range concepts {
			concepts[i] = strings.TrimSpace(c)
		}
		template.Concepts = concepts
	}
	
	// Extract template parameters
	paramRegex := regexp.MustCompile(`\${([a-zA-Z_][a-zA-Z0-9_]*)}`)
	matches := paramRegex.FindAllStringSubmatch(queryContent, -1)
	
	for _, match := range matches {
		if len(match) >= 2 {
			paramName := match[1]
			template.Parameters[paramName] = struct{}{}
		}
	}
	
	return template, nil
}

// SubstituteParameters replaces template parameters with actual variable names
func SubstituteParameters(template *QueryTemplate, conceptMatches map[string][]ConceptMatch) (*ParameterizedQuery, error) {
	paramQuery := &ParameterizedQuery{
		Template:   template,
		Parameters: make(map[string]string),
	}
	
	processedQuery := template.Original
	
	// Check if we have matches for all required concepts
	for _, concept := range template.Concepts {
		matches, found := conceptMatches[concept]
		if !found || len(matches) == 0 {
			return nil, fmt.Errorf("no matches found for required concept: %s", concept)
		}
		
		// Use the best match (highest similarity score)
		bestMatch := matches[0]
		paramName := concept
		varName := bestMatch.Variable.Name
		
		// Store the parameter substitution
		paramQuery.Parameters[paramName] = varName
		
		// Replace in the query string
		placeholder := fmt.Sprintf("${%s}", paramName)
		processedQuery = strings.ReplaceAll(processedQuery, placeholder, varName)
	}
	
	// Store the processed query
	paramQuery.ProcessedQuery = processedQuery
	
	return paramQuery, nil
}

// ProcessTemplatedQueries takes a set of query templates and processes them with matched variables
func ProcessTemplatedQueries(queryTemplates map[string]*QueryTemplate, 
                           conceptMatches map[string][]ConceptMatch) []*ParameterizedQuery {
	var processed []*ParameterizedQuery
	
	for _, template := range queryTemplates {
		// Skip templates with no concepts
		if len(template.Concepts) == 0 {
			continue
		}
		
		// Try to substitute parameters
		paramQuery, err := SubstituteParameters(template, conceptMatches)
		if err != nil {
			// Log the error but continue with other templates
			fmt.Printf("Warning: Skipping template %s: %v\n", template.Name, err)
			continue
		}
		
		processed = append(processed, paramQuery)
	}
	
	return processed
}