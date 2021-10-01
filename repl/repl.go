package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/yourfavoritedev/golang-interpreter/evaluator"
	"github.com/yourfavoritedev/golang-interpreter/lexer"
	"github.com/yourfavoritedev/golang-interpreter/object"
	"github.com/yourfavoritedev/golang-interpreter/parser"
)

const PROMPT = ">> "
const MONKEY_FACE = "@(^_^)@\n"

func Start(in io.Reader, out io.Writer) {
	// scanner helps intake standard input (from user) as a data stream
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

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

		// evaluate the AST
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			// write program string to output
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
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
