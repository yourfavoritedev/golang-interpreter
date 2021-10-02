package object

// Environment employ a hashmap to keep track of evaluated values for expressions.
// Each value (Object) is associated with a name, typically the same name of the Identifier
// it was original bound too.
type Environment struct {
	store map[string]Object
	// The environment that encloses this one. Outer will be set to "nil" if no enclosing environment.
	outer *Environment
}

// Get uses the given name to find an associated Object in the Environment store.
// If the name cannot be found in the current environment, and the current environment has
// an outer environment, then try to Get the Object from the outer environment.
// This repeats and surfaces up the Environment tree until an associated Object is found
// or when there are no enclosing Environments left (reached the root environment).
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set will use the given name to update the associated entry in the
// Environment store with the new value
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// NewEnvironment creates a new instance of an Environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment extends the given Environment (outer).
// We create a new instance of an Environment with a pointer to the environment it should extend.
// By doing that, we enclose a fresh and empty environment with an existing one (outer).
// This allows us to preserve previous bindings while at the same time making new ones available,
// effectively introducing the concept of variable "scoping" ie: Identifiers of the same name
// can exist in different environments and hold different values:
/*
  // The outer enclosing environment
	let x = 5
	let printNum = fn(x) {
		// The inner environment
		puts(x)
	}

	printNum(10) ----> 10
	puts(x) ---->  5
*/
// The inner environment can always Get and reference the store of its outer environment
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
