package compiler

import (
	"fmt"
	"sort"

	"github.com/yourfavoritedev/golang-interpreter/ast"
	"github.com/yourfavoritedev/golang-interpreter/code"
	"github.com/yourfavoritedev/golang-interpreter/object"
)

// Compiler will create Bytecode for the VM to execute.
// The Compiler will leverage the evaluated abstract-syntax-tree to
// compile the necessary attributes for Bytecode. This includes the
// instructions (generated bytecode) and constants (the constant pool).
// symbolTable keeps track of the identifiers observed by the compiler.
// scopes is a stack used to keep record of unique scopes as their instructions are being compiled
// scopeIndex refers to the current scope being compiled
type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable
	scopes      []CompilationScope
	scopeIndex  int
}

// EmittedInstruction is the struct that describes an instruction that was
// emitted by the compiler
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// CompilationScope is the struct used to designate a unique scope for the compilation
// of a node. Nodes like *ast.FunctionLiteral and others that require their own scope,
// need to emit and keep track of their own instructions so that
// they don't become entangled in the parent/global scope.
// LastInstruction is the very last instruction that was omitted.
// PreviousInstruction is the one before that.
type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// New simply initializes a new Compiler
func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	// initialize symbol table with built-in functions
	symbolTable := NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

// currentInstructions simply returns the instructions of the current scope
func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// Compile builds the instructions and constants for the Compiler
// to generate bytecode. It is a recursive function that navigates the AST,
// evaluates the nodes and transform them into constants (object.Objects)
// to be added to the constants pool, and builds the necessary instructions
// for the VM to execute.
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	// our starting point
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	// compile expression statement - work our way down to the expression
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	// compile infix expression - work our way down to the literals
	case *ast.InfixExpression:
		// when a "<" operator is encountered, we want to simply apply the
		// comparison in reverse to keep logic succinct. To the VM, its as if the
		// "<" operator does not exist, all it should worry about is the OpGreaterThan instructions.
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	// compile prefix expression - work our way down to the literals
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}

	// compile an if expression - work our way down conditions and block statements
	// when compiling an if expression we run into the question of how to effectively compile
	// the right instructions given that an if-condition by nature has a binary evaluation,
	// a consequence and an alternative. We need to compile the if-expression instructions
	// while also providing additional instructions on where the VM needs to jump as a result
	// of the binary evaluation to execute the right block. This is represented by the
	// code.OpJumpNotTruthy and code.OpJump instructions that get compiled during this step.
	// code.OpJumpNotTruthy - jump over the compiled consequence
	// code.OpJump - jump over the compiled alternative
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an 'OpJumpNotTruthy' with a bogus operand value (absolute offset byte position) which we will resolve through backpatching
		// Recall that code.OpJumpNotTruthy has a single operand that indicates where in the instructions to jump to if the condition is not truthy
		jumpNotTruhyPos := c.emit(code.OpJumpNotTruthy, 9999)

		// compile the consequence
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// The consequence is an *ast.ExpressionStatement which emits an OpPop instruction.
		// We want to keep the constant on the stack to have a value for statements that use
		// it as an expression (let x = 5), so we must remove the last pop instruction.
		// OpPop instruction
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// the code.OpJump instruction is emitted directly after emitting the consequence (almost like its part of the consequence
		// when the consequence is executed by the VM, it knows to jump over the alternative instruction or over a OpNull instruction.
		// when an alternative block exists - it tells us the VM it can skip over the alternative).
		// the OpJump instruction itself is direcly before the alternative or OpNull instruction.
		// The code.OpJump operand will be backpatched with the position of the instruction to be jumped over
		// Emit an `OpJump with bogus value` to be backpatched
		jumpPos := c.emit(code.OpJump, 9999)
		// as soon as the consequence is emitted, we know exactly what to change the code.OpJumpNotTruthy operand to
		// knowing that we need to skip over this truthy instruction (consequence) because OpJumpNotTruthy should execute when the condition is falsey.
		// afterConsequencePos should now be the position of the alternative or OpNull instructiom.
		afterConsequencePos := len(c.currentInstructions())
		// replace code.OpJumpNotTruthy's operand with the new position, the position of the alternatve or OpNull instruction (afterConsequencePos)
		c.changeOperand(jumpNotTruhyPos, afterConsequencePos)

		// if we have no node.Alternative then we need to emit code.OpNull. In the event that the condition is falsey,
		// we need a OpNull instruction for the VM to be able to execute and pop.
		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			// compile the alternative
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			// same reasoning as above, we want to prevent the constant generated from the alternative
			// from popping so it can be used in the future
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		// as soon as the alternative or OpNull instruction is emitted, we know exactly what to backpatch the code.OpJump operand to
		// knowing that we need to skip over these falsey instructions because OpJump should execute as part of the consequence when the condition is truthy.
		// afterAlternativePos should now be the position after the alternative or OpNull instructiom.
		afterAlternativePos := len(c.currentInstructions())
		// replace code.OpJump's operand with the new position, the position after the alternative or OpNull instruction (afterAlternativePos)
		c.changeOperand(jumpPos, afterAlternativePos)

	// compile a block statement
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	// compile a let statement and update the symbolTable
	case *ast.LetStatement:
		// define the identifier in the symbol table
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		// the symbol for that identifier now has an index, which we use as an operand
		// to construct the instruction
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	// compile an identifier, it should look into the symbolTable to validate that the identifier has
	// been previously associated with a symbol.
	case *ast.Identifier:
		// grab the identiier from the symbol table
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}

		// construct an instruction with the symbol's index as the operand
		c.loadSymbol(symbol)

	// compile an array literal, it should cosntruct an OpArray instruction with the operand
	// being the number of elements in the array.
	case *ast.ArrayLiteral:
		// compile all elements in the array. The elements themselves are expressions.
		for _, e := range node.Elements {
			err := c.Compile(e)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	// compile a hash literal, it should construct an OpHash instruction with the operand
	// being the combined number of keys and values in the hash
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		// get keys from hash
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		// sort keys in descending/increasing order to guarantee a consistent order before we compile them
		// our tests assert that the instructions are generated in a specific order
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		// build Opcode instructions for keys and their values which should lead to a series
		// of OpConstants if the hash is not empty
		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	// compile an index expression. it should simply compile the object being indexed and then the index itself,
	// then finally emit an OpIndex instruction.
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	// compile a function literal. It should create a unique scope for the function and compile its body into
	// instructions, use those instructions to build a object.CompiledFunction, push that object to the
	// constants pool and finally emit an OpClosure instruction for the function literal.
	case *ast.FunctionLiteral:
		c.enterScope()

		// define function's name to symbol table if it exists
		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		// bind parameters to the function's symbole table
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// remove OpPop instruction (if there is one) from the function body's instructions, when the VM executes the body,
		// we don't want to pop the returnable value from the stack. Instead we want to replace OpPop
		// with the desired OpReturnValue instruction so the VM can actually return the value.
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		// when the function does not have a returnable value and therefore not an OpReturnValue instruction,
		// we want to add a code.OpReturn to the end of its instructions so the VM can simply return the function.
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		// Before leaving the inner-function's scope, we stored its free-variables in freeSymbols.
		// Now in the enclosing scope, we need to emit instructions to load these free-variables for the inner function.
		// The free-variables are inherited from the enclosing scope, so it has the responsibility of loading
		// them with the right Opcode instructions. ie: fn(a) { fn() { a; }} The closure for fn(a) should have instructions to load
		// a on to the stack so that  fn() { a; } can execute correctly.
		// These new instructions will belong to the enclosing scope and they will be in a compiledFunction constant in the constants pool
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		// add the compiledFn into the constants pool and use its index as the first operand
		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))

	// compile a return statement, it should emit an OpReturnValue instruction
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	// compile a call expression
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		// compile the function's arguments and emit their instructions
		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	// compile an integer literal
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	// compile a string literal
	case *ast.StringLiteral:
		s := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(s))

	// compile a boolean literal
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}

	return nil
}

