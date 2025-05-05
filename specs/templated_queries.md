# Templated Queries Design Document

## Overview

The templated queries system will transform static Tree-sitter queries into dynamic templates that can be populated with variable names identified through our embedding system. This addresses the hardcoded identifier limitation in the current implementation, making the tool more adaptable to different coding styles and naming conventions.

## Requirements

1. Define a template syntax for Tree-sitter queries
2. Implement a parser for templated queries
3. Develop a mechanism to substitute template parameters with actual variable names
4. Maintain compatibility with the existing query execution system
5. Provide clear error handling for template syntax and substitution issues

## Design Approach

### 1. Template Syntax

We will use a simple, distinct syntax for template parameters that won't conflict with Tree-sitter's query syntax:

```
${concept_name}
```

Where `concept_name` is a security concept like "active", "locked", "initialized", etc.

Example of a templated query:

```scheme
; Name: Missing Market Activation Check
; Description: Functions that should check market activation status but don't
; Concepts: active

(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|withdraw|flashLoan|borrow)$"))
  body: (block) @func_body
  (#not-match? @func_body "${active}"))
```

In this example, `${active}` is a template parameter that will be replaced with the actual variable name that represents the "active" concept in the analyzed contract.

### 2. Query Template Data Structure

```go
// QueryTemplate represents a Tree-sitter query with template parameters
type QueryTemplate struct {
    Name           string            // Query name
    Description    string            // Query description
    Concepts       []string          // Required concepts
    OriginalQuery  string            // Original query string with templates
    Parameters     map[string]string // Parameter placeholders to descriptions
}

// ParameterizedQuery represents a query with actual parameter values
type ParameterizedQuery struct {
    Template    *QueryTemplate       // Original template
    Parameters  map[string]string    // Parameter values
    ProcessedQuery string            // Query with parameters substituted
}
```

### 3. Template Parsing

```go
// ParseQueryTemplate parses a query string to extract template information
func ParseQueryTemplate(queryContent string) (*QueryTemplate, error) {
    template := &QueryTemplate{
        Parameters: make(map[string]string),
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
    
    // Store the original query
    template.OriginalQuery = queryContent
    
    // Extract template parameters
    paramRegex := regexp.MustCompile(`\${([a-zA-Z_][a-zA-Z0-9_]*)}`)
    matches := paramRegex.FindAllStringSubmatch(queryContent, -1)
    
    for _, match := range matches {
        if len(match) >= 2 {
            paramName := match[1]
            template.Parameters[paramName] = "" // We'll fill descriptions later
        }
    }
    
    return template, nil
}
```

### 4. Parameter Substitution

```go
// SubstituteParameters replaces template parameters with actual variable names
func SubstituteParameters(template *QueryTemplate, 
                         conceptMatches map[string][]ConceptMatch) (*ParameterizedQuery, error) {
    paramQuery := &ParameterizedQuery{
        Template:   template,
        Parameters: make(map[string]string),
    }
    
    processedQuery := template.OriginalQuery
    
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
```

### 5. Query Execution with Templates

```go
// ExecuteParameterizedQuery runs a parameterized query against source code
func ExecuteParameterizedQuery(source []byte, tree *tree_sitter.Tree, 
                              paramQuery *ParameterizedQuery) ([]QueryResult, error) {
    // Extract just the query part (remove comments)
    queryPattern := extractQueryPattern(paramQuery.ProcessedQuery)
    
    // Compile and execute the query (similar to current implementation)
    lang := tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language()))
    query, err := tree_sitter.NewQuery(lang, queryPattern)
    if err != nil {
        return nil, fmt.Errorf("error compiling query: %w", err)
    }
    defer query.Close()
    
    var results []QueryResult
    
    qc := tree_sitter.NewQueryCursor()
    defer qc.Close()
    matches := qc.Matches(query, tree.RootNode(), source)
    
    // Process matches (similar to current implementation)
    for match := matches.Next(); match != nil; match = matches.Next() {
        for _, capture := range match.Captures {
            node := capture.Node
            text := string(source[node.StartByte():node.EndByte()])
            
            // Get the line number for better reporting
            startPosition := node.StartPosition()
            
            results = append(results, QueryResult{
                QueryName:   paramQuery.Template.Name,
                QueryFile:   "templated", // Or store original file name
                Description: paramQuery.Template.Description,
                LineNumber:  uint32(startPosition.Row) + 1,
                Code:        text,
                // Include parameter substitutions in the result
                Parameters:  paramQuery.Parameters,
            })
            
            // We only need one capture per match to report the issue
            break
        }
    }
    
    return results, nil
}
```

