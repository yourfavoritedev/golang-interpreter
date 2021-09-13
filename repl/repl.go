package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/yourfavoritedev/golang-interpreter/lexer"
	"github.com/yourfavoritedev/golang-interpreter/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	// scanner helps intake standard input (from user) as a data stream
	scanner := bufio.NewScanner(in)

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

		// print all tokens from the lexer until we reach the end of the input (EOF)
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			// print to standard output (typically the terminal viewed by the user)
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
