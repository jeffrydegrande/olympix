package cairo

/*
#include "parser.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func Language() *C.TSLanguage {
	return C.tree_sitter_cairo()
}

// Parse parses Cairo source code and returns a Tree-sitter tree
func Parse(source []byte) (*tree_sitter.Tree, error) {
	parser := tree_sitter.NewParser()
	defer parser.Close()

	err := parser.SetLanguage(tree_sitter.NewLanguage(unsafe.Pointer(Language())))
	if err != nil {
		return nil, fmt.Errorf("error setting language: %w", err)
	}
	
	tree := parser.Parse(source, nil)
	
	return tree, nil
}