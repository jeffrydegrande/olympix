# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Solidair is a static analysis tool for Cairo smart contracts designed to detect vulnerability patterns that led to the zkLend hack. It uses Tree-sitter to parse Cairo code and identify potential security issues through pattern matching.

## Commands

### Build and Run

```bash
# Build the project
go build -o solidair

# Run the tool against a Cairo file
./solidair examples/bad.cairo

# Run the tool against a good example
./solidair examples/good.cairo
```

### Development

```bash
# Run tests
go test ./...

# Get dependencies
go get -u ./...

# Format code
go fmt ./...
```

## Code Architecture

Solidair is structured as follows:

1. **Parser Layer**: Uses Tree-sitter to parse Cairo files into an AST
   - Located in `/cairo` directory
   - Interfaces with the Tree-sitter C library through CGO

2. **Query Engine**: Loads and executes Tree-sitter queries against the parsed AST
   - Defined in `main.go`
   - Reads query definitions from the `queries` directory

3. **Vulnerability Detection**: Patterns defined as Tree-sitter queries (`.scm` files)
   - Organized by vulnerability category in the `queries` directory
   - Each query includes metadata like name and description

4. **Version 2 Enhancements** (Planned):
   - Variable name extraction to identify semantically similar variables
   - Embedding generation for semantic matching of variable names
   - Templated queries that can be parameterized with extracted variables

## Project Structure

- **`main.go`**: Entry point, contains the query execution logic
- **`cairo/`**: Cairo language binding for Tree-sitter
  - `cairo.go`: Go wrapper for the Tree-sitter Cairo parser
  - `parser.c`/`parser.h`: C interface to the Tree-sitter parser
  - `api.h`: Tree-sitter API definitions
- **`queries/`**: Tree-sitter query definitions for vulnerability patterns
  - `initialization/`: Queries for detecting market initialization issues
  - `race_conditions/`: Queries for detecting race conditions
  - `rounding/`: Queries for detecting rounding and precision issues
- **`examples/`**: Sample Cairo files for testing
  - `bad.cairo`: Example with vulnerabilities
  - `good.cairo`: Example with proper safeguards
- **`docs/`**: Documentation and design files
  - Design documents for Version 2 enhancements
  - Background information on the zkLend hack

## Working with Tree-sitter Queries

Tree-sitter queries use a Scheme-like syntax to match patterns in the AST. Key concepts:

1. **Node Matching**: Match specific node types in the AST
2. **Capture Groups**: Capture specific nodes with `@name` syntax
3. **Predicates**: Filter matches with conditions like `(#match?)` or `(#eq?)`

Example query to find functions without activation checks:
```scheme
(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|withdraw|flashLoan)$"))
  body: (block) @func_body
  (#not-match? @func_body "is_active|active|isActive"))
```

## Version 2 Planned Enhancements

The project is evolving to include:

1. **Variable Name Extraction**: Extract variable names from Cairo contracts to identify security-related variables regardless of naming conventions.

2. **Embedding-Based Matching**: Generate semantic embeddings for variable names to match them against security concepts.

3. **Templated Queries**: Convert static queries to templates that can be parameterized with extracted variable names.

These enhancements are designed to make the tool more flexible and capable of adapting to different coding styles and naming conventions.

## Development Guidelines

- Add proper metadata to new query files (Name, Description, Severity)
- Test new queries against both good and bad examples
- Follow Go best practices for new code
- Document new features in the docs directory