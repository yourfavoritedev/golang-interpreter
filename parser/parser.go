package parser

import (
	"fmt"
	"strconv"

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

// a map of the token infix operators and their precedences
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

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
	// register token types with their designated parsing function to the maps.
	// now whenever we encounter that token.Type as part of an expression (foobar in let x = foobar;),
	// we can call its parsing function
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	// register infixParseFns as well
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	// register boolean parsing functions
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	// register grouped parsing function
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	// register ifExpression parsing function
	p.registerPrefix(token.IF, p.parseIfExpression)

	return p
}

// parseIdentifier uses the parser's current token to construct an Identifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// nextToken sets the tokens of the Parser to the next sequential tokens provided by the lexer.
// When called, the lexer will examine the next character, produce a new token
// and advance its position.
func (p *Parser) nextToken() {
	defer untrace(trace("nextToken"))
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
// Depending on the current token type, it will use a designated parsing function to construct the Expression
// The Expression is then set on the ExpressionStatement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))

	stmt := &ast.ExpressionStatement{Token: p.curToken}
	// parseExpression will determine what parsing function to use
	stmt.Expression = p.parseExpression(LOWEST)

	// advance tokens if peekToken is a semicolon.
	// we can assume everything before the semicolon has been examined (foobar;),
	// semicolons are optional and not required by expression statements
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// noPrefixParseFnError appends a new error to the parser's errors
// when the parser encounters a token in the expresson
// that does not have a prefix parse function
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// parseExpression checks whether a parsing function is
// mapped to the current token. If one exists,
// it calls the parsing function (func () ast.Expression) and returns the Expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))
	// check whether a parsing function exists for this token type
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// Call parsing function for this token
	leftExp := prefix()
	// if the next token is not a SEMICOLON, we can assume that
	// there is more to parse in the expression. The next token's precedence
	// must be higher than the previous token's, this allows us to construct higher
	// precedent expressions first, working our way from highest inner expressions to lowest
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// identify if there is an infix parsing function for the next token
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// advance the token to the next operator (left-binding-power)
		p.nextToken()

		// we called this infix but its really just parseInfixExpression
		// this takes the currently parsed expression and recursively builds the
		// infix expressions by highest precedence first and working our way out of them
		// until we complete the entire expression statement
		// (parseInfixExpression calls parseExpression)
		leftExp = infix(leftExp)
	}

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

// parseIntegerLiteral will construct an IntegerLiteral.
// It uses the current token and converts its literal value into an integer.
// The IntegerLiteral implements the Expression interface.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q a integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// parsePrefixExpression constructs an AST node as a PrefixExpression.
// It uses the current token and token literal to construct the PrefixExpression,
// { Token: { Type: Token.BANG, Literal: "!" }}
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	// initially the current token is the prefix operator (token.BANG or token.MINUS),
	// we must advance the tokens to be ready to consume the remaining part (Right value) of the prefix expression (5 of -5)
	p.nextToken()

	// call parseExpression again to consume the Right Value of the prefix expression
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// peekPrecedence finds the precedence of the peekToken and returns it
// otherwise return the LOWEST precedence
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence finds the precedence of the curToken and returns it
// otherwise return the LOWEST precedence
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// parsePrefixExpression constructs an AST node as an InfixExpression.
// It uses the current token, "Left" value and token literal to construct the PrefixExpression,
// { Token: { Type: Token.PLUS, Literal: "+" }, Operator: "+", Left: 5}
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	// can assume when this is called, we are already on the infix operator token
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// get precedence of current operator
	precedence := p.curPrecedence()
	// advance token past operator to the next expression
	p.nextToken()
	// set Right part of Expression
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseBoolean uses the parser's current token to construct a Boolean expression
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression constructs a Grouped Expression by
// advancing the current token "(" and calling parseExpression to construct
// the expression. It expects the parser to have parsed an expression up until the ")" token.
func (p *Parser) parseGroupedExpression() ast.Expression {
	defer untrace(trace("parseGroupedExpression"))
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	// We dont want to add an additional ")" to the end of an expression which already has ")".
	// Therefore, expectPeek will advance the current token to ")" after verifying it is the peekToken.
	// The ")" token is never actually consumed in the expression during parsing, we "skip it",
	// which will be evident as we traverse through parseExpression!
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression constructs an IfExpression by verifying
// that all the conditions to create the expression are met
func (p *Parser) parseIfExpression() ast.Expression {
	defer untrace(trace("parseIfExpression"))
	// iniitalize expression with current token (if)
	expression := &ast.IfExpression{Token: p.curToken}

	// expect next token to be "(", advance to that token
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// advance past the current token "(" to start constructing inner-expression
	p.nextToken()
	// construct expression, parseExpression will parse the token up until ")"
	expression.Condition = p.parseExpression(LOWEST)

	// expect next token to be ")", the end of the condition, advance to that token
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// expect next token to be "{", the start of the block statement, advance to that token
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// construct consequence block statement
	expression.Consequence = p.parseBlockStatement()

	// Consequence should be constructed and current token should now be "}"
	// verfify next token is "ELSE"
	if p.peekTokenIs(token.ELSE) {
		// advance to "ELSE" token
		p.nextToken()

		// verify next token is "{" and advance to that token
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// construct alternative block statement
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseBlockStatement calls parseStatement until it encounters either a },
// which signifies the end of the block statement or a token.EOF, which
// tells us there are no more tokens left to parse
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	defer untrace(trace("parseBlockStatement"))
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	// advance token past "{"
	p.nextToken()

	// build statements until parser encounters "}" token
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}
