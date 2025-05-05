# Analysis of Version 2 Requirements

## Current Implementation (Version 1)

The current implementation (v1) of Solidair is a static analysis tool for Cairo smart contracts that uses Tree-sitter to parse Cairo code and detect vulnerability patterns through pattern matching queries defined in Scheme files (.scm).

Key components:
1. Tree-sitter parser for Cairo (written in C, with Go bindings)
2. Query engine that reads and executes Tree-sitter pattern matching queries
3. Query files organized by vulnerability categories (initialization, race_conditions, rounding)
4. Simple CLI that parses Cairo files and reports matches

Limitations (as noted in README.md):
- Not semantic – it's purely syntax-based (no types or scopes)
- Hardcoded identifiers – fixed names like is_active, marketActive, etc.
- No rule composition – can't AND/OR queries programmatically

## Version 2 Requirements

Version 2 aims to address the limitation of hardcoded identifiers in queries. The core problem is that the current implementation assumes specific naming conventions for variables (e.g., "is_active", "marketActive"), but real-world contracts may use different variable names for the same concepts.

The plan for Version 2 involves:

1. **Variable Name Extraction**
   - Implement functionality to extract variable names from smart contracts using Tree-sitter
   - This will allow the tool to identify variables regardless of naming conventions

2. **Embedding Generation**
   - Calculate vector embeddings for extracted variable names
   - This will enable semantic matching rather than string-matching
   - Need to select or implement an embedding model suitable for code variable names

3. **Parameterized Queries**
   - Modify the query system to treat queries as templates
   - Add support for passing dynamic values (identified variable names) to queries
   - This allows queries to be more flexible and adapt to different naming conventions

## Implementation Considerations

### 1. Variable Name Extraction
- Need to create Tree-sitter queries to extract variable names from different contexts:
  - Storage variables
  - Function parameters
  - Local variables
  - State variables
- Need to capture variable context (e.g., storage variable vs local variable)
- Potentially track variable types if possible

### 2. Embedding Generation
- Need to select or implement an embedding model suitable for code
- Options include:
  - Integrating with an external embedding API (e.g., OpenAI embeddings)
  - Using a local embedding model
  - Implementing a simpler semantic matching approach (e.g., word similarity)
- Need to store reference embeddings for common security-related concepts:
  - "active", "initialized", "locked", etc.

### 3. Parameterized Query System
- Need to modify the query parser to support template parameters
- Design a template syntax for queries (e.g., `{active_flag}` instead of hardcoded "is_active")
- Create a mapping system that links template parameters to identified variables
- Update the query execution to substitute parameters with actual variable names

## Technical Challenges

1. **Embedding Quality**
   - General-purpose embeddings might not capture semantic relationships specific to code
   - Need to evaluate how well embeddings can match semantically similar variable names

2. **Variable Scoping**
   - Need to handle variables with the same name in different scopes
   - May need to track variable context (struct field vs local variable)

3. **Template Syntax**
   - Need to design a template syntax that works well with Tree-sitter queries
   - Must ensure that parameter substitution doesn't break the query syntax

4. **Performance**
   - Computing embeddings can be computationally expensive
   - May need to implement caching or optimize the embedding generation

## Next Steps

1. Design and implement variable extraction queries
2. Research and select an embedding approach
3. Design the template syntax for parameterized queries
4. Implement the parameter substitution mechanism
5. Update the main program to incorporate these new features
6. Test with real-world contracts to evaluate effectiveness

This approach should address the limitations of hardcoded identifiers while maintaining the simplicity and performance advantages of the current implementation.