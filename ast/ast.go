package ast

import (
	"bytes"
	"fmt"
	"strings"

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

// Boolean holds a Token field (Token{TokenType, Literal}) for the boolean and
// a Value field for the actual bool value
type Boolean struct {
	Token token.Token
	Value bool
}

// expressionNode is implemented to allow Boolean to be served as an Expression
func (b *Boolean) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the the Boolean
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// String returns the literal value (Token.Literal) for the the Boolean
func (b *Boolean) String() string { return b.Token.Literal }

// IfExpression holds the necessary information
// to construct an if-expression
type IfExpression struct {
	Token       token.Token     // The 'if' token
	Condition   Expression      // The condition to be evalated
	Consequence *BlockStatement // The collection of statements which is a direct result of the passing Condition
	Alternative *BlockStatement // The alternative statements should the condition not pass
}

// expressionNode is implemented to allow IfExpression to be served as an Expression
func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the the if token
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// String will contruct the entire IfExpression as a string
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// BlockStatement holds the necessary information
// to construct a statement(s) that exist within an IfExpression or Function Literal
type BlockStatement struct {
	Token      token.Token // the "{" token
	Statements []Statement
}

// statementNode is implemented to allow BlockStatement to be served as a Statement
func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the "{" token
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// String will construct the entire BlockStatement as a string by
// iterating through and stringifying all its statements. The statements can
// be any combination of LET, RETURN or ExpressionStatements
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// FunctionLiteral holds the necessary information to
// construct a function-literal expression
type FunctionLiteral struct {
	Token      token.Token     // The 'fn' token
	Parameters []*Identifier   // The parameters of the function
	Body       *BlockStatement // The collection of statements in the body of the function
	Name       string          // The name the function is bound to
}

// expressionNode is implemented to allow FunctionLiteral to be served as an Expression
func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the "fn" token
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// String builds the entire FunctionLiteral as a string,
// first by stringifying all its params, then building the string with
// the FunctionLiterals expected components
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}

	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(fmt.Sprintf("%s", fl.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// CallExpression consist of an expression that results in a function when evaluated
// and a list of expressions that are the arguments of this function call
type CallExpression struct {
	Token     token.Token  // The "(" token
	Function  Expression   // Identifier of Function Literal
	Arguments []Expression // The list of expressions that are arguments to the function call
}

// expressionNode is implemented to allow CallExpression to be served as an Expression
func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the "(" token
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// String builds the entire CallExpression as a string,
// first by stringifying all its arguments, then building the string
// with the CallExpressions expected components
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	// stringify all arguments
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	// build string
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// StringLiteral holds a Token field (Token{TokenType, Literal}) for the lexed string and
// a Value field for the actual string value
type StringLiteral struct {
	Token token.Token
	Value string
}

// expressionNode is implemented to allow StringLiteral to be served as an Expression
func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the string
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// String returns the literal value (Token.Literal) for the the StringLiteral
func (sl *StringLiteral) String() string { return sl.Token.Literal }

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

// ArrayLiteral is used to construct an ast.Node for array literals ([1,2,3]).
// Parsing the tokens of an array literal should return an ArrayLiteral struct.
// ArrayLiteral is a valid expression node within the abstract-syntax tree.
type ArrayLiteral struct {
	Token    token.Token // the "[" token
	Elements []Expression
}

// expressionNode is implemented to allow IntegerLiteral to be served as an Expression
func (al *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the opening bracket of the array
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

// String builds the entire ArrayLiteral as a string,
// first by stringifying all its elments, then building the string
// with the ArrayLiteral expected components
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// IndexExpression is used to construct an ast.Node for index operator expressions ([1, 2, 3][1])
// Parsing the tokens of an index operator expression should return an IndexExpression struct.
// IndexExpression is a valid expression node within the abstract-syntax tree.
type IndexExpression struct {
	Token token.Token // The [ Token
	Left  Expression
	Index Expression
}

// expressionNode is implemented to allow IndexExpression to be served as an Expression
func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the opening bracket of the index operation
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// String builds the entire IndexExpression as a string.
// It stringifies the array and index, then builds the string
// with the IndexExpression expected components
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	// stringify the "array" being indexed. This "array"
	// could take the form of an identifier, array literal a function call,
	// or any expression.
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	// stringify the "index" which can take the form of any valid expression.
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// HashLiteral is used to construct an ast.Node for hash literals ({ "a": 1 })
// Parsing the tokens of a hash literal should return an HashLiteral struct.
// HashLiteral is a valid expression node within the abstract-syntax tree.
type HashLiteral struct {
	Token token.Token               // the '{' token
	Pairs map[Expression]Expression // the key value pairs of the hash
}

// expressionNode is implemented to allow HashLiteral to be served as an Expression
func (hl *HashLiteral) expressionNode() {}

// TokenLiteral returns the literal value (Token.Literal) for the opening brace of the hash literal
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }

// String builds the entire HashLiteral as a string.
// It stringifies the key value pairs, then builds the string
// with the String expected components
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
