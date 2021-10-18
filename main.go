package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/yourfavoritedev/golang-interpreter/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	// the os package has access to the current context that is running this program
	// if running in a terminal, os.Stdin and os.Stdout will be the terminal's
	// open data-streams for standard input and output
	repl.Start(os.Stdin, os.Stdout)
}
