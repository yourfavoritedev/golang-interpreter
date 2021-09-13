package parser

import (
	"fmt"

	"github.com/yourfavoritedev/golang-interpreter/ast"
	"github.com/yourfavoritedev/golang-interpreter/lexer"
	"github.com/yourfavoritedev/golang-interpreter/token"
)

// Parser holds information on the lexer that is producing tokens,
// the current token being parsed and the next token (peekToken)
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

// New creates a new instance of a Parser with the first two tokens read
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
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
		return nil
	}
}

// parseLetStatement constructs a statement struct with the attributes of a LET statement
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

// parseReturnStatement constructs a statement struct with the attributes of a RETURN statement
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

// curTokenIs verifies whether t and the parser's peek token type are the same
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
