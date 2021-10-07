package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yourfavoritedev/golang-interpreter/ast"
)

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
)

// ObjectType is the type that represents an evaluated value as a string
type ObjectType string

// Object is the interface that represents every value
// we encounter when evaluating Monkey source code.
// Every value will be wrapped inside a stuct, which fulfills
// this Object interface. Tt is foundation for our object system.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer is the referenced struct for Integer Literals in our object system.
// The struct holds the evaluated value of the Integer Literal.
type Integer struct {
	Value int64 // the evaluated value
}

// Inspect returns the Integer struct's Value as a string
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type returns the ObjectType (INTEGER_OBJ) associated with the referenced Integer struct
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Boolean is the referenced struct for Boolean Literals in our object system.
// The struct holds the evaluated value of the Boolean Literal.
type Boolean struct {
	Value bool // the evaluated value
}

// Inspect returns the Boolean struct's Value as a string
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Type returns the ObjectType (BOOLEAN_OBJ) associated with the referenced Boolean struct
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Null is the referenced struct for Null Literals in our object system.
// By nature it has no value, since it represents the absence of any value.
type Null struct{}

// Inspect returns a literal "null" string as there is no value to stringify on Null structs
func (n *Null) Inspect() string { return "null" }

// Type returns the ObjectType (NULL_OBJ) associated with the referenced Null struct
func (n *Null) Type() ObjectType { return NULL_OBJ }

// ReturnValue wraps the intended return value inside an Object,
// giving us the ability to keep track of it. Keeping track of it helps
// us later decide whether to stop evalution or not.
type ReturnValue struct {
	Value Object
}

// Type returns the ObjectType (RETURN_VALUE_OBJ) associated with the referenced ReturnValue struct
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Inspect returns the ReturnValue struct's Value as a string. Since the
// Value is of type Object (interface), we can call Inspect() from the
// underlying struct which implemeneted the Object interface.
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Error contains the Message corresponding to an error that
// was encountered while evaluating the AST
type Error struct {
	Message string
}

// Type returns the ObjectType (ERROR_OBJ) associated with the referenced Error struct
func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Inspect returns the Error struct's Message as a formatted string
// to print out the error message
func (e *Error) Inspect() string { return "ERROR: " + e.Message }

// Function is the referenced struct for Function Literals in our object system.
// The struct holds the function's parameters and body to be later evaluated
// when referenced in its respective environment in a function call
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// Type returns the ObjectType (FUNCTION_OBJ) associated with the referenced Function struct
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect will construct the Function as a string by stringifying its components,
// the parameters and body, and concatenating them into the expected function format.
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	// build params, convert ast.Identifiers to strings
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	// construct function literal as string
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// String is the referenced struct for String Literals in our object system.
// The struct holds the evaluated value of the String Literal.
type String struct {
	Value string
}

// Type returns the ObjectType (STRING_OBJ) associated with the referenced String struct
func (s *String) Type() ObjectType { return STRING_OBJ }

// Inspect returns the String struct's Value which is of type string
func (s *String) Inspect() string { return s.Value }

// BuiltinFunction is used to create built-in functions that can be called in the interpretor.
// The functions are defined by us and can be called by the user. A built-in function can be
// constructed with any number of arguments of the type Object, but it must return an Object.
type BuiltinFunction func(args ...Object) Object

// Builtin is the referenced struct for built-in functions in our object system.
// The struct holds the defined built-in function.
type Builtin struct {
	Fn BuiltinFunction
}

// Type returns the ObjectType (BUILTIN_OBJ) associated with the referenced Builtin struct
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

// Inspect returns a static string for the Builtin struct
func (b *Builtin) Inspect() string { return "builtin function" }

// Array is the referenced struct for Array Literals in our object system.
// The struct holds the evaluated elements of the array literal
type Array struct {
	Elements []Object
}

// Type returns the ObjectType (ARRAY_OBJ) associated with the referenced Array struct
func (ao *Array) Type() ObjectType { return ARRAY_OBJ }

// Inspect will construct the Array as a string by stringifying its elements,
// and concatenating them into the expected array format.
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}

	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
