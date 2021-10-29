package compiler

// SymbolScope is the unique scope a symbol belongs to
type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

// Symbol is the struct that holds all the necessary information about a symbol
// thats associated with an identifier.
// It contains information such as its name (the identifier, x in let x), the scope it belongs to
// and its unique number (index) in a SymbolTable. The index enables the VM to store
// and retrieve values.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable helps associate identifiers with a scope and unique number.
// The store maps the identifiers (strings) with their corresponding Symbol.
// numDefinitions simply refers to the total number of unique definitions in the store.
// It helps us do two things:
//
// 1. Define - Associate identifiers in the global scope with a unique number.
//
// 2. Resolve - Get the previously associated number for a given identifier.
type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

// NewSymbolTable creates a new SymbolTable with an empty store
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// Define sets an identifier/symbol association in the SymbolTable's store.
// Upon setting an association, we increment the number of definitions. A new
// Symbol is constructed for the given identifier and its Index is set to
// the number of defnitions the store had before adding this new association.
func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Scope: GlobalScope, Index: st.numDefinitions}
	st.store[name] = symbol
	st.numDefinitions++
	return symbol
}

// Resolve uses the given name to find a Symbol in the SymbolTable's store.
func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := st.store[name]
	return symbol, ok
}
