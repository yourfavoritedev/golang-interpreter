package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/yourfavoritedev/golang-interpreter/ast"
	"github.com/yourfavoritedev/golang-interpreter/code"
)

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	HASH_OBJ              = "HASH"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
	CLOSURE_OBJ           = "CLOSURE"
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

// HashKey constructs an integer hash-key for a Hash. It uses the Integer's Value
// as the HashKey value. This HashKey struct will be used as a key in the evaluated Hash Literal.
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// Boolean is the referenced struct for Boolean Literals in our object system.
// The struct holds the evaluated value of the Boolean Literal.
type Boolean struct {
	Value bool // the evaluated value
}

// Inspect returns the Boolean struct's Value as a string
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Type returns the ObjectType (BOOLEAN_OBJ) associated with the referenced Boolean struct
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// HashKey constructs a boolean hash-key for a Hash. The hash-key can
// be a binary of 1 or 0 depending on the Boolean's Value (true or false).
// This HashKey struct will be used as a key in the evaluated Hash Literal.
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

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

// HashKey constructs a string hash-key for a Hash. The hash-key uses the String's Value
// to construct a new 64-bit hash and converts it to a primitive uint64 for the hash-key.
// This helps resolve the issue where &Object.Strings have the same Value, but have different
// memory allocations. There is an inequality when comparing the index operation's value,
// map[&Object.String{Value:"key"}] to a key in the map, {&Object.String{Value:"key"}: "value"},
// we are comparing two different pointer addresses. Instead we should compare two structs.
// By using a HashKey, we convert the &Object.String.Value to a uint64 value and use that as the key.
// A unique input (string) will always return the same corresponding unique output (uint64).
// We take the inequal pointers out of the equation and now simply compare the two structs.
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

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

// HashPair is the referenced struct used as the designated value to HashKeys.
// It helps us print the values of the map in a more practial manner by
// containing both the objects that generated the keys and values of the map.
type HashPair struct {
	Key   Object
	Value Object
}

// HashKey is the referenced struct for a hash-key used in a Hash.
// It helps us effectively look up keys in the Hash.Pairs. Type refers to the different
// object types a hash-key can have (string, integer, boolean) before being converted to a uint64.
// Value refers to the actual literal value of the key, the key in the key-value pair.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Hash is the referenced strsuct for Hash Literals in our object system
// The Pairs field holds the evaluated map of the hash literal.
type Hash struct {
	Pairs map[HashKey]HashPair
}

// Type returns the ObjectType (HASH_OBJ) associated with the referenced Hash struct
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect will construct the Hash as a string by stringifying its key-value pairs,
// and concatenating them into the expected hash format.
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// Hashable is the interface used in our evaluator to check if the given object is
// usable as a hash key when we evaluate hash literals or index expressions for hashes.
type Hashable interface {
	HashKey() HashKey
}

// CompiledFunction is the referenced struct for compiled functions in our object system.
// The Instructions field holds the bytecode instructions from compiling a function literal.
// NumLocals is the number of local bindings in the function.
// CompiledFunction is intended to be a bytecode constant, it will be loaded on to
// to the stack and eventually used by the VM when it executes the function as a call expression instruction (OpCall).
type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

// Type returns the ObjectType (COMPILED_FUNCTION_OBJ) associated with the referenced CompiledFunction struct
func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }

// Inspect will simply return a preformatted string for the CompiledFunction with its memory-address.
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

// Closure is the referenced struct for closures in the object system.
// Fn points to the CompiledFunction enclosed by the closure.
// Free is a slice that keeps track of the free-variables relevant to the closure.
// A Closure struct will be constructed during run-time (when the VM executes)
// The Compiler provides an OpClosure instruction and the VM executes it,
// the will wrap an *object.CompiledFunction from the constants pool in a new Closure and put it on the stack.
// NOTE: All *object.CompiledFunctions will be wrapped by a Closure.
type Closure struct {
	Fn   *CompiledFunction
	Free []Object
}

// Type returns the ObjectType (CLOSURE_OBJ) associated with the referenced CLOSURE_OBJ struct
func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }

// Inspect will simply return a preformatted string for the Closure with its memory-address.
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}