// addConstant will add the given obj to the end of the constant pool and
// will return the index of that obj, that index can be used as an identifier
// to find obj in the pool.
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit generates an instruction for the compiler using the given params
// and then returns the starting position of the new instruction. The Compiler
// will keep track of the instruction it last emitted.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

// addInstruction builds to the compiler's current instructions. It takes
// the given instruction (ins) and appends it. It returns the starting position of the
// new instruction which should just be where the instructions initially ended + 1 position.
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

// setLastInstruction helps the compiler keep track of the instructions that
// it has emitted. When a new instruction is emitted, the lastInstructon recorded
// will become the previousInstruction and the new instruction will
// become the lastInstruction
func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	// only update the instructions belonging to the current scope
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

// lastInstructionIs simply checks whether the last emitted instruction
// has a matching Opcode with op
func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

// removeLastPop is used to remove an OpPop instruction from the compiler,
// shortening the instructions to everything up until the OpPop instruction.
func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

// replaceInstruction will replace an instruction starting at the absolute offset (pos)
// with a new instruction
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// replaceLastPopWithReturn will replace an OpPop instruction with an OpReturnValue instruction
func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

// NewWithState keeps global state in the REPL so the compiler can continue
// to run with the generated bytecode from a previous compilation.
func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

// changeOperand will replace the operand of an instruction at the absolute offset (opPos)
// with a new operand. It recreates the instructions with the new operand and swaps
// it with the old instruction.
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

// enterScope creates a new unique scope for the Compiler and enters it.
// Upon entering a scope, the compiler will use a new, enclosed SymbolTable.
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope tells the Compiler to leave the current scope and return
// the instructions that were created from that scope. When leaving the scope,
// the compiler will `un-enclose` the SymbolTable and use the outer SymbolTable from its parent scope.
func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	// remove the current scope (which will always be the last scope) from the scopes stack
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions
}

// loadSymbol uses the scope of the given Symbol to determine what Opcode instruction to emit
func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}

// Bytecode constructs a Bytecode struct using the Compiler's
// instructions and constants
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

// Bytecode is the struct for the representation of bytecode that
// will be passed to the VM. The Compiler will generate the Instructions
// and the Constants that were evaluated.
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
