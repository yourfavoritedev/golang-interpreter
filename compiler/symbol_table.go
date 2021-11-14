package compiler

// SymbolScope is the unique scope a symbol belongs to
type SymbolScope string

const (
	LocalScope    SymbolScope = "LOCAL"
	GlobalScope   SymbolScope = "GLOBAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTIOn"
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
// FreeSymbols refers to the free-variables defined in the Symbol Tables enclosing scopes (if any).
type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol
}

// NewSymbolTable creates a new SymbolTable with an empty store
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
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

// DefineBuiltin sets an identifier/symbol association for a builtin function in the SymbolTable's store.
// It uses the index of the builtin function in Builtins and its name to create a new symbol with the BuiltinScope
func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	st.store[name] = symbol
	return symbol
}

// SymbolTable sets an identifier/symbol association for a function in the SymbolTable's store.
// There can only ever be one symbol in the FunctionScope for a SymbolTable.
func (st *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	st.store[name] = symbol
	return symbol
}

// Resolve uses the given name to find a Symbol in the SymbolTable's store.
// If the SymbolTable is enclosed, it will recursively call the Outer table's Resolve
// method until the symbol is found or when there is no longer an enclosing Table.
func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := st.store[name]
	if !ok && st.Outer != nil {
		symbol, ok = st.Outer.Resolve(name)
		if !ok {
			return symbol, ok
		}

		if symbol.Scope == GlobalScope || symbol.Scope == BuiltinScope {
			return symbol, ok
		}

		// at this point, if the symbol was found (is ok) and none of the scopes above match,
		// we need to add the symbol to the current symbolTable's FreeSymbols and return a
		// FreeScoped version of it
		free := st.defineFree(symbol)
		return free, true
	}
	return symbol, ok
}

// defineFree adds a identifier/symbol association in the SymbolTable's store.
// It adds original, a Symbol from the enclosing scope into the symbolTables FreeSymbols.
// It returns a FreeScope version of the original symbol with the index updated to reflect
// the position of the newly added symbol in FreeSymbols
func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	s.store[original.Name] = symbol
	return symbol
}

// NewEnclosedSymbolTable creates a new SymbolTable enclosed by an outer SymbolTable.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}
