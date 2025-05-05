package cairo

/*
#include "parser.h"
*/
import "C"

func Language() *C.TSLanguage {
	return C.tree_sitter_cairo()
}
