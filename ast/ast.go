package ast

import (
	"bytes"

	"github.com/yourfavoritedev/golang-interpreter/token"
)

// Node is the interface that wraps TokenLiteral method
//
// TokenLiteral() should return a token's literal value (Token.Literal) in a Node
// String() should return the AST node as a string
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is the interface that embeds the Node interface and the statementNode method
//
// statementNode() is a dummy method that does not do anything practical at the moment,
// it simply distinguishes the Statement node
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface that embeds the Node interface and the expressionNode method
//
// expressionNode() is a dummy method that does not do anything practical at the moment,
// it simply distinguishes the Expression node
type Expression interface {
	Node
	expressionNode()
}

// LetStatement holds the name and value for a let statement
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier // Name holds the identifer of the binding
	// Value is the expression that produces the value eg: 5 in `let x = 5`.
	// Coincidentally, it can also be an identifier for an expression in a different statement
	// eg: valueProducingIdentifier in `let x = valueProducingIdentifier` is an identifier
	// for a different statement eg: `let valueProducingIdentifier = 5`
	// Therefore in this statement, `let x = valueProducingIdentifier`, valueProducingIdentifer
	// is an identifier that serves as an expression - it produces a value
	Value Expression
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the literal value (Token.Literal) for a token of type Token.LET
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// String constructs the entire LetStatement node as a string
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// Identifier holds the identifier of a binding eg: x in `let x = 5` and implmements the Expression interface
type Identifier struct {
	Token token.Token // the token.IDENT token
	// Value is used to represent the name in a variable binding x in `let x = 5`,
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for a token of type Token.IDENT
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// String() returns the identifier's name value (x in let x = 5)
func (i *Identifier) String() string { return i.Value }

// ReturnStatement holds a Token field for the return token
// and a ReturnValue field for the expression that's to be returned
type ReturnStatement struct {
	Token       token.Token // the token.RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the literal value (Token.Literal) for a token of type token.RETURN
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// String constructs the entire ReturnStatement node as a string
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement holds a Token field and
// an Expression field for the expression.
// It implements the Node and Statement interfaces.
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the first token in the expression
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// String constructs the entire ExpressionStatement node as a string
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Program serves as the root node of every AST a parser produces.
type Program struct {
	Statements []Statement // Statements are just a slice of AST nodes
}

// TokenLiteral returns the token literal for the first statement in the program
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// String creates a buffer and writes the return value of each statement's String()
// method to it. Finally it returns the buffer as a string.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}
