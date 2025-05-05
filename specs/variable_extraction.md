# Variable Name Extraction Design

## Overview

The variable name extraction component will identify and extract variable names from Cairo smart contracts. This is a crucial step in making our static analysis tool more flexible by removing the dependency on hardcoded identifier names in queries.

## Requirements

1. Extract variable names from different contexts:
   - Storage variables
   - Function parameters
   - Local variables
   - State variables
   - Constants

2. Capture context information:
   - Variable scope/location (storage, local, parameter)
   - Variable type (if available)
   - Associated comments (for potential hint extraction)

3. Organize variables in a structured format for later embedding and matching

## Implementation Approach

### 1. Tree-sitter Queries for Variable Extraction

We'll define Tree-sitter queries to identify and extract variables from Cairo code. Each query will focus on a specific context where variables are declared.

#### a. Storage Variables Query

```scheme
; Extract storage variables from struct declarations
(struct_item
  name: (type_identifier) @struct_name
  (#match? @struct_name "Storage")
  body: (field_declaration_list
    (field_declaration
      name: (identifier) @storage_var
      type: (_) @storage_type)))
```

#### b. Function Parameters Query

```scheme
; Extract function parameters
(function_item
  (function
    name: (identifier) @func_name
    parameters: (parameters
      (parameter
        name: (identifier) @param_name
        type: (_) @param_type))))
```

#### c. Local Variables Query

```scheme
; Extract local variables from function bodies
(function_item
  (function
    name: (identifier) @func_name
    body: (block
      (let_statement
        pattern: (identifier) @local_var
        type: (_)? @local_type))))
```

#### d. Constants Query

```scheme
; Extract constants
(const_declaration
  name: (identifier) @const_name
  type: (_)? @const_type
  value: (_) @const_value)
```

### 2. Data Structure for Variable Information

```go
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
    Filename string         // Source filename
    Variables []VariableInfo // All extracted variables
}
```

### 3. Variable Extraction Function

```go
// ExtractVariables extracts variables from the parsed source code
func ExtractVariables(source []byte, tree *tree_sitter.Tree) (*ExtractedVariables, error) {
    vars := &ExtractedVariables{
        Variables: make([]VariableInfo, 0),
    }
    
    // Load and execute queries for different variable types
    err := extractStorageVariables(source, tree, vars)
    if err != nil {
        return nil, err
    }
    
    err = extractFunctionParameters(source, tree, vars)
    if err != nil {
        return nil, err
    }
    
    err = extractLocalVariables(source, tree, vars)
    if err != nil {
        return nil, err
    }
    
    err = extractConstants(source, tree, vars)
    if err != nil {
        return nil, err
    }
    
    return vars, nil
}
```

### 4. Query Execution Helper

```go
// executeVariableQuery is a helper function for executing a query and processing its results
func executeVariableQuery(source []byte, tree *tree_sitter.Tree, queryStr string, 
                         context string, varsResult *ExtractedVariables) error {
    lang := tree_sitter.NewLanguage(unsafe.Pointer(cairo.Language()))
    query, err := tree_sitter.NewQuery(lang, queryStr)
    if err != nil {
        return err
    }
    defer query.Close()
    
    qc := tree_sitter.NewQueryCursor()
    defer qc.Close()
    matches := qc.Matches(query, tree.RootNode(), source)
    
    // Process matches and add to vars
    for match := matches.Next(); match != nil; match = matches.Next() {
        // Extract variable info from captures
        // Add to varsResult.Variables
    }
    
    return nil
}
```

### 5. Integration with Main Program

The variable extraction will be integrated into the main program flow:

1. Parse the Cairo source file (existing functionality)
2. Extract variables using the new extraction functions
3. Process the extracted variables for embedding
4. Use the variables in parameterized queries

## Implementation Steps

1. Define and implement the VariableInfo and ExtractedVariables structs
2. Create Tree-sitter queries for each variable context
3. Implement the extraction functions
4. Add tests with sample Cairo code
5. Integrate with the main program flow

## Testing Strategy

1. Create sample Cairo files with various variable declarations
2. Run the extraction on these files
3. Verify that all expected variables are extracted with the correct context
4. Check edge cases like variables with unusual names or in nested scopes

## Potential Challenges

1. **Complex Scoping**: Cairo may have complex scoping rules that make variable context tracking difficult
2. **Type Inference**: Extracting accurate type information may be challenging
3. **Comments Association**: Linking variables with their comments requires tracking source positions
4. **Query Complexity**: Tree-sitter queries may become complex for nested structures

## Next Steps

After implementing variable extraction, we will:
1. Process the extracted variables to generate embeddings
2. Create a system for matching extracted variables to query parameters
3. Update the query engine to use the matched variables