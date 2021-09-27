package object

import "fmt"

const (
	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
	NULL_OBJ    = "NULL"
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
