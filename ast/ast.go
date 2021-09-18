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
// it simply distinguishes the Statement interface
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface that embeds the Node interface and the expressionNode method
//
// expressionNode() is a dummy method that does not do anything practical at the moment,
// it simply distinguishes the Expression interface
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

// statementNode is implemented to allow LetStatement to be served as a Statement
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

// Identifier holds the identifier of a binding eg: x in `let x = 5` and implements the Expression interface
type Identifier struct {
	Token token.Token // the token.IDENT token
	// Value is used to represent the name in a variable binding x in `let x = 5`,
	Value string
}

// expressionNode is implemented to allow Identifier to be served as an Expression
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

// statementNode is implmented to allow ReturnStatement to be served as a Statement
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

// statementNode is implemented to allow ExpressionStatement to be served as a Statement
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

// IntegerLiteral holds a Token field (Token{TokenType, Literal}) for the integer and
// a Value field for the actual integer value
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

// expressionNode is implemented to allow IntegerLiteral to be served as an Expression
func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the the integer
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// String constructs the integer value as a string
func (il *IntegerLiteral) String() string { return il.Token.Literal }

// PrefixExpression holds a Token field for the input,
// Operator is a string that contains either "-" or "!" and
// Right contains the expression to the right of the operator.
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

// expressionNode is implemented to allow PrefixExpression to be served as an Expression
func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the the PrefixExpression input
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// String constructs a string for the PrefixExpression,
// explicitly adding paranthesis around the constructed string to distinguish it from other expressions
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	// typically Right will be an IntegerLiteral, so we can call the string method on it
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// PrefixExpression holds a Token field for the input,
// Left contains the expression to the right of the operator.
// Operator is a string that contains either "-" or "!" and
// Right contains the expression to the right of the operator.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

// expressionNode is implemented to allow InfixExpression to be served as an Expression
func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the the InfixExpression input
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// String constructs a string for the InfixExpression,
// explicitly adding paranthesis around the constructed string to distinguish it from other expressions
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
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
