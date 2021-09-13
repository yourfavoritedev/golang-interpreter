package ast

import (
	"testing"

	"github.com/yourfavoritedev/golang-interpreter/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			// initialize program with a LetStatement: "let myVar = anotherVar"
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				// Identifier implements Expression so it can be used as a LetStatement Value
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
			// initialize program with a ReturnStatement: "return myVar"
			&ReturnStatement{
				Token: token.Token{Type: token.RETURN, Literal: "return"},
				// Identifier implements Expression so it can be used as a ReturnStatement ReturnValue
				ReturnValue: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
			},
		},
	}

	if program.String() != "let myVar = anotherVar;return myVar;" {
		t.Errorf("program.String() wrong. got %q", program.String())
	}
}
