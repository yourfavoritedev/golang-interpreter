package object

import "fmt"

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
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
