package evaluator

import (
	"github.com/yourfavoritedev/golang-interpreter/object"
)

// builtins is a dictionary that keeps track of built-in functions
var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltInByName("len"),
	"first": object.GetBuiltInByName("first"),
	"last":  object.GetBuiltInByName("last"),
	"rest":  object.GetBuiltInByName("rest"),
	"push":  object.GetBuiltInByName("push"),
	"puts":  object.GetBuiltInByName("puts"),
}
