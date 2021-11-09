package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/yourfavoritedev/golang-interpreter/compiler"
	"github.com/yourfavoritedev/golang-interpreter/lexer"
	"github.com/yourfavoritedev/golang-interpreter/object"
	"github.com/yourfavoritedev/golang-interpreter/parser"
	"github.com/yourfavoritedev/golang-interpreter/vm"
)

const PROMPT = ">> "
const MONKEY_FACE = "@(^_^)@\n"

func Start(in io.Reader, out io.Writer) {
	// scanner helps intake standard input (from user) as a data stream
	scanner := bufio.NewScanner(in)

	// helps us preserve the work when running multiple compilations
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	// keep accepting standard input until the user forcefully stops the program
	for {
		// Display prompt to signal start of input after ">> "
		fmt.Fprintf(out, PROMPT)
		// Scan loops until it receives input (from user), then makes the input available to its other methods
		scanned := scanner.Scan()

		// Exit program when no active data-stream left to scan
		if !scanned {
			return
		}

		// get the entire newly scanned input
		line := scanner.Text()
		// create mew lexer using input
		l := lexer.New(line)
		// create new parser using lexer
		p := parser.New(l)

		// initialize program
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// compile the program
		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		// execute the program
		code := comp.Bytecode()
		constants = code.Constants
		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		// write program string to output
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

// printParserErrors writes the parser errors to the output
func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, "parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