### 6. Query Template Loading

```go
// LoadQueryTemplates reads all query template files from a directory
func LoadQueryTemplates(queryDir string) (map[string]*QueryTemplate, error) {
    templates := make(map[string]*QueryTemplate)
    
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
            
            // Parse the template
            template, err := ParseQueryTemplate(string(queryContent))
            if err != nil {
                return fmt.Errorf("error parsing query template %s: %w", path, err)
            }
            
            // Use relative path from queryDir as the key
            relPath, err := filepath.Rel(queryDir, path)
            if err != nil {
                return fmt.Errorf("error getting relative path for %s: %w", path, err)
            }
            
            templates[relPath] = template
        }
        
        return nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("error walking query directory: %w", err)
    }
    
    return templates, nil
}
```

### 7. Extended QueryResult Structure

```go
// Extended QueryResult to include parameter information
type QueryResult struct {
    QueryName   string
    QueryFile   string
    Description string
    LineNumber  uint32
    Code        string
    Parameters  map[string]string  // Added field for parameter substitutions
}
```

## Integration with Main Program Flow

The templated query system will integrate into the main program flow:

1. Load query templates from the query directory
2. Extract variables from the source code
3. Generate embeddings and match variables to concepts
4. For each template:
   a. Substitute parameters with matched variable names
   b. Execute the parameterized query
   c. Collect and report results

```go
// RunTemplatedQueries processes and runs templated queries
func RunTemplatedQueries(source []byte, tree *tree_sitter.Tree, 
                         queryDir string) ([]QueryResult, error) {
    // Load query templates
    templates, err := LoadQueryTemplates(queryDir)
    if err != nil {
        return nil, err
    }
    
    // Extract variables
    variables, err := ExtractVariables(source, tree)
    if err != nil {
        return nil, err
    }
    
    // Initialize embedding cache
    cache := &EmbeddingCache{
        Variables: make(map[string]Embedding),
        Concepts:  make(map[string]Embedding),
        Scores:    make(map[string]map[string]float64),
    }
    
    // Match variables to concepts
    allMatches := MatchVariablesToConcepts(variables.Variables, cache)
    
    // Organize matches by concept
    conceptMatches := make(map[string][]ConceptMatch)
    for _, match := range allMatches {
        conceptMatches[match.Concept] = append(conceptMatches[match.Concept], match)
    }
    
    // Process each template
    var allResults []QueryResult
    
    for _, template := range templates {
        // Skip templates with no concepts
        if len(template.Concepts) == 0 {
            continue
        }
        
        // Try to substitute parameters
        paramQuery, err := SubstituteParameters(template, conceptMatches)
        if err != nil {
            // Log the error but continue with other templates
            log.Printf("Warning: Skipping template %s: %v", template.Name, err)
            continue
        }
        
        // Execute the parameterized query
        results, err := ExecuteParameterizedQuery(source, tree, paramQuery)
        if err != nil {
            log.Printf("Error executing query %s: %v", template.Name, err)
            continue
        }
        
        // Add results to the combined list
        allResults = append(allResults, results...)
    }
    
    return allResults, nil
}
```

## Compatibility with Existing Queries

For backward compatibility, we'll support both templated and non-templated queries:

1. If a query contains template parameters and the required concepts are found, use the templated execution path
2. If a query doesn't contain template parameters, or if substitution fails, fall back to the original execution path

## Error Handling

The system will handle several error conditions:

1. **Missing Concept Matches**: Log a warning and skip queries that require concepts with no matches
2. **Template Syntax Errors**: Validate template syntax during loading and report issues
3. **Query Compilation Errors**: Handle parsing errors gracefully and report detailed information
4. **Substitution Issues**: If parameter substitution results in an invalid query, report the specific issue

## Testing Strategy

1. **Unit Tests**:
   - Test the template parsing logic
   - Test parameter substitution
   - Test error handling

2. **Integration Tests**:
   - Create test cases with templated queries and matching variables
   - Verify that templates correctly substitute and execute
   - Test the fallback to non-templated execution

3. **Real-world Tests**:
   - Convert existing queries to templates
   - Run against actual Cairo contracts
   - Verify that results match expectations

## Next Steps

After implementing the templated query system:
1. Convert existing queries to templates
2. Create reference documentation for template syntax
3. Consider additional template features like conditional sections or operator variations