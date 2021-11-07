package compiler

// SymbolScope is the unique scope a symbol belongs to
type SymbolScope string

const (
	LocalScope  SymbolScope = "LOCAL"
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
// Outer points to the SymbolTable that encloses the current one.
type SymbolTable struct {
	Outer          *SymbolTable
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
	symbol := Symbol{Name: name, Index: st.numDefinitions}
	// if the Symboltable does not have an outer (enclosing) table, then it belongs to the outer scope
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	st.store[name] = symbol
	st.numDefinitions++
	return symbol
}

// Resolve uses the given name to find a Symbol in the SymbolTable's store.
// If the SymbolTable is enclosed, it will recursively call the Outer table's Resolve
// method until the symbol is found or when there is no longer an enclosing Table.
func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := st.store[name]
	if !ok && st.Outer != nil {
		symbol, ok = st.Outer.Resolve(name)
		return symbol, ok
	}
	return symbol, ok
}

// NewEnclosedSymbolTable creates a new SymbolTable enclosed by an outer SymbolTable.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}
