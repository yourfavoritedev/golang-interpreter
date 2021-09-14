package parser

import (
	"fmt"

	"github.com/yourfavoritedev/golang-interpreter/ast"
	"github.com/yourfavoritedev/golang-interpreter/lexer"
	"github.com/yourfavoritedev/golang-interpreter/token"
)

// Define precedences. Using iota to give constants incrementing numbers as values.
// The higher the value, the higher precedence it has.
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// Parser holds information on the lexer that is producing tokens,
// the current token being parsed (curToken), the next token (peekToken),
// errors that were encourntered during parsing
// and maps of its tokens with their parsing functions.
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	// prefixParseFn is called when the token type is in the prefix position (--5)
	prefixParseFn func() ast.Expression
	// infixParseFn is called when the token type is in the infix position.
	// It expects the left-side of the infix operator as an argument
	infixParseFn func(ast.Expression) ast.Expression
)

// New creates a new instance of a Parser with the first two tokens read
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	// initialize prefixParseFns map on the parser
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// register token.IDENT with the parseIdentifier function to the map.
	// Now whenever we encounter a token of type token.IDENT it will call
	// the parseIdentifier parsing function
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	return p
}

// parseIdentifier uses the parser's current token to construct an Identifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// nextToken simply sets the tokens of p to the next sequential tokens
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// parseStatement checks the parser's current token type to determine what statement operation to run and return
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement constructs a Statement with the attributes of a LetStatement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	// construct initial LetStatement node with the starting token (token.LET)
	stmt := &ast.LetStatement{Token: p.curToken}
	// should expect next token type to be token.IDENT `x in let x = 5`
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// construct the Identifier node with the attributes of the initial token.LET
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// should expect LetStatement to use an assignment `=`
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: We're skipping the expression until we encounter a semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement constructs a Statement with the attributes of a ReturnStatement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// construct initial returnStatement node with the starting token (token.RETURn)
	stmt := &ast.ReturnStatement{Token: p.curToken}
	// advance the parser to start examining the proceeding expression
	p.nextToken()

	// TODO: We're skipping the expression until we encounter a semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// curTokenIs verifies whether t and the parser's current token type are the same
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs verifies whether t and the parser's peek token type are the same
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek asserts the correctness of the order of tokens
// by checking the type of the next token.
// It will advance the parser's tokens if t and the parser's next token are the same token type.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		// advance token after verifying the peekToken is correct
		p.nextToken()
		return true
	} else {
		// peekToken does not match expectations, create error message in parser
		p.peekError(t)
		return false
	}
}

// parseExpressionStatement constructs a Statement with the attributes of an ExpressionStatement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// advance tokens if peekToken is a semicolon.
	// we can assume everything before the semicolon has been examined (5;),
	// semicolons are optional and not required by expression statements
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression checks whether a parsing function is
// mapped to the current token. If one exists,
// tt calls the parsing function (func () ast.Expression) and returns the Expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// check whether a parsing function exists for this token type
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	// Call parsing function for this token
	leftExp := prefix()
	return leftExp
}

// Errors returns the errors in the parser
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError adds an error message (string) to the parser's errors ([]string)
// when the peekToken does not match the expected token.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// ParseProgram constructs the root node of a AST an *ast.Program.
// It then iterates over every token in the input until it encounters an token.EOF token.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		// parse statement and add them to the program's Statements
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// registerPrefix assigns a key-value pair to the parser's prefixParseFns map
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix assigns a key-value pair to the parser's infixParseFns map
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
