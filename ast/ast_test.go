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

	// validate the first statement is a LetStatement
	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not LetStatement. got=%T", letStmt)
	}

	// validate the second statement is a ReturnStatement
	returnStmt, ok := program.Statements[1].(*ReturnStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not ReturnStatement. got=%T", returnStmt)
	}

	// validate the Program string matches expectations
	if program.String() != "let myVar = anotherVar;return myVar;" {
		t.Errorf("program.String() wrong. got %q", program.String())
	}
}
